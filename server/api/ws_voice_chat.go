package api

import (
	"context"
	"encoding/json"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"server"
	"server/db/boiler"
	"server/gamedb"
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
	ListenURL          string `json:"listen_url"`
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
		gamelog.L.Error().Str("log_name", "battle arena").Str("msg", string(payload)).Err(err).Msg("unable to unmarshal queue join")
		return err
	}
	activeVoiceStreams, err := boiler.VoiceStreams(
		boiler.VoiceStreamWhere.FactionID.EQ(factionID),
		boiler.VoiceStreamWhere.IsActive.EQ(true),
		boiler.VoiceStreamWhere.ArenaID.EQ(req.ArenaID),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get active voice streams")
	}

	respVoiceStream := []*VoiceStreamResp{}

	for _, stream := range activeVoiceStreams {
		rvs := &VoiceStreamResp{
			ListenURL:          stream.ListenStreamURL,
			IsFactionCommander: stream.SenderType == boiler.VoiceSenderTypeFACTION_COMMANDER,
		}

		if user.ID == stream.OwnerID {
			rvs.SendURL = stream.SendStreamURL
		}

		respVoiceStream = append(respVoiceStream, rvs)
	}

	reply(respVoiceStream)

	return nil
}

func (vcs *VoiceStreamController) JoinFactionCommander(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {

	return nil
}
