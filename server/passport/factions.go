package passport

import (
	"math/rand"
	"server"

	"github.com/gofrs/uuid"
)

func RandomFaction() *server.Faction {
	randomIndex := rand.Intn(len(FakeFactions))
	return FakeFactions[randomIndex]
}

// NOTE: This is a set of dummy functions that demonstrate passport server actions

var FakeFactions = []*server.Faction{
	{
		ID:     server.FactionID(uuid.Must(uuid.FromString("98bf7bb3-1a7c-4f21-8843-458d62884060"))),
		Label:  "Red Mountain Offworld Mining Corporation",
		Colour: "#BB1C2A",
	},
	{
		ID:     server.FactionID(uuid.Must(uuid.FromString("7c6dde21-b067-46cf-9e56-155c88a520e2"))),
		Label:  "Boston Cybernetics",
		Colour: "#03AAF9",
	},
	{
		ID:     server.FactionID(uuid.Must(uuid.FromString("880db344-e405-428d-84e5-6ebebab1fe6d"))),
		Label:  "Zaibatsu Heavy Industries",
		Colour: "#263D4D",
	},
}
