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
	api.SecureUserCommand(server.HubKeyRepairOfferList, api.RepairOfferList)
	api.SecureUserCommand(server.HubKeyRepairOfferIssue, api.RepairOfferIssue)
	api.SecureUserCommand(server.HubKeyRepairOfferClose, api.RepairOfferClose)
	api.SecureUserCommand(server.HubKeyRepairAgentRegister, api.RepairAgentRegister)
	api.SecureUserCommand(server.HubKeyRepairAgentComplete, api.RepairAgentComplete)
}

type RepairListRequest struct {
	Payload struct {
		OrderBy    string   `json:"order_by"`
		OrderDir   string   `json:"order_dir"`
		IsExpired  bool     `json:"is_expired"`
		PageSize   int      `json:"page_size"`
		PageNumber int      `json:"page_number"`
		MaxReward  null.Int `json:"max_reward"`
		MinReward  null.Int `json:"min_reward"`
	} `json:"payload"`
}

type RepairOfferListResponse struct {
	Offers []*server.RepairOffer `json:"offers"`
	Total  int64                 `json:"total"`
}

func (api *API) RepairOfferList(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &RepairListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	resp := &RepairOfferListResponse{
		Offers: []*server.RepairOffer{},
		Total:  0,
	}
	queries := []qm.QueryMod{
		boiler.RepairOfferWhere.OfferedByID.IsNotNull(), // only get non-system generated offers
	}

	if req.Payload.MinReward.Valid {
		queries = append(queries, qm.Where(
			fmt.Sprintf(
				"%s/%s >= ?",
				qm.Rels(boiler.TableNames.RepairOffers, boiler.RepairOfferColumns.OfferedSupsAmount),
				qm.Rels(boiler.TableNames.RepairOffers, boiler.RepairOfferColumns.BlocksTotal),
			),
			decimal.New(int64(req.Payload.MinReward.Int), 18).StringFixed(0),
		))
	}

	if req.Payload.MaxReward.Valid {
		queries = append(queries, qm.Where(
			fmt.Sprintf(
				"%s/%s <= ?",
				qm.Rels(boiler.TableNames.RepairOffers, boiler.RepairOfferColumns.OfferedSupsAmount),
				qm.Rels(boiler.TableNames.RepairOffers, boiler.RepairOfferColumns.BlocksTotal),
			),
			decimal.New(int64(req.Payload.MaxReward.Int), 18).StringFixed(0),
		))
	}

	if req.Payload.IsExpired {
		queries = append(queries, boiler.RepairOfferWhere.ClosedAt.IsNotNull())
	} else {
		queries = append(queries,
			boiler.RepairOfferWhere.ExpiresAt.GT(time.Now()),
			boiler.RepairOfferWhere.ClosedAt.IsNull(),
		)
	}

	resp.Total, err = boiler.RepairOffers(queries...).Count(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to query offer list.")
		return terror.Error(err, "Failed to get the offer list.")
	}

	// validate order direction
	switch req.Payload.OrderDir {
	case "DESC", "ASC":
	default:
		return terror.Error(fmt.Errorf("invalid order direction"), "Invalid order direction.")
	}

	// validate order by column
	switch req.Payload.OrderBy {
	case boiler.RepairOfferColumns.ExpiresAt, boiler.RepairOfferColumns.OfferedSupsAmount, boiler.RepairOfferColumns.CreatedAt:
	default:
		return terror.Error(fmt.Errorf("invalid order option"), "Invalid order option.")
	}

	queries = append(queries,
		qm.OrderBy(fmt.Sprintf("%s %s", req.Payload.OrderBy, req.Payload.OrderDir)),
		qm.Limit(req.Payload.PageSize),
		qm.Offset(req.Payload.PageNumber*req.Payload.PageSize),
		qm.Load(
			boiler.RepairOfferRels.OfferedBy,
			qm.Select(
				boiler.PlayerColumns.ID,
				boiler.PlayerColumns.Username,
				boiler.PlayerColumns.Gid,
				boiler.PlayerColumns.FactionID,
			),
		),
	)

	ros, err := boiler.RepairOffers(queries...).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to query offer list from db.")
		return terror.Error(err, "Failed to get offer list.")
	}

	for _, ro := range ros {
		resp.Offers = append(resp.Offers, &server.RepairOffer{
			RepairOffer: ro,
			JobOwner:    ro.R.OfferedBy,
		})
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
var mechRepairOfferBucket = leakybucket.NewCollector(0.5, 1, true)

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

	offeredSups := req.Payload.OfferedSups.Mul(decimal.New(1, 18))

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
	charges := offeredSups.Mul(decimal.NewFromFloat(1.1)).Round(0)

	// pay sups to offer repair job
	_, err = api.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           uuid.Must(uuid.FromString(user.ID)),
		ToUserID:             uuid.Must(uuid.FromString(server.SupremacyGameUserID)),
		Amount:               charges.String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("create_repair_offer|%s|%d", ro.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupRepair),
		Description:          "create a repair offer",
		NotSafe:              true,
	})
	if err != nil {
		gamelog.L.Error().Str("player_id", user.ID).Str("repair offer id", ro.ID).Str("amount", charges.String()).Err(err).Msg("Failed to pay sups for offering repair job")
		return terror.Error(err, "Failed to pay sups for offering repair job.")
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
		return terror.Error(err, "Failed to offer repair contract.")
	}

	sro := server.RepairOffer{
		RepairOffer:          ro,
		BlocksRequiredRepair: mrc.BlocksRequiredRepair,
		BlocksRepaired:       mrc.BlocksRepaired,
		SupsWorthPerBlock:    offeredSups.Div(decimal.NewFromInt(int64(ro.BlocksTotal))),
		WorkingAgentCount:    0,
	}

	//  broadcast to repair offer market
	ws.PublishMessage("/public/repair_offer/new", server.HubKeyNewRepairOfferSubscribe, sro)
	ws.PublishMessage(fmt.Sprintf("/public/repair_offer/%s", ro.ID), server.HubKeyRepairOfferSubscribe, sro)
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
	} `json:"payload"`
}

