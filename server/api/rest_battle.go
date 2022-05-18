package api

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"server/battle"
	"server/helpers"
)

// BattleController holds connection data for handlers
type BattleController struct {
	BattleArena *battle.Arena
}

func BattleRouter(battleArena *battle.Arena) chi.Router {
	c := &BattleController{
		BattleArena: battleArena,
	}
	r := chi.NewRouter()
	r.Get("/mech/{id}/destroyed_detail", WithError(c.MechDestroyedDetail))

	return r
}

// MechDestroyedDetail return mech destroyed record if exists
func (bc *BattleController) MechDestroyedDetail(w http.ResponseWriter, r *http.Request) (int, error) {
	mechID := chi.URLParam(r, "id")

	if destroyedRecord := bc.BattleArena.WarMachineDestroyedDetail(mechID); destroyedRecord != nil {
		return helpers.EncodeJSON(w, destroyedRecord)
	}

	return http.StatusOK, nil
}
