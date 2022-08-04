package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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
	"server"
	"server/battle"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"time"
)

func NewMechRepairController(api *API) {
	api.SecureUserCommand(server.HubKeyRepairOfferIssue, api.RepairOfferIssue)
	api.SecureUserCommand(server.HubKeyRepairOfferClose, api.RepairOfferClose)
	api.SecureUserCommand(server.HubKeyRepairAgentRegister, api.RepairAgentRegister)
	api.SecureUserCommand(server.HubKeyRepairAgentRecord, api.RepairAgentRecord)
	api.SecureUserCommand(server.HubKeyRepairAgentComplete, api.RepairAgentComplete)
	api.SecureUserCommand(server.HubKeyRepairAgentAbandon, api.RepairAgentAbandon)
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
		MechID         string          `json:"mech_id"`
		LastForMinutes int             `json:"last_for_minutes"`
		OfferedSups    decimal.Decimal `json:"offered_sups"` // the amount that excluded tax
	} `json:"payload"`
}

// prevent owner issue multi repair offer on the same mech
var mechRepairOfferBucket = leakybucket.NewCollector(2, 1, true)

func (api *API) RepairOfferIssue(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	now := time.Now()

	req := &RepairOfferIssueRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if mechRepairOfferBucket.Add(user.ID+req.Payload.MechID, 1) == 0 {
		return terror.Error(fmt.Errorf("too many request"), "Too many mech repair request.")
	}

	// validate ownership
	ci, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
		boiler.CollectionItemWhere.ItemID.EQ(req.Payload.MechID),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("item type", boiler.ItemTypeMech).Str("mech id", req.Payload.MechID).Msg("Failed to query war machine collection item")
		return terror.Error(err, "Failed to load war machine detail.")
	}

	if ci.OwnerID != user.ID {
		return terror.Error(fmt.Errorf("do not own the mech"), "The mech is not owned by you.")
	}

	// look for repair cases
	mrc, err := boiler.RepairCases(
		boiler.RepairCaseWhere.MechID.EQ(req.Payload.MechID),
		boiler.RepairCaseWhere.CompletedAt.IsNull(),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("mech id", req.Payload.MechID).Msg("Failed to query mech repair case.")
		return terror.Error(err, "Failed to load mech repair case.")
	}

	if mrc == nil {
		return terror.Error(fmt.Errorf("mech does not have repair case"), "The mech does not need to be repaired.")
	}

	if mrc.BlocksRequiredRepair == mrc.BlocksRepaired {
		return terror.Error(fmt.Errorf("mech already repaired"), "The mech has already repaired.")
	}

	unclosedOffer, err := mrc.RepairOffers(
		boiler.RepairOfferWhere.OfferedByID.IsNotNull(),
		boiler.RepairOfferWhere.ClosedAt.IsNull(), // check any unclosed offer
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("mech repair case", mrc.ID).Msg("Failed to queries repair offer.")
		return terror.Error(err, "There is check unclosed repair offer.")
	}

	if unclosedOffer != nil {
		return terror.Error(fmt.Errorf("unclosed offer exists"), "Cannot offer a new repair contract if the previous offer has not ended yet.")
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to begin db transaction.")
		return terror.Error(err, "Failed to offer repair job.")
	}

	defer tx.Rollback()

	// calculate required point
	err = mrc.Reload(tx)
	if err != nil {
		gamelog.L.Error().Err(err).Str("repair case id", mrc.ID).Msg("Failed to reload repair case.")
		return terror.Error(err, "Failed to load repair case.")
	}

	offeredSups := req.Payload.OfferedSups.Mul(decimal.New(1, 18)).Round(0)

	// remain hours
	// register a new repair offer
	ro := &boiler.RepairOffer{
		OfferedByID:       null.StringFrom(user.ID),
		RepairCaseID:      mrc.ID,
		BlocksTotal:       mrc.BlocksRequiredRepair - mrc.BlocksRepaired,
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
		FromUserID:           uuid.Must(uuid.FromString(user.ID)),
		ToUserID:             uuid.Must(uuid.FromString(server.RepairCenterUserID)),
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

	ro.PaidTXID = null.StringFrom(offerTXID)

	// pay tax to XSYN treasury
	offerTaxTXID, err := api.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           uuid.Must(uuid.FromString(server.RepairCenterUserID)),
		ToUserID:             uuid.UUID(server.XsynTreasuryUserID),
		Amount:               tax.String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("repair_offer_tax|%s|%d", ro.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupRepair),
		Description:          "repair offer tax",
	})
	if err != nil {
		gamelog.L.Error().Str("player_id", user.ID).Str("repair offer id", ro.ID).Str("amount", tax.String()).Err(err).Msg("Failed to pay tax for offering repair job")
		return terror.Error(err, "Failed to pay sups for offering repair job.")
	}

	ro.TaxTXID = null.StringFrom(offerTaxTXID)

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
		return terror.Error(err, "Failed to offer repair contract.")
	}

	_, err = ro.Update(gamedb.StdConn, boil.Whitelist(
		boiler.RepairOfferColumns.PaidTXID,
		boiler.RepairOfferColumns.TaxTXID,
	))
	if err != nil {
		gamelog.L.Error().Err(err).Interface("repair offer", ro).Msg("Failed to update repair offer transaction id.")
	}

	sro := &server.RepairOffer{
		RepairOffer:          ro,
		BlocksRequiredRepair: mrc.BlocksRequiredRepair,
		BlocksRepaired:       mrc.BlocksRepaired,
		SupsWorthPerBlock:    offeredSups.Div(decimal.NewFromInt(int64(ro.BlocksTotal))),
		WorkingAgentCount:    0,
		JobOwner:             server.PublicPlayerFromBoiler(user),
	}

	//  broadcast to repair offer market
	ws.PublishMessage("/public/repair_offer/new", server.HubKeyNewRepairOfferSubscribe, sro)
	ws.PublishMessage(fmt.Sprintf("/public/repair_offer/%s", ro.ID), server.HubKeyRepairOfferSubscribe, sro)
	ws.PublishMessage("/public/repair_offer/update", server.HubKeyRepairOfferUpdateSubscribe, []*server.RepairOffer{sro})
	ws.PublishMessage(fmt.Sprintf("/public/mech/%s/active_repair_offer", mrc.MechID), server.HubKeyMechActiveRepairOffer, sro)

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

	if mechRepairOfferBucket.Add(user.ID+req.Payload.RepairOfferID, 1) == 0 {
		return terror.Error(fmt.Errorf("too many request"), "Too many mech repair request.")
	}

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

	// close offer
	api.BattleArena.RepairOfferCloseChan <- &battle.RepairOfferClose{
		OfferIDs:          []string{ro.ID},
		OfferClosedReason: boiler.RepairFinishReasonSTOPPED,
		AgentClosedReason: boiler.RepairAgentFinishReasonEXPIRED,
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
		ws.PublishMessage(fmt.Sprintf("/public/repair_offer/%s", repairOfferID), server.HubKeyRepairOfferSubscribe, sro)
		ws.PublishMessage("/public/repair_offer/update", server.HubKeyRepairOfferUpdateSubscribe, []*server.RepairOffer{sro})
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

	switch req.Payload.TriggerWith {
	case boiler.RepairTriggerWithTypeSPACE_BAR, boiler.RepairTriggerWithTypeLEFT_CLICK, boiler.RepairTriggerWithTypeTOUCH:
	default:
		gamelog.L.Debug().Str("repair agent id", req.Payload.RepairAgentID).Msg("Unknown trigger type is detected.")
		return nil
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

	// log path
	ral, err := boiler.RepairAgentLogs(
		boiler.RepairAgentLogWhere.RepairAgentID.EQ(ra.ID),
		qm.OrderBy(boiler.RepairAgentLogColumns.CreatedAt),
	).All(gamedb.StdConn)
	if err != nil {
		L.Error().Err(err).Msg("failed to log mini-game records")
		return terror.Error(err, "Failed to load repair records.")
	}

	err = BlockStackingGameVerification(ra, ral)
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
			ws.PublishMessage(fmt.Sprintf("/public/repair_offer/%s", ro.ID), server.HubKeyRepairOfferSubscribe, ro)
			ws.PublishMessage("/public/repair_offer/update", server.HubKeyRepairOfferUpdateSubscribe, []*server.RepairOffer{ro})
			ws.PublishMessage(fmt.Sprintf("/public/mech/%s/active_repair_offer", ro.ID), server.HubKeyMechActiveRepairOffer, ro)
		}
	}

	// broadcast result if repair is not completed
	if rc.BlocksRepaired < rc.BlocksRequiredRepair {
		ws.PublishMessage(fmt.Sprintf("/public/mech/%s/repair_case", rc.MechID), server.HubKeyMechRepairCase, rc)
		reply(true)
		return nil
	}

	// clean up repair case if repair is completed
	ws.PublishMessage(fmt.Sprintf("/public/mech/%s/repair_case", rc.MechID), server.HubKeyMechRepairCase, nil)

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

	ids := []string{}
	for _, ro := range ros {
		ids = append(ids, ro.ID)
	}

	api.BattleArena.RepairOfferCloseChan <- &battle.RepairOfferClose{
		OfferIDs:          ids,
		OfferClosedReason: boiler.RepairAgentFinishReasonSUCCEEDED,
		AgentClosedReason: boiler.RepairAgentFinishReasonEXPIRED,
	}

	reply(true)

	return nil
}

