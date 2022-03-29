package api

import (
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/sasha-s/go-deadlock"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type ActivePlayers struct {
	FactionID string
	Map       map[string]*ActiveStat
	deadlock.Mutex
}

type ActiveStat struct {
	ActivedAt time.Time
	ExpiredAt time.Time
}

func (api *API) FactionActivePlayerSetup() {
	// get factions
	factions, err := boiler.Factions().All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to setup faction punish vote tracker")
		return
	}

	for _, f := range factions {
		ap := &ActivePlayers{
			FactionID: f.ID,
			Map:       make(map[string]*ActiveStat),
		}

		ap.Run()

		api.FactionActivePlayers[f.ID] = ap
	}
}

func (ap *ActivePlayers) Run() {
	for {
		// run check every minute
		time.Sleep(1 * time.Minute)
		ap.CheckExpiry()
	}
}

func (ap *ActivePlayers) CheckExpiry() {
	ap.Lock()
	defer ap.Unlock()

	now := time.Now()

	for playerID, activeStat := range ap.Map {

		// skip, if active stat is not expired
		if activeStat.ExpiredAt.After(now) {
			continue
		}

		// Otherwise, remove player from the list

		// get player active log
		pvl, err := boiler.PlayerActiveLogs(
			boiler.PlayerActiveLogWhere.PlayerID.EQ(playerID),
			boiler.PlayerActiveLogWhere.InactiveAt.IsNull(),
			qm.OrderBy(boiler.PlayerActiveLogColumns.ActiveAt+" DESC"),
		).One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Str("player id", playerID).Err(err).Msg("Failed to get player active log")
		}

		// update player active log
		pvl.InactiveAt = null.TimeFrom(time.Now())
		_, err = pvl.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerActiveLogColumns.InactiveAt))
		if err != nil {
			gamelog.L.Error().Str("player id", playerID).Err(err).Msg("Failed update player inactive log")
		}

		delete(ap.Map, playerID)
	}
}

func (ap *ActivePlayers) Set(playerID string, isActive bool) error {
	ap.Lock()
	defer ap.Unlock()

	if isActive {
		err := ap.Add(playerID)
		if err != nil {
			return terror.Error(err, "Failed to add player onto active player map")
		}
	} else {
		err := ap.Remove(playerID)
		if err != nil {
			return terror.Error(err, "Failed to remove player from active player map")
		}
	}

	return nil
}

func (ap *ActivePlayers) Add(playerID string) error {
	now := time.Now()

	// check player's active stat is in the list
	as, ok := ap.Map[playerID]
	if ok {
		// if exists, expend player expiry for another two minutes
		as.ExpiredAt = now.Add(2 * time.Minute)
		return nil
	}

	// Otherwise, add player into the map
	ap.Map[playerID] = &ActiveStat{
		ActivedAt: now,
		ExpiredAt: now.Add(2 * time.Minute),
	}

	// get player
	player, err := boiler.FindPlayer(gamedb.StdConn, playerID)
	if err != nil {
		return terror.Error(err, "Failed to get player from db")
	}

	// store user active log into db
	pvl := &boiler.PlayerActiveLog{
		PlayerID:  player.ID,
		FactionID: player.FactionID,
		ActiveAt:  now,
	}
	err = pvl.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Str("player id", playerID).Err(err).Msg("Failed to store active player into db")
		return terror.Error(err, "Failed to store active player into db")
	}
	return nil
}

func (ap *ActivePlayers) Remove(playerID string) error {
	// check player is in the list
	if _, ok := ap.Map[playerID]; !ok {
		// skip, if player is not in the list
		return nil
	}

	// remove player from the list
	delete(ap.Map, playerID)

	// get player active log
	pvl, err := boiler.PlayerActiveLogs(
		boiler.PlayerActiveLogWhere.PlayerID.EQ(playerID),
		boiler.PlayerActiveLogWhere.InactiveAt.IsNull(),
		qm.OrderBy(boiler.PlayerActiveLogColumns.ActiveAt+" DESC"),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get player active log")
	}

	// update player active log
	pvl.InactiveAt = null.TimeFrom(time.Now())
	_, err = pvl.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerActiveLogColumns.InactiveAt))
	if err != nil {
		return terror.Error(err, "Failed update player inactive log")
	}

	return nil
}

// GetPlayerIDs return a copy of current active player id list
func (ap *ActivePlayers) GetPlayerIDs() []string {
	ap.Lock()
	defer ap.Unlock()
	ids := []string{}
	for playerID := range ap.Map {
		ids = append(ids, playerID)
	}
	return ids
}
