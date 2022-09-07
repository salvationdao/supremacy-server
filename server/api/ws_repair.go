package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"server"
	"server/battle"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/kevinms/leakybucket-go"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
)

func NewMechRepairController(api *API) {
	api.SecureUserCommand(server.HubKeyRepairOfferIssue, api.RepairOfferIssue)
	api.SecureUserCommand(server.HubKeyRepairOfferClose, api.RepairOfferClose)
	api.SecureUserCommand(server.HubKeyRepairAgentRegister, api.RepairAgentRegister)
	api.SecureUserCommand(server.HubKeyRepairAgentRecord, api.RepairAgentRecord)
	api.SecureUserCommand(server.HubKeyRepairAgentComplete, api.RepairAgentComplete)
	api.SecureUserCommand(server.HubKeyRepairAgentAbandon, api.RepairAgentAbandon)

	api.SecureUserCommand(server.HubKeyMechRepairSlotInsert, api.MechRepairSlotInsert)
	api.SecureUserCommand(server.HubKeyMechRepairSlotRemove, api.MechRepairSlotRemove)
	api.SecureUserCommand(server.HubKeyMechRepairSlotSwap, api.MechRepairSlotSwap)

}

func (api *API) RepairOfferList(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	ros, err := boiler.RepairOffers(
		boiler.RepairOfferWhere.ExpiresAt.GT(time.Now()),
		boiler.RepairOfferWhere.ClosedAt.IsNull(),
		boiler.RepairOfferWhere.OfferedByID.IsNotNull(),
		qm.Load(boiler.RepairOfferRels.RepairCase, boiler.RepairCaseWhere.CompletedAt.IsNull()),
		qm.Load(boiler.RepairOfferRels.RepairAgents, boiler.RepairAgentWhere.FinishedAt.IsNull()),
		qm.Load(boiler.RepairOfferRels.OfferedBy),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to load repair offer detail")
	}

	resp := []*server.RepairOffer{}
	for _, ro := range ros {
		if ro.R == nil || ro.R.RepairCase == nil {
			continue
		}

		rc := ro.R.RepairCase

		sro := &server.RepairOffer{
			RepairOffer:          ro,
			BlocksRequiredRepair: rc.BlocksRequiredRepair,
			BlocksRepaired:       rc.BlocksRepaired,
			SupsWorthPerBlock:    ro.OfferedSupsAmount.Div(decimal.NewFromInt(int64(ro.BlocksTotal))),
			WorkingAgentCount:    0,
			JobOwner:             server.PublicPlayerFromBoiler(ro.R.OfferedBy),
		}

		if ro.R.RepairAgents != nil {
			sro.WorkingAgentCount = len(ro.R.RepairAgents)
		}

		resp = append(resp, sro)
	}

	reply(resp)

	return nil
}

type RepairOfferIssueRequest struct {
	Payload struct {
		MechIDs             []string        `json:"mech_ids"`
		LastForMinutes      int             `json:"last_for_minutes"`
		OfferedSupsPerBlock decimal.Decimal `json:"offered_sups_per_block"` // the amount that excluded tax
	} `json:"payload"`
}

