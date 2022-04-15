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
	"server/rpcclient"
	"sync"
	"time"

	"github.com/volatiletech/null/v8"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/atomic"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

var SupremacyBattleUserID = uuid.Must(uuid.FromString("87c60803-b051-4abb-aa60-487104946bd7"))
var SupremacyUserID = uuid.Must(uuid.FromString("4fae8fdf-584f-46bb-9cb9-bb32ae20177e"))

type SpoilsOfWar struct {
	_battle       *Battle
	flushCh       chan bool
	flushed       *atomic.Bool
	tickSpeed     time.Duration
	transactSpeed time.Duration
	maxTicks      int
	sync.RWMutex
}

func (sow *SpoilsOfWar) battle() *Battle {
	sow.RLock()
	defer sow.RUnlock()
	return sow._battle
}

func (sow *SpoilsOfWar) storeBattle(btl *Battle) {
	sow.Lock()
	defer sow.Unlock()
	sow._battle = btl
}

func NewSpoilsOfWar(btl *Battle, transactSpeed time.Duration, dripSpeed time.Duration, maxTicks int) *SpoilsOfWar {
	spw := &SpoilsOfWar{
		_battle:       btl,
		transactSpeed: transactSpeed,
		flushCh:       make(chan bool),
		flushed:       atomic.NewBool(false),
		tickSpeed:     dripSpeed,
		maxTicks:      maxTicks,
	}

	sow, err := boiler.SpoilsOfWars(boiler.SpoilsOfWarWhere.BattleID.EQ(btl.BattleID)).One(gamedb.StdConn)
	if errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Info().Err(err).Msgf("spoil of war not found. this is expected.")
	} else if err != nil {
		gamelog.L.Info().Err(err).Msgf("spoil of war not found. strange error.")
	}

	if sow == nil {
		sow = &boiler.SpoilsOfWar{
			BattleID:     btl.ID,
			BattleNumber: btl.BattleNumber,
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

func (sow *SpoilsOfWar) End() {
	sow.flushCh <- true
}

func (sow *SpoilsOfWar) Run() {
	gamelog.L.Debug().Msg("starting spoils of war service")
	t := time.NewTicker(sow.transactSpeed)

	mismatchCount := atomic.NewInt32(0)
	defer t.Stop()

	for {
		select {
		case <-sow.flushCh:
			if sow.flushed.Load() {
				return
			}
			// Runs at the end of each battle, called with sow.Flush(
			gamelog.L.Debug().Msg("running full flush and returning out")

			sow.flushed.Store(true)
			sow.storeBattle(nil)
			sow = nil
			return
		case <-t.C:
			// terminate ticker if the battle mismatch
			if sow.battle() != sow.battle().arena.currentBattle() {
				mismatchCount.Add(1)
				gamelog.L.Warn().
					Int32("times", mismatchCount.Load()).
					Msg("battle mismatch is detected on spoil of war ticker")
				if mismatchCount.Load() < 10 {
					continue
				}

				gamelog.L.Info().Msg("detect battle mismatch 10 times, cleaning up the spoil of war")
				sow.storeBattle(nil)
				sow = nil
				return
			}

			err := sow.Drip()
			if err != nil {
				gamelog.L.Err(err).Msg("push transactions over rpc")
				continue
			}
		}
	}
}

func (sow *SpoilsOfWar) Drip() error {
	if sow.battle() == nil {
		return nil
	}

	yesterday := time.Now().Add(time.Hour * -24)

	// get all sow with spoils left on them
	spoilsOfWars, err := boiler.SpoilsOfWars(
		boiler.SpoilsOfWarWhere.CreatedAt.GT(yesterday),
		boiler.SpoilsOfWarWhere.BattleID.NEQ(sow.battle().ID),
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
		userSpoils, err := boiler.UserSpoilsOfWars(
			boiler.UserSpoilsOfWarWhere.BattleID.EQ(spoils.BattleID),
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
			return sow.battle().arena.RPCClient.SpendSupMessage(rpcclient.SpendSupsReq{
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

		// the below code cleans up legacy spoils paymout methods, can be removed in future deployments
		if userSpoils == nil || len(userSpoils) == 0 {
			multipliers, err := multipliers.GetPlayersMultiplierSummaryForBattle(spoils.BattleNumber)
			if err != nil {
				gamelog.L.Error().
					Err(err).
					Int("spoils.BattleNumber", spoils.BattleNumber).
					Msg("failed to get player multipliers for battle")
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
					func(userID uuid.UUID) bool { return sow.battle().isOnline(userID) },
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

	if amountLeft.IsZero() || totalMultis.IsZero() {
		return spoils
	}

	for _, multi := range multipliers {
		totalMultis = totalMultis.Add(multi.TotalMultiplier)
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
	user *boiler.UserSpoilsOfWar,
	spoils *boiler.SpoilsOfWar,
	isOnline func(userID uuid.UUID) bool,
	sendSups func(userID uuid.UUID, amount string, txr string) (string, error),
) (*boiler.UserSpoilsOfWar, *boiler.SpoilsOfWar) {
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
		gamelog.L.Error().
			Err(fmt.Errorf("warChestSpoilsLeft.LessThan(user.TickAmount)")).
			Str("battle_id", spoils.BattleID).
			Str("warChestSpoilsLeft", warChestSpoilsLeft.String()).
			Str("user.TickAmount", user.TickAmount.String()).
			Msg("not enough spoils to pay out a user spoil tick (issue!)")
		return user, spoils
	}

	// check paying this tick out doesn't over pay them
	if user.PaidSow.Add(user.TickAmount).GreaterThan(user.TotalSow) {
		gamelog.L.Error().
			Err(fmt.Errorf("user.PaidSow.Add(user.TickAmount).GreaterThan(user.TotalSow)")).
			Str("battle_id", spoils.BattleID).
			Str("warChestSpoilsLeft", warChestSpoilsLeft.String()).
			Str("user.PaidSow", user.PaidSow.String()).
			Str("user.TickAmount", user.TickAmount.String()).
			Str("user.TotalSow", user.TotalSow.String()).
			Msg("paying the user this tick over pays them")
		return user, spoils
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
	userSpoils []*boiler.UserSpoilsOfWar,
	spoils *boiler.SpoilsOfWar,
	sendSups func(userID uuid.UUID, amount string, txr string) (string, error),
) ([]*boiler.UserSpoilsOfWar, *boiler.SpoilsOfWar) {
	// get remaining, send it to supremacy user
	remainingSpoils := spoils.Amount.Sub(spoils.AmountSent)
	txr := fmt.Sprintf("spoils_of_war_leftovers|%s|%d", spoils.BattleID, time.Now().UnixNano())
	txid, err := sendSups(SupremacyUserID, remainingSpoils.String(), txr)
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
	spoils.LeftoversTransactionID = null.StringFrom(txid)

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
			Msg("issue with the remaining/lost spoils")
	}

	return userSpoils, spoils
}
