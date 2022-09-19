package api

import (
	"context"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"server"
	"server/db/boiler"
	"server/gamedb"
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

func (vcs *VoiceStreamController) VoiceStreamSubscribe(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	activeVoiceStreams, err := boiler.VoiceStreams(
		boiler.VoiceStreamWhere.FactionID.EQ(factionID),
		boiler.VoiceStreamWhere.IsActive.EQ(true),
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
