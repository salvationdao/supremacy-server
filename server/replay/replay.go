package replay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"net/http"
	"server/db"
	"server/db/boiler"
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

func RecordReplayRequest(battle boiler.Battle, environment string, action RecordController) error {
	req := RecordingRequest{
		ID: strconv.Itoa(battle.BattleNumber),
		Stream: OvenmediaRecordingStream{
			Name: battle.ArenaID,
		},
		FilePath: fmt.Sprintf("/recordings/%s/${Stream}/%s.mp4", environment, strconv.Itoa(battle.BattleNumber)),
		InfoPath: fmt.Sprintf("/info/%s/${Stream}/%s.xml", environment, strconv.Itoa(battle.BattleNumber)),
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
		return terror.Error(fmt.Errorf("response for replay recording status not 200"))
	}

	return nil
}
