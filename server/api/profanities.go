package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ninja-software/terror/v2"
)

type AddPhraseToProfanityDictionaryRequest struct {
	Phrase string `json:"phrase"`
}

func (api *API) AddPhraseToProfanityDictionary(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &AddPhraseToProfanityDictionaryRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("invaid request %w", err))
	}

	defer r.Body.Close()

	if req.Phrase == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("phrase cannot be empty"))
	}

	err = api.ProfanityManager.AddToDictionary(req.Phrase)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, fmt.Sprintf("failed to add phrase to dictionary %s", req.Phrase))
	}

	return http.StatusOK, nil
}
