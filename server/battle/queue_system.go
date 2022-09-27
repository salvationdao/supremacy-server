package battle

import (
	"database/sql"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"
)

func (am *ArenaManager) ScheduleLobbyChecker() {
	l := gamelog.L.With().Str("func", "ScheduleLobbyChecker").Logger()

	now := time.Now()
	defaultWaitTime := 10 * time.Second
	currentCheckpoint := now.Add(defaultWaitTime)

	// set timer
	timer := time.NewTimer(currentCheckpoint.Sub(now))

	for {
		select {

		// channel for update timer
		case <-am.ScheduledLobbyCheckpointChan:
			earliestScheduledBattleLobby, err := boiler.BattleLobbies(
				boiler.BattleLobbyWhere.ReadyAt.IsNotNull(),
				boiler.BattleLobbyWhere.WillNotStartUntil.IsNotNull(),
				boiler.BattleLobbyWhere.AssignedToBattleID.IsNull(),
				boiler.BattleLobbyWhere.WillNotStartUntil.LT(null.TimeFrom(currentCheckpoint)),
				boiler.BattleLobbyWhere.EndedAt.IsNull(),
				qm.OrderBy(boiler.BattleLobbyColumns.WillNotStartUntil),
			).One(gamedb.StdConn)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				l.Error().Err(err).Msg("Failed to load earliest scheduled battle lobby.")
			}

			if earliestScheduledBattleLobby != nil {
				newCheckpoint := earliestScheduledBattleLobby.WillNotStartUntil.Time.Add(defaultWaitTime)
				if newCheckpoint.After(time.Now()) && newCheckpoint.Before(currentCheckpoint) {
					currentCheckpoint = newCheckpoint
					timer.Reset(time.Now().Sub(currentCheckpoint))
				}
			}

		case <-timer.C:
			// kick idle arenas
			am.KickIdleArenas()

			// get the next scheduled battle lobby
			nextScheduledBattleLobby, err := boiler.BattleLobbies(
				boiler.BattleLobbyWhere.ReadyAt.IsNotNull(),
				boiler.BattleLobbyWhere.WillNotStartUntil.IsNotNull(),
				boiler.BattleLobbyWhere.WillNotStartUntil.GT(null.TimeFrom(currentCheckpoint)),
				boiler.BattleLobbyWhere.EndedAt.IsNull(),
				qm.OrderBy(boiler.BattleLobbyColumns.WillNotStartUntil),
			).One(gamedb.StdConn)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				l.Error().Err(err).Msg("Failed to load next scheduled battle lobby.")
			}

			if nextScheduledBattleLobby != nil {
				newCheckpoint := nextScheduledBattleLobby.WillNotStartUntil.Time.Add(defaultWaitTime)
				currentCheckpoint = newCheckpoint
				timer.Reset(time.Now().Sub(newCheckpoint))
			}
		}
	}
}

func (am *ArenaManager) SendBattleQueueFunc(fn func() error) error {
	am.BattleQueueFuncMx.Lock()
	defer am.BattleQueueFuncMx.Unlock()
	return fn()
}

func BroadcastBattleLobbyUpdate(battleLobbyIDs ...string) {
	if battleLobbyIDs == nil || len(battleLobbyIDs) == 0 {
		return
	}

	bls, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.ID.IN(battleLobbyIDs),
		qm.Load(boiler.BattleLobbyRels.HostBy),
		qm.Load(boiler.BattleLobbyRels.GameMap),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Strs("battle lobby id list", battleLobbyIDs).Msg("Failed to query battle lobby")
		return
	}

	if bls == nil {
		return
	}

	resp, err := server.BattleLobbiesFromBoiler(bls)
	if err != nil {
		return
	}

	ws.PublishMessage("/secure/battle_lobbies", server.HubKeyBattleLobbyListUpdate, resp)
}

// SetDefaultPublicBattleLobbies ensure there are enough battle lobbies when server start
func (am *ArenaManager) SetDefaultPublicBattleLobbies() error {
	// check once when server start
	err := am.DefaultPublicLobbiesCheck()
	if err != nil {
		return err
	}

	// check every minutes
	go func() {
		ticker := time.NewTicker(1 * time.Minute)

		for {
			<-ticker.C
			err = am.DefaultPublicLobbiesCheck()
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to check default public lobbies.")
			}
		}
	}()

	return nil
}

// DefaultPublicLobbiesCheck check there are enough public lobbies
func (am *ArenaManager) DefaultPublicLobbiesCheck() error {
	// load default public lobbies amount
	publicLobbiesCount := db.GetIntWithDefault(db.KeyDefaultPublicLobbyCount, 20)

	// lock queue func
	am.BattleQueueFuncMx.Lock()
	defer am.BattleQueueFuncMx.Unlock()

	bls, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.ReadyAt.IsNull(),
		boiler.BattleLobbyWhere.GeneratedBySystem.EQ(true),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load active public battle lobbies.")
		return terror.Error(err, "Failed to load active public battle lobbies.")
	}

	count := len(bls)

	if count >= publicLobbiesCount {
		return nil
	}

	// fill up battle lobbies
	for i := 0; i < publicLobbiesCount-count; i++ {
		bl := &boiler.BattleLobby{
			HostByID:              server.SupremacyBattleUserID,
			EntryFee:              decimal.Zero, // free to join
			FirstFactionCut:       decimal.NewFromFloat(0.75),
			SecondFactionCut:      decimal.NewFromFloat(0.25),
			ThirdFactionCut:       decimal.Zero,
			EachFactionMechAmount: 3,
			GeneratedBySystem:     true,
		}

		err = bl.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to insert public battle lobbies.")
			return terror.Error(err, "Failed to insert public battle lobbies.")
		}
	}

	return nil
}

func BroadcastMechQueueStatus(playerID string, mechIDs ...string) {
	if len(mechIDs) == 0 {
		return
	}

	mechInfo, err := db.OwnedMechsBrief(playerID, mechIDs...)
	if err != nil {
		return
	}

	ws.PublishMessage(fmt.Sprintf("/secure/user/%s/owned_mechs", playerID), server.HubKeyPlayerMechsBrief, mechInfo)
}

func BroadcastPlayerQueueStatus(playerID string) {
	resp := &server.PlayerQueueStatus{
		TotalQueued: 0,
		QueueLimit:  db.GetIntWithDefault(db.KeyPlayerQueueLimit, 10),
	}

	blms, err := boiler.BattleLobbiesMechs(
		boiler.BattleLobbiesMechWhere.OwnerID.EQ(playerID),
		boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
		boiler.BattleLobbiesMechWhere.EndedAt.IsNull(),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("player id", playerID).Err(err).Msg("Failed to load player battle queue mechs")
		return
	}

	if blms != nil {
		resp.TotalQueued = len(blms)
	}

	ws.PublishMessage(fmt.Sprintf("/secure/user/%s/queue_status", playerID), server.HubKeyPlayerQueueStatus, resp)
}
