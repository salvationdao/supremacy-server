package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/helpers"

	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v4"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

type PassportWebhookController struct {
	Conn db.Conn
	Log  *zerolog.Logger
	API  *API
}

func PassportWebhookRouter(log *zerolog.Logger, conn db.Conn, webhookSecret string, api *API) chi.Router {
	c := &PassportWebhookController{
		Conn: conn,
		Log:  log,
		API:  api,
	}
	r := chi.NewRouter()
	r.Post("/auth_ring_check", WithPassportSecret(webhookSecret, WithError(c.AuthRingCheck)))
	r.Post("/user_update", WithPassportSecret(webhookSecret, WithError(c.UserUpdated)))
	r.Post("/user_enlist_faction", WithPassportSecret(webhookSecret, WithError(c.UserEnlistFaction)))
	r.Post("/user_stat", WithPassportSecret(webhookSecret, WithError(c.UserStatGet)))
	r.Post("/faction_stat", WithPassportSecret(webhookSecret, WithError(c.FactionStatGet)))

	return r
}

type UserUpdateRequest struct {
	User *server.User `json:"user"`
}

// UserUpdated update user detail
func (pc *PassportWebhookController) UserUpdated(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &UserUpdateRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	// update users
	user, err := pc.API.UserMap.GetUserDetailByID(req.User.ID)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "User not found")
	}

	// update user
	user.FirstName = req.User.FirstName
	user.LastName = req.User.LastName
	user.Username = req.User.Username
	user.AvatarID = req.User.AvatarID
	user.FactionID = req.User.FactionID
	if !user.FactionID.IsNil() {
		faction, err := boiler.FindFaction(gamedb.StdConn, user.FactionID.String())
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err, "faction not found")
		}
		user.Faction = faction
	}

	broadcastData, err := json.Marshal(&BroadcastPayload{
		Key:     HubKeyUserSubscribe,
		Payload: user,
	})
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}
	for _, cl := range pc.API.UserMap.Update(user) {
		go cl.Send(broadcastData)
	}

	return helpers.EncodeJSON(w, struct {
		IsSuccess bool `json:"is_success"`
	}{
		IsSuccess: true,
	})
}

type UserEnlistFactionRequest struct {
	UserID    server.UserID    `json:"user_id"`
	FactionID server.FactionID `json:"faction_id"`
}

func (pc *PassportWebhookController) UserEnlistFaction(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &UserEnlistFactionRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	if req.FactionID.IsNil() || !req.FactionID.IsValid() {
		return http.StatusBadRequest, terror.Error(err, "Faction id is required")
	}

	// get player
	player, err := db.PlayerGet(req.UserID.String())
	if err != nil {
		return http.StatusBadRequest, terror.Error(err)
	}

	player.FactionID = null.StringFrom(req.FactionID.String())

	err = db.PlayerUpdateFactionID(player)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	pc.API.MessageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, player.ID)), player)

	return helpers.EncodeJSON(w, struct {
		IsSuccess bool `json:"is_success"`
	}{
		IsSuccess: true,
	})
}

type WarMachineJoinRequest struct {
	WarMachineMetadata *server.WarMachineMetadata `json:"warMachineMetadata"`
	NeedInsured        bool                       `json:"needInsured"`
}

type WarMachineJoinResp struct {
	Position       *int            `json:"position"`
	ContractReward decimal.Decimal `json:"contractReward"`
}

