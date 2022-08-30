package replay

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"net/http"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"strconv"
	"time"
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
	// load battle reply from given replay id
	replay, err := boiler.FindBattleReplay(gamedb.StdConn, replayID)
	if err != nil {
		return terror.Error(err, "Failed to load battle replay.")
	}

	canRecord := db.GetBoolWithDefault(db.KeyCanRecordReplayStatus, false)
	switch replay.RecordingStatus {
	case boiler.RecordingStatusIDLE:
		if !canRecord {
			// do not log error if it is not prod
			if !server.IsProductionEnv() {
				return ErrDontLogRecordingStatus
			}

			// return error
			gamelog.L.Info().Msg("recording replay is turned off in kv. consider turning it on")
			return fmt.Errorf("recording replay is turned off in kv. consider turning it on")
		}
	case boiler.RecordingStatusRECORDING:
		// return error if action is not "stop recording"
		if action != StopRecording {
			return terror.Error(fmt.Errorf("video recording is already started"), "Video recording is already started.")
		}

		// skip, if can not record
		if !canRecord {
			return nil
		}
	case boiler.RecordingStatusSTOPPED:
		return terror.Error(fmt.Errorf("video record is already stopped"), "Video record is already stopped.")
	}

	err = recordPostRequest(replayID, battle.ArenaID, battle.BattleNumber, action)
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
	client := &http.Client{
		Transport: tr,
		Timeout:   3 * time.Second, // set timeout prevent hanging
	}

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
		qm.Load(boiler.BattleReplayRels.Battle),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get all active recording while stopping recordings")
	}

	for _, recording := range activeRecordings {
		err = RecordReplayRequest(recording.R.Battle, recording.ID, StopRecording)
		if err != nil {
			gamelog.L.Error().Err(err).Str("recording_id", recording.ID).Msg("failed to stop battle recording")
		}

		recording.StoppedAt = null.TimeFrom(time.Now())
		recording.RecordingStatus = boiler.RecordingStatusSTOPPED
		_, err = recording.Update(
			gamedb.StdConn,
			boil.Whitelist(
				boiler.BattleReplayColumns.StoppedAt,
				boiler.BattleReplayColumns.RecordingStatus,
			),
		)
		if err != nil {
			gamelog.L.Error().Str("battle_id", recording.BattleID).Str("replay_id", recording.ID).Err(err).Msg("Failed to update replay session")
		}
	}

	return nil
}
