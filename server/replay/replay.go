package replay

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"net/http"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"strconv"
)

type RecordingRequest struct {
	ID       string                   `json:"id"`
	Stream   OvenmediaRecordingStream `json:"stream"`
	FilePath string                   `json:"filePath"`
	InfoPath string                   `json:"infoPath"`
}
type OvenmediaRecordingStream struct {
	Name string `json:"name"`
}

type RecordController string

const StartRecording RecordController = "startRecord"
const StopRecording RecordController = "stopRecord"

// ErrDontLogRecordingStatus For development and staging environments where we wouldn't want to log recording status
var ErrDontLogRecordingStatus = fmt.Errorf("can record is false and not logged")

func RecordReplayRequest(battle *boiler.Battle, replayID string, action RecordController) error {
	environment := server.Env()
	canRecord := db.GetBoolWithDefault(db.KeyCanRecordReplayStatus, false)
	if !canRecord {
		if !server.IsProductionEnv() {
			return ErrDontLogRecordingStatus
		}
		gamelog.L.Info().Msg("recording replay is turned off in kv. consider turning it on")
		return fmt.Errorf("recording replay is turned off in kv. consider turning it on")
	}

	req := RecordingRequest{
		ID: strconv.Itoa(battle.BattleNumber),
		Stream: OvenmediaRecordingStream{
			Name: fmt.Sprintf("%s-%s", server.Env(), battle.ArenaID),
		},
		FilePath: fmt.Sprintf("/recordings/%s/${Stream}/%s-%s.mp4", environment, replayID, strconv.Itoa(battle.BattleNumber)),
		InfoPath: fmt.Sprintf("/info/%s/${Stream}/%s-%s.xml", environment, replayID, strconv.Itoa(battle.BattleNumber)),
	}

	body, err := json.Marshal(req)
	if err != nil {
		return terror.Error(err, "Failed to marshal stream recording json")
	}

	baseURL := db.GetStrWithDefault(db.KeyOvenmediaAPIBaseUrl, "https://stream2.supremacy.game")

	resp, err := http.Post(fmt.Sprintf("%s/v1/vhosts/stream2.supremacy.game/apps/app:%s", baseURL, action), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return terror.Error(err, "Failed to start post recording")
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		gamelog.L.Error().Err(fmt.Errorf("response for replay recording status not 200")).Msg("ovenmedia returned a not 200 response while attempting recording")
		return terror.Error(fmt.Errorf("response for replay recording status not 200"))
	}
	gamelog.L.Info().Msg(fmt.Sprintf("Ovenmedia Recording Status: %s", action))

	return nil
}

func StopAllActiveRecording() error {
	activeRecordings, err := boiler.BattleReplays(
		boiler.BattleReplayWhere.RecordingStatus.EQ(boiler.RecordingStatusRECORDING),
	).All(gamedb.StdConn)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}

		return terror.Error(err, "Failed to get all active recording while stopping recordings")
	}

	for _, recording := range activeRecordings {
		battle, err := boiler.Battles(boiler.BattleWhere.ArenaID.EQ(recording.ArenaID), boiler.BattleWhere.ID.EQ(recording.BattleID)).One(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Str("recording_id", recording.ID).Msg("failed to get battle while stopping recording")
			continue
		}

		err = RecordReplayRequest(battle, recording.ID, StopRecording)
		if err != nil {
			gamelog.L.Error().Err(err).Str("recording_id", recording.ID).Msg("failed to stop battle recording")
		}
	}

	return nil
}