func (api *API) RepairOfferIssue(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	now := time.Now()

	req := &RepairOfferIssueRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if len(req.Payload.MechIDs) == 0 {
		return terror.Error(fmt.Errorf("missing mech id"), "Mech id is not provided.")
	}

	// validate ownership
	cis, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
		boiler.CollectionItemWhere.ItemID.IN(req.Payload.MechIDs),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("item type", boiler.ItemTypeMech).Strs("mech id list", req.Payload.MechIDs).Msg("Failed to query war machine collection item")
		return terror.Error(err, "Failed to load war machine detail.")
	}

	if len(req.Payload.MechIDs) != len(cis) {
		return terror.Error(fmt.Errorf("contain non-mech asset"), "Request contain non-mech asset.")
	}

	for _, ci := range cis {
		if ci.OwnerID != user.ID {
			return terror.Error(fmt.Errorf("do not own the mech"), "The mech is not owned by you.")
		}
	}

	// send repair offer func in channel
	err = api.ArenaManager.SendRepairFunc(func() error {
		// look for repair cases
		mrcs, err := boiler.RepairCases(
			boiler.RepairCaseWhere.MechID.IN(req.Payload.MechIDs),
			boiler.RepairCaseWhere.CompletedAt.IsNull(),
			qm.Load(
				boiler.RepairCaseRels.RepairOffers,
				boiler.RepairOfferWhere.OfferedByID.IsNotNull(),
				boiler.RepairOfferWhere.ClosedAt.IsNull(),
			),
		).All(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Err(err).Strs("mech ids", req.Payload.MechIDs).Msg("Failed to query mech repair case.")
			return terror.Error(err, "Failed to load mech repair case.")
		}

		if len(mrcs) != len(cis) {
			return terror.Error(fmt.Errorf("not all the mech need to be repaired"), "List contains mech which doesn't need to be repaired.")
		}

		for _, mrc := range mrcs {
			if mrc == nil {
				return terror.Error(fmt.Errorf("mech does not have repair case"), "The mech does not need to be repaired.")
			}

			if mrc.BlocksRequiredRepair == mrc.BlocksRepaired {
				return terror.Error(fmt.Errorf("mech already repaired"), "The mech has already repaired.")
			}

			if mrc.R != nil && mrc.R.RepairOffers != nil && len(mrc.R.RepairOffers) > 0 {
				return terror.Error(fmt.Errorf("unclosed offer exists"), "Cannot offer a new repair contract if the previous offer has not ended yet.")
			}
		}

		// register a new repair offer
		sros := []*server.RepairOffer{}
		for _, mrc := range mrcs {
			err = func() error {
				tx, err := gamedb.StdConn.Begin()
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to begin db transaction.")
					return terror.Error(err, "Failed to offer repair job.")
				}

				defer tx.Rollback()

				blocksTotal := mrc.BlocksRequiredRepair - mrc.BlocksRepaired

				offeredSups := req.Payload.OfferedSupsPerBlock.Mul(decimal.New(int64(blocksTotal), 18)).Round(0)

				ro := &boiler.RepairOffer{
					OfferedByID:       null.StringFrom(user.ID),
					RepairCaseID:      mrc.ID,
					BlocksTotal:       blocksTotal,
					OfferedSupsAmount: offeredSups,
					ExpiresAt:         now.Add(time.Duration(req.Payload.LastForMinutes) * time.Minute),
				}
				err = ro.Insert(tx, boil.Infer())
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to insert repair offer.")
					return terror.Error(err, "Failed to offer repair job.")
				}

				// offering price plus 10%
				tax := offeredSups.Mul(decimal.NewFromFloat(0.1)).Round(0)

				// pay sups to offer repair job
				offerTXID, err := api.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
					FromUserID:           uuid.FromStringOrNil(user.ID),
					ToUserID:             uuid.FromStringOrNil(server.RepairCenterUserID),
					Amount:               offeredSups.Add(tax).String(),
					TransactionReference: server.TransactionReference(fmt.Sprintf("create_repair_offer|%s|%d", ro.ID, time.Now().UnixNano())),
					Group:                string(server.TransactionGroupSupremacy),
					SubGroup:             string(server.TransactionGroupRepair),
					Description:          "create repair offer including 10% GST",
				})
				if err != nil {
					gamelog.L.Error().Str("player_id", user.ID).Str("repair offer id", ro.ID).Str("amount", offeredSups.Add(tax).String()).Err(err).Msg("Failed to pay sups for offering repair job")
					return terror.Error(err, "Failed to pay sups for offering repair job.")
				}

				refundOfferSupsFunc := func() {
					_, err = api.Passport.RefundSupsMessage(offerTXID)
					if err != nil {
						gamelog.L.Error().Str("tx id", offerTXID).Err(err).Msg("Failed to refund sups for offering repair job")
					}
				}

				ro.PaidTXID = null.StringFrom(offerTXID)

				// pay tax to XSYN treasury
				offerTaxTXID, err := api.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
					FromUserID:           uuid.FromStringOrNil(server.RepairCenterUserID),
					ToUserID:             uuid.FromStringOrNil(server.SupremacyChallengeFundUserID), // NOTE: send fees to challenge fund for now. (was to treasury)
					Amount:               tax.String(),
					TransactionReference: server.TransactionReference(fmt.Sprintf("repair_offer_tax|%s|%d", ro.ID, time.Now().UnixNano())),
					Group:                string(server.TransactionGroupSupremacy),
					SubGroup:             string(server.TransactionGroupRepair),
					Description:          "repair offer tax",
				})
				if err != nil {
					refundOfferSupsFunc()
					gamelog.L.Error().Str("player_id", user.ID).Str("repair offer id", ro.ID).Str("amount", tax.String()).Err(err).Msg("Failed to pay tax for offering repair job")
					return terror.Error(err, "Failed to pay sups for offering repair job.")
				}

				// trigger challenge fund update
				defer func() {
					api.ArenaManager.ChallengeFundUpdateChan <- true
				}()

				refundTaxFunc := func() {
					_, err = api.Passport.RefundSupsMessage(offerTaxTXID)
					if err != nil {
						gamelog.L.Error().Str("tx id", offerTaxTXID).Err(err).Msg("Failed to refund tax")
					}
				}

				ro.TaxTXID = null.StringFrom(offerTaxTXID)

				_, err = ro.Update(tx, boil.Whitelist(
					boiler.RepairOfferColumns.PaidTXID,
					boiler.RepairOfferColumns.TaxTXID,
				))
				if err != nil {
					refundTaxFunc()
					refundOfferSupsFunc()
					gamelog.L.Error().Err(err).Interface("repair offer", ro).Msg("Failed to update repair offer transaction id.")
					return terror.Error(err, "Failed to update sups transaction id")
				}

				err = tx.Commit()
				if err != nil {
					refundTaxFunc()
					refundOfferSupsFunc()
					gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
					return terror.Error(err, "Failed to offer repair contract.")
				}

				sro := &server.RepairOffer{
					RepairOffer:          ro,
					BlocksRequiredRepair: mrc.BlocksRequiredRepair,
					BlocksRepaired:       mrc.BlocksRepaired,
					SupsWorthPerBlock:    req.Payload.OfferedSupsPerBlock.Mul(decimal.New(1, 18)),
					WorkingAgentCount:    0,
					JobOwner:             server.PublicPlayerFromBoiler(user),
				}

				ws.PublishMessage(fmt.Sprintf("/secure/repair_offer/%s", ro.ID), server.HubKeyRepairOfferSubscribe, sro)
				ws.PublishMessage(fmt.Sprintf("/secure/mech/%s/active_repair_offer", mrc.MechID), server.HubKeyMechActiveRepairOffer, sro)

				sros = append(sros, sro)

				return nil
			}()

			if err != nil {
				if len(sros) == 0 {
					return terror.Error(err, "Failed to offer repair job")
				}

				// if repair jobs are partially offered
				if len(sros) > 0 {
					//  broadcast to repair offer list update to market
					ws.PublishMessage("/secure/repair_offer/update", server.HubKeyRepairOfferUpdateSubscribe, sros)
					return terror.Error(err, "Failed to offer all the repair jobs.")
				}
			}
		}

		ws.PublishMessage("/secure/repair_offer/update", server.HubKeyRepairOfferUpdateSubscribe, sros)
		return nil
	})
	if err != nil {
		return err
	}

	reply(true)

	return nil
}

type RepairOfferCancelRequest struct {
	Payload struct {
		RepairOfferID string `json:"repair_offer_id"`
	} `json:"payload"`
}

func (api *API) RepairOfferClose(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &RepairOfferCancelRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	err = api.ArenaManager.SendRepairFunc(func() error {
		ro, err := boiler.FindRepairOffer(gamedb.StdConn, req.Payload.RepairOfferID)
		if err != nil {
			return terror.Error(err, "Failed to get repair offer id.")
		}

		if ro.OfferedByID.String != user.ID {
			return terror.Error(fmt.Errorf("cannot cancel others offer"), "Can only cancel the offer which is issued by yourself.")
		}

		if ro.ExpiresAt.Before(time.Now()) {
			return terror.Error(fmt.Errorf("offer is expired"), "The offer is already expired.")
		}

		if ro.ClosedAt.Valid {
			return terror.Error(fmt.Errorf("offer is closed"), "The offer is already closed.")
		}

		err = api.ArenaManager.CloseRepairOffers([]string{ro.ID}, boiler.RepairFinishReasonSTOPPED, boiler.RepairAgentFinishReasonEXPIRED)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("repair offer", ro).Msg("Failed to close repair offer.")
			return terror.Error(err, "Failed to close repair offer.")
		}

		return nil
	})
	if err != nil {
		return err
	}

	reply(true)

	return nil
}

type RepairAgentRegisterRequest struct {
	Payload struct {
		// this is for player who want to repair their mech themselves
		RepairCaseID string `json:"repair_case_id"`

		// this is for player who grab offer from the job list
		RepairOfferID string `json:"repair_offer_id"`

		CaptchaToken string `json:"captcha_token"`
	} `json:"payload"`
}

var mechRepairAgentBucket = leakybucket.NewCollector(2, 1, true)

