package battle

//
//import (
//	"context"
//	"database/sql"
//	"encoding/json"
//	"errors"
//	"fmt"
//	"server"
//	"server/db"
//	"server/db/boiler"
//	"server/gamedb"
//	"server/gamelog"
//	"server/xsyn_rpcclient"
//	"time"
//
//	"github.com/gofrs/uuid"
//	"github.com/ninja-software/terror/v2"
//	"github.com/ninja-syndicate/ws"
//	"github.com/shopspring/decimal"
//	"github.com/volatiletech/null/v8"
//	"github.com/volatiletech/sqlboiler/v4/boil"
//)
//
//const (
//	RepairModeStandard = "STANDARD"
//	RepairModeFast     = "FAST"
//)
//
//func (btl *Battle) processWarMachineRepair() {
//	for _, wm := range btl.WarMachines {
//		id := wm.ID
//		mh := wm.MaxHealth
//		rh := wm.Health
//
//		// register mech repair case
//		go func() {
//			err := btl.arena.RepairSystem.RegisterMechRepairCase(id, mh, rh)
//			if err != nil {
//				gamelog.L.Error().Err(err).Str("mech id", id).Msg("Failed to register repair case")
//			}
//		}()
//	}
//}
//
//// InsurancePrice handle price calculation
//func (arena *Arena) InsurancePrice(mechID string) decimal.Decimal {
//	// get insurance price from mech
//
//	// else get current global insurance price
//
//	return decimal.New(10, 18)
//}
//
//type AssetRepairPayFeeRequest struct {
//	Payload struct {
//		MechID string `json:"mech_id"`
//	} `json:"payload"`
//}
//
//const HubKeyAssetRepairPayFee = "ASSET:REPAIR:PAY:FEE"
//
//func (arena *Arena) AssetRepairPayFeeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
//	req := &AssetRepairPayFeeRequest{}
//	err := json.Unmarshal(payload, req)
//	if err != nil {
//		return terror.Error(err, "Invalid request received")
//	}
//
//	playerID := uuid.FromStringOrNil(user.ID)
//	if playerID.IsNil() {
//		return terror.Error(terror.ErrForbidden, "You are not login")
//	}
//
//	// get mech
//	mech, err := boiler.FindMech(gamedb.StdConn, req.Payload.MechID)
//	if err != nil {
//		return terror.Error(err, "Failed to get mech from db")
//	}
//
//	ci, err := db.CollectionItemFromItemID(nil, mech.ID)
//	if err != nil {
//		return terror.Error(err, "Failed to get mech from db")
//	}
//
//	if ci.OwnerID != user.ID {
//		return terror.Error(terror.ErrForbidden, "You are not the owner of the mech")
//	}
//
//	now := time.Now()
//
//	// check repair center
//	ar, err := boiler.MechRepairs(
//		boiler.MechRepairWhere.MechID.EQ(mech.ID),
//		boiler.MechRepairWhere.RepairCompleteAt.GT(now),
//	).One(gamedb.StdConn)
//	if err != nil {
//		if !errors.Is(err, sql.ErrNoRows) {
//			return terror.Error(err, "Your mech is not in the repair center")
//		}
//
//		return terror.Error(err, "Failed to get asset repair record from db")
//	}
//
//	// calculate pay fee
//	fullDurationSecond := decimal.NewFromFloat(ar.RepairCompleteAt.Sub(ar.CreatedAt).Seconds())
//	alreadyPassedSeconds := decimal.NewFromFloat(now.Sub(ar.CreatedAt).Seconds())
//	remainSeconds := fullDurationSecond.Sub(alreadyPassedSeconds)
//	ratio := remainSeconds.Div(fullDurationSecond)
//
//	fee := ar.FullRepairFee.Mul(ratio)
//
//	tx, err := gamedb.StdConn.Begin()
//	if err != nil {
//		return terror.Error(err, "Failed to start db transaction")
//	}
//
//	defer tx.Rollback()
//
//	ar.RepairCompleteAt = now
//	_, err = ar.Update(tx, boil.Whitelist(boiler.MechRepairColumns.RepairCompleteAt))
//	if err != nil {
//		return terror.Error(err, "Failed to update asset repair")
//	}
//
//	// get syndicate account
//	factionAccountID, ok := server.FactionUsers[factionID]
//	if !ok {
//		gamelog.L.Error().Str("log_name", "battle arena").
//			Str("player id", playerID.String()).
//			Str("faction ID", factionID).
//			Err(fmt.Errorf("failed to get hard coded syndicate player id")).
//			Msg("unable to get hard coded syndicate player ID from faction ID")
//		return terror.Error(err, "Failed to load syndicate id")
//	}
//
//	// pay sups
//	txID, err := arena.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
//		FromUserID:           playerID,
//		ToUserID:             uuid.FromStringOrNil(factionAccountID),
//		Amount:               fee.StringFixed(0),
//		TransactionReference: server.TransactionReference(fmt.Sprintf("pay_asset_repair_fee|%s|%d", ar.ID, time.Now().UnixNano())),
//		Group:                string(server.TransactionGroupBattle),
//		Description:          "Paying asset repair fee " + ar.ID + ".",
//		NotSafe:              true,
//	})
//	if err != nil {
//		gamelog.L.Error().Str("log_name", "battle arena").Str("asset repair id", ar.ID).Err(err).Msg("Failed to pay asset repair fee")
//		return terror.Error(err, "Failed to pay asset repair fee")
//	}
//
//	err = tx.Commit()
//	if err != nil {
//		return terror.Error(err, "Failed to commit db transaction")
//	}
//
//	ar.PayToRepairTXID = null.StringFrom(txID)
//	_, err = ar.Update(gamedb.StdConn, boil.Whitelist(boiler.MechRepairColumns.PayToRepairTXID))
//	if err != nil {
//		return terror.Error(err, "Failed to update asset repair")
//	}
//
//	reply(true)
//
//	return nil
//}
//
//type AssetRepairStatusRequest struct {
//	Payload struct {
//		MechID string `json:"mech_id"`
//	} `json:"payload"`
//}
//
//type AssetRepairStatusResponse struct {
//	TotalRequiredSeconds int             `json:"total_required_seconds"`
//	RemainSeconds        int             `json:"remain_seconds"`
//	FullRepairFee        decimal.Decimal `json:"full_repair_fee"`
//}
//
//const HubKeyAssetRepairStatus = "ASSET:REPAIR:STATUS"
//
//func (arena *Arena) AssetRepairStatusHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
//	req := &AssetRepairStatusRequest{}
//	err := json.Unmarshal(payload, req)
//	if err != nil {
//		return terror.Error(err, "Invalid request received")
//	}
//
//	playerID := uuid.FromStringOrNil(user.ID)
//	if playerID.IsNil() {
//		return terror.Error(terror.ErrForbidden, "You are not login")
//	}
//
//	// get mech
//	mech, err := boiler.FindMech(gamedb.StdConn, req.Payload.MechID)
//	if err != nil {
//		return terror.Error(err, "Failed to get mech from db")
//	}
//
//	ci, err := db.CollectionItemFromItemID(nil, mech.ID)
//	if err != nil {
//		return terror.Error(err, "Failed to get mech from db")
//	}
//
//	if ci.OwnerID != user.ID {
//		return terror.Error(terror.ErrForbidden, "You are not the owner of the mech")
//	}
//
//	now := time.Now()
//
//	// check repair center
//	ar, err := boiler.MechRepairs(
//		boiler.MechRepairWhere.MechID.EQ(mech.ID),
//		boiler.MechRepairWhere.RepairCompleteAt.GT(now),
//	).One(gamedb.StdConn)
//	if err != nil && !errors.Is(err, sql.ErrNoRows) {
//		return terror.Error(err, "Failed to get asset repair record from db")
//	}
//
//	if ar == nil {
//		reply(nil)
//		return nil
//	}
//
//	// reply asset repair status
//	reply(&AssetRepairStatusResponse{
//		TotalRequiredSeconds: int(ar.RepairCompleteAt.Sub(ar.CreatedAt).Seconds()),
//		RemainSeconds:        int(ar.RepairCompleteAt.Sub(now).Seconds()),
//		FullRepairFee:        ar.FullRepairFee,
//	})
//
//	return nil
//}
