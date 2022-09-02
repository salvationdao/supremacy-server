package battle

import (
	"github.com/gofrs/uuid"
	"github.com/sasha-s/go-deadlock"
	"server/helpers"
)

type MapEventType byte

const (
	MapEventTypeAirstrikeExplosions MapEventType = iota // The locations of airstrike missile impacts.
	MapEventTypeLandmineActivations                     // The id, location and faction of a mine that got activated.
	MapEventTypeLandmineExplosions                      // The ids of mines that exploded.
	MapEventTypeHiveState                               // The full state of The Hive map.
	MapEventTypeHiveHexRaised                           // The ids of the hexes that have recently raised.
	MapEventTypeHiveHexLowered                          // The ids of the hexes that have recently lowered.
)

const TheHiveMapID string = "bf84dd8e-e124-4c77-99a1-f515a81752b1"

type MapEventList struct {
	Landmines map[uint16]Landmine
	HiveState []bool

	mapID string

	deadlock.RWMutex
}

func NewMapEventList(mapID uuid.UUID) *MapEventList {
	return &MapEventList{
		Landmines: make(map[uint16]Landmine),
		HiveState: make([]bool, 589),
		mapID:     mapID.String(),
	}
}

type MapEvent interface {
	Pack() []byte
}

type Landmine struct {
	ID        uint16 `json:"id"`
	FactionNo byte   `json:"faction"`
	X         int32  `json:"x"`
	Y         int32  `json:"y"`
}

func (e *Landmine) Pack() []byte {
	var bytes []byte
	bytes = append(bytes)

	return bytes
}

func (mel *MapEventList) MapEventsUnpack(payload []byte) {
	offset := 0

	count := int(payload[offset])
	offset++
	for c := 0; c < count; c++ {
		mapEventType := MapEventType(payload[offset])
		offset++
		switch mapEventType {
		case MapEventTypeLandmineActivations:
			// Add Landmine
			landMineCount := int(helpers.BytesToUInt16(payload[offset : offset+2]))
			offset += 2
			factionNo := payload[offset]
			offset++
			for l := 0; l < landMineCount; l++ {
				landmineID := helpers.BytesToUInt16(payload[offset : offset+2])
				offset += 3 // (uint16 + skip byte) skip time offset as server doesn't need to know about it
				x := helpers.BytesToInt(payload[offset : offset+4])
				offset += 4
				y := helpers.BytesToInt(payload[offset : offset+4])
				offset += 4

				mel.AddLandmine(Landmine{
					ID:        landmineID,
					FactionNo: factionNo,
					X:         x,
					Y:         y,
				})

			}

		case MapEventTypeLandmineExplosions:
			// Remove Landmine
			landMineCount := int(helpers.BytesToUInt16(payload[offset : offset+2]))
			offset += 2
			for l := 0; l < landMineCount; l++ {
				landmineID := helpers.BytesToUInt16(payload[offset : offset+2])
				offset += 3 // (uint16 + skip byte) skip time offset as server doesn't need to know about it
				mel.RemoveLandmine(landmineID)
			}

		case MapEventTypeHiveHexRaised:
			hexes := int(helpers.BytesToUInt16(payload[offset : offset+2]))
			offset += 2
			for i := 0; i <= hexes; i++ {
				hexID := helpers.BytesToUInt16(payload[offset : offset+2])
				offset += 3 // (skip time offset)
				mel.HiveState[hexID] = true
			}
		case MapEventTypeHiveHexLowered:
			hexes := int(helpers.BytesToUInt16(payload[offset : offset+2]))
			offset += 2
			for i := 0; i <= hexes; i++ {
				hexID := helpers.BytesToUInt16(payload[offset : offset+2])
				offset += 3 // (skip time offset)
				mel.HiveState[hexID] = false
			}
		}
	}
}

func (mel *MapEventList) AddLandmine(landmine Landmine) {
	mel.Lock()
	defer mel.Unlock()

	mel.Landmines[landmine.ID] = landmine
}

func (mel *MapEventList) RemoveLandmine(landmineID uint16) {
	mel.Lock()
	defer mel.Unlock()

	delete(mel.Landmines, landmineID)
}

// Pack all information a new frontend client needs to know (eg: landmine, pickup locations and the hive state)
func (mel *MapEventList) Pack() (bool, []byte) {
	mel.Lock()
	defer mel.Unlock()

	payload := []byte{0} // prepend message count
	var messageCount byte = 0

	// Landmines
	if len(mel.Landmines) > 0 {
		// Group landmines by faction (MapEventTypeLandmineActivations sends each faction's landmines separately for optimised byte size messages)
		var landminesPerFaction [3][]Landmine
		for _, landmine := range mel.Landmines {
			if landmine.FactionNo == 0 || landmine.FactionNo > 3 {
				continue
			}
			index := landmine.FactionNo - 1
			landminesPerFaction[index] = append(landminesPerFaction[index], landmine)
		}

		for factionNo, landmines := range landminesPerFaction {
			landmineCount := len(landmines)
			if landmineCount == 0 {
				continue
			}

			payload = append(payload, byte(MapEventTypeLandmineActivations))
			payload = append(payload, helpers.UInt16ToBytes(uint16(landmineCount))...)
			payload = append(payload, byte(factionNo+1))

			for _, landmine := range landmines {
				payload = append(payload, helpers.UInt16ToBytes(landmine.ID)...)
				payload = append(payload, 255) // Time offset never go past 250, so 255 is used to mark an event that will spawn instantly with no animation
				payload = append(payload, helpers.IntToBytes(landmine.X)...)
				payload = append(payload, helpers.IntToBytes(landmine.Y)...)
			}
			messageCount++
		}
	}

	// The Hive State
	if mel.mapID == TheHiveMapID {
		payload = append(payload, byte(MapEventTypeHiveState))
		packedHiveState := helpers.PackBooleansIntoBytes(mel.HiveState)
		payload = append(payload, packedHiveState...)
	}

	if messageCount == 0 {
		return false, nil
	}
	payload[0] = messageCount

	return true, payload
}
