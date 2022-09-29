package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"net/http"
	"server"
	"server/db"
	"server/helpers"
)

func BattleRouter(api *API) chi.Router {
	r := chi.NewRouter()
	r.Get("/mech/{id}/destroyed_detail", WithError(api.MechDestroyedDetail))
	r.Get("/challenge_fund_amount", WithError(api.ChallengeFundAmount))

	return r
}

// MechDestroyedDetail return mech destroyed record if exists
func (api *API) MechDestroyedDetail(w http.ResponseWriter, r *http.Request) (int, error) {
	mechID := chi.URLParam(r, "id")

	if destroyedRecord := api.ArenaManager.WarMachineDestroyedDetail(mechID); destroyedRecord != nil {
		return helpers.EncodeJSON(w, destroyedRecord)
	}

	return http.StatusOK, nil
}

func (api *API) ChallengeFundAmount(w http.ResponseWriter, r *http.Request) (int, error) {
	challengeFundBalance := api.Passport.UserBalanceGet(uuid.FromStringOrNil(server.SupremacyChallengeFundUserID))
	bonusSupPerWinner := db.GetDecimalWithDefault(db.KeyBattleSupsRewardBonus, decimal.New(45, 18))

	return helpers.EncodeJSON(w, struct {
		ChallengeFundBalance decimal.Decimal `json:"challenge_fund_balance"`
		BonusSupsPerWinner   decimal.Decimal `json:"bonus_sups_per_winner"`
	}{
		challengeFundBalance,
		bonusSupPerWinner,
	})
}
