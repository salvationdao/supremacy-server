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
	"server/passport"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
)

type SpoilsOfWar struct {
	battle        *Battle
	flushCh       chan bool
	tickSpeed     time.Duration
	transactSpeed time.Duration
}

func NewSpoilsOfWar(btl *Battle, transactSpeed time.Duration, dripSpeed time.Duration) *SpoilsOfWar {
	spw := &SpoilsOfWar{
		battle:        btl,
		transactSpeed: transactSpeed,
		flushCh:       make(chan bool),
		tickSpeed:     dripSpeed,
	}

	amnt := decimal.New(int64(rand.Intn(2000)+500), 18)

	sow := &boiler.SpoilsOfWar{
		BattleID:     btl.ID,
		BattleNumber: btl.BattleNumber,
		Amount:       amnt,
		AmountSent:   decimal.New(0, 18),
	}

	txr := fmt.Sprintf("spoils_of_war_fill_up|%s|%d", server.XsynTreasuryUserID, time.Now().UnixNano())

	_, err := btl.arena.ppClient.SpendSupMessage(passport.SpendSupsReq{
		FromUserID:           uuid.UUID(server.XsynTreasuryUserID),
		ToUserID:             SupremacyBattleUserID,
		Amount:               amnt.String(),
		TransactionReference: server.TransactionReference(txr),
		Group:                "spoil of war",
		SubGroup:             "system",
		Description:          "system",
		NotSafe:              false,
	})

	if err != nil {
		gamelog.L.Warn().Err(err).Msgf("transferring to spoils failed")
	}

	_ = sow.Insert(gamedb.StdConn, boil.Infer())

	go spw.Run()

	return spw
}

func (sow *SpoilsOfWar) End() {
	sow.flushCh <- true
}