func (pc *PassportWebhookController) WarMachineJoin(w http.ResponseWriter, r *http.Request) (int, error) {
	//TODO ALEX fix
	return 0, nil
	//req := &WarMachineJoinRequest{}
	//err := json.NewDecoder(r.Body).Decode(req)
	//if err != nil {
	//	return http.StatusInternalServerError, terror.Error(err)
	//}
	//
	//if req.WarMachineMetadata.FactionID.IsNil() {
	//	return http.StatusBadRequest, terror.Error(fmt.Errorf("Non-faction war machine is not able to join"))
	//}
	//
	//err = pc.API.BattleArena.WarMachineQueue.Join(req.WarMachineMetadata, req.NeedInsured)
	//if err != nil {
	//	return http.StatusBadRequest, terror.Error(err, err.Error())
	//}
	//
	//// broadcast price change
	//factionQueuePrice := &passport.FactionQueuePriceUpdateReq{
	//	FactionID: req.WarMachineMetadata.FactionID,
	//}
	//switch req.WarMachineMetadata.FactionID {
	//case server.RedMountainFactionID:
	//	factionQueuePrice.QueuingLength = pc.API.BattleArena.WarMachineQueue.RedMountain.QueuingLength()
	//case server.BostonCyberneticsFactionID:
	//	factionQueuePrice.QueuingLength = pc.API.BattleArena.WarMachineQueue.Boston.QueuingLength()
	//case server.ZaibatsuFactionID:
	//	factionQueuePrice.QueuingLength = pc.API.BattleArena.WarMachineQueue.Zaibatsu.QueuingLength()
	//}
	//pc.API.Passport.FactionQueueCostUpdate(factionQueuePrice)
	//
	//errChan := make(chan error)
	//
	//// fire a payment to passport
	//pc.API.Passport.SpendSupMessage(passport.SpendSupsReq{
	//	FromUserID:           req.WarMachineMetadata.OwnedByID,
	//	ToUserID:             &server.XsynTreasuryUserID,
	//	Amount:               req.WarMachineMetadata.Fee.String(),
	//	TransactionReference: server.TransactionReference(fmt.Sprintf("war_machine_queuing_fee|%s", uuid.Must(uuid.NewV4()))),
	//	Group:                "Supremacy",
	//	Description:          "Adding war machine to queue.",
	//}, func(transaction string) {
	//	errChan <- nil
	//}, func(reqErr error) {
	//	// check faction id
	//	switch req.WarMachineMetadata.FactionID {
	//	case server.RedMountainFactionID:
	//		err = pc.API.BattleArena.WarMachineQueue.RedMountain.Leave(req.WarMachineMetadata.Hash)
	//		if err != nil {
	//			pc.Log.Err(err).Msg("")
	//		}
	//	case server.BostonCyberneticsFactionID:
	//		err = pc.API.BattleArena.WarMachineQueue.Boston.Leave(req.WarMachineMetadata.Hash)
	//		if err != nil {
	//			pc.Log.Err(err).Msg("")
	//		}
	//	case server.ZaibatsuFactionID:
	//		err = pc.API.BattleArena.WarMachineQueue.Zaibatsu.Leave(req.WarMachineMetadata.Hash)
	//		if err != nil {
	//			pc.Log.Err(err).Msg("")
	//		}
	//	}
	//	pc.API.Passport.SupremacyQueueUpdate(&server.SupremacyQueueUpdateReq{
	//		Hash: req.WarMachineMetadata.Hash,
	//	})
	//	errChan <- reqErr
	//})
	//
	//err = <-errChan
	//if err != nil {
	//	return http.StatusInternalServerError, terror.Error(err, "Issue joining queue")
	//}
	//
	//// prepare response
	//resp := &WarMachineJoinResp{}
	//// set insurance flag
	//warMachinePosition, _ := pc.API.BattleArena.WarMachineQueue.GetWarMachineQueue(req.WarMachineMetadata.FactionID, req.WarMachineMetadata.Hash)
	//if err != nil {
	//	return http.StatusInternalServerError, terror.Error(err)
	//}
	//
	//resp.Position = warMachinePosition
	//resp.ContractReward = decimal.New(int64((*warMachinePosition+1)*2), 0)
	//
	//// get contract reward
	//queuingStat, err := db.AssetQueuingStat(context.Background(), pc.Conn, req.WarMachineMetadata.Hash)
	//if err != nil && !errors.Is(err, pgx.ErrNoRows) {
	//	return http.StatusInternalServerError, terror.Error(err)
	//}
	//
	//queueingContractReward, err := decimal.NewFromString(queuingStat.ContractReward)
	//if err != nil {
	//	return http.StatusInternalServerError, terror.Error(err)
	//}
	//resp.ContractReward = decimal.Zero
	//if queuingStat != nil {
	//	resp.ContractReward = queueingContractReward
	//}
	//
	//// return current queuing position
	//return helpers.EncodeJSON(w, resp)
}

type UserSupsMultiplierGetRequest struct {
	UserID server.UserID `json:"user_id"`
}

type UserStatGetRequest struct {
	UserID server.UserID `json:"user_id"`
}

