package api

import (
	"context"
	"net/http"
	"server"
	"server/battle"
	"server/db"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"

	DatadogTracer "github.com/ninja-syndicate/hub/ext/datadog"
)

// CheckController holds connection data for handlers
type CheckController struct {
	Conn        db.Conn
	Log         *zerolog.Logger
	BattleArena *battle.Arena
	Telegram    server.Telegram
}

func CheckRouter(log *zerolog.Logger, conn db.Conn, battleArena *battle.Arena, telegram server.Telegram) chi.Router {
	c := &CheckController{
		Conn:        conn,
		Log:         log,
		BattleArena: battleArena,
		Telegram:    telegram,
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
		_, wErr := w.Write([]byte(err.Error()))
		if wErr != nil {
			DatadogTracer.HttpFinishSpan(r.Context(), http.StatusInternalServerError, wErr)
			c.Log.Err(err).Msg("failed to send")
		} else {
			DatadogTracer.HttpFinishSpan(r.Context(), http.StatusInternalServerError, err)
		}
		return
	}

	//// get current battle
	//ba, err := boiler.Battles(qm.OrderBy(fmt.Sprintf("%s DESC", boiler.BattleColumns.BattleNumber)), qm.Limit(1)).One(gamedb.StdConn)
	//if err != nil {
	//	c.Log.Err(err).Msg("failed to retrieve battle")
	//	return
	//}

	//now := time.Now()
	//diff := now.Sub(ba.StartedAt)
	//
	//// check if current battle over 15 mins
	//if diff.Minutes() > 15 {
	//	ok = false
	//	w.WriteHeader(http.StatusGone)
	//	msg := fmt.Sprintf("current battle over 15 mins, battle started at: %s (%f mins ago)",
	//		ba.StartedAt.String(),
	//		diff.Minutes())
	//
	//	c.Log.Err(err).Str("battle_no", fmt.Sprintf("%d", ba.BattleNumber)).Msg(msg)
	//	_, err = w.Write([]byte(msg))
	//	if err != nil {
	//		c.Log.Err(err).Str("battle_no", fmt.Sprintf("%d", ba.BattleNumber)).Msg("failed to send")
	//	}
	//}
	//
	//// get contributions for the last  2 mins
	//btlContributions, err := ba.BattleContributions(boiler.BattleContributionWhere.ContributedAt.GT(now.Add(-10 * time.Minute))).All(gamedb.StdConn)
	//if err != nil && !errors.Is(err, sql.ErrNoRows) {
	//	c.Log.Err(err).Str("battle_no", fmt.Sprintf("%d", ba.BattleNumber)).Msg("failed to get battle contributions")
	//
	//}
	//
	//if len(btlContributions) <= 0 {
	//	ok = false
	//	w.WriteHeader(http.StatusGone)
	//	msg := "there has been no contributions on the last 10 mins"
	//	c.Log.Err(err).Str("battle_no", fmt.Sprintf("%d", ba.BattleNumber)).Msg(msg)
	//	_, err = w.Write([]byte("\n" + msg))
	//	if err != nil {
	//		c.Log.Err(err).Str("battle_no", fmt.Sprintf("%d", ba.BattleNumber)).Msg("failed to send")
	//	}
	//}

	_, err = w.Write([]byte("\nok"))
	if err != nil {
		c.Log.Err(err).Msg("failed to send")
	}
	DatadogTracer.HttpFinishSpan(r.Context(), http.StatusOK, nil)
}

// CheckGame return a game stat check
func (c *CheckController) CheckGame(w http.ResponseWriter, r *http.Request) {
	err := check(r.Context(), c.Conn)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, wErr := w.Write([]byte(err.Error()))
		if wErr != nil {
			c.Log.Err(err).Msg("failed to send")
			DatadogTracer.HttpFinishSpan(r.Context(), http.StatusInternalServerError, wErr)
		} else {
			DatadogTracer.HttpFinishSpan(r.Context(), http.StatusInternalServerError, err)
		}
		return
	}
	_, err = w.Write([]byte("ok"))
	if err != nil {
		c.Log.Err(err).Msg("failed to send")
		DatadogTracer.HttpFinishSpan(r.Context(), http.StatusInternalServerError, err)
	}
	DatadogTracer.HttpFinishSpan(r.Context(), http.StatusOK, nil)
}
