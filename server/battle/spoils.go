package battle

import (
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpcclient"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

type SpoilsOfWar struct {
	battle        *Battle
	rpcClient     *rpcclient.XrpcClient
	l             zerolog.Logger
	flushCh       chan bool
	tickSpeed     time.Duration
	transactSpeed time.Duration
}

func NewSpoilsOfWar(btl *Battle, rpcClient *rpcclient.XrpcClient, transactSpeed time.Duration, dripSpeed time.Duration, dripMax decimal.Decimal) *SpoilsOfWar {
	l := gamelog.L.With().Str("svc", "spoils_of_war").Logger()

	return &SpoilsOfWar{
		battle:        btl,
		rpcClient:     rpcClient,
		l:             l,
		transactSpeed: transactSpeed,
		flushCh:       make(chan bool),
		tickSpeed:     dripSpeed,
	}
}

func (sow *SpoilsOfWar) End() {
	err := sow.ProcessSpoils()
	if err != nil {
		sow.l.Error().Err(err).Msg("unable to process spoils")
	}
	sow.flushCh <- true
}

func (sow *SpoilsOfWar) Start() {
	sow.l.Debug().Msg("starting spoils of war service")
	t := time.NewTicker(sow.transactSpeed)

	for {
		select {
		case <-sow.flushCh:
			// Runs at the end of each battle, called with sow.Flush()
			sow.l.Debug().Msg("running full flush")
			return
		case <-t.C:
			// Push all pending transactions to passport server
			sow.l.Debug().Msg("running transaction pusher")
			err := sow.ProcessPendingTransactions()
			if err != nil {
				sow.l.Err(err).Msg("push transactions over rpc")
				continue
			}
		}
	}
}

func (sow *SpoilsOfWar) ProcessSpoils() error {
	contributions, sumSpoils, err := db.Spoils(sow.battle.ID.String())
	if err != nil {
		return terror.Error(err, "calculate total spoils for last battle failed")
	}

	spoils := &boiler.SpoilsOfWar{
		BattleID:   sow.battle.ID.String(),
		Amount:     sumSpoils,
		AmountSent: decimal.New(0, 18),
	}
	err = spoils.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "unable to insert spoils of war")
	}
	for _, contrib := range contributions {
		err = db.MarkContributionProcessed(uuid.Must(uuid.FromString(contrib.ID)))
		if err != nil {
			return terror.Error(err, "mark single contribution processed")
		}
	}
	return nil
}

func (sow *SpoilsOfWar) ProcessPendingTransactions() error {
	txes, err := db.UnprocessedPendingTransactions()
	if err != nil {
		return terror.Error(err, "get unprocessed transactions")
	}

	if len(txes) <= 0 {
		sow.l.Debug().Msg("no txes to process")
		return nil
	}

	//spoil := SpoilsOfWar{}

	if err != nil {
		return terror.Error(err, "insert transactions on passport server")
	}

	for _, tx := range txes {
		err = db.MarkPendingTransactionProcessed(uuid.Must(uuid.FromString(tx.ID)))
		if err != nil {
			return terror.Error(err, "mark transactions as processed")
		}
	}
	return nil
}

//func (sow *SpoilsOfWar) Process(battle *boiler.Battle, spoils decimal.Decimal, multipliers []*db.Multipliers) error {
//	totalShares := decimal.Zero
//	onlineUsers := []*db.Multipliers{}
//	for _, player := range multipliers {
//		if sow.battle.isOnline(player.PlayerID) {
//			totalShares = totalShares.Add(player.TotalMultiplier)
//			onlineUsers = append(onlineUsers, player)
//		}
//	}
//
//	tx, err := gamedb.StdConn.Begin()
//	if err != nil {
//		return err
//	}
//	defer tx.Rollback()
//	for _, player := range onlineUsers {
//		// Create a tx per player
//		playerReward := spoils.Mul(player.TotalMultiplier).Div(totalShares)
//		record := &boiler.PendingTransaction{
//			Amount:               playerReward,
//			FromUserID:           server.SupremacyGameUserID,
//			ToUserID:             player.PlayerID.String(),
//			TransactionReference: fmt.Sprintf("BATTLE#%d|SPOILS_OF_WAR|%d", battle.BattleNumber, time.Now().UnixNano()),
//			Group:                "Battle",
//			Description:          fmt.Sprintf("Spoils of war from battle %d", battle.BattleNumber),
//			Subgroup:             strconv.Itoa(battle.BattleNumber),
//		}
//		err := record.Insert(tx, boil.Infer())
//		if err != nil {
//			return err
//		}
//	}
//	tx.Commit()
//
//	return nil
//}
