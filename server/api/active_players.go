package api

import (
	"time"

	"github.com/sasha-s/go-deadlock"
)

type ActivePlayers struct {
	Map map[string]*ActiveStat
	deadlock.Mutex
}

type ActiveStat struct {
	ActivedAt time.Time
	ExpiredAt time.Time
}

func (ap *ActivePlayers) Set(playerID string, isActive bool) {
	ap.Lock()
	defer ap.Unlock()

	if isActive {
		// update player active time

		// check player's active stat is in the list
		as, ok := ap.Map[playerID]
		if !ok {
			now := time.Now()
			// add player into the map
			ap.Map[playerID] = &ActiveStat{
				ActivedAt: now,
				ExpiredAt: now.Add(2 * time.Minute),
			}

			// store user active log into db

			return
		}

		// if exists, expend player expiry for another two minutes
		as.ExpiredAt = time.Now().Add(2 * time.Minute)

		return
	}

	// remove player from the list

}
