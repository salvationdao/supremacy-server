package api

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"server/battle"
	"server/helpers"
)

// BattleController holds connection data for handlers
type BattleController struct {
	ArenaManager *battle.ArenaManager
}

func BattleRouter(arenaManager *battle.ArenaManager) chi.Router {
	c := &BattleController{
		ArenaManager: arenaManager,
	}
	r := chi.NewRouter()
	r.Get("/mech/{id}/destroyed_detail", WithError(c.MechDestroyedDetail))

	return r
}

// MechDestroyedDetail return mech destroyed record if exists
func (bc *BattleController) MechDestroyedDetail(w http.ResponseWriter, r *http.Request) (int, error) {
	mechID := chi.URLParam(r, "id")

	if destroyedRecord := bc.ArenaManager.WarMachineDestroyedDetail(mechID); destroyedRecord != nil {
		return helpers.EncodeJSON(w, destroyedRecord)
	}

	return http.StatusOK, nil
}
