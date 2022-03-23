package api

import (
	"context"
	"fmt"
	"net/http"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
)

type AssetStatsController struct {
	Conn db.Conn
	Log  *zerolog.Logger
	API  *API
}

func AssetStatsRouter(log *zerolog.Logger, conn db.Conn, api *API) chi.Router {
	c := &AssetStatsController{
		Conn: conn,
		Log:  log,
		API:  api,
	}
	r := chi.NewRouter()
	r.Get("/mech", WithError(c.GetMechStatPercentage))

	return r
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

	var rank int64
	err = gamedb.Conn.QueryRow(context.Background(), fmt.Sprintf(`
	SELECT
		count(id)
	FROM
		chassis
	WHERE
		"%s" >= $1
`, stat), valueInt).Scan(&rank)
	if err != nil {
		gamelog.L.Error().
			Str("stat", stat).
			Str("value", value).
			Str("global", global).
			Str("db func", "ChassisStatRank").Err(err).Msg("unable to get rank of chassis stat")
		return http.StatusInternalServerError, err
	}

	var total int64
	var max int64
	err = gamedb.Conn.QueryRow(context.Background(), fmt.Sprintf(`
	SELECT
		count(id),
		max("%s")
	FROM
		chassis
`, stat)).Scan(&total, &max)
	if err != nil {
		gamelog.L.Error().
			Str("stat", stat).
			Str("value", value).
			Str("global", global).
			Str("db func", "ChassisStatRank").Err(err).Msg("unable to get max value of chassis stat")
		return http.StatusInternalServerError, err
	}

	return helpers.EncodeJSON(w, struct {
		Total      int64 `json:"total"`
		Percentile uint8 `json:"percentile"`
		Percentage uint8 `json:"percentage"`
	}{
		total,
		uint8((rank * 100) / total),
		uint8((int64(valueInt) * 100) / max),
	})
}
