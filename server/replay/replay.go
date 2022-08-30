package replay

import (
	"bytes"
	"crypto/tls"
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

var OvenMediaAuthKey string

type RecordController string

const StartRecording RecordController = "startRecord"
const StopRecording RecordController = "stopRecord"

// ErrDontLogRecordingStatus For development and staging environments where we wouldn't want to log recording status
var ErrDontLogRecordingStatus = fmt.Errorf("can record is false and not logged")

func RecordReplayRequest(battle *boiler.Battle, replayID string, action RecordController) error {
	canRecord := db.GetBoolWithDefault(db.KeyCanRecordReplayStatus, false)
	if !canRecord {
		// stop recording already running recording
		replay, err := boiler.FindBattleReplay(gamedb.StdConn, replayID)
		if err == nil && err != sql.ErrNoRows {
			if replay.RecordingStatus == boiler.RecordingStatusRECORDING && action == StopRecording {
				err := recordPostRequest(replayID, battle.ArenaID, battle.BattleNumber, StopRecording)
				if err != nil {
					gamelog.L.Error().Err(err).Str("replay_id", replayID).Str("battle_id", battle.ID).Msg("Failed to post recording request")
					return err
				}
				// return here to stop recording
				return nil
			}
		}
		if !server.IsProductionEnv() {
			return ErrDontLogRecordingStatus
		}
		gamelog.L.Info().Msg("recording replay is turned off in kv. consider turning it on")
		return fmt.Errorf("recording replay is turned off in kv. consider turning it on")
	}

	err := recordPostRequest(replayID, battle.ArenaID, battle.BattleNumber, action)
	if err != nil {
		gamelog.L.Error().Err(err).Str("replay_id", replayID).Str("battle_id", battle.ID).Msg("Failed to post recording request")
		return err
	}

	gamelog.L.Info().Msg(fmt.Sprintf("Ovenmedia Recording Status: %s", action))

	return nil
}

func recordPostRequest(replayID, arenaID string, battleNumber int, action RecordController) error {
	req := RecordingRequest{
		ID: replayID,
		Stream: OvenmediaRecordingStream{
			Name: fmt.Sprintf("%s-%s", server.Env(), arenaID),
		},
		FilePath: fmt.Sprintf("/recordings/%s/${Stream}/%s-%s.mp4", server.Env(), replayID, strconv.Itoa(battleNumber)),
		InfoPath: fmt.Sprintf("/info/%s/${Stream}/%s-%s.xml", server.Env(), replayID, strconv.Itoa(battleNumber)),
	}

	body, err := json.Marshal(req)
	if err != nil {
		return terror.Error(err, "Failed to marshal stream recording json")
	}

	baseURL := db.GetStrWithDefault(db.KeyOvenmediaAPIBaseUrl, "https://stream2.supremacy.game:8082")

	request, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/vhosts/stream2.supremacy.game/apps/app:%s", baseURL, action), bytes.NewBuffer(body))
	if err != nil {
		return terror.Error(err, "Failed to start post recording")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	request.Header.Set("Authorization", OvenMediaAuthKey)
	request.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		return terror.Error(err, "Failed to start recording")
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
