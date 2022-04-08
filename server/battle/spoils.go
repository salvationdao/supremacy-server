package battle

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpcclient"
	"sync"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/atomic"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
)

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

	amnt := decimal.New(int64(rand.Intn(150)), 18)

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
			Amount:       amnt,
			AmountSent:   decimal.New(0, 18),
		}

		txr := fmt.Sprintf("spoils_of_war_fill_up|%s|%d", server.XsynTreasuryUserID, time.Now().UnixNano())

		_, err := btl.arena.RPCClient.SpendSupMessage(rpcclient.SpendSupsReq{
			FromUserID:           uuid.UUID(server.XsynTreasuryUserID),
			ToUserID:             SupremacyBattleUserID,
			Amount:               amnt.String(),
			TransactionReference: server.TransactionReference(txr),
			Group:                string(server.TransactionGroupBattle),
			SubGroup:             "system",
			Description:          "system",
			NotSafe:              false,
		})

		if err != nil {
			gamelog.L.Warn().Err(err).Msgf("transferring to spoils failed")
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
			// Push all pending transactions to passport server
			gamelog.L.Debug().Msg("running transaction pusher")
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
	var err error

	yesterday := time.Now().Add(time.Hour * -24)

	warchests, err := boiler.SpoilsOfWars(
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

	dripAllocations := 20

	totalAmount := decimal.NewFromInt(0)
	for _, warchest := range warchests {
		totalAmount = totalAmount.Add(warchest.Amount).Sub(warchest.AmountSent)
	}

	dripAmount := totalAmount.Div(decimal.NewFromInt(int64(dripAllocations)))

	multipliers, err := db.PlayerMultipliers(sow.battle().battleSeconds().IntPart())
	if err != nil {
		return terror.Error(err, "unable to retrieve multipliers")
	}

	totalShares := decimal.Zero
	onlineUsers := []*db.Multipliers{}
	for _, player := range multipliers {
		if sow.battle().isOnline(player.PlayerID) {
			if player.TotalMultiplier.LessThanOrEqual(decimal.Zero) {
				continue
			}
			totalShares = totalShares.Add(player.TotalMultiplier)
			onlineUsers = append(onlineUsers, player)
		}
	}

	if totalShares.LessThanOrEqual(decimal.Zero) {
		gamelog.L.Warn().Msgf("total shares is less than or equal to zero")
		return nil
	}
	subgroup := fmt.Sprintf("Spoils of War from Battle #%d", sow.battle().BattleNumber-1)
	amountRemaining := totalAmount.Copy()

	onShareSups := dripAmount.Div(totalShares)
	for _, player := range onlineUsers {
		userDrip := onShareSups.Mul(player.TotalMultiplier)

		amountRemaining = amountRemaining.Sub(userDrip)
		if amountRemaining.LessThan(userDrip) {
			gamelog.L.Warn().Msg("not enough funds in the spoils of war to do a tick")
			return nil
		}

		sendSups := func(warchest *boiler.SpoilsOfWar, dripAmnt decimal.Decimal) error {
			warchest.AmountSent = warchest.AmountSent.Add(dripAmnt)
			_, err = warchest.Update(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Msg("unable to update spoils of war")
				warchest = nil
				return err
			}

			txr := fmt.Sprintf("spoils_of_war|%s|%d", player.PlayerID, time.Now().UnixNano())

			_, err := sow.battle().arena.RPCClient.SpendSupMessage(rpcclient.SpendSupsReq{
				FromUserID:           SupremacyBattleUserID,
				ToUserID:             player.PlayerID,
				Amount:               dripAmnt.StringFixed(18),
				TransactionReference: server.TransactionReference(txr),
				Group:                string(server.TransactionGroupBattle),
				SubGroup:             subgroup,
				Description:          subgroup,
				NotSafe:              false,
			})
			if err != nil {
				gamelog.L.Error().Err(err).Msg("unable to send spoils of war transaction")
				return err
			}

			pt := boiler.PendingTransaction{
				FromUserID:           SupremacyBattleUserID.String(),
				ToUserID:             player.PlayerID.String(),
				Amount:               dripAmnt,
				TransactionReference: txr,
				Group:                string(server.TransactionGroupBattle),
				Subgroup:             subgroup,
				ProcessedAt:          null.TimeFrom(time.Now()),
				Description:          subgroup,
			}
			err = pt.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Msg("unable to save spoils of war transaction")
			}
			totalAmount = totalAmount.Add(warchest.Amount).Sub(warchest.AmountSent)
			userDrip = userDrip.Sub(dripAmnt)
			return nil
		}

		for _, warchest := range warchests {

			if userDrip.LessThanOrEqual(decimal.Zero) {
				break
			}
			remaining := warchest.Amount.Sub(warchest.AmountSent)
			if remaining.LessThanOrEqual(decimal.Zero) {
				continue
			}
			if remaining.LessThan(userDrip) {
				userDrip = userDrip.Sub(remaining)

				err = sendSups(warchest, remaining)
				if err != nil {
					return err
				}
				continue
			}

			err = sendSups(warchest, userDrip)
			if err != nil {
				return err
			}
			break
		}

		if userDrip.GreaterThan(decimal.Zero) {
			gamelog.L.Info().Str("battle_id", sow.battle().ID).Msg("no more money in spoils of war for distribution")
			return nil
		}
	}

	return nil
}

var SupremacyBattleUserID = uuid.Must(uuid.FromString("87c60803-b051-4abb-aa60-487104946bd7"))
