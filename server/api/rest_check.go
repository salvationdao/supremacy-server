package api

import (
	"fmt"
	"net/http"
	"server"
	"server/battle"
	"server/db/boiler"
	"server/gamedb"
	"time"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"server/gamelog"

	"github.com/go-chi/chi/v5"

	DatadogTracer "github.com/ninja-syndicate/hub/ext/datadog"
)

// CheckController holds connection data for handlers
type CheckController struct {
	BattleArena       *battle.Arena
	Telegram          server.Telegram
	IsClientConnected func() error
}

func CheckRouter(battleArena *battle.Arena, telegram server.Telegram, IsClientConnected func() error) chi.Router {
	c := &CheckController{
		BattleArena:       battleArena,
		Telegram:          telegram,
		IsClientConnected: IsClientConnected,
	}
	r := chi.NewRouter()
	r.Get("/", c.Check)
	r.Get("/game-connection", c.CheckGameConnection)
	r.Get("/game-length", c.CheckGameLength)
	r.Get("/game-contributes", c.CheckGameContributes)

	return r
}

func (c *CheckController) Check(w http.ResponseWriter, r *http.Request) {
	err := check()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, wErr := w.Write([]byte(err.Error()))
		if wErr != nil {
			DatadogTracer.HttpFinishSpan(r.Context(), http.StatusInternalServerError, wErr)
			gamelog.L.Err(err).Msg("failed to send")
		} else {
			DatadogTracer.HttpFinishSpan(r.Context(), http.StatusInternalServerError, err)
		}
		return
	}

	_, err = w.Write([]byte("\nok"))
	if err != nil {
		gamelog.L.Err(err).Msg("failed to send")
		DatadogTracer.HttpFinishSpan(r.Context(), http.StatusInternalServerError, err)
		return
	}
	DatadogTracer.HttpFinishSpan(r.Context(), http.StatusOK, nil)
}

// CheckGameConnection returns if gameclient is connected
func (c *CheckController) CheckGameConnection(w http.ResponseWriter, r *http.Request) {
	// check connection to gameclient
	err := c.IsClientConnected()
	if err != nil {
		gamelog.L.Err(err).Msg("gameclient not online in check/game call")
		w.WriteHeader(http.StatusFailedDependency)
		_, wErr := w.Write([]byte(err.Error()))
		if wErr != nil {
			DatadogTracer.HttpFinishSpan(r.Context(), http.StatusFailedDependency, wErr)
			gamelog.L.Err(err).Msg("failed to send")
		} else {
			DatadogTracer.HttpFinishSpan(r.Context(), http.StatusFailedDependency, err)
		}
		return
	}

	_, err = w.Write([]byte("ok"))
	if err != nil {
		gamelog.L.Err(err).Msg("failed to send")
		DatadogTracer.HttpFinishSpan(r.Context(), http.StatusInternalServerError, err)
		return
	}
	DatadogTracer.HttpFinishSpan(r.Context(), http.StatusOK, nil)
}

// CheckGameLength returns game length
func (c *CheckController) CheckGameLength(w http.ResponseWriter, r *http.Request) {
	// get latest current battle
	latestBattle, err := boiler.Battles(qm.OrderBy(fmt.Sprintf("%s DESC", boiler.BattleColumns.BattleNumber)), qm.Limit(1)).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Err(err).Msg("failed to retrieve battle in check/game call")
		w.WriteHeader(http.StatusFailedDependency)
		_, wErr := w.Write([]byte(err.Error()))
		if wErr != nil {
			DatadogTracer.HttpFinishSpan(r.Context(), http.StatusFailedDependency, wErr)
			gamelog.L.Err(err).Msg("failed to send")
		} else {
			DatadogTracer.HttpFinishSpan(r.Context(), http.StatusFailedDependency, err)
		}
		return
	}

	minutesAgoBattleStarted := time.Now().Sub(latestBattle.StartedAt).Minutes()

	//if battle started longer than 10 minutes ago, possible issue
	if minutesAgoBattleStarted >= 10 {
		err := fmt.Errorf("battle started %f minutes ago, possible hung battle (battle number: %d, battle id: %s)", minutesAgoBattleStarted, latestBattle.BattleNumber, latestBattle.ID)
		gamelog.L.Err(err).
			Float64("minutesAgoBattleStarted", minutesAgoBattleStarted).
			Str("latestBattle.StartedAt", latestBattle.StartedAt.String()).
			Str("battle id", latestBattle.ID).
			Int("battle number", latestBattle.BattleNumber).
			Msg("minutesAgoBattleStarted >= 10 in check/game call")
		w.WriteHeader(http.StatusFailedDependency)
		_, wErr := w.Write([]byte(err.Error()))
		if wErr != nil {
			DatadogTracer.HttpFinishSpan(r.Context(), http.StatusFailedDependency, wErr)
			gamelog.L.Err(err).Msg("failed to send")
		} else {
			DatadogTracer.HttpFinishSpan(r.Context(), http.StatusFailedDependency, err)
		}
		return
	}

	_, err = w.Write([]byte("ok"))
	if err != nil {
		gamelog.L.Err(err).Msg("failed to send")
		DatadogTracer.HttpFinishSpan(r.Context(), http.StatusInternalServerError, err)
		return
	}
	DatadogTracer.HttpFinishSpan(r.Context(), http.StatusOK, nil)
}

