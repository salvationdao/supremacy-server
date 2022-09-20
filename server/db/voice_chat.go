package db

import (
	"github.com/ninja-software/terror/v2"
	"golang.org/x/exp/slices"
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

	checkList := []string{}

	for _, stream := range activeVoiceStreams {
		if slices.Index(checkList, stream.OwnerID) != -1 {
			continue
		}

		checkList = append(checkList, stream.OwnerID)

		rvs := &api.VoiceStreamResp{
			IsFactionCommander: stream.SenderType == boiler.VoiceSenderTypeFACTION_COMMANDER,
		}

		if userID == stream.OwnerID {
			rvs.SendURL = stream.SendStreamURL
		} else {
			rvs.ListenURL = stream.ListenStreamURL
		}

		vcr = append(vcr, rvs)
	}

	return vcr, nil
}
