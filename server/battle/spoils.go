package battle

import (
	"database/sql"
	"errors"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/multipliers"
	"server/xsyn_rpcclient"
	"sync"
	"time"

	"github.com/ninja-syndicate/hub/ext/messagebus"

	"github.com/volatiletech/null/v8"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

var SupremacyBattleUserID = uuid.Must(uuid.FromString("87c60803-b051-4abb-aa60-487104946bd7"))
var SupremacyUserID = uuid.Must(uuid.FromString("4fae8fdf-584f-46bb-9cb9-bb32ae20177e"))

type SpoilsOfWar struct {
	messageBus   *messagebus.MessageBus
	passport     *xsyn_rpcclient.XsynXrpcClient
	isOnline     func(userID uuid.UUID) bool
	battleID     string
	battleNumber int
	tickSpeed    time.Duration
	maxTicks     int
	cleanUp      chan bool
	sync.RWMutex
}

func (sow *SpoilsOfWar) End() {
	sow.cleanUp <- true
}
func (sow *SpoilsOfWar) BattleID() string {
	sow.RLock()
	defer sow.RUnlock()
	return sow.battleID
}

func (sow *SpoilsOfWar) BattleNumber() int {
	sow.RLock()
	defer sow.RUnlock()
	return sow.battleNumber
}

func NewSpoilsOfWar(
	passport *xsyn_rpcclient.XsynXrpcClient,
	messageBus *messagebus.MessageBus,
	isOnline func(userID uuid.UUID) bool,
	battleID string,
	battleNumber int,
	dripSpeed time.Duration,
	maxTicks int,
) *SpoilsOfWar {
	spw := &SpoilsOfWar{
		isOnline:     isOnline,
		battleID:     battleID,
		battleNumber: battleNumber,
		messageBus:   messageBus,
		passport:     passport,
		tickSpeed:    dripSpeed,
		maxTicks:     maxTicks,
		cleanUp:      make(chan bool),
	}

	sow, err := boiler.SpoilsOfWars(boiler.SpoilsOfWarWhere.BattleID.EQ(spw.BattleID())).One(gamedb.StdConn)
	if errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Info().Err(err).Msgf("spoil of war not found. this is expected.")
	} else if err != nil {
		gamelog.L.Info().Err(err).Msgf("spoil of war not found. strange error.")
	}

	if sow == nil {
		sow = &boiler.SpoilsOfWar{
			BattleID:     spw.BattleID(),
			BattleNumber: spw.BattleNumber(),
			Amount:       decimal.Zero,
			AmountSent:   decimal.Zero,
			CurrentTick:  0,
			MaxTicks:     maxTicks,
		}

		_ = sow.Insert(gamedb.StdConn, boil.Infer())
	}
	go spw.Run()

	return spw
}

func (sow *SpoilsOfWar) Run() {
	gamelog.L.Debug().Msg("starting spoils of war service")
	t := time.NewTicker(sow.tickSpeed)
	defer t.Stop()

	for {
		select {
		case <-sow.cleanUp:
			gamelog.L.Debug().Msg("cleaning up spoils of war service")
			return
		case <-t.C:
			err := sow.Drip()
			if err != nil {
				gamelog.L.Err(err).Interface("sow", sow).Msg("failed to drip spoils of war")
				continue
			}
		}
	}
}

