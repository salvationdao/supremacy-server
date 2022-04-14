package battle

import (
	"database/sql"
	"errors"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpcclient"
	"sync"
	"time"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/atomic"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

var SupremacyBattleUserID = uuid.Must(uuid.FromString("87c60803-b051-4abb-aa60-487104946bd7"))

type SpoilsOfWar struct {
	_battle       *Battle
	flushCh       chan bool
	flushed       *atomic.Bool
	tickSpeed     time.Duration
	transactSpeed time.Duration
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

func NewSpoilsOfWar(btl *Battle, transactSpeed time.Duration, dripSpeed time.Duration) *SpoilsOfWar {
	spw := &SpoilsOfWar{
		_battle:       btl,
		transactSpeed: transactSpeed,
		flushCh:       make(chan bool),
		flushed:       atomic.NewBool(false),
		tickSpeed:     dripSpeed,
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
			MaxTicks:     20,
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

		// if current tick = last tick, dump the rest to online users
		if spoils.CurrentTick == spoils.MaxTicks {
			userSpoils, spoils = payoutRemainingUserSpoils(
				userSpoils,
				spoils,
				func(userID uuid.UUID) bool { return sow.battle().isOnline(userID) },
				sendSups,
			)
		} else { // else drip out normally
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

		// update spoils (payoutRemainingUserSpoils/payoutUserSpoils mutates it)
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

// payoutRemainingUserSpoils takes the leftover spoils and gives them out to online users,
// mutates and returns userSpoils and spoils
func payoutRemainingUserSpoils(
	userSpoils []*boiler.UserSpoilsOfWar,
	spoils *boiler.SpoilsOfWar,
	isOnline func(userID uuid.UUID) bool,
	sendSups func(userID uuid.UUID, amount string, txr string) (string, error),
) ([]*boiler.UserSpoilsOfWar, *boiler.SpoilsOfWar) {
	var onlineUserSpoils []*boiler.UserSpoilsOfWar
	var offlineUserSpoils []*boiler.UserSpoilsOfWar
	for _, user := range userSpoils {
		userID, err := uuid.FromString(user.PlayerID)
		if err != nil {
			gamelog.L.Error().
				Err(err).
				Str("user.PlayerID", user.PlayerID).
				Msg("failed to create uuid from player id")
			continue
		}
		if isOnline(userID) {
			onlineUserSpoils = append(onlineUserSpoils, user)
		} else {
			offlineUserSpoils = append(offlineUserSpoils, user)
		}
	}

	spoilsAmountLeft := spoils.Amount.Sub(spoils.AmountSent)
	totalMulties := decimal.Zero
	for _, user := range onlineUserSpoils {
		totalMulties = totalMulties.Add(decimal.NewFromInt(int64(user.TotalMultiplierForBattle)))
	}

	oneMultiWorth := spoilsAmountLeft.Div(totalMulties)

	for _, user := range onlineUserSpoils {
		usersPayout := oneMultiWorth.Mul(decimal.NewFromInt(int64(user.TotalMultiplierForBattle)))
		userID, err := uuid.FromString(user.PlayerID)
		if err != nil {
			gamelog.L.Error().
				Err(err).
				Str("user.PlayerID", user.PlayerID).
				Msg("failed to create uuid from player id")
			continue
		}
		txr := fmt.Sprintf("spoils_of_war|%s|%d", userID, time.Now().UnixNano())
		txID, err := sendSups(userID, usersPayout.String(), txr)
		if err != nil {
			gamelog.L.Error().
				Err(err).
				Str("FromUserID", SupremacyBattleUserID.String()).
				Str("ToUserID", userID.String()).
				Str("Amount", user.TickAmount.String()).
				Str("TransactionReference", txr).
				Msg("unable to send spoils of war transaction")
			continue
		}
		spoils.AmountSent = spoils.AmountSent.Add(usersPayout)                // add this battles amount sent
		user.PaidSow = user.PaidSow.Add(usersPayout)                          // add this user sow amount sent
		user.RelatedTransactionIds = append(user.RelatedTransactionIds, txID) // add the tx id from sending spoils
	}

	// for all the offline users set their low sow
	for _, user := range offlineUserSpoils {
		user.LostSow = user.TotalSow.Sub(user.PaidSow)
	}

	return append(onlineUserSpoils, offlineUserSpoils...), spoils
}