func (api *API) RepairAgentRegister(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	if mechRepairAgentBucket.Add(user.ID, 1) == 0 {
		return terror.Error(fmt.Errorf("too many request"), "Too many request.")
	}

	req := &RepairAgentRegisterRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	queries := []qm.QueryMod{
		boiler.RepairOfferWhere.ClosedAt.IsNull(),
	}

	if req.Payload.RepairCaseID != "" {
		// check person is the mech owner
		isOwner, err := db.IsRepairCaseOwner(req.Payload.RepairCaseID, user.ID)
		if err != nil {
			return err
		}

		if !isOwner {
			return terror.Error(fmt.Errorf("only owner can repair their mech themselves"), "Only mech owner can repair their mech themselves.")
		}

		queries = append(queries,
			boiler.RepairOfferWhere.RepairCaseID.EQ(req.Payload.RepairCaseID),
			boiler.RepairOfferWhere.OfferedByID.IsNull(), // system generated offer
		)
	} else {
		queries = append(queries, boiler.RepairOfferWhere.ID.EQ(req.Payload.RepairOfferID))
	}

	// get repair offer
	ro, err := boiler.RepairOffers(queries...).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("repair offer id", req.Payload.RepairOfferID).Msg("Failed to get repair offer from id")
		return terror.Error(err, "Failed to get repair offer")
	}

	if ro == nil {
		return terror.Error(err, "Repair offer does not exist.")
	}

	// get the last registered repair agent of the player
	lastRegister, err := boiler.RepairAgents(
		boiler.RepairAgentWhere.PlayerID.EQ(user.ID),
		qm.OrderBy(boiler.RepairAgentColumns.CreatedAt+" DESC"),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to load repair agent.")
	}

	// verify token, if players have not done any repair, or they are doing different offer
	if lastRegister == nil || lastRegister.RepairOfferID != ro.ID {
		err = api.captcha.verify(req.Payload.CaptchaToken)
		if err != nil {
			return terror.Error(err, "Failed to complete captcha verification.")
		}
	}

	// abandon last repair agent
	if lastRegister != nil && !lastRegister.FinishedAt.Valid {
		lastRegister.FinishedAt = null.TimeFrom(time.Now())
		lastRegister.FinishedReason = null.StringFrom(boiler.RepairAgentFinishReasonABANDONED)
		_, err = lastRegister.Update(gamedb.StdConn, boil.Whitelist(
			boiler.RepairAgentColumns.FinishedAt,
			boiler.RepairAgentColumns.FinishedReason,
		))
		if err != nil {
			gamelog.L.Error().Err(err).Str("player id", user.ID).Msg("Failed to close repair agents.")
			return terror.Error(err, "Failed to abandon repair job")
		}

		// broadcast changes if targeting different repair offer
		if lastRegister.RepairOfferID != ro.ID {
			err = api.broadcastRepairOffer(lastRegister.RepairOfferID)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to broadcast updated repair offer.")
				return terror.Error(err, "Failed to broadcast updated repair offer")
			}
		}
	}

	// insert repair agent
	ra := &boiler.RepairAgent{
		RepairCaseID:   ro.RepairCaseID,
		RepairOfferID:  ro.ID,
		PlayerID:       user.ID,
		RequiredStacks: db.GetIntWithDefault(db.KeyRequiredRepairStacks, 50),
	}

	err = ra.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("repair agent", ra).Msg("Failed to register repair agent")
		return terror.Error(err, "Failed to register repair agent")
	}

	go func() {
		err = api.broadcastRepairOffer(ro.ID)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to broadcast updated repair offer.")
			return
		}
	}()

	reply(ra)

	return nil
}

func (api *API) broadcastRepairOffer(repairOfferID string) error {
	sro, err := db.RepairOfferDetail(repairOfferID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msg("Failed to load updated repair offer")
		return terror.Error(err, "Failed to load repair offer")
	}

	if sro != nil {
		ws.PublishMessage(fmt.Sprintf("/secure/repair_offer/%s", repairOfferID), server.HubKeyRepairOfferSubscribe, sro)
		ws.PublishMessage("/secure/repair_offer/update", server.HubKeyRepairOfferUpdateSubscribe, []*server.RepairOffer{sro})
	}

	return nil
}

type RepairAgentRecordRequest struct {
	Payload struct {
		RepairAgentID string                 `json:"repair_agent_id"`
		TriggerWith   string                 `json:"trigger_with"`
		Score         int                    `json:"score"`
		IsFailed      bool                   `json:"is_failed"`
		Dimension     MiniGameStackDimension `json:"dimension"`
	} `json:"payload"`
}

type MiniGameStackDimension struct {
	Width decimal.Decimal `json:"width"`
	Depth decimal.Decimal `json:"depth"`
}

func (api *API) RepairAgentRecord(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &RepairAgentRecordRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// skip, if it is an initial block
	if req.Payload.Score == 0 {
		reply(false)
		return nil
	}

	switch req.Payload.TriggerWith {
	case boiler.RepairTriggerWithTypeSPACE_BAR, boiler.RepairTriggerWithTypeLEFT_CLICK, boiler.RepairTriggerWithTypeTOUCH:
	default:
		gamelog.L.Debug().Str("repair agent id", req.Payload.RepairAgentID).Msg("Unknown trigger type is detected.")
		return terror.Error(fmt.Errorf("invalid trigger type"), "Unknown trigger type is detected.")
	}

	// log record
	ral := boiler.RepairAgentLog{
		RepairAgentID: req.Payload.RepairAgentID,
		TriggeredWith: req.Payload.TriggerWith,
		Score:         req.Payload.Score,
		BlockWidth:    req.Payload.Dimension.Width,
		BlockDepth:    req.Payload.Dimension.Depth,
		IsFailed:      req.Payload.IsFailed,
	}

	err = ral.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to insert repair agent request")
	}

	if req.Payload.IsFailed {
		reply(false)
		return nil
	}
	reply(true)

	return nil
}

type RepairAgentCompleteRequest struct {
	Payload struct {
		RepairAgentID string `json:"repair_agent_id"`
	} `json:"payload"`
}

