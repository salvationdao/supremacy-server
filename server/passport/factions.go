package passport

import (
	"context"
	"server"

	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/hub"
)

type FactionAllResponse struct {
	Factions []*server.Faction `json:"payload"`
}

// FactionAll get all the factions from passport server
func (pp *Passport) FactionAll(callback func(msg []byte)) {
	pp.send <- &Message{
		Key:           "FACTION:ALL",
		TransactionID: uuid.Must(uuid.NewV4()).String(),
		Callback:      callback,
	}
}

type FactionStatSend struct {
	FactionStat     *server.FactionStat `json:"factionStat"`
	ToUserID        *server.UserID      `json:"toUserID,omitempty"`
	ToUserSessionID *hub.SessionID      `json:"toUserSessionID,omitempty"`
}

// FactionStatsSend send faction stat to passport serer
func (pp *Passport) FactionStatsSend(ctx context.Context, factionStatSends []*FactionStatSend) error {

	pp.send <- &Message{
		Key: "SUPREMACY:FACTION_STAT_SEND",
		Payload: struct {
			FactionStatSends []*FactionStatSend `json:"factionStatSends"`
		}{
			FactionStatSends: factionStatSends,
		},
		TransactionID: uuid.Must(uuid.NewV4()).String(),
	}
	return nil
}
