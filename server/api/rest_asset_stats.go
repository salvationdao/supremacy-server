package api

import (
	"context"
	"fmt"
	"net/http"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/ninja-software/terror/v2"
)

type AssetStatsController struct {
	API *API
}

func AssetStatsRouter(api *API) chi.Router {
	c := &AssetStatsController{
		API: api,
	}
	r := chi.NewRouter()
	r.Get("/mech", WithError(c.GetMechStatPercentage))

	return r
}

type GetMechStatPercentageResponse struct {
	Total      int64 `json:"total"`
	Percentile uint8 `json:"percentile"`
	Percentage uint8 `json:"percentage"`
}

func (sc *AssetStatsController) GetMechStatPercentage(w http.ResponseWriter, r *http.Request) (int, error) {
	stat := r.URL.Query().Get("stat")     // the stat identifier e.g. speed
	value := r.URL.Query().Get("value")   // the value of the stat e.g. 2000
	global := r.URL.Query().Get("global") // indicates whether or not to compare stats from all mechs or mechs of that type e.g. true

	valueInt, err := strconv.Atoi(value)
	if err != nil {
		gamelog.L.Error().
			Str("stat", stat).
			Str("value", value).
			Str("global", global).
			Str("db func", "ChassisStatRank").Err(err).Msg("unable to convert stat value to int")
	}

	switch stat {
	case boiler.ChassisColumns.ShieldRechargeRate:
	case boiler.ChassisColumns.HealthRemaining:
	case boiler.ChassisColumns.WeaponHardpoints:
	case boiler.ChassisColumns.TurretHardpoints:
	case boiler.ChassisColumns.UtilitySlots:
	case boiler.ChassisColumns.Speed:
	case boiler.ChassisColumns.MaxHitpoints:
	case boiler.ChassisColumns.MaxShield:
		break
	default:
		gamelog.L.Error().Str("stat", stat).Msg("invalid mech stat identifier")
		return http.StatusBadRequest, terror.Error(fmt.Errorf("invalid mech stat identifier"))
	}

	var total int
	var max int
	var min int
	err = gamedb.Conn.QueryRow(context.Background(), fmt.Sprintf(`
	SELECT
		count(id),
		max("%[1]s"),
		min("%[1]s")
	FROM
		chassis
`, stat)).Scan(&total, &max, &min)
	if err != nil {
		gamelog.L.Error().
			Str("stat", stat).
			Str("value", value).
			Str("global", global).
			Str("db func", "ChassisStatRank").Err(err).Msg("unable to get max value of chassis stat")
		return http.StatusInternalServerError, err
	}

	if max-min <= 0 {
		return helpers.EncodeJSON(w, GetMechStatPercentageResponse{
			int64(total),
			0,
			100,
		})
	}

	percentage := uint8(float64(valueInt-min) * 100 / float64(max-min))
	return helpers.EncodeJSON(w, GetMechStatPercentageResponse{
		int64(total),
		100 - percentage,
		percentage,
	})
}