func (api *API) RepairAgentComplete(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	if mechRepairAgentBucket.Add(user.ID, 1) == 0 {
		return nil
	}
	L := gamelog.L.With().Str("func", "RepairAgentComplete").Interface("user", user).Logger()

	time.Sleep(1 * time.Second)

	req := &RepairAgentCompleteRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	L = L.With().Interface("payload", req.Payload).Logger()

	ra, err := boiler.RepairAgents(
		boiler.RepairAgentWhere.ID.EQ(req.Payload.RepairAgentID),
		qm.Load(boiler.RepairAgentRels.RepairOffer),
	).One(gamedb.StdConn)
	if err != nil {
		L.Error().Err(err).Msg("failed to find repair agent")
		return terror.Error(err, "Failed to load repair agent.")
	}

	L = L.With().Interface("repair agent", ra).Logger()

	if ra.PlayerID != user.ID {
		L.Error().Err(err).Msg(" wrong repair agent")
		return terror.Error(fmt.Errorf("agnet id not match"), "Repair agent id mismatch")
	}

	if ra.FinishedAt.Valid {
		L.Error().Err(err).Msg("already finished")
		return terror.Error(fmt.Errorf("agent finalised"), "This repair agent is already finalised.")
	}

	err = BlockStackingGameVerification(ra)
	if err != nil {
		L.Error().Err(err).Msg("failed BlockStackingGameVerification")
		return err
	}

	rb := boiler.RepairBlock{
		RepairAgentID: ra.ID,
		RepairCaseID:  ra.RepairCaseID,
		RepairOfferID: ra.RepairOfferID,
	}

	err = rb.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		L.Warn().Err(err).Msg("unable to write block")
		if err.Error() == "unable to write block" {
			return terror.Error(err, "repair offer is already closed.")
		}

		return terror.Error(err, "Failed to complete repair agent task.")
	}

	// check repair case after insert
	rc, err := ra.RepairCase().One(gamedb.StdConn)
	if err != nil {
		L.Error().Err(err).Msg("failed to load repair case")
		return terror.Error(err, "Failed to load repair case.")
	}

	L = L.With().Interface("repair case", rc).Logger()

	// if it is not a self repair
	if ra.R.RepairOffer.OfferedByID.Valid {
		// claim sups
		ro, err := db.RepairOfferDetail(ra.RepairOfferID)
		if err != nil {
			L.Error().Err(err).Msg("failed to load repair offer")
			return terror.Error(err, "Failed to load repair offer")
		}

		L = L.With().Interface("repair offer", ro).Logger()

		// if it is not a self offer, pay the agent
		if ro.SupsWorthPerBlock.GreaterThan(decimal.Zero) {
			// claim reward
			payoutTXID, err := api.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
				FromUserID:           uuid.Must(uuid.FromString(server.RepairCenterUserID)),
				ToUserID:             uuid.Must(uuid.FromString(user.ID)),
				Amount:               ro.SupsWorthPerBlock.StringFixed(0),
				TransactionReference: server.TransactionReference(fmt.Sprintf("claim_repair_offer_reward|%s|%d", ro.ID, time.Now().UnixNano())),
				Group:                string(server.TransactionGroupSupremacy),
				SubGroup:             string(server.TransactionGroupRepair),
				Description:          "claim repair offer reward.",
			})
			if err != nil {
				L.Error().Err(err).Msg("failed to pay sups for offering repair job")
				return terror.Error(err, "Failed to pay sups for offering repair job.")
			}

			ra.PayoutTXID = null.StringFrom(payoutTXID)
			_, err = ra.Update(gamedb.StdConn, boil.Whitelist(boiler.RepairAgentColumns.PayoutTXID))
			if err != nil {
				L.Error().Err(err).Msg("failed to update repair agent payout tx id")
			}

		}

		// broadcast result if repair is not completed
		if rc.BlocksRepaired < rc.BlocksRequiredRepair {
			ws.PublishMessage(fmt.Sprintf("/secure/repair_offer/%s", ro.ID), server.HubKeyRepairOfferSubscribe, ro)
			ws.PublishMessage("/secure/repair_offer/update", server.HubKeyRepairOfferUpdateSubscribe, []*server.RepairOffer{ro})
			ws.PublishMessage(fmt.Sprintf("/secure/mech/%s/active_repair_offer", ro.ID), server.HubKeyMechActiveRepairOffer, ro)
		}

		// if repair for others
		if ra.R.RepairOffer.OfferedByID.String != user.ID {
			api.questManager.RepairQuestCheck(user.ID)
		}

	}

	// broadcast result if repair is not completed
	if rc.BlocksRepaired < rc.BlocksRequiredRepair {
		canDeployRatio := db.GetDecimalWithDefault(db.KeyCanDeployDamagedRatio, decimal.NewFromFloat(0.5))

		totalBlocks := db.TotalRepairBlocks(rc.MechID)

		// broadcast current mech stat if damage blocks is less than or equal to deploy ratio
		if decimal.NewFromInt(int64(rc.BlocksRequiredRepair - rc.BlocksRepaired)).Div(decimal.NewFromInt(int64(totalBlocks))).LessThanOrEqual(canDeployRatio) {
			go BroadcastMechQueueStat(rc.MechID)
		}
		ws.PublishMessage(fmt.Sprintf("/secure/mech/%s/repair_case", rc.MechID), server.HubKeyMechRepairCase, rc)
		reply(true)
		return nil
	}

	// clean up repair case if repair is completed
	ws.PublishMessage(fmt.Sprintf("/secure/mech/%s/repair_case", rc.MechID), server.HubKeyMechRepairCase, nil)

	// broadcast current mech stat
	go BroadcastMechQueueStat(rc.MechID)

	// close repair case
	rc.CompletedAt = null.TimeFrom(time.Now())
	_, err = rc.Update(gamedb.StdConn, boil.Whitelist(boiler.RepairCaseColumns.CompletedAt))
	if err != nil {
		L.Error().Err(err).Msg("failed to update repair case")
		return terror.Error(err, "Failed to close repair case.")
	}

	// close offer, self and non-self
	ros, err := rc.RepairOffers(
		boiler.RepairOfferWhere.ClosedAt.IsNull(),
	).All(gamedb.StdConn)
	if err != nil {
		L.Error().Err(err).Msg("failed to load incomplete repair offer")
		return terror.Error(err, "Failed to load incomplete repair offer")
	}

	if len(ros) == 0 {
		reply(true)
		return nil
	}

	roIDs := []string{}
	for _, ro := range ros {
		roIDs = append(roIDs, ro.ID)
	}

	now := time.Now()
	nextRepairDurationSeconds := db.GetIntWithDefault(db.KeyAutoRepairDurationSeconds, 600)

	err = api.ArenaManager.SendRepairFunc(func() error {
		// check current mech is in active repair slot
		pmr, err := boiler.PlayerMechRepairSlots(
			boiler.PlayerMechRepairSlotWhere.MechID.EQ(rc.MechID),
			boiler.PlayerMechRepairSlotWhere.Status.NEQ(boiler.RepairSlotStatusDONE),
		).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Err(err).Interface("repair case", rc).Msg("Failed to check repair slot from repair case.")
			return terror.Error(err, "Failed to check repair slot")
		}

		// clean up repair slot, if exist
		if pmr != nil {
			func() {
				tx, err := gamedb.StdConn.Begin()
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to start db transaction.")
					return
				}

				defer tx.Rollback()

				if pmr.Status == boiler.RepairSlotStatusREPAIRING {
					// set next
					nextSlot, err := boiler.PlayerMechRepairSlots(
						boiler.PlayerMechRepairSlotWhere.PlayerID.EQ(pmr.PlayerID),
						boiler.PlayerMechRepairSlotWhere.Status.EQ(boiler.RepairSlotStatusPENDING),
						qm.OrderBy(boiler.PlayerMechRepairSlotColumns.SlotNumber),
					).One(tx)
					if err != nil && !errors.Is(err, sql.ErrNoRows) {
						gamelog.L.Error().Str("player id", pmr.PlayerID).Err(err).Msg("Failed to load player mech repair bays.")
						return
					}

					// upgrade next "PENDING" slot to "REPAIRING" slot
					if nextSlot != nil {
						nextSlot.Status = boiler.RepairSlotStatusREPAIRING
						nextSlot.NextRepairTime = null.TimeFrom(now.Add(time.Duration(nextRepairDurationSeconds) * time.Second))
						_, err = nextSlot.Update(tx, boil.Whitelist(
							boiler.PlayerMechRepairSlotColumns.Status,
							boiler.PlayerMechRepairSlotColumns.NextRepairTime,
						))
						if err != nil {
							gamelog.L.Error().Interface("repair slot", nextSlot).Err(err).Msg("Failed to update next repair slot.")
							return
						}
					}
				}

				// decrement slot number from current slot
				err = db.DecrementRepairSlotNumber(tx, pmr.PlayerID, pmr.SlotNumber)
				if err != nil {
					gamelog.L.Error().Err(err).Str("player id", pmr.PlayerID).Msg("Failed to decrement slot number.")
					return
				}

				// mark current slot as "DONE"
				pmr.Status = boiler.RepairSlotStatusDONE
				pmr.SlotNumber = 0
				pmr.NextRepairTime = null.TimeFromPtr(nil)
				_, err = pmr.Update(tx,
					boil.Whitelist(
						boiler.PlayerMechRepairSlotColumns.Status,
						boiler.PlayerMechRepairSlotColumns.SlotNumber,
						boiler.PlayerMechRepairSlotColumns.NextRepairTime,
					),
				)
				if err != nil {
					gamelog.L.Error().Err(err).Str("player id", pmr.PlayerID).Msg("Failed to decrement slot number.")
					return
				}

				err = tx.Commit()
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
					return
				}

				// broadcast current repair bay
				go battle.BroadcastRepairBay(pmr.PlayerID)
			}()
		}

		// close repair offer
		err = api.ArenaManager.CloseRepairOffers(roIDs, boiler.RepairAgentFinishReasonSUCCEEDED, boiler.RepairAgentFinishReasonEXPIRED)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("repair offers", ros).Msg("Failed to close repair offer.")
			return terror.Error(err, "Failed to close repair offer.")
		}
		return nil
	})
	if err != nil {
		return err
	}

	reply(true)

	return nil
}