func BlockStackingGameVerification(ra *boiler.RepairAgent, gps []*boiler.RepairAgentLog) error {
	startTime := ra.StartedAt
	endTime := time.Now()

	// check each pattern is within the time frame
	prevScore := 0
	failedLastTime := false
	prevStackAt := time.Now()
	totalStack := 0
	for i, gp := range gps {
		if i > 0 {
			// valid score pattern
			// 1. current score equal to previous score + 1
			// 2. current score equal to previous score, and current stack is failed
			// 3. current score equal to zero, and previous stack is failed

			isValidScorePattern := false
			if gp.Score == prevScore+1 {
				// meet RULE 1
				isValidScorePattern = true

			} else if gp.Score == prevScore && gp.IsFailed {
				// meet RULE 2
				isValidScorePattern = true

			} else if gp.Score == 0 && failedLastTime {
				// meet RULE 3
				isValidScorePattern = true
			}

			// if score pattern does not match
			if !isValidScorePattern {
				return terror.Error(fmt.Errorf("invalid game score"), "Invalid game pattern detected.")
			}

			if !prevStackAt.Before(gp.CreatedAt) {
				return terror.Error(fmt.Errorf("invalid stack time"), "Invalid game pattern detected.")
			}

		}

		// set initial score and failed stat
		prevScore = gp.Score
		failedLastTime = gp.IsFailed
		prevStackAt = gp.CreatedAt

		if gp.CreatedAt.Before(startTime) || gp.CreatedAt.After(endTime) {
			return terror.Error(fmt.Errorf("pattern is outside of time frame"), "Invalid game pattern detected.")
		}

		// reduce the invalid score
		if gp.IsFailed || gp.Score == 0 {
			continue
		}

		// increment score
		totalStack += 1
	}

	// check the stack amount match
	if totalStack < ra.RequiredStacks {
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