func (sow *SpoilsOfWar) Drip() error {
	if sow.BattleID() == "" || sow.battleNumber == 0 {
		return nil
	}

	yesterday := time.Now().Add(time.Hour * -24)

	// get all sow with spoils left on them
	spoilsOfWars, err := boiler.SpoilsOfWars(
		boiler.SpoilsOfWarWhere.CreatedAt.GT(yesterday),
		boiler.SpoilsOfWarWhere.BattleID.NEQ(sow.BattleID()),
		boiler.SpoilsOfWarWhere.LeftoversTransactionID.IsNull(),
		boiler.SpoilsOfWarWhere.CurrentTick.LTE(sow.maxTicks),
		qm.And(`amount_sent < amount`),
		qm.OrderBy(fmt.Sprintf("%s %s", boiler.SpoilsOfWarColumns.CreatedAt, "DESC")),
	).All(gamedb.StdConn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		gamelog.L.Error().Err(err).Msg("unable to retrieve spoils of war")
		return err
	}

	for _, spoils := range spoilsOfWars {
		spoils.CurrentTick++

		// get all user spoils of war for this battle
		userSpoils, err := boiler.PlayerSpoilsOfWars(
			boiler.PlayerSpoilsOfWarWhere.BattleID.EQ(spoils.BattleID),
		).All(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().
				Err(err).
				Str("spoils.BattleID", spoils.BattleID).
				Msg("failed to get user spoils for battle")
			continue
		}

		subgroup := fmt.Sprintf("Spoils of War from Battle #%d", spoils.BattleNumber)
		sendSups := func(userID uuid.UUID, amount string, txr string) (string, error) {
			return sow.passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
				FromUserID:           SupremacyBattleUserID,
				ToUserID:             userID,
				Amount:               amount,
				TransactionReference: server.TransactionReference(txr),
				Group:                string(server.TransactionGroupBattle),
				SubGroup:             subgroup,
				Description:          subgroup,
				NotSafe:              false,
			})
		}

		// the below code cleans up legacy spoils payout methods, can be removed in future deployments
		if userSpoils == nil || len(userSpoils) == 0 {
			multipliers, err := multipliers.GetPlayersMultiplierSummaryForBattle(spoils.BattleNumber)
			if err != nil {
				gamelog.L.Error().
					Err(err).
					Int("spoils.BattleNumber", spoils.BattleNumber).
					Msg("failed to get player multipliers for battle")
				continue
			}

			// if no multipliers flush out sups to us
			// this can happen when there are spoils but no contributor put in more than 3 sups
			if len(multipliers) == 0 {
				amountLeft := spoils.Amount.Sub(spoils.AmountSent)
				txr := fmt.Sprintf("spoils_of_war_leftovers|%s|%d", spoils.BattleID, time.Now().UnixNano())
				txID, err := sendSups(SupremacyUserID, amountLeft.String(), txr)
				if err != nil {
					gamelog.L.Error().
						Err(err).
						Str("FromUserID", SupremacyBattleUserID.String()).
						Str("ToUserID", SupremacyUserID.String()).
						Str("Amount", amountLeft.String()).
						Str("TransactionReference", txr).
						Msg("unable to send left over spoils")
					continue
				}
				spoils.LeftoverAmount = amountLeft
				spoils.LeftoversTransactionID = null.StringFrom(txID)
				continue
			}

			spoils = flushOutOldSpoils(spoils, multipliers, sendSups)

			_, err = spoils.Update(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().
					Err(err).
					Interface("battle spoils", spoils).
					Msg("failed to update battle spoils")
			}
			continue
		}

		// if the last tick has passed, flush all remaining sups to the supremacy user! $$$
		if spoils.CurrentTick > spoils.MaxTicks {
			userSpoils, spoils = takeRemainingSpoils(
				userSpoils,
				spoils,
				sendSups,
			)
		} else {
			for _, user := range userSpoils {
				user, spoils = payoutUserSpoils(
					user,
					spoils,
					sow.isOnline,
					sendSups,
				)
			}
		}

		// for each user, update em! (payoutRemainingUserSpoils/payoutUserSpoils mutates them)
		for _, user := range userSpoils {
			_, err = user.Update(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().
					Err(err).
					Interface("user spoils", user).
					Msg("failed to update user spoils")
				continue
			}
		}

		_, err = spoils.Update(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().
				Err(err).
				Interface("battle spoils", spoils).
				Msg("failed to update battle spoils")
		}
	}

	return nil
}