// BroadcastMechQueueStat broadcast current mech queue stat
func BroadcastMechQueueStat(mechID string) {
	ci, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
		boiler.CollectionItemWhere.ItemID.EQ(mechID),
		qm.Load(boiler.CollectionItemRels.Owner),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return
	}

	if ci != nil && ci.R != nil && ci.R.Owner != nil && ci.R.Owner.FactionID.Valid {
		owner := ci.R.Owner
		queueDetails, err := db.GetCollectionItemStatus(*ci)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to get mech arena status")
			return
		}

		ws.PublishMessage(fmt.Sprintf("/faction/%s/queue/%s", owner.FactionID.String, mechID), battle.WSPlayerAssetMechQueueSubscribe, queueDetails)
	}
}

func BlockStackingGameVerification(ra *boiler.RepairAgent) error {
	// log path
	gps, err := ra.RepairAgentLogs(
		boiler.RepairAgentLogWhere.RepairAgentID.EQ(ra.ID),
		boiler.RepairAgentLogWhere.Score.GT(0),
		qm.OrderBy(boiler.RepairAgentLogColumns.CreatedAt),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to log mini-game records")
		return terror.Error(err, "Failed to load repair records.")
	}

	startTime := ra.StartedAt
	endTime := time.Now()

	failedRate := db.GetDecimalWithDefault(db.KeyRepairMiniGameFailedRate, decimal.NewFromFloat(0.25))

	// check each pattern is within the time frame
	failedCount := 0

	prevScore := 0
	totalStack := 0
	lastStackFailed := false
	for i, gp := range gps {
		if i > 0 {
			// valid score pattern
			// 1. current score equal to previous score + 1
			// 2. current score equal to previous score, and current stack is failed
			// 3. current score equal to 1, and last stack failed

			isValidScorePattern := false
			if gp.Score == prevScore+1 {
				// meet RULE 1
				isValidScorePattern = true

			} else if gp.Score == prevScore && gp.IsFailed {
				// meet RULE 2
				isValidScorePattern = true

			} else if gp.Score == 1 && lastStackFailed {
				// meet RULE 3
				isValidScorePattern = true

			}

			// if score pattern does not match
			if !isValidScorePattern {
				gamelog.L.Debug().Interface("current stack", gp).Int("prev score", prevScore).Msg("Invalid game pattern detected")
				return terror.Error(fmt.Errorf("invalid game score, current score: %d, prev score: %d, current failed: %v, agent id: %s", gp.Score, prevScore, gp.IsFailed, gp.RepairAgentID), "Invalid game pattern detected.")
			}
		}

		// set initial score and failed stat
		prevScore = gp.Score
		lastStackFailed = gp.IsFailed

		if gp.CreatedAt.Before(startTime) || gp.CreatedAt.After(endTime) {
			gamelog.L.Debug().Time("current stack time", gp.CreatedAt).Time("start time", startTime).Time("end time", endTime).Msg("Invalid game pattern detected")
			return terror.Error(fmt.Errorf("pattern is outside of time frame, stack time: %v, start time: %v, end time: %v, agent id: %s", gp.CreatedAt, startTime, endTime, gp.RepairAgentID), "Game stack is outside of the time frame.")
		}

		// increase failed count, if failed
		if gp.IsFailed {
			failedCount += 1
			continue
		}

		// increment score
		totalStack += 1
	}

	// if player failed 25% of the clicks
	if decimal.NewFromInt(int64(failedCount)).GreaterThanOrEqual(decimal.NewFromInt(int64(ra.RequiredStacks)).Mul(failedRate)) {
		return terror.Error(fmt.Errorf("stack failed"), "Too many failed stacks.")
	}

	// check the stack amount match
	if totalStack < ra.RequiredStacks {
		gamelog.L.Warn().
			Err(fmt.Errorf("stack not complete")).
			Int("totalStack", totalStack).
			Int("requiredStacks", ra.RequiredStacks).
			Interface("gps", gps).
			Interface("repair agent", ra).
			Msg("totalStack less than required stacks")
		return terror.Error(fmt.Errorf("stack not complete"), "The task is not completed.")
	}

	return nil
}

