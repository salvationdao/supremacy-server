package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
)

func BattleLobbyController(api *API) {
	api.SecureUserFactionCommand(HubKeyBattleLobbyCreate, api.BattleLobbyCreate)
	api.SecureUserFactionCommand(HubKeyBattleLobbyJoin, api.BattleLobbyJoin)

}

type BattleLobbyCreateRequest struct {
	Payload struct {
		MechIDs          []string        `json:"mechIDs"`
		EntryFee         decimal.Decimal `json:"entry_fee"`
		FirstFactionCut  decimal.Decimal `json:"first_faction_cut"`
		SecondFactionCut decimal.Decimal `json:"second_faction_cut"`
		ThirdFactionCut  decimal.Decimal `json:"third_faction_cut"`
	} `json:"payload"`
}

const HubKeyBattleLobbyCreate = "BATTLE:LOBBY:CREATE"

func (api *API) BattleLobbyCreate(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &BattleLobbyCreateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// initial mech amount check
	if len(req.Payload.MechIDs) == 0 {
		return terror.Error(fmt.Errorf("mech id list not provided"), "Initial mech is not provided.")
	}

	// NOTE: three mech per faction by default
	if len(req.Payload.MechIDs) > db.FACTION_MECH_LIMIT {
		return terror.Error(fmt.Errorf("mech more than 3"), "Maximum 3 mech per faction.")
	}

	// ownership check
	mcis, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.IN(req.Payload.MechIDs),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Strs("mech ids", req.Payload.MechIDs).Err(err).Msg("unable to retrieve mech collection item from hash")
		return err
	}

	if len(mcis) != len(req.Payload.MechIDs) {
		return terror.Error(fmt.Errorf("contain non-mech assest"), "The list contains non-mech asset.")
	}

	for _, mci := range mcis {
		if mci.XsynLocked {
			err := fmt.Errorf("mech is locked to xsyn locked")
			gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mci.ItemID).Err(err).Msg("war machine is xsyn locked")
			return err
		}

		if mci.LockedToMarketplace {
			err := fmt.Errorf("mech is listed in marketplace")
			gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mci.ItemID).Err(err).Msg("war machine is listed in marketplace")
			return err
		}

		battleReady, err := db.MechBattleReady(mci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to load battle ready status")
			return err
		}

		if !battleReady {
			gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mci.ItemID).Err(err).Msg("war machine is not available for queuing")
			return fmt.Errorf("mech is cannot be used")
		}

		if mci.OwnerID != user.ID {
			return terror.Error(fmt.Errorf("does not own the mech"), "This mech is not owned by you")
		}
	}

	// entry fee check
	if req.Payload.EntryFee.IsNegative() {
		return terror.Error(fmt.Errorf("negative entry fee"), "Entry fee cannot be negative.")
	}

	// reward cut check
	if !req.Payload.FirstFactionCut.IsPositive() || !req.Payload.SecondFactionCut.IsPositive() || !req.Payload.ThirdFactionCut.IsPositive() {
		return terror.Error(fmt.Errorf("invalid reward cut"), "Reward cut must be positive.")
	}

	if !req.Payload.FirstFactionCut.Add(req.Payload.SecondFactionCut).Add(req.Payload.ThirdFactionCut).Equal(decimal.NewFromInt(100)) {
		return terror.Error(fmt.Errorf("total must be 100"), "The total of the reward cut must equal 100.")
	}

	// start process
	err = api.ArenaManager.SendBattleLobbyFunc(func() error {
		tx, err := gamedb.StdConn.Begin()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to start db transaction.")
			return terror.Error(err, "Failed to create battle lobby.")
		}

		defer tx.Rollback()

		bl := boiler.BattleLobby{
			HostByID:              user.ID,
			EntryFee:              req.Payload.EntryFee,
			FirstFactionCut:       req.Payload.FirstFactionCut,
			SecondFactionCut:      req.Payload.SecondFactionCut,
			ThirdFactionCut:       req.Payload.ThirdFactionCut,
			EachFactionMechAmount: db.FACTION_MECH_LIMIT,
		}

		err = bl.Insert(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Interface("battle lobby", bl).Msg("Failed to insert battle lobby")
			return terror.Error(err, "Failed to create battle lobby")
		}

		// insert battle mechs
		for _, mechID := range req.Payload.MechIDs {
			blm := boiler.BattleLobbiesMech{
				BattleLobbyID: bl.ID,
				MechID:        mechID,
				OwnerID:       user.ID,
				FactionID:     factionID,
			}

			err = blm.Insert(tx, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Interface("battle lobby mech", blm).Msg("Failed to insert battle lobbies mech")
				return terror.Error(err, "Failed to insert mechs into battle lobby.")
			}
		}

		err = tx.Commit()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
			return terror.Error(err, "Failed to create battle lobby.")
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

const HubKeyBattleLobbyJoin = "BATTLE:LOBBY:JOIN"

func (api *API) BattleLobbyJoin(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {

	return nil
}

// subscriptions

const HubKeyBattleLobbyListUpdate = "BATTLE:LOBBY:LIST:UPDATE"

func (api *API) LobbyListUpdate(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {

	return nil
}
