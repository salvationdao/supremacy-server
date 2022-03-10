package battle

import (
	"database/sql"
	"errors"
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/passport"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

type SpoilsOfWar struct {
	battle        *Battle
	l             zerolog.Logger
	flushCh       chan bool
	tickSpeed     time.Duration
	transactSpeed time.Duration
	warchest      *boiler.SpoilsOfWar
}

func NewSpoilsOfWar(btl *Battle, transactSpeed time.Duration, dripSpeed time.Duration) *SpoilsOfWar {
	l := gamelog.L.With().Str("svc", "spoils_of_war").Logger()

	return &SpoilsOfWar{
		battle:        btl,
		l:             l,
		transactSpeed: transactSpeed,
		flushCh:       make(chan bool),
		tickSpeed:     dripSpeed,
	}
}

func (sow *SpoilsOfWar) End() {
	sow.flushCh <- true
}

func (sow *SpoilsOfWar) Start() {
	sow.l.Debug().Msg("starting spoils of war service")
	t := time.NewTicker(sow.transactSpeed)

	for {
		select {
		case <-sow.flushCh:
			// Runs at the end of each battle, called with sow.Flush()
			sow.l.Debug().Msg("running full flush and returning out")
			err := sow.Flush()
			if err != nil {
				sow.l.Err(err).Msg("blast out remainder failed of spoils of war")
				continue
			}
			return
		case <-t.C:
			// Push all pending transactions to passport server
			sow.l.Debug().Msg("running transaction pusher")
			err := sow.Drip()
			if err != nil {
				sow.l.Err(err).Msg("push transactions over rpc")
				continue
			}
		}
	}
}

func (sow *SpoilsOfWar) Flush() error {
	warchest, err := sow.ProcessSpoils(sow.battle.BattleNumber)
	if err != nil {
		return terror.Error(err, "can't retrieve last battle's spoils")
	}

	multipliers, err := db.PlayerMultipliers(sow.battle.BattleNumber - 1)
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
	amount := warchest.Amount.Sub(warchest.AmountSent)
	amount = amount.Div(totalShares)

	subgroup := fmt.Sprintf("Spoils of War from Battle #%d", sow.battle.BattleNumber-1)

	for _, player := range onlineUsers {
		txr := fmt.Sprintf("spoils_of_war|%s|%d", player.PlayerID, time.Now().UnixNano())
		userAmount := amount.Mul(player.TotalMultiplier)
		_, err := sow.battle.arena.ppClient.SpendSupMessage(passport.SpendSupsReq{
			FromUserID:           SupremacyBattleUserID,
			ToUserID:             &player.PlayerID,
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
				sow.l.Error().Err(err).Msg("unable to update spoils of war")
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
				sow.l.Error().Err(err).Msg("unable to save spoils of war transaction")
			}
		}
	}
	return nil
}

//ProcessSpoils work out how much was spent last battle
func (sow *SpoilsOfWar) ProcessSpoils(battleNumber int) (*boiler.SpoilsOfWar, error) {
	contributions, sumSpoils, err := db.Spoils(sow.battle.ID.String())
	if err != nil {
		return nil, terror.Error(err, "calculate total spoils for last battle failed")
	}

	battle, err := boiler.Battles(qm.Where(`battle_number = ?`, battleNumber)).One(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err, "unable to retrieve battle from battle number")
	}

	spoils, err := boiler.SpoilsOfWars(qm.Where(`battle_number = ?`, battleNumber)).One(gamedb.StdConn)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		spoils = &boiler.SpoilsOfWar{
			BattleID:     battle.ID,
			BattleNumber: battleNumber,
		}
	} else if err != nil {
		return nil, terror.Error(err, "unable to retrieve spoils from battle number")
	}

	spoils.Amount = sumSpoils

	err = spoils.Upsert(gamedb.StdConn, true, []string{
		boiler.PlayerColumns.PublicAddress,
	},
		boil.Whitelist(
			boiler.SpoilsOfWarColumns.BattleID,
			boiler.SpoilsOfWarColumns.BattleNumber,
		), boil.Infer())
	if err != nil {
		return nil, terror.Error(err, "unable to insert spoils of war")
	}

	for _, contrib := range contributions {
		err = db.MarkContributionProcessed(uuid.Must(uuid.FromString(contrib.ID)))
		if err != nil {
			return nil, terror.Error(err, "mark single contribution processed")
		}
	}
	return nil, nil
}

func (sow *SpoilsOfWar) Drip() error {
	var err error
	if sow.warchest == nil {
		sow.warchest, err = sow.ProcessSpoils(sow.battle.BattleNumber - 1)
		if err != nil {
			sow.l.Error().Err(err).Msg("unable to process spoils")
		}
		return err
	}

	dripAllocations := sow.tickSpeed / time.Minute * 5
	dripAmount := sow.warchest.Amount.Div(decimal.NewFromInt(int64(dripAllocations)))

	multipliers, err := db.PlayerMultipliers(sow.battle.BattleNumber - 1)
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

	subgroup := fmt.Sprintf("Spoils of War from Battle #%d", sow.battle.BattleNumber-1)
	amountRemaining := sow.warchest.Amount.Sub(sow.warchest.AmountSent)
	for _, player := range onlineUsers {
		userDrip := dripAmount.Div(player.TotalMultiplier)

		amountRemaining = amountRemaining.Sub(userDrip)
		if amountRemaining.LessThan(userDrip) {
			sow.l.Warn().Msg("not enough funds in the spoils of war to do a tick")
			return nil
		}

		txr := fmt.Sprintf("spoils_of_war|%s|%d", player.PlayerID, time.Now().UnixNano())

		_, err := sow.battle.arena.ppClient.SpendSupMessage(passport.SpendSupsReq{
			FromUserID:           SupremacyBattleUserID,
			ToUserID:             &player.PlayerID,
			Amount:               userDrip.StringFixed(18),
			TransactionReference: server.TransactionReference(txr),
			Group:                "spoil of war",
			SubGroup:             subgroup,
			Description:          subgroup,
			NotSafe:              false,
		})
		if err != nil {
			sow.l.Error().Err(err).Msg("unable to send spoils of war transaction")
			continue
		} else {
			sow.warchest.AmountSent = sow.warchest.AmountSent.Add(userDrip)
			_, err = sow.warchest.Update(gamedb.StdConn, boil.Infer())
			if err != nil {
				sow.l.Error().Err(err).Msg("unable to update spoils of war")
				sow.warchest = nil
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
				sow.l.Error().Err(err).Msg("unable to save spoils of war transaction")
			}
		}
	}

	return nil
}

var SupremacyBattleUserID = uuid.Must(uuid.FromString("87c60803-b051-4abb-aa60-487104946bd7"))