type RepairAgentAbandonRequest struct {
	Payload struct {
		RepairAgentID string `json:"repair_agent_id"`
	} `json:"payload"`
}

func (api *API) RepairAgentAbandon(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	if mechRepairAgentBucket.Add(user.ID, 1) == 0 {
		return nil
	}
	L := gamelog.L.With().Str("func", "RepairAgentAbandon").Interface("user", user).Logger()

	req := &RepairAgentAbandonRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	ra, err := boiler.RepairAgents(
		boiler.RepairAgentWhere.ID.EQ(req.Payload.RepairAgentID),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to load repair agent")
	}

	if ra.PlayerID != user.ID {
		return terror.Error(fmt.Errorf("player not match"), "Player does not match")
	}

	if ra.FinishedAt.Valid {
		return terror.Error(fmt.Errorf("repair agent is already closed"), "Repair agent is already closed.")
	}

	ra.FinishedAt = null.TimeFrom(time.Now())
	ra.FinishedReason = null.StringFrom(boiler.RepairAgentFinishReasonABANDONED)
	_, err = ra.Update(gamedb.StdConn, boil.Whitelist(boiler.RepairAgentColumns.FinishedAt, boiler.RepairAgentColumns.FinishedReason))
	if err != nil {
		L.Error().Err(err).Interface("repair agent", ra).Msg("Failed to close repair agent.")
		return terror.Error(err, "Failed to abandon the repair agent.")
	}

	err = api.broadcastRepairOffer(ra.RepairOfferID)
	if err != nil {
		return err
	}

	reply(true)

	return nil
}

type MechRepairSlotInsertRequest struct {
	Payload struct {
		MechIDs []string `json:"mech_ids"`
	} `json:"payload"`
}

func (api *API) MechRepairSlotInsert(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	L := gamelog.L.With().Str("func", "MechRepairSlotInsert").Interface("user", user).Logger()

	req := &MechRepairSlotInsertRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// validate ownership
	cis, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
		boiler.CollectionItemWhere.ItemID.IN(req.Payload.MechIDs),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("item type", boiler.ItemTypeMech).Strs("mech id list", req.Payload.MechIDs).Msg("Failed to query war machine collection item")
		return terror.Error(err, "Failed to load war machine detail.")
	}

	if len(req.Payload.MechIDs) != len(cis) {
		return terror.Error(fmt.Errorf("contain non-mech asset"), "Request contain non-mech asset.")
	}

	for _, ci := range cis {
		if ci.OwnerID != user.ID {
			return terror.Error(fmt.Errorf("do not own the mech"), "The mech is not owned by you.")
		}
	}

	maximumRepairSlotCount := db.GetIntWithDefault(db.KeyAutoRepairSlotCount, 5)
	nextRepairDurationSeconds := db.GetIntWithDefault(db.KeyAutoRepairDurationSeconds, 600)
	now := time.Now()

	shouldBroadcast := false
	err = api.ArenaManager.SendRepairFunc(func() error {
		// check remain slots
		occupiedSlotCount, err := boiler.PlayerMechRepairSlots(
			boiler.PlayerMechRepairSlotWhere.PlayerID.EQ(user.ID),
			boiler.PlayerMechRepairSlotWhere.Status.NEQ(boiler.RepairSlotStatusDONE),
		).Count(gamedb.StdConn)
		if err != nil {
			L.Error().Err(err).Str("player id", user.ID).Msg("Failed to check remain repair slot count.")
			return terror.Error(err, "Failed to check remain repair slot count.")
		}

		// return if no slot left
		if maximumRepairSlotCount <= int(occupiedSlotCount) {
			return nil
		}

		remainSlot := maximumRepairSlotCount - int(occupiedSlotCount)

		// filter out mechs by db query
		rcs, err := boiler.RepairCases(
			boiler.RepairCaseWhere.MechID.IN(req.Payload.MechIDs),
			boiler.RepairCaseWhere.CompletedAt.IsNull(),

			// filter out mechs which are in queue
			qm.Where(
				fmt.Sprintf(
					"NOT EXISTS ( SELECT 1 FROM %s WHERE %s = %s )",
					boiler.TableNames.BattleQueue,
					qm.Rels(boiler.TableNames.BattleQueue, boiler.BattleQueueColumns.MechID),
					qm.Rels(boiler.TableNames.RepairCases, boiler.RepairCaseColumns.MechID),
				),
			),

			// filter out mechs which are already in slot
			qm.Where(
				fmt.Sprintf(
					"NOT EXISTS ( SELECT 1 FROM %s WHERE %s = %s AND %s != ?)",
					boiler.TableNames.PlayerMechRepairSlots,
					qm.Rels(boiler.TableNames.PlayerMechRepairSlots, boiler.PlayerMechRepairSlotColumns.MechID),
					qm.Rels(boiler.TableNames.RepairCases, boiler.RepairCaseColumns.MechID),
					qm.Rels(boiler.TableNames.PlayerMechRepairSlots, boiler.PlayerMechRepairSlotColumns.Status),
				),
				boiler.RepairSlotStatusDONE,
			),
		).All(gamedb.StdConn)
		if err != nil {
			L.Error().Err(err).Strs("mech id list", req.Payload.MechIDs).Msg("Failed to check mechs' repair case.")
			return terror.Error(err, "Failed to check mechs' repair case.")
		}

		// return, if no mechs are available
		if rcs == nil {
			return nil
		}
		// index for inserted mech
		insertedIndex := 0
		// insert into slots in the input order
		for _, mechID := range req.Payload.MechIDs {
			// continue, if not reach remain slot count
			if insertedIndex >= remainSlot {
				break
			}

			// skip, if not available
			idx := slices.IndexFunc(rcs, func(rc *boiler.RepairCase) bool { return rc.MechID == mechID })
			if idx == -1 {
				continue
			}

			pmr := boiler.PlayerMechRepairSlot{
				PlayerID:     user.ID,
				MechID:       mechID,
				RepairCaseID: rcs[idx].ID,
				Status:       boiler.RepairSlotStatusPENDING,
				SlotNumber:   int(occupiedSlotCount) + (insertedIndex + 1),
			}

			// if this is the first slot, and currently no slot is occupied
			if pmr.SlotNumber == 1 {
				pmr.Status = boiler.RepairSlotStatusREPAIRING
				pmr.NextRepairTime = null.TimeFrom(now.Add(time.Duration(nextRepairDurationSeconds) * time.Second))
			}

			err = pmr.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Interface("repair slot", pmr).Msg("Failed to insert repair slot")
				return terror.Error(err, "Failed to insert repair slot.")
			}

			shouldBroadcast = true
			insertedIndex++
		}

		return nil
	})

	// broadcast changes, if slot changed
	if shouldBroadcast {
		go battle.BroadcastRepairBay(user.ID)
	}

	if err != nil {
		return err
	}

	reply(true)

	return nil
}