// CheckGameContributes returns error if length > 2 and 0 contributes
func (c *CheckController) CheckGameContributes(w http.ResponseWriter, r *http.Request) {

	// get latest current battle
	latestBattle, err := boiler.Battles(qm.OrderBy(fmt.Sprintf("%s DESC", boiler.BattleColumns.BattleNumber)), qm.Limit(1)).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Err(err).Msg("failed to retrieve battle in check/game call")
		w.WriteHeader(http.StatusFailedDependency)
		_, wErr := w.Write([]byte(err.Error()))
		if wErr != nil {
			DatadogTracer.HttpFinishSpan(r.Context(), http.StatusFailedDependency, wErr)
			gamelog.L.Err(err).Msg("failed to send")
		} else {
			DatadogTracer.HttpFinishSpan(r.Context(), http.StatusFailedDependency, err)
		}
		return
	}

	minutesAgoBattleStarted := time.Now().Sub(latestBattle.StartedAt).Minutes()

	if minutesAgoBattleStarted >= 2 {
		contributes, err := boiler.BattleContributions(boiler.BattleContributionWhere.BattleID.EQ(latestBattle.ID)).Count(gamedb.StdConn)
		if err != nil {
			gamelog.L.Err(err).
				Str("battle id", latestBattle.ID).
				Int("battle number", latestBattle.BattleNumber).
				Msg("failed to retrieve battle contribute count in check/game call")
			w.WriteHeader(http.StatusFailedDependency)
			_, wErr := w.Write([]byte(err.Error()))
			if wErr != nil {
				DatadogTracer.HttpFinishSpan(r.Context(), http.StatusFailedDependency, wErr)
				gamelog.L.Err(err).Msg("failed to send")
			} else {
				DatadogTracer.HttpFinishSpan(r.Context(), http.StatusFailedDependency, err)
			}
			return
		}
		if contributes == 0 {
			err := fmt.Errorf("0 contributes after battle started %f minutes ago, possible hung battle (battle number: %d, battle id: %s)", minutesAgoBattleStarted, latestBattle.BattleNumber, latestBattle.ID)
			gamelog.L.Err(err).
				Float64("minutesAgoBattleStarted", minutesAgoBattleStarted).
				Str("latestBattle.StartedAt", latestBattle.StartedAt.String()).
				Str("battle id", latestBattle.ID).
				Int("battle number", latestBattle.BattleNumber).
				Msg("minutesAgoBattleStarted >= 10 in check/game call")
			w.WriteHeader(http.StatusFailedDependency)
			_, wErr := w.Write([]byte(err.Error()))
			if wErr != nil {
				DatadogTracer.HttpFinishSpan(r.Context(), http.StatusFailedDependency, wErr)
				gamelog.L.Err(err).Msg("failed to send")
			} else {
				DatadogTracer.HttpFinishSpan(r.Context(), http.StatusFailedDependency, err)
			}
			return
		}
	}

	_, err = w.Write([]byte("ok"))
	if err != nil {
		gamelog.L.Err(err).Msg("failed to send")
		DatadogTracer.HttpFinishSpan(r.Context(), http.StatusInternalServerError, err)
		return
	}
	DatadogTracer.HttpFinishSpan(r.Context(), http.StatusOK, nil)
}
