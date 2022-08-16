package player_abilities

import (
	"math"
	"server"
	"time"

	"github.com/sasha-s/go-deadlock"
)

type BlackoutEntry struct {
	GameCoords server.GameLocation
	CellCoords server.CellLocation
	ExpiresAt  time.Time

	deadlock.RWMutex
}

const BlackoutRadius = 20000

type ReadBlackoutEntryFunc func(*BlackoutEntry)

func (be *BlackoutEntry) Read(fn ReadBlackoutEntryFunc) {
	be.RLock()
	defer be.RUnlock()

	fn(be)
}

// ContainsPosition returns true if the specified position is within the blackout area
func (be *BlackoutEntry) ContainsPosition(position server.GameLocation) bool {
	be.RLock()
	defer be.RUnlock()

	c1 := position
	c2 := be.GameCoords
	d := math.Sqrt(math.Pow(float64(c2.X)-float64(c1.X), 2) + math.Pow(float64(c2.Y)-float64(c1.Y), 2))
	if d < float64(BlackoutRadius) {
		return true
	}
	return false
}

func (be *BlackoutEntry) IsExpired() bool {
	be.RLock()
	defer be.RUnlock()

	return time.Now().After(be.ExpiresAt)
}
