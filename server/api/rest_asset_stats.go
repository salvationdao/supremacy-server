package api

import (
	"net/http"

	"github.com/go-chi/chi"
)

type AssetStatsController struct {
	API *API
}

func AssetStatsRouter(api *API) chi.Router {
	c := &AssetStatsController{
		API: api,
	}
	r := chi.NewRouter()
	r.Get("/chassis", WithError(c.GetChassisStatPercentage))

	return r
}

type GetMechStatPercentageResponse struct {
	Total      int64 `json:"total"`
	Percentile uint8 `json:"percentile"`
	Percentage uint8 `json:"percentage"`
}

func (sc *AssetStatsController) GetChassisStatPercentage(w http.ResponseWriter, r *http.Request) (int, error) {
	// TODO: fix this
	//stat := r.URL.Query().Get("stat")   // the stat identifier e.g. speed
	//value := r.URL.Query().Get("value") // the value of the stat e.g. 2000
	//model := r.URL.Query().Get("model") // if provided compare to given model

	//valueInt, err := strconv.Atoi(value)
	//if err != nil {
	//	gamelog.L.Error().
	//		Str("stat", stat).
	//		Str("value", value).
	//		Str("model", model).
	//		Str("db func", "ChassisStatRank").Err(err).Msg("unable to convert stat value to int")
	//	return http.StatusBadRequest, terror.Error(err, "Invalid value provided.")
	//}
	//
	//// validate stat column
	//switch stat {
	//case boiler.ChassisColumns.ShieldRechargeRate:
	//case boiler.ChassisColumns.HealthRemaining:
	//case boiler.ChassisColumns.WeaponHardpoints:
	//case boiler.ChassisColumns.TurretHardpoints:
	//case boiler.ChassisColumns.UtilitySlots:
	//case boiler.ChassisColumns.Speed:
	//case boiler.ChassisColumns.MaxHitpoints:
	//case boiler.ChassisColumns.MaxShield:
	//	break
	//default:
	//	gamelog.L.Error().Str("stat", stat).Msg("invalid mech stat identifier")
	//	return http.StatusBadRequest, terror.Error(fmt.Errorf("invalid mech stat identifier"), "Invalid mech stat identifier.")
	//}
	//
	//modelCondition := ""
	//
	//if model != "" {
	//	// validate model
	//	exists, err := boiler.Chasses(boiler.ChassisWhere.Model.EQ(model)).Exists(gamedb.StdConn)
	//	if err != nil {
	//		gamelog.L.Error().Err(err).Str("model", model).Msg("invalid model provided")
	//		return http.StatusBadRequest, terror.Error(err, "Invalid model provided.")
	//	}
	//	if !exists {
	//		gamelog.L.Error().Err(fmt.Errorf("model doesn't exist")).Str("model", model).Msg("model doesn't exist")
	//		return http.StatusBadRequest, terror.Error(fmt.Errorf("model doesn't exist"), "Invalid model provided.")
	//	}
	//
	//	modelCondition = fmt.Sprintf(`WHERE model ilike '%s'`, model)
	//}
	//
	//var total int
	//var max int
	//var min int
	//
	//q := fmt.Sprintf(`
	//     	SELECT
	//     		count(id),
	//     		max("%[1]s"),
	//     		min("%[1]s")
	//     	FROM chassis
	//		%s
	//     `, stat, modelCondition)
	//
	//err = gamedb.Conn.QueryRow(context.Background(), q).Scan(&total, &max, &min)
	//if err != nil {
	//	gamelog.L.Error().
	//		Str("stat", stat).
	//		Str("value", value).
	//		Str("model", model).
	//		Str("db func", "ChassisStatRank").Err(err).Msg("unable to get max value of chassis stat")
	//	return http.StatusInternalServerError, err
	//}
	//
	//if max-min <= 0 {
	//	return helpers.EncodeJSON(w, GetMechStatPercentageResponse{
	//		int64(total),
	//		0,
	//		100,
	//	})
	//}
	//
	//percentage := uint8(float64(valueInt-min) * 100 / float64(max-min))
	//return helpers.EncodeJSON(w, GetMechStatPercentageResponse{
	//	int64(total),
	//	100 - percentage,
	//	percentage,
	//})
	return 200, nil
}
