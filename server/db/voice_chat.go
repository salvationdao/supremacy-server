package db

import (
	"github.com/ninja-software/terror/v2"
	"server/api"
	"server/db/boiler"
	"server/gamedb"
)

func GetActiveVoiceChat(userID, factionID, arenaID string) ([]*api.VoiceStreamResp, error) {
	vcr := []*api.VoiceStreamResp{}

	activeVoiceStreams, err := boiler.VoiceStreams(
		boiler.VoiceStreamWhere.FactionID.EQ(factionID),
		boiler.VoiceStreamWhere.IsActive.EQ(true),
		boiler.VoiceStreamWhere.ArenaID.EQ(arenaID),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err, "Failed to get active voice streams")
	}

	for _, stream := range activeVoiceStreams {
		rvs := &api.VoiceStreamResp{
			ListenURL:          stream.ListenStreamURL,
			IsFactionCommander: stream.SenderType == boiler.VoiceSenderTypeFACTION_COMMANDER,
		}

		if userID == stream.OwnerID {
			rvs.SendURL = stream.SendStreamURL
		}

		vcr = append(vcr, rvs)
	}

	return vcr, nil
}
