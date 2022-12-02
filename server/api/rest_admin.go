package api

import "github.com/go-chi/chi/v5"

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

	return r
}