type MechRepairSlotRemoveRequest struct {
	Payload struct {
		MechIDs []string `json:"mech_ids"`
	} `json:"payload"`
}

func (api *API) MechRepairSlotRemove(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	L := gamelog.L.With().Str("func", "MechRepairSlotRemove").Interface("user", user).Logger()

	req := &MechRepairSlotRemoveRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// validate ownership
	cis, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
		boiler.CollectionItemWhere.ItemID.IN(req.Payload.MechIDs),
	).All(gamedb.StdConn)
	if err != nil {
		L.Error().Err(err).Str("item type", boiler.ItemTypeMech).Strs("mech id list", req.Payload.MechIDs).Msg("Failed to query war machine collection item")
		return terror.Error(err, "Failed to load war machine detail.")
	}

	if len(req.Payload.MechIDs) != len(cis) {
		return terror.Error(fmt.Errorf("contain non-mech asset"), "Request contain non-mech asset.")
	}

	for _, ci := range cis {
		if ci.OwnerID != user.ID {
			return terror.Error(fmt.Errorf("do not own the mech"), "The mech is not owned by you.")
		}
	}

	nextRepairDurationSeconds := db.GetIntWithDefault(db.KeyAutoRepairDurationSeconds, 600)
	now := time.Now()

	err = api.ArenaManager.SendRepairFunc(func() error {
		tx, err := gamedb.StdConn.Begin()
		if err != nil {
			L.Error().Err(err).Msg("Failed to start db transaction.")
			return terror.Error(err, "Failed to start db transaction")
		}

		defer tx.Rollback()

		count, err := boiler.PlayerMechRepairSlots(
			boiler.PlayerMechRepairSlotWhere.MechID.IN(req.Payload.MechIDs),
			boiler.PlayerMechRepairSlotWhere.Status.NEQ(boiler.RepairSlotStatusDONE),
		).UpdateAll(
			tx,
			boiler.M{
				boiler.PlayerMechRepairSlotColumns.Status:         boiler.RepairSlotStatusDONE,
				boiler.PlayerMechRepairSlotColumns.SlotNumber:     0,
				boiler.PlayerMechRepairSlotColumns.NextRepairTime: null.TimeFromPtr(nil),
			},
		)
		if err != nil {
			L.Error().Err(err).Strs("mech id list", req.Payload.MechIDs).Msg("Failed to update repair slot.")
			return terror.Error(err, "Failed to update repair slot")
		}

		// update remain slots and broadcast
		resp := []*boiler.PlayerMechRepairSlot{}
		if count > 0 {
			pms, err := boiler.PlayerMechRepairSlots(
				boiler.PlayerMechRepairSlotWhere.PlayerID.EQ(user.ID),
				boiler.PlayerMechRepairSlotWhere.Status.NEQ(boiler.RepairSlotStatusDONE),
				qm.OrderBy(boiler.PlayerMechRepairSlotColumns.SlotNumber),
			).All(tx)
			if err != nil {
				L.Error().Err(err).Msg("Failed to load player mech repair slots.")
				return terror.Error(err, "Failed to load repair slots")
			}

			for i, pm := range pms {
				shouldUpdate := false

				// check slot number
				if pm.SlotNumber != i+1 {
					pm.SlotNumber = i + 1
					shouldUpdate = true
				}

				if pm.SlotNumber == 1 {
					if pm.Status != boiler.RepairSlotStatusREPAIRING {
						pm.Status = boiler.RepairSlotStatusREPAIRING
						shouldUpdate = true
					}

					if !pm.NextRepairTime.Valid {
						pm.NextRepairTime = null.TimeFrom(now.Add(time.Duration(nextRepairDurationSeconds) * time.Second))
						shouldUpdate = true
					}
				} else {
					if pm.Status != boiler.RepairSlotStatusPENDING {
						pm.Status = boiler.RepairSlotStatusPENDING
						shouldUpdate = true
					}

					if pm.NextRepairTime.Valid {
						pm.NextRepairTime = null.TimeFromPtr(nil)
						shouldUpdate = true
					}
				}

				if shouldUpdate {
					_, err = pm.Update(tx,
						boil.Whitelist(
							boiler.PlayerMechRepairSlotColumns.SlotNumber,
							boiler.PlayerMechRepairSlotColumns.Status,
							boiler.PlayerMechRepairSlotColumns.NextRepairTime,
						),
					)
					if err != nil {
						return terror.Error(err, "Failed to update repair slot")
					}
				}

				resp = append(resp, pm)
			}
		}

		err = tx.Commit()
		if err != nil {
			L.Error().Err(err).Msg("Failed to commit db transaction.")
			return terror.Error(err, "Failed to commit db transaction.")
		}

		// broadcast new list
		ws.PublishMessage(fmt.Sprintf("/secure/user/%s/repair_bay", user.ID), server.HubKeyMechRepairSlots, resp)

		return nil
	})
	if err != nil {
		return err
	}

	reply(true)
	return nil
}

type MechRepairSlotSwapRequest struct {
	Payload struct {
		FromMechID string `json:"from_mech_id"`
		ToMechID   string `json:"to_mech_id"`
	} `json:"payload"`
}

