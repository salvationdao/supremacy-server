package api

import (
	"encoding/json"
	"net/http"
	"server/db"

	"github.com/go-chi/chi/v5"
	"github.com/ninja-syndicate/ws"
)

func AdminRoutes(api *API, key string) chi.Router {
	r := chi.NewRouter()

	r.Post("/global_announcement", WithToken(key, WithError(api.GlobalAnnouncementSend)))
	r.Delete("/global_announcement", WithToken(key, WithError(api.GlobalAnnouncementDelete)))

	r.Post("/chat_shadowban", WithToken(key, WithError(api.ShadowbanChatPlayer)))
	r.Post("/chat_shadowban/remove", WithToken(key, WithError(api.ShadowbanChatPlayerRemove)))
	r.Get("/chat_shadowban/list", WithToken(key, WithError(api.ShadowbanChatPlayerList)))

	r.Post("/prod/give-crate", WithToken(key, WithError(api.ProdGiveCrate)))

	r.Post("/video_server", WithToken(key, WithError(api.CreateStreamHandler)))
	r.Get("/video_server", WithError(api.GetStreamsHandler))

	r.Delete("/video_server", WithToken(key, WithError(api.DeleteStreamHandler)))
	r.Post("/close_stream", WithToken(key, WithError(api.CreateStreamCloseHandler)))

	r.Post("/livestream", WithToken(key, WithError(api.LivestreamUpdate)))

	return r
}

type LivestreamReq struct {
	LivestreamURL string `json:"livestream_url"`
}

func (api *API) LivestreamUpdate(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &LivestreamReq{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusBadRequest, err
	}

	db.PutStr(db.KeyLivestreamURL, req.LivestreamURL)

	ws.PublishMessage("/public/livestream", HubKeyLivestream, req.LivestreamURL)

	return http.StatusOK, nil
}
