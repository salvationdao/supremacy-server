package api

import (
	"context"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"server/db/boiler"
	"server/gamedb"
)

type VoiceStreamController struct {
	API *API
}

func NewVoiceStreamController(api *API) *VoiceStreamController {
	vcs := &VoiceStreamController{API: api}

	return vcs
}

func (vcs *VoiceStreamController) VoiceStreamSubscribe(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	activeVoiceStreams, err := boiler.VoiceStreams(
		boiler.VoiceStreamWhere.FactionID.EQ(factionID),
		boiler.VoiceStreamWhere.IsActive.EQ(true),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get active voice streams")
	}

	respVoiceStream := []*boiler.VoiceStream{}

	for _, stream := range activeVoiceStreams {
		rvs := &boiler.VoiceStream{
			ListenStreamURL: stream.ListenStreamURL,
		}

		if user.ID == stream.OwnerID {
			rvs.SendStreamURL = stream.SendStreamURL
		}

		respVoiceStream = append(respVoiceStream, rvs)
	}

	reply(respVoiceStream)

	return nil
}