func (pc *PassportWebhookController) UserStatGet(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &UserSupsMultiplierGetRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	if req.UserID.IsNil() {
		return http.StatusBadRequest, terror.Error(terror.ErrInvalidInput, "User id is required")
	}

	userStat, err := db.UserStatGet(context.Background(), pc.Conn, req.UserID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get user stat")
	}

	if userStat == nil {
		// build a empty user stat if there is no user stat in db
		userStat = &server.UserStat{
			ID: req.UserID,
		}
	}

	return helpers.EncodeJSON(w, userStat)
}

type FactionStatGetRequest struct {
	FactionID server.FactionID `json:"faction_id"`
}

func (pc *PassportWebhookController) FactionStatGet(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &FactionStatGetRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	if req.FactionID.IsNil() {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("faction id is empty"), "Faction id is required")
	}

	factionStat := &server.FactionStat{
		ID: req.FactionID,
	}

	err = db.FactionStatGet(context.Background(), pc.Conn, factionStat)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, fmt.Sprintf("Failed to get faction %s stat", req.FactionID))
	}

	return helpers.EncodeJSON(w, factionStat)
}

type WarMachineQueuePositionRequest struct {
	FactionID server.FactionID `json:"factionID"`
	AssetHash string           `json:"assethash"`
}

// PassportWarMachineQueuePositionHandler return the list of user's war machines in the queue
func (pc *PassportWebhookController) WarMachineQueuePositionGet(w http.ResponseWriter, r *http.Request) (int, error) {
	return 0, nil
	//TODO ALEX: fix
	//req := &WarMachineQueuePositionRequest{}
	//err := json.NewDecoder(r.Body).Decode(req)
	//if err != nil {
	//	return http.StatusInternalServerError, terror.Error(err)
	//}
	//
	//position, contractReward := pc.API.BattleArena.WarMachineQueue.GetWarMachineQueue(req.FactionID, req.AssetHash)
	//if err != nil {
	//	return http.StatusInternalServerError, terror.Error(err)
	//}
	//
	//return helpers.EncodeJSON(w, struct {
	//	Position       *int            `json:"position"`
	//	ContractReward decimal.Decimal `json:"contractReward"`
	//}{
	//	Position:       position,
	//	ContractReward: contractReward,
	//})
}

type AssetRepairStatRequest struct {
	Hash string `json:"hash"`
}

func (pc *PassportWebhookController) AssetRepairStatGet(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &AssetRepairStatRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	record := &server.AssetRepairRecord{
		Hash: req.Hash,
	}
	err = db.AssetRepairIncompleteGet(context.Background(), pc.Conn, record)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return helpers.EncodeJSON(w, struct {
				AssetRepairRecord *server.AssetRepairRecord `json:"asset_repair_record"`
			}{
				AssetRepairRecord: &server.AssetRepairRecord{},
			})
		}

		return http.StatusInternalServerError, terror.Error(err)
	}
	return helpers.EncodeJSON(w, struct {
		AssetRepairRecord *server.AssetRepairRecord `json:"asset_repair_record"`
	}{
		AssetRepairRecord: record,
	})
}

type AuthRingCheckRequest struct {
	User                *server.User `json:"user"`
	GameserverSessionID string       `json:"gameserver_session_id"`
}

func (pc *PassportWebhookController) AuthRingCheck(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &AuthRingCheckRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	if req.GameserverSessionID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("no auth ring check key provided"), "Ring check key is required")
	}

	if !req.User.PublicAddress.Valid {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing user public address"), "User public address is required")
	}

	client, err := pc.API.RingCheckAuthMap.Check(req.GameserverSessionID)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Hub client not found")
	}

	// set client identifier
	client.SetIdentifier(req.User.ID.String())

	var factionID *string
	if !req.User.FactionID.IsNil() {
		str := req.User.FactionID.String()
		factionID = &str

		// get faction detail
		faction, err := db.FactionGet(str)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err)
		}

		req.User.Faction = faction
	}

	// store user into player table
	err = db.UpsertPlayer(&boiler.Player{
		ID:            req.User.ID.String(),
		Username:      null.StringFrom(req.User.Username),
		PublicAddress: req.User.PublicAddress,
		FactionID:     null.StringFromPtr(factionID),
	})
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	b, err := json.Marshal(&BroadcastPayload{
		Key:     HubKeyUserRingCheck,
		Payload: req.User,
	})
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	go client.Send(b)

	return helpers.EncodeJSON(w, struct {
		IsSuccess bool `json:"is_success"`
	}{
		IsSuccess: true,
	})

}

type FactionQueueCostGetRequest struct {
	FactionID server.FactionID `json:"factionID"`
}
