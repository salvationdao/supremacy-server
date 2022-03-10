package spoilsofwar

import (
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpcclient"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type SOW struct {
	rpcClient     *rpcclient.XrpcClient
	l             zerolog.Logger
	flushCh       chan bool
	tickSpeed     time.Duration
	transactSpeed time.Duration
}

func New(rpcClient *rpcclient.XrpcClient, transactSpeed time.Duration, dripSpeed time.Duration, dripMax decimal.Decimal) *SOW {
	l := gamelog.L.With().Str("svc", "spoils_of_war").Logger()

	return &SOW{
		rpcClient:     rpcClient,
		l:             l,
		transactSpeed: transactSpeed,
		flushCh:       make(chan bool),
		tickSpeed:     dripSpeed,
	}
}

func (sow *SOW) Flush() {
	sow.flushCh <- true
}

func (sow *SOW) Run() {
	sow.l.Debug().Msg("starting spoils of war service")
	t := time.NewTicker(sow.tickSpeed)
	t2 := time.NewTicker(sow.transactSpeed)
	for {
		select {
		case <-sow.flushCh:
			// Runs at the end of each battle, called with sow.Flush()
			sow.l.Debug().Msg("running full flush")
			err := sow.RunFullFlush()
			if err != nil {
				sow.l.Err(err).Msg("process end of battle spoils")
				continue
			}
		case <-t.C:
			// Runs often with a capped limit
			sow.l.Debug().Msg("running spoils processor")
			err := sow.ProcessCappedSpoils()
			if err != nil {
				sow.l.Err(err).Msg("process capped spoils")
				continue
			}
		case <-t2.C:
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

func (sow *SOW) RunFullFlush() error {
	totalSpoils, err := db.TotalSpoils()
	if err != nil {
		return terror.Error(err, "get total spoils")
	}

	latestBattle, noBattles, err := db.LatestBattleNumber()
	if err != nil {
		return terror.Error(err, "get latest battle number")
	}

	if noBattles {
		sow.l.Debug().Msg("no battles yet")
		return nil
	}
	multipliers, err := db.PlayerMultipliers(latestBattle)
	if err != nil {
		return terror.Error(err, "get player multipliers")
	}

	err = sow.Process(latestBattle, totalSpoils, multipliers)
	if err != nil {
		return terror.Error(err, "process flush spoils")
	}
	err = db.MarkAllContributionsProcessed()
	if err != nil {
		return terror.Error(err, "mark flush contributions processed")
	}
	return nil
}

func (sow *SOW) ProcessCappedSpoils() error {
	config, err := boiler.Configs().One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "get global config")
	}

	contributions, sumCappedSpoils, err := db.CappedSpoils(config.SupsPerTick)
	if err != nil {
		return terror.Error(err, "get total spoils")
	}

	latestBattle, noBattles, err := db.LatestBattleNumber()
	if err != nil {
		return terror.Error(err, "get latest battle number")
	}

	if noBattles {
		sow.l.Debug().Msg("no battles yet")
		return nil
	}

	multipliers, err := db.PlayerMultipliers(latestBattle)
	if err != nil {
		return terror.Error(err, "get player multipliers")
	}

	err = sow.Process(latestBattle, sumCappedSpoils, multipliers)
	if err != nil {
		return terror.Error(err, "process tick spoils")
	}
	for _, contrib := range contributions {
		err = db.MarkContributionProcessed(uuid.Must(uuid.FromString(contrib.ID)))
		if err != nil {
			return terror.Error(err, "mark single contribution processed")
		}
	}
	return nil
}

func (sow *SOW) ProcessPendingTransactions() error {
	txes, err := db.UnprocessedPendingTransactions()
	if err != nil {
		return terror.Error(err, "get unprocessed transactions")
	}

	if len(txes) <= 0 {
		sow.l.Debug().Msg("no txes to process")
		return nil
	}

	err = sow.rpcClient.Call(
		"s.InsertTransactions",
		rpcclient.InsertTransactionsReq{Transactions: txes},
		&rpcclient.InsertTransactionsResp{},
	)

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

func (sow *SOW) Process(battleNumber int, spoils decimal.Decimal, multipliers []*db.Multipliers) error {
	totalShares := decimal.Zero
	for _, player := range multipliers {
		totalShares = totalShares.Add(player.TotalMultiplier)
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, player := range multipliers {
		// Create a tx per player
		playerReward := spoils.Mul(player.TotalMultiplier).Div(totalShares)
		record := &boiler.PendingTransaction{
			Amount:               playerReward,
			FromUserID:           server.XsynTreasuryUserID.String(),
			ToUserID:             player.PlayerID.String(),
			TransactionReference: fmt.Sprintf("BATTLE#%d|SPOILS_OF_WAR|%d", battleNumber, time.Now().UnixNano()),
			Group:                "Battle",
			Subgroup:             strconv.Itoa(battleNumber),
		}
		err := record.Insert(tx, boil.Infer())
		if err != nil {
			return err
		}
	}
	tx.Commit()

	return nil
}
