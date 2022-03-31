package api

import (
	"context"
	"encoding/json"
	"server/db/boiler"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/rs/zerolog"
)

type BattleControllerWS struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

func NewBattleController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *BattleControllerWS {
	bc := &BattleControllerWS{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "twitch_hub"),
		API:  api,
	}

	api.Command(HubKeyBattleMechHistoryList, bc.BattleMechHistoryListHandler)

	return bc
}

const HubKeyBattleMechHistoryList = hub.HubCommandKey("BATTLE:MECH:HISTORY:LIST")

type BattleMechHistoryRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		MechID string `json:"mech_id"`
	} `json:"payload"`
}

type BattleMechHistoryResponse struct {
	Total           int                      `json:"total"`
	BattleContracts []*boiler.BattleContract `json:"battle_contracts"`
}

func (bc *BattleControllerWS) BattleMechHistoryListHandler(ctx context.Context, hub *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &BattleMechHistoryRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	battleContracts := make([]*boiler.BattleContract, 0)
	q := `
		select *
		from battle_contracts
		where mech_id = $1
		order by queued_at desc
		limit 10
	`
	err = pgxscan.Get(ctx, bc.Conn, battleContracts, q, req.Payload.MechID)
	if err != nil {
		return err
	}

	reply(BattleMechHistoryResponse{
		len(battleContracts),
		battleContracts,
	})

	return nil
}