var mechRepairAgentBucket = leakybucket.NewCollector(0.5, 1, true)

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

	// abandon any unfinished repair task
	_, err = boiler.RepairAgents(
		boiler.RepairAgentWhere.PlayerID.EQ(user.ID),
		boiler.RepairAgentWhere.FinishedAt.IsNull(),
	).UpdateAll(gamedb.StdConn,
		boiler.M{
			boiler.RepairAgentColumns.FinishedAt:     null.TimeFrom(time.Now()),
			boiler.RepairAgentColumns.FinishedReason: null.StringFrom(boiler.RepairAgentFinishReasonABANDONED),
		},
	)
	if err != nil {
		gamelog.L.Error().Err(err).Str("player id", user.ID).Msg("Failed to close repair agents.")
		return terror.Error(err, "Failed to abandon repair job")
	}

	// insert repair agent
	ra := &boiler.RepairAgent{
		RepairCaseID:  ro.RepairCaseID,
		RepairOfferID: ro.ID,
		PlayerID:      user.ID,
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
	}

	return nil
}

type RepairAgentCompleteRequest struct {
	Payload struct {
		RepairAgentID string `json:"repair_agent_id"`
	} `json:"payload"`
}

func (api *API) RepairAgentComplete(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	if mechRepairAgentBucket.Add(user.ID, 1) == 0 {
		return terror.Error(fmt.Errorf("too many request"), "Too many request.")
	}

	req := &RepairAgentCompleteRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	ra, err := boiler.FindRepairAgent(gamedb.StdConn, req.Payload.RepairAgentID)
	if err != nil {
		return terror.Error(err, "Failed to load repair agent.")
	}

	if ra.PlayerID != user.ID {
		return terror.Error(fmt.Errorf("agnet id not match"), "Repair agent id mismatch")
	}

	if ra.FinishedAt.Valid {
		return terror.Error(fmt.Errorf("agent finalised"), "This repair agent is already finalised.")
	}

	rb := boiler.RepairBlock{
		RepairAgentID: ra.ID,
		RepairCaseID:  ra.RepairCaseID,
		RepairOfferID: ra.RepairOfferID,
	}

	err = rb.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		if err.Error() == "unable to write block" {
			return terror.Error(err, "repair offer is already closed.")
		}

		return terror.Error(err, "Failed to complete repair agent task.")
	}

	// check repair case after insert
	rc, err := ra.RepairCase().One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to load repair case.")
	}

	// claim sups
	ro, err := db.RepairOfferDetail(ra.RepairOfferID)
	if err != nil {
		return terror.Error(err, "Failed to load repair offer")
	}

	// if it is a not self offer
	if ro.OfferedByID.Valid && ro.OfferedSupsAmount.GreaterThan(decimal.Zero) {
		amount := ro.OfferedSupsAmount.Div(decimal.NewFromInt(int64(ro.BlocksRequiredRepair))).StringFixed(0)

		// claim reward
		_, err = api.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
			FromUserID:           uuid.Must(uuid.FromString(server.SupremacyGameUserID)),
			ToUserID:             uuid.Must(uuid.FromString(user.ID)),
			Amount:               amount,
			TransactionReference: server.TransactionReference(fmt.Sprintf("claim_repair_offer_reward|%s|%d", ro.ID, time.Now().UnixNano())),
			Group:                string(server.TransactionGroupSupremacy),
			SubGroup:             string(server.TransactionGroupRepair),
			Description:          "claim repair offer reward.",
			NotSafe:              true,
		})
		if err != nil {
			gamelog.L.Error().Str("player_id", user.ID).Str("repair offer id", ro.ID).Str("amount", amount).Err(err).Msg("Failed to pay sups for offering repair job")
			return terror.Error(err, "Failed to pay sups for offering repair job.")
		}
	}

	// broadcast result if repair is not completed
	if rc.BlocksRepaired < rc.BlocksRequiredRepair {
		ws.PublishMessage(fmt.Sprintf("/public/repair_offer/%s", ro.ID), server.HubKeyRepairOfferSubscribe, ro)
		if ro.OfferedByID.Valid {
			ws.PublishMessage(fmt.Sprintf("/public/mech/%s/active_repair_offer", ro.ID), server.HubKeyMechActiveRepairOffer, ro)
		}
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
		gamelog.L.Error().Err(err).Msg("Failed to update repair case.")
		return terror.Error(err, "Failed to close repair case.")
	}

	// close offer, self and non-self
	ros, err := rc.RepairOffers(
		boiler.RepairOfferWhere.ClosedAt.IsNull(),
	).All(gamedb.StdConn)
	if err != nil {
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

// subscription

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
			boiler.RepairCaseRels.RepairAgents,
			boiler.RepairAgentWhere.FinishedAt.IsNull(),
		),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to load mech repair case.")
	}

	if rc != nil && rc.R != nil && rc.R.RepairOffers != nil && len(rc.R.RepairOffers) > 0 {
		ro := rc.R.RepairOffers[0]
		workingAgentCount := 0
		if rc.R.RepairAgents != nil {
			workingAgentCount = len(rc.R.RepairAgents)
		}
		reply(server.RepairOffer{
			RepairOffer:          ro,
			BlocksRequiredRepair: rc.BlocksRequiredRepair,
			BlocksRepaired:       rc.BlocksRepaired,
			SupsWorthPerBlock:    ro.OfferedSupsAmount.Div(decimal.NewFromInt(int64(ro.BlocksTotal))),
			WorkingAgentCount:    workingAgentCount,
		})
	}

	return nil
}
