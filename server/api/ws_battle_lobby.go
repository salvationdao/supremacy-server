package api

import (
	"context"
	"github.com/ninja-syndicate/ws"
	"server/db/boiler"
)

func BattleLobbyController(api *API) {
	api.SecureUserFactionCommand(HubKeyBattleLobbyCreate, api.BattleLobbyCreate)
	api.SecureUserFactionCommand(HubKeyBattleLobbyJoin, api.BattleLobbyJoin)

}

type BattleLobbyCreateRequest struct {
	Payload struct {
		MechIDs []string `json:"mechIDs"`
	} `json:"payload"`
}

const HubKeyBattleLobbyCreate = "BATTLE:LOBBY:CREATE"

func (api *API) BattleLobbyCreate(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {

	return nil
}

const HubKeyBattleLobbyJoin = "BATTLE:LOBBY:JOIN"

func (api *API) BattleLobbyJoin(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {

	return nil
}

// subscriptions

const HubKeyBattleLobbyListUpdate = "BATTLE:LOBBY:LIST:UPDATE"

func (api *API) LobbyListUpdate(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {

	return nil
}
