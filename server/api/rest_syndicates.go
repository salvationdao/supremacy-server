package api

import (
	"net/http"
	"server"
)

func (api *API) SyndicateMotionIssue(user *server.Player, w http.ResponseWriter, r *http.Request) (int, error) {

	return http.StatusOK, nil
}