func (sow *SpoilsOfWar) Run() {
	gamelog.L.Debug().Msg("starting spoils of war service")
	t := time.NewTicker(sow.transactSpeed)

	for {
		select {
		case <-sow.flushCh:
			// Runs at the end of each battle, called with sow.Flush()
			t.Stop()
			gamelog.L.Debug().Msg("running full flush and returning out")
			err := sow.Flush()
			if err != nil {
				gamelog.L.Err(err).Msg("blast out remainder failed of spoils of war")
				continue
			}
			gamelog.L.Info().Msgf("spoils system has been cleaned up: %s", sow.battle.ID)

			close(sow.flushCh)
			return
		case <-t.C:
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

func (sow *SpoilsOfWar) Flush() error {
	bn := sow.battle.BattleNumber - 1

	warchest, err := boiler.SpoilsOfWars(boiler.SpoilsOfWarWhere.BattleNumber.EQ(bn)).One(gamedb.StdConn)

	if err != nil {
		return terror.Error(err, "can't retrieve last battle's spoils")
	}

	multipliers, err := db.PlayerMultipliers(sow.battle.BattleNumber)
	if err != nil {
		return terror.Error(err, "unable to retrieve multipliers")
	}

	totalShares := decimal.Zero
	onlineUsers := []*db.Multipliers{}
	for _, player := range multipliers {
		if sow.battle.isOnline(player.PlayerID) {
			totalShares = totalShares.Add(player.TotalMultiplier)
			onlineUsers = append(onlineUsers, player)
		}
	}

	if warchest.Amount.LessThanOrEqual(decimal.Zero) {
		gamelog.L.Warn().Msgf("warchest amount is less than or equal to zero")
		return nil
	}
	amount := warchest.Amount.Sub(warchest.AmountSent)

	if totalShares.LessThanOrEqual(decimal.Zero) {
		gamelog.L.Warn().Msgf("total share is less than or equal to zero")
		return nil
	}
	amount = amount.Div(totalShares)

	subgroup := fmt.Sprintf("Spoils of War from Battle #%d", sow.battle.BattleNumber-1)

	for _, player := range onlineUsers {
		txr := fmt.Sprintf("spoils_of_war|%s|%d", player.PlayerID, time.Now().UnixNano())
		userAmount := amount.Mul(player.TotalMultiplier)
		_, err := sow.battle.arena.ppClient.SpendSupMessage(passport.SpendSupsReq{
			FromUserID:           SupremacyBattleUserID,
			ToUserID:             player.PlayerID,
			Amount:               userAmount.StringFixed(18),
			TransactionReference: server.TransactionReference(txr),
			Group:                "spoil of war",
			SubGroup:             subgroup,
			Description:          subgroup,
			NotSafe:              false,
		})
		if err != nil {
			return terror.Error(err, "unable to send sups spoil of war flush")
		} else {
			warchest.AmountSent = warchest.AmountSent.Add(userAmount)
			_, err = warchest.Update(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Msg("unable to update spoils of war")
				warchest = nil
				return err
			}
			pt := boiler.PendingTransaction{
				FromUserID:           SupremacyBattleUserID.String(),
				ToUserID:             player.PlayerID.String(),
				Amount:               userAmount,
				TransactionReference: txr,
				Group:                "spoil of war",
				Subgroup:             subgroup,
				ProcessedAt:          null.TimeFrom(time.Now()),
				Description:          subgroup,
			}
			err = pt.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Msg("unable to save spoils of war transaction")
			}
		}
	}
	return nil
}

func (sow *SpoilsOfWar) Drip() error {
	var err error
	bn := sow.battle.BattleNumber - 1

	warchest, err := boiler.SpoilsOfWars(boiler.SpoilsOfWarWhere.BattleNumber.EQ(bn)).One(gamedb.StdConn)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		gamelog.L.Error().Err(err).Msg("unable to retrieve spoils of war")
		return err
	}

	if warchest.Amount.LessThanOrEqual(decimal.Zero) {
		gamelog.L.Warn().Msgf("warchest amount is less than or equal to zero")
		return nil
	}

	dripAllocations := 300

	dripAmount := warchest.Amount.Div(decimal.NewFromInt(int64(dripAllocations)))

	multipliers, err := db.PlayerMultipliers(sow.battle.BattleNumber)
	if err != nil {
		return terror.Error(err, "unable to retrieve multipliers")
	}

	totalShares := decimal.Zero
	onlineUsers := []*db.Multipliers{}
	for _, player := range multipliers {
		if sow.battle.isOnline(player.PlayerID) {
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
	subgroup := fmt.Sprintf("Spoils of War from Battle #%d", sow.battle.BattleNumber-1)
	amountRemaining := warchest.Amount.Sub(warchest.AmountSent)

	onShareSups := dripAmount.Div(totalShares)
	for _, player := range onlineUsers {
		userDrip := onShareSups.Mul(player.TotalMultiplier)

		amountRemaining = amountRemaining.Sub(userDrip)
		if amountRemaining.LessThan(userDrip) {
			gamelog.L.Warn().Msg("not enough funds in the spoils of war to do a tick")
			return nil
		}

		txr := fmt.Sprintf("spoils_of_war|%s|%d", player.PlayerID, time.Now().UnixNano())

		_, err := sow.battle.arena.ppClient.SpendSupMessage(passport.SpendSupsReq{
			FromUserID:           SupremacyBattleUserID,
			ToUserID:             player.PlayerID,
			Amount:               userDrip.StringFixed(18),
			TransactionReference: server.TransactionReference(txr),
			Group:                "spoil of war",
			SubGroup:             subgroup,
			Description:          subgroup,
			NotSafe:              false,
		})
		if err != nil {
			gamelog.L.Error().Err(err).Msg("unable to send spoils of war transaction")
			continue
		} else {
			warchest.AmountSent = warchest.AmountSent.Add(userDrip)
			_, err = warchest.Update(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Msg("unable to update spoils of war")
				warchest = nil
				return err
			}
			pt := boiler.PendingTransaction{
				FromUserID:           SupremacyBattleUserID.String(),
				ToUserID:             player.PlayerID.String(),
				Amount:               userDrip,
				TransactionReference: txr,
				Group:                "spoil of war",
				Subgroup:             subgroup,
				ProcessedAt:          null.TimeFrom(time.Now()),
				Description:          subgroup,
			}
			err = pt.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Msg("unable to save spoils of war transaction")
			}
		}
	}

	return nil
}

var SupremacyBattleUserID = uuid.Must(uuid.FromString("87c60803-b051-4abb-aa60-487104946bd7"))
