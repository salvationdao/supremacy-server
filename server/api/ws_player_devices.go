package api

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-syndicate/ws"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/ninja-software/terror/v2"
)

type PlayerDevicesControllerWS struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

func NewPlayerDevicesController(api *API) *PlayerDevicesControllerWS {
	pdc := &PlayerDevicesControllerWS{
		API: api,
	}

	api.SecureUserCommand(HubKeyPlayerDeviceList, pdc.PlayerDevicesListHandler)

	return pdc
}

const HubKeyPlayerDeviceList = "PLAYER:DEVICE:LIST"

type PlayerDevice struct {
	Name string `json:"name"`
}

type PlayerDeviceListResp struct {
	Total   int64           `json:"total"`
	Devices []*PlayerDevice `json:"devices"`
}

func (pdc *PlayerDevicesControllerWS) PlayerDevicesListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {

	// Get players devices
	playerDevices, err := boiler.Devices(
		qm.Select(boiler.DeviceColumns.Name),
		boiler.DeviceWhere.PlayerID.EQ(user.ID),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("db func", "PlayerDevicesList").Str("userID", user.ID).Err(err).Msg("unable to get player devices")
		return terror.Error(err, "Unable to retrieve abilities, try again or contact support.")
	}

	reply(playerDevices)
	return nil
}
