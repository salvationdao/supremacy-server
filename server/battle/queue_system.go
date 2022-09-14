package battle

import (
	"database/sql"
	"errors"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"
)

func (am *ArenaManager) SendBattleQueueFunc(fn func() error) error {
	am.BattleQueueFuncMx.Lock()
	defer am.BattleQueueFuncMx.Unlock()
	return fn()
}

func BroadcastBattleLobbyUpdate(battleLobbyID string) {
	bl, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.ID.EQ(battleLobbyID),
		qm.Load(boiler.BattleLobbyRels.HostBy),
		qm.Load(boiler.BattleLobbyRels.GameMap),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("battle lobby id", battleLobbyID).Msg("Failed to query battle lobby")
		return
	}

	if bl == nil {
		return
	}

	resp, err := server.BattleLobbiesFromBoiler([]*boiler.BattleLobby{bl})
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

	// load game maps
	gameMaps, err := boiler.GameMaps().All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load game map.")
		return terror.Error(err, "Failed to load game map.")
	}

	// fill up battle lobbies
	for i := 0; i < publicLobbiesCount-count; i++ {
		gameMap := gameMaps[i%len(gameMaps)]

		bl := &boiler.BattleLobby{
			HostByID:              server.SupremacyBattleUserID,
			EntryFee:              decimal.Zero, // free to join
			FirstFactionCut:       decimal.NewFromFloat(0.75),
			SecondFactionCut:      decimal.NewFromFloat(0.25),
			ThirdFactionCut:       decimal.Zero,
			EachFactionMechAmount: 3,
			GameMapID:             gameMap.ID,
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
