package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server/db/boiler"
	"server/gamedb"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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

type BattleDetailed struct {
	*boiler.Battle
	GameMap *boiler.GameMap `json:"game_map"`
}

type BattleMechDetailed struct {
	*boiler.BattleMech
	Battle *BattleDetailed `json:"battle"`
}

type BattleMechHistoryResponse struct {
	Total         int                  `json:"total"`
	BattleHistory []BattleMechDetailed `json:"battle_history"`
}

func (bc *BattleControllerWS) BattleMechHistoryListHandler(ctx context.Context, hub *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &BattleMechHistoryRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	fmt.Println("-----------mech id-----------")
	fmt.Println(req.Payload.MechID)
	// battleMechs := make([]*boiler.BattleMech, 0)
	// q := `
	// 	select *
	// 	from battle_mechs
	// 	where mech_id = $1
	// 	order by created_at desc
	// 	limit 10
	// `
	// err = pgxscan.Select(ctx, bc.Conn, &battleMechs, q, req.Payload.MechID)
	// if err != nil && !errors.Is(err, pgx.ErrNoRows) {
	// 	return err
	// }
	battleMechs, err := boiler.BattleMechs(boiler.BattleMechWhere.MechID.EQ(req.Payload.MechID), qm.Limit(10), qm.Load(qm.Rels(boiler.BattleMechRels.Battle, boiler.BattleRels.GameMap))).All(gamedb.StdConn)

	output := []BattleMechDetailed{}
	for _, o := range battleMechs {
		output = append(output, BattleMechDetailed{
			BattleMech: o,
			Battle: &BattleDetailed{
				Battle:  o.R.Battle,
				GameMap: o.R.Battle.R.GameMap,
			},
		})
	}

	reply(BattleMechHistoryResponse{
		len(output),
		output,
	})

	return nil
}
