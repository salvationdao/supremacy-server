package db

import (
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
	"server"
	"server/db/boiler"
	"server/gamedb"
)

func GetActiveVoiceChat(userID, factionID, arenaID string) ([]*server.VoiceStreamResp, error) {
	vcr := []*server.VoiceStreamResp{}

	activeVoiceStreams, err := boiler.VoiceStreams(
		boiler.VoiceStreamWhere.FactionID.EQ(factionID),
		boiler.VoiceStreamWhere.IsActive.EQ(true),
		boiler.VoiceStreamWhere.ArenaID.EQ(arenaID),
		qm.Load(boiler.VoiceStreamRels.Owner),
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

		rvs := &server.VoiceStreamResp{
			IsFactionCommander: stream.SenderType == boiler.VoiceSenderTypeFACTION_COMMANDER,
		}

		if stream.R.Owner != nil {
			rvs.Username = stream.R.Owner.Username
			rvs.UserGID = stream.R.Owner.Gid
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
