package api

import (
	"context"
	"encoding/json"
	"github.com/ninja-syndicate/ws"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamelog"
)

type VoiceStreamController struct {
	API *API
}

func NewVoiceStreamController(api *API) *VoiceStreamController {
	vcs := &VoiceStreamController{API: api}

	api.SecureUserFactionCommand(server.HubKeyVoiceStreamJoinFactionCommander, vcs.JoinFactionCommander)

	return vcs
}

type VoiceStreamResp struct {
	ListenURL          string `json:"listen_url,omitempty"`
	SendURL            string `json:"send_url,omitempty"`
	IsFactionCommander bool   `json:"is_faction_commander"`
}

type VoiceStreamReq struct {
	ArenaID string `json:"arena_id"`
}

func (vcs *VoiceStreamController) VoiceStreamSubscribe(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &VoiceStreamReq{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		gamelog.L.Error().Str("msg", string(payload)).Err(err).Msg("unable to unmarshal voice stream")
		return err
	}

	rvs, err := db.GetActiveVoiceChat(user.ID, factionID, req.ArenaID)
	if err != nil {
		gamelog.L.Error().Str("user_id", user.ID).Err(err).Msg("failed to get active voice chats")
	}

	reply(rvs)

	return nil
}

func (vcs *VoiceStreamController) JoinFactionCommander(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	
	return nil
}