// flushOutOldSpoils gets all user multipliers for that battle and flushes out the remaining spoils
func flushOutOldSpoils(
	spoils *boiler.SpoilsOfWar,
	multipliers []*multipliers.MultiplierSummary,
	sendSups func(userID uuid.UUID, amount string, txr string) (string, error),
) *boiler.SpoilsOfWar {
	amountLeft := spoils.Amount.Sub(spoils.AmountSent)
	oneMultiWorth := decimal.Zero
	totalMultis := decimal.Zero

	for _, multi := range multipliers {
		totalMultis = totalMultis.Add(multi.TotalMultiplier)
	}

	if amountLeft.IsZero() || totalMultis.IsZero() {
		return spoils
	}

	oneMultiWorth = amountLeft.Div(totalMultis)

	for _, multi := range multipliers {
		userAmount := oneMultiWorth.Mul(multi.TotalMultiplier).RoundDown(0)
		playerUUID, err := uuid.FromString(multi.PlayerID)
		if err != nil {
			gamelog.L.Error().
				Err(err).
				Str("multi.PlayerID", multi.PlayerID).
				Msg("failed to make uuid from string")
			continue
		}

		if spoils.AmountSent.Add(userAmount).GreaterThan(spoils.Amount) {
			gamelog.L.Error().
				Err(err).
				Str("userAmount", userAmount.String()).
				Str("spoils.AmountSent", spoils.AmountSent.String()).
				Str("spoils.Amount", spoils.Amount.String()).
				Msg("spoils doesn't have enough to make this payout")
			continue
		}

		_, err = sendSups(playerUUID, userAmount.String(), fmt.Sprintf("spoils_of_war|%s|%d", multi.PlayerID, time.Now().UnixNano()))
		if err != nil {
			gamelog.L.Error().
				Err(err).
				Str("playerUUID", playerUUID.String()).
				Str("userAmount.", userAmount.String()).
				Str("txr", "").
				Msg("failed to send user their spoils")
			continue
		}

		spoils.AmountSent = spoils.AmountSent.Add(userAmount)
	}

	return spoils
}

// payoutUserSpoils checks the user is online,
// checks there is enough spoils left,
// checks paying the user doesn't over pay them for the given spoils
// then calls the sendSups function
// mutates and returns userSpoils and spoils
func payoutUserSpoils(
	user *boiler.PlayerSpoilsOfWar,
	spoils *boiler.SpoilsOfWar,
	isOnline func(userID uuid.UUID) bool,
	sendSups func(userID uuid.UUID, amount string, txr string) (string, error),
) (*boiler.PlayerSpoilsOfWar, *boiler.SpoilsOfWar) {
	userID, err := uuid.FromString(user.PlayerID)
	if err != nil {
		gamelog.L.Error().
			Err(err).
			Str("user.PlayerID", user.PlayerID).
			Msg("failed to create uuid from player id")
		return user, spoils
	}

	if !isOnline(userID) {
		return user, spoils
	}

	warChestSpoilsLeft := spoils.Amount.Sub(spoils.AmountSent)

	// check the spoils for this battle have enough left for a tick (always should)
	if warChestSpoilsLeft.LessThan(user.TickAmount) {
		difference := user.TickAmount.Sub(warChestSpoilsLeft)
		// if difference is > than 1000, throw an error since that is more than a rounding issue.
		if difference.GreaterThan(decimal.NewFromInt(1000)) {
			gamelog.L.Error().
				Err(fmt.Errorf("warChestSpoilsLeft.LessThan(user.TickAmount)")).
				Str("battle_id", spoils.BattleID).
				Str("warChestSpoilsLeft", warChestSpoilsLeft.String()).
				Str("user.TickAmount", user.TickAmount.String()).
				Msg("not enough spoils to pay out a user spoil tick (issue!)")
			return user, spoils
		}
		// if difference is < than 1000, give them what we can
		user.TickAmount = user.TickAmount.Sub(difference)
	}

	// check paying this tick out doesn't over pay them
	if user.PaidSow.Add(user.TickAmount).GreaterThan(user.TotalSow) {
		// sometimes on the last tick it can just be rounding issues,
		// so check if it's less than a 0.000000000000010000 sup difference and if so just pay them what is left
		difference := user.TotalSow.Sub(user.PaidSow.Add(user.TickAmount))
		if difference.GreaterThan(decimal.NewFromInt(1000)) {
			gamelog.L.Error().
				Err(fmt.Errorf("user.PaidSow.Add(user.TickAmount).GreaterThan(user.TotalSow)")).
				Str("battle_id", spoils.BattleID).
				Str("warChestSpoilsLeft", warChestSpoilsLeft.String()).
				Str("user.PaidSow", user.PaidSow.String()).
				Str("user.TickAmount", user.TickAmount.String()).
				Str("user.TotalSow", user.TotalSow.String()).
				Msg("paying the user this tick over pays them by more than 0.000000000000010000 possibly not a rounding error")
			return user, spoils
		}
		// if just rounder error give them what we can
		user.TickAmount = user.TotalSow.Sub(user.PaidSow)
	}

	txr := fmt.Sprintf("spoils_of_war|%s|%d", userID, time.Now().UnixNano())
	txID, err := sendSups(userID, user.TickAmount.String(), txr)
	if err != nil {
		gamelog.L.Error().
			Err(err).
			Str("FromUserID", SupremacyBattleUserID.String()).
			Str("ToUserID", userID.String()).
			Str("Amount", user.TickAmount.String()).
			Str("TransactionReference", txr).
			Msg("unable to send spoils of war transaction")
		return user, spoils
	}

	// update relative fields
	spoils.AmountSent = spoils.AmountSent.Add(user.TickAmount)            // add this battles amount sent
	user.PaidSow = user.PaidSow.Add(user.TickAmount)                      // add this user sow amount sent
	user.RelatedTransactionIds = append(user.RelatedTransactionIds, txID) // add the tx id from sending spoils

	return user, spoils
}

