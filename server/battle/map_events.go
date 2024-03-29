package battle

import (
	"github.com/sasha-s/go-deadlock"
	"server/gamelog"
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

const TheHiveMapName string = "TheHive" // Would prefer to check uuid but it changes between seeds

type MapEventList struct {
	Landmines map[uint16]Landmine
	HiveState []bool

	mapName string

	deadlock.RWMutex
}

func NewMapEventList(mapName string) *MapEventList {
	return &MapEventList{
		Landmines: make(map[uint16]Landmine),
		HiveState: make([]bool, 589),
		mapName:   mapName,
	}
}

type Landmine struct {
	ID        uint16 `json:"id"`
	FactionNo byte   `json:"faction"`
	X         int32  `json:"x"`
	Y         int32  `json:"y"`
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
			for i := 0; i < hexes; i++ {
				hexID := helpers.BytesToUInt16(payload[offset : offset+2])
				offset += 3 // (skip time offset)
				if hexID > 589 {
					gamelog.L.Warn().Msgf(`MapEventTypeHiveHexRaised received invalid id: %v`, hexID)
					break
				}
				mel.UpdateHexState(hexID, true)
			}

		case MapEventTypeHiveHexLowered:
			hexes := int(helpers.BytesToUInt16(payload[offset : offset+2]))
			offset += 2
			for i := 0; i < hexes; i++ {
				hexID := helpers.BytesToUInt16(payload[offset : offset+2])
				offset += 3 // (skip time offset)
				if hexID > 589 {
					gamelog.L.Warn().Msgf(`MapEventTypeHiveHexLowered received invalid id: %v`, hexID)
					break
				}
				mel.UpdateHexState(hexID, false)
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

func (mel *MapEventList) UpdateHexState(hexID uint16, raised bool) {
	mel.Lock()
	defer mel.Unlock()

	mel.HiveState[hexID] = raised
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
	if mel.mapName == TheHiveMapName {
		payload = append(payload, byte(MapEventTypeHiveState))
		packedHiveState := helpers.PackBooleansIntoBytes(mel.HiveState)
		payload = append(payload, packedHiveState...)
		messageCount++
	}

	if messageCount == 0 {
		return false, nil
	}
	payload[0] = messageCount

	return true, payload
}
