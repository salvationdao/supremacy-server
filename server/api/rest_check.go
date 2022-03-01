package api

import (
	"context"
	"net/http"
	"server/db"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

// CheckController holds connection data for handlers
type CheckController struct {
	Conn db.Conn
	Log  *zerolog.Logger
}

func CheckRouter(log *zerolog.Logger, conn db.Conn) chi.Router {
	c := &CheckController{
		Conn: conn,
		Log:  log,
	}
	r := chi.NewRouter()
	r.Get("/", c.Check)

	return r
}

func (c *CheckController) Check(w http.ResponseWriter, r *http.Request) {
	err := check(context.Background(), c.Conn)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			c.Log.Err(err).Msg("failed to send")
			return
		}
	}
	_, err = w.Write([]byte("ok"))
	if err != nil {
		c.Log.Err(err).Msg("failed to send")
	}
}
