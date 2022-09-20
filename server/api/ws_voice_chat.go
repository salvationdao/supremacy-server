package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamelog"
)

func NewVoiceStreamController(api *API) {
	api.SecureUserFactionCommand(server.HubKeyVoiceStreamJoinFactionCommander, api.JoinFactionCommander)
}

func (api *API) VoiceStreamSubscribe(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	arenaID, ok := ctx.Value("arena_id").(string)
	if !ok || arenaID == "" {
		return terror.Error(fmt.Errorf("missing arena id"), "Missing arena id")
	}

	rvs, err := db.GetActiveVoiceChat(user.ID, factionID, arenaID)
	if err != nil {
		gamelog.L.Error().Str("user_id", user.ID).Err(err).Msg("failed to get active voice chats")
	}

	reply(rvs)

	return nil
}

type VoiceStreamReq struct {
	Payload struct {
		ArenaID string `json:"arena_id"`
	} `json:"payload"`
}

func (api *API) JoinFactionCommander(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &VoiceStreamReq{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	arena, err := api.ArenaManager.GetArena(req.Payload.ArenaID)
	if err != nil {
		return err
	}

	arena.VoiceChannel.Lock()
	defer arena.VoiceChannel.Unlock()

	// check if there is a faction commander

	// create one if there is no faction commander

	//

	return nil
}
