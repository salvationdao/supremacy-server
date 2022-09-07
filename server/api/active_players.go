package api

import (
	"fmt"
	"github.com/sasha-s/go-deadlock"
	"golang.org/x/exp/slices"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/ninja-syndicate/ws"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type ActivePlayers struct {
	FactionID string
	Map       map[string]*ActiveStat
	deadlock.RWMutex

	// channel for debounce broadcast
	ActivePlayerListChan chan *ActivePlayerBroadcast
}

type ActiveStat struct {
	// player stat
	Player *server.PublicPlayer

	// active stat
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
			FactionID:            f.ID,
			Map:                  make(map[string]*ActiveStat),
			ActivePlayerListChan: make(chan *ActivePlayerBroadcast),
		}

		go ap.Run()

		go ap.debounceBroadcastActivePlayers()

		api.FactionActivePlayers[f.ID] = ap
	}

	ap := &ActivePlayers{
		FactionID:            "GLOBAL",
		Map:                  make(map[string]*ActiveStat),
		ActivePlayerListChan: make(chan *ActivePlayerBroadcast),
	}

	go ap.Run()

	go ap.debounceBroadcastActivePlayers()

	api.FactionActivePlayers["GLOBAL"] = ap

}

// CurrentFactionActivePlayer return a copy of current faction active player list
func (ap *ActivePlayers) CurrentFactionActivePlayer() []server.PublicPlayer {
	ap.RLock()
	defer ap.RUnlock()

	var players []server.PublicPlayer
	for _, as := range ap.Map {
		players = append(players, *as.Player)
	}

	return players
}

type ActivePlayerBroadcast struct {
	Players []server.PublicPlayer
}

func (ap *ActivePlayers) debounceBroadcastActivePlayers() {
	var result *ActivePlayerBroadcast

	interval := 500 * time.Millisecond
	timer := time.NewTimer(interval)

	for {
		select {
		case result = <-ap.ActivePlayerListChan:
			timer.Reset(interval)
		case <-timer.C:
			if result != nil {
				ws.PublishMessage(fmt.Sprintf("/faction/%s", ap.FactionID), HubKeyFactionActivePlayersSubscribe, result.Players)
				result = nil
			}
		}
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

	// collect active player list for broadcast
	var players []server.PublicPlayer

	// update player username
	ids := []string{}
	for playerID := range ap.Map {
		ids = append(ids, playerID)
	}

	if len(ids) == 0 {
		return
	}

	ps, err := boiler.Players(boiler.PlayerWhere.ID.IN(ids)).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Strs("player id list", ids).Err(err).Msg("Failed to get players")
		return
	}

	for playerID, activeStat := range ap.Map {
		// update player detail
		if idx := slices.IndexFunc(ps, func(p *boiler.Player) bool { return p.ID == playerID }); idx != -1 {
			activeStat.Player = server.PublicPlayerFromBoiler(ps[idx])
		}

		// skip, if active stat is not expired
		if activeStat.ExpiredAt.After(now) {

			players = append(players, *activeStat.Player)
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
		} else {
			// update player active log
			pvl.InactiveAt = null.TimeFrom(time.Now())
			_, err = pvl.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerActiveLogColumns.InactiveAt))
			if err != nil {
				gamelog.L.Error().Str("player id", playerID).Err(err).Msg("Failed update player inactive log")
			}
		}

		delete(ap.Map, playerID)
	}

	// broadcast current online player
	ap.ActivePlayerListChan <- &ActivePlayerBroadcast{
		Players: players,
	}
}

func (ap *ActivePlayers) activePlayerUpdate(p *boiler.Player) {
	ap.Lock()
	defer ap.Unlock()

	var players []server.PublicPlayer
	for playerID, activeStat := range ap.Map {
		if playerID == p.ID {
			activeStat.Player = server.PublicPlayerFromBoiler(p)
		}
		players = append(players, *activeStat.Player)
	}

	// broadcast current online player
	ap.ActivePlayerListChan <- &ActivePlayerBroadcast{
		Players: players,
	}
}

func (ap *ActivePlayers) Set(playerID string, isActive bool) error {
	ap.Lock()
	defer ap.Unlock()

	if isActive {
		err := ap.add(playerID)
		if err != nil {
			return terror.Error(err, "Failed to add player onto active player map")
		}
	} else {
		err := ap.remove(playerID)
		if err != nil {
			return terror.Error(err, "Failed to remove player from active player map")
		}
	}

	return nil
}

func (ap *ActivePlayers) add(playerID string) error {
	now := time.Now()

	// check player's active stat is in the list
	as, ok := ap.Map[playerID]
	if ok {
		// if exists, expend player expiry for another two minutes
		as.ExpiredAt = now.Add(2 * time.Minute)
		return nil
	}

	// Otherwise, add player into the map

	// get player
	player, err := boiler.Players(
		qm.Select(
			boiler.PlayerColumns.ID,
			boiler.PlayerColumns.Username,
			boiler.PlayerColumns.Gid,
			boiler.PlayerColumns.FactionID,
		),
		boiler.PlayerWhere.ID.EQ(playerID),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("player id", playerID).Err(err).Msg("Failed to get player from db")
		return terror.Error(err, "Failed to get player from db")
	}

	pp := &server.PublicPlayer{
		ID:        player.ID,
		Username:  player.Username,
		Gid:       player.Gid,
		FactionID: player.FactionID,
		AboutMe:   player.AboutMe,
		Rank:      player.Rank,
		CreatedAt: player.CreatedAt,
	}

	ap.Map[playerID] = &ActiveStat{
		Player:    pp,
		ActivedAt: now,
		ExpiredAt: now.Add(2 * time.Minute),
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

func (ap *ActivePlayers) remove(playerID string) error {
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
