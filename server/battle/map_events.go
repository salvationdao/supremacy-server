package battle

import (
	"github.com/sasha-s/go-deadlock"
	"server/helpers"
)

type MapEventType byte

const (
	MapEventTypeAirstrikeExplosions MapEventType = iota
	MapEventTypeLandmineActivations
	MapEventTypeLandmineExplosions
	MapEventTypePickupLanded
	MapEventTypePickupUsed
	MapEventTypeHiveHexUpdate
)

type MapEventList struct {
	Landmines map[uint16]Landmine
	deadlock.RWMutex
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
				offset += 4 // (uint16 + skip uint16) skip time offset as server doesn't need to know about it
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
				offset += 4 // (uint16 + skip uint16) skip time offset as server doesn't need to know about it
				mel.RemoveLandmine(landmineID)
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
func (mel *MapEventList) Pack() [][]byte {
	mel.Lock()
	defer mel.Unlock()

	var mapEvents [][]byte

	// Landmines
	if len(mel.Landmines) > 0 {
		// Group landmines by faction (MapEventType_LandmineActivations sends each faction's landmines separately for optimised byte size messages)
		var landminesPerFaction [3][]Landmine
		for _, landmine := range mel.Landmines {
			if landmine.FactionNo == 0 || landmine.FactionNo > 3 {
				continue
			}
			index := landmine.FactionNo + 1
			landminesPerFaction[index] = append(landminesPerFaction[index], landmine)
		}

		for _, landmines := range landminesPerFaction {
			landminesEvent := []byte{
				byte(MapEventTypeLandmineExplosions),
			}
			landminesEvent = append(landminesEvent, helpers.UInt16ToBytes(uint16(len(landmines)))...)

			for _, landmine := range landmines {
				landminesEvent = append(landminesEvent, helpers.UInt16ToBytes(landmine.ID)...)
				landminesEvent = append(landminesEvent, helpers.UInt16ToBytes(0)...) // Time Offset
				landminesEvent = append(landminesEvent, helpers.IntToBytes(landmine.X)...)
				landminesEvent = append(landminesEvent, helpers.IntToBytes(landmine.Y)...)
			}

			mapEvents = append(mapEvents, landminesEvent)
		}
	}

	return mapEvents
}