func (api *API) MechRepairSlotSwap(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	L := gamelog.L.With().Str("func", "MechRepairSlotRemove").Interface("user", user).Logger()

	req := &MechRepairSlotSwapRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	mechIDs := []string{req.Payload.FromMechID, req.Payload.ToMechID}

	// validate ownership
	cis, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
		boiler.CollectionItemWhere.ItemID.IN(mechIDs),
	).All(gamedb.StdConn)
	if err != nil {
		L.Error().Err(err).Str("item type", boiler.ItemTypeMech).Strs("mech id list", mechIDs).Msg("Failed to query war machine collection item")
		return terror.Error(err, "Failed to load war machine detail.")
	}

	if len(mechIDs) != len(cis) {
		return terror.Error(fmt.Errorf("contain non-mech asset"), "Request contain non-mech asset.")
	}

	for _, ci := range cis {
		if ci.OwnerID != user.ID {
			return terror.Error(fmt.Errorf("do not own the mech"), "The mech is not owned by you.")
		}
	}

	nextRepairDurationSeconds := db.GetIntWithDefault(db.KeyAutoRepairDurationSeconds, 600)
	now := time.Now()

	err = api.ArenaManager.SendRepairFunc(func() error {
		tx, err := gamedb.StdConn.Begin()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to start db transaction.")
			return terror.Error(err, "Failed to start db transaction")
		}

		defer tx.Rollback()

		pms, err := boiler.PlayerMechRepairSlots(
			boiler.PlayerMechRepairSlotWhere.MechID.IN(mechIDs),
			boiler.PlayerMechRepairSlotWhere.Status.NEQ(boiler.RepairSlotStatusDONE),
		).All(tx)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to load player mech repair slots.")
			return terror.Error(err, "Failed to load repair slots")
		}

		if len(pms) != len(mechIDs) {
			return terror.Error(fmt.Errorf("mech not found"), "The mech is not in the list")
		}

		slotOne := pms[0]
		slotTwo := pms[1]

		newRepairSlots := []*boiler.PlayerMechRepairSlot{
			{
				// slot 1 id
				ID: slotOne.ID,

				// slot 2 details
				Status:         slotTwo.Status,
				SlotNumber:     slotTwo.SlotNumber,
				NextRepairTime: null.TimeFromPtr(nil),
			},
			{
				// slot 2 id
				ID: slotTwo.ID,

				// slot 1 details
				Status:         slotOne.Status,
				SlotNumber:     slotOne.SlotNumber,
				NextRepairTime: null.TimeFromPtr(nil),
			},
		}

		for _, slot := range newRepairSlots {
			// set next repair time, if status is repairing
			if slot.Status == boiler.RepairSlotStatusREPAIRING {
				slot.NextRepairTime = null.TimeFrom(now.Add(time.Duration(nextRepairDurationSeconds) * time.Second))
			}

			// update repair slot
			_, err = slot.Update(tx, boil.Whitelist(
				boiler.PlayerMechRepairSlotColumns.SlotNumber,
				boiler.PlayerMechRepairSlotColumns.Status,
				boiler.PlayerMechRepairSlotColumns.NextRepairTime,
			))
			if err != nil {
				gamelog.L.Error().Err(err).Interface("repair slot", slot).Msg("Failed to update repair slot.")
				return terror.Error(err, "Failed to update repair slot")
			}
		}

		err = tx.Commit()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
			return terror.Error(err, "Failed to commit db transaction.")
		}

		// broadcast new repair bay
		go battle.BroadcastRepairBay(user.ID)

		return nil
	})
	if err != nil {
		return err
	}

	reply(true)

	return nil
}

// subscription

// RepairOfferSubscribe return the detail of the offer
func (api *API) RepairOfferSubscribe(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	offerID := chi.RouteContext(ctx).URLParam("offer_id")
	if offerID == "" {
		return fmt.Errorf("offer id is required")
	}

	ro, err := db.RepairOfferDetail(offerID)
	if err != nil {
		return terror.Error(err, "Failed to load repair offer.")
	}

	reply(ro)

	return nil
}

// MechRepairCaseSubscribe return the ongoing repair case of the mech
func (api *API) MechRepairCaseSubscribe(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	mechID := chi.RouteContext(ctx).URLParam("mech_id")
	if mechID == "" {
		return fmt.Errorf("offer id is required")
	}

	rc, err := boiler.RepairCases(
		boiler.RepairCaseWhere.MechID.EQ(mechID),
		boiler.RepairCaseWhere.CompletedAt.IsNull(),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to load mech repair case.")
	}

	reply(rc)

	return nil
}

// MechActiveRepairOfferSubscribe show the active repair offer of the given mech
func (api *API) MechActiveRepairOfferSubscribe(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	mechID := chi.RouteContext(ctx).URLParam("mech_id")
	if mechID == "" {
		return fmt.Errorf("offer id is required")
	}

	rc, err := boiler.RepairCases(
		boiler.RepairCaseWhere.MechID.EQ(mechID),
		boiler.RepairCaseWhere.CompletedAt.IsNull(),
		qm.Load(
			boiler.RepairCaseRels.RepairOffers,
			boiler.RepairOfferWhere.ClosedAt.IsNull(),
			boiler.RepairOfferWhere.OfferedByID.IsNotNull(),
		),
		qm.Load(
			qm.Rels(boiler.RepairCaseRels.RepairOffers, boiler.RepairOfferRels.OfferedBy),
			qm.Select(
				boiler.PlayerColumns.ID,
				boiler.PlayerColumns.Gid,
				boiler.PlayerColumns.Username,
				boiler.PlayerColumns.FactionID,
			),
		),
		qm.Load(
			boiler.RepairCaseRels.RepairAgents,
			boiler.RepairAgentWhere.FinishedAt.IsNull(),
		),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to load mech repair case.")
	}

	if rc != nil && rc.R != nil && rc.R.RepairOffers != nil && len(rc.R.RepairOffers) > 0 {
		ro := rc.R.RepairOffers[0]
		sro := server.RepairOffer{
			RepairOffer:          ro,
			BlocksRequiredRepair: rc.BlocksRequiredRepair,
			BlocksRepaired:       rc.BlocksRepaired,
			SupsWorthPerBlock:    ro.OfferedSupsAmount.Div(decimal.NewFromInt(int64(ro.BlocksTotal))),
			WorkingAgentCount:    0,
		}
		if rc.R.RepairAgents != nil {
			sro.WorkingAgentCount = len(rc.R.RepairAgents)
		}
		if ro.R != nil && ro.R.OfferedBy != nil {
			sro.JobOwner = server.PublicPlayerFromBoiler(ro.R.OfferedBy)
		}

		reply(sro)
	}

	return nil
}

// PlayerMechRepairSlots return current player repair bay status
func (api *API) PlayerMechRepairSlots(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	l := gamelog.L.With().Str("player id", user.ID).Str("func name", "PlayerMechRepairSlots").Logger()

	resp := []*boiler.PlayerMechRepairSlot{}
	pms, err := boiler.PlayerMechRepairSlots(
		boiler.PlayerMechRepairSlotWhere.PlayerID.EQ(user.ID),
		boiler.PlayerMechRepairSlotWhere.Status.NEQ(boiler.RepairSlotStatusDONE),
		qm.OrderBy(boiler.PlayerMechRepairSlotColumns.SlotNumber),
	).All(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("Failed to load player mech .")
		return err
	}

	for _, pm := range pms {
		resp = append(resp, pm)
	}

	reply(resp)

	return nil
}
