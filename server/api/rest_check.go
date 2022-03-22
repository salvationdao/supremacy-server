package api

import (
	"context"
	"fmt"
	"net/http"
	"server/battle"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"time"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

// CheckController holds connection data for handlers
type CheckController struct {
	Conn        db.Conn
	Log         *zerolog.Logger
	BattleArena *battle.Arena
}

func CheckRouter(log *zerolog.Logger, conn db.Conn, battleArena *battle.Arena) chi.Router {
	c := &CheckController{
		Conn:        conn,
		Log:         log,
		BattleArena: battleArena,
	}
	r := chi.NewRouter()
	r.Get("/", c.Check)
	r.Get("/game", c.CheckGame)
	r.Get("/game/kill/next", c.CheckGame)

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

	// get current battle
	ba := c.BattleArena.Battle()

	if ba != nil {
		now := time.Now()
		diff := now.Sub(ba.StartedAt)

		// check if current battle over 15 mins
		if diff.Minutes() > 15 {
			msg := fmt.Sprintf("current battle over 15 mins, battle started at: %s (%f mins ago)",
				ba.StartedAt.String(),
				diff.Minutes())

			c.Log.Err(err).Msg(msg)
			_, err = w.Write([]byte(msg))
			if err != nil {
				c.Log.Err(err).Msg("failed to send")
			}

		}

		// get contributions for the last  2 mins
		_, err := ba.BattleContributions(boiler.BattleContributionWhere.ContributedAt.GT(now.Add(-2 * time.Minute))).One(gamedb.StdConn)
		if err != nil {
			msg := "there has been no contributions on the last 2 mins"
			c.Log.Err(err).Msg(msg)
			_, err = w.Write([]byte("\n" + msg))
			if err != nil {
				c.Log.Err(err).Msg("failed to send")
			}
		}

	}

	_, err = w.Write([]byte("\nok"))
	if err != nil {
		c.Log.Err(err).Msg("failed to send")
	}

}

// CheckGame return a game stat check
func (c *CheckController) CheckGame(w http.ResponseWriter, r *http.Request) {
	err := check(r.Context(), c.Conn)
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
