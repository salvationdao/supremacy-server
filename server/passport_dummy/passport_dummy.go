package passport_dummy

import (
	"math/rand"
	"server"

	"github.com/gofrs/uuid"
)

func RandomFaction() *server.Faction {
	randomIndex := rand.Intn(len(FakeFactions))
	return FakeFactions[randomIndex]
}

type PassportDummy struct {
	authStuff string
}

func NewPassportDummy(authStuff string) *PassportDummy {
	newPP := &PassportDummy{authStuff: authStuff}

	return newPP
}

type PassportUser struct {
	ID            server.UserID   `json:"id"`
	Faction       *server.Faction `json:"faction"`
	ConnectPoint  int64           `json:"connectPoint"`
	SupremacyCoin int64           `json:"supremacyCoin"`
	PassportURL   string          `json:"passportURL"`
}

var defaultNamespaceUUID = uuid.Must(uuid.FromString("8f2d7180-bbe3-47b0-96ef-ee3e64697387"))

func (pp *PassportDummy) FakeUserLoginWithFaction(twitchUserID string) *server.User {
	// we will auth with passport, and we'll get a passport user and convert it to a server.user
	return &server.User{
		ID:            server.UserID(uuid.NewV3(defaultNamespaceUUID, twitchUserID)),
		Faction:       FakeFactions[0],
		FactionID:     FakeFactions[0].ID,
		ConnectPoint:  12345,
		SupremacyCoin: 12345,
		PassportURL:   "", // url we get from passport which has a token in the url to take them to their passport page
	}
}

func (pp *PassportDummy) FakeUserLoginWithoutFaction(twitchUserID string) *server.User {
	// we will auth with passport, and we'll get a passport user and convert it to a server.user
	return &server.User{
		ID:            server.UserID(uuid.NewV3(defaultNamespaceUUID, twitchUserID)),
		ConnectPoint:  12345,
		SupremacyCoin: 12345,
		PassportURL:   "https://dev.supremacygame.io/api/temp-random-faction", // url we get from passport which has a token in the url to take them to their passport page
	}
}