func takeRemainingSpoils(
	userSpoils []*boiler.PlayerSpoilsOfWar,
	spoils *boiler.SpoilsOfWar,
	sendSups func(userID uuid.UUID, amount string, txr string) (string, error),
) ([]*boiler.PlayerSpoilsOfWar, *boiler.SpoilsOfWar) {
	// get remaining, send it to supremacy user
	remainingSpoils := spoils.Amount.Sub(spoils.AmountSent)
	txr := fmt.Sprintf("spoils_of_war_leftovers|%s|%d", spoils.BattleID, time.Now().UnixNano())
	txID, err := sendSups(SupremacyUserID, remainingSpoils.String(), txr)
	if err != nil {
		gamelog.L.Error().
			Err(err).
			Str("FromUserID", SupremacyBattleUserID.String()).
			Str("ToUserID", SupremacyUserID.String()).
			Str("Amount", remainingSpoils.String()).
			Str("TransactionReference", txr).
			Msg("unable to send left over spoils")
		return userSpoils, spoils
	}

	spoils.LeftoverAmount = remainingSpoils
	spoils.LeftoversTransactionID = null.StringFrom(txID)

	// update all users with their lost spoils
	totalLostSpoils := decimal.Zero
	for _, user := range userSpoils {
		user.LostSow = user.TotalSow.Sub(user.PaidSow) // we loop over users on line 236 to update them in db
		totalLostSpoils = totalLostSpoils.Add(user.LostSow)
	}

	// if lost spoils doesn't match remaining spoils, something is wrong!
	if !remainingSpoils.Equal(totalLostSpoils) {
		gamelog.L.Error().Err(fmt.Errorf("remainingSpoils not equal totalLostSpoils")).
			Str("remainingSpoils", remainingSpoils.String()).
			Str("totalLostSpoils", totalLostSpoils.String()).
			Str("spoils.BattleID", spoils.BattleID).
			Int("spoils.BattleNumber", spoils.BattleNumber).
			Msg("issue with the remaining/lost spoils")
	}

	return userSpoils, spoils
}
