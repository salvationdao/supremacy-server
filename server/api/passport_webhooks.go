package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"server"
	"server/db"
	"server/helpers"
	"server/passport"

	"github.com/go-chi/chi"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/rs/zerolog"
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
	r.Post("/user_multiplier", WithPassportSecret(webhookSecret, WithError(c.UserSupsMultiplierGet)))

	r.Post("/war_machine_join", WithPassportSecret(webhookSecret, WithError(c.WarMachineJoin)))
	r.Post("/war_machine_queue_position", WithPassportSecret(webhookSecret, WithError(c.WarMachineQueuePositionGet)))
	r.Post("/asset_repair_stat", WithPassportSecret(webhookSecret, WithError(c.AssetRepairStatGet)))

	r.Post("/user_stat", WithPassportSecret(webhookSecret, WithError(c.UserStatGet)))
	r.Post("/faction_stat", WithPassportSecret(webhookSecret, WithError(c.FactionStatGet)))
	r.Post("/faction_contract_reward", WithPassportSecret(webhookSecret, WithError(c.FactionContractRewardGet)))

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
		user.Faction = pc.API.factionMap[user.FactionID]
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

	return helpers.EncodeJSON(w, true)
}

type UserEnlistFactionRequest struct {
	UserID    server.UserID    `json:"userID"`
	FactionID server.FactionID `json:"factionID"`
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

	user, err := pc.API.UserMap.GetUserDetailByID(req.UserID)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "User not found")
	}

	// swap from non faction to faction
	pc.API.ViewerLiveCount.Swap(user.FactionID, req.FactionID)

	// update user faction
	user.FactionID = req.FactionID
	user.Faction = pc.API.factionMap[user.FactionID]

	broadcastData, err := json.Marshal(&BroadcastPayload{
		Key:     HubKeyUserSubscribe,
		Payload: user,
	})
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	// broadcast to all the client
	for _, cl := range pc.API.UserMap.Update(user) {
		go cl.Send(broadcastData)
	}

	return helpers.EncodeJSON(w, true)
}

type WarMachineJoinRequest struct {
	WarMachineMetadata *server.WarMachineMetadata `json:"warMachineMetadata"`
	NeedInsured        bool                       `json:"needInsured"`
}

func (pc *PassportWebhookController) WarMachineJoin(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &WarMachineJoinRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	if req.WarMachineMetadata.FactionID.IsNil() {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("Non-faction war machine is not able to join"))
	}

	err = pc.API.BattleArena.WarMachineQueue.Join(req.WarMachineMetadata, req.NeedInsured)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, err.Error())
	}

	// set insurance flag
	userWarMachinePostion, err := pc.API.BattleArena.WarMachineQueue.GetUserWarMachineQueue(req.WarMachineMetadata.FactionID, req.WarMachineMetadata.OwnedByID)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	// return current queuing position
	return helpers.EncodeJSON(w, userWarMachinePostion)
}

type UserSupsMultiplierGetRequest struct {
	UserID server.UserID `json:"userID"`
}

// UserSupsMultiplierGet return the sups multiplier of the given user
func (pc *PassportWebhookController) UserSupsMultiplierGet(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &UserSupsMultiplierGetRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		fmt.Println(err)
		return http.StatusInternalServerError, terror.Error(err)
	}

	return helpers.EncodeJSON(w, pc.API.UserMultiplier.UserMultiplierGet(req.UserID))
}

type UserStatGetRequest struct {
	UserID server.UserID `json:"userID"`
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

	userStat, err := db.UserStatGet(r.Context(), pc.Conn, req.UserID)
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
	FactionID server.FactionID `json:"factionID"`
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

	err = db.FactionStatGet(r.Context(), pc.Conn, factionStat)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, fmt.Sprintf("Failed to get faction %s stat", req.FactionID))
	}

	return helpers.EncodeJSON(w, factionStat)
}

type FactionContractRewardGetRequest struct {
	FactionID server.FactionID `json:"factionID"`
}

func (pc *PassportWebhookController) FactionContractRewardGet(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &FactionContractRewardGetRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	if req.FactionID.IsNil() || !req.FactionID.IsValid() {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("faction id is empty"), "Faction id is required")
	}

	contractReward := "0"
	switch req.FactionID {
	case server.RedMountainFactionID:
		contractReward = pc.API.BattleArena.WarMachineQueue.RedMountain.GetContractReward()
	case server.BostonCyberneticsFactionID:
		contractReward = pc.API.BattleArena.WarMachineQueue.Boston.GetContractReward()
	case server.ZaibatsuFactionID:
		contractReward = pc.API.BattleArena.WarMachineQueue.Zaibatsu.GetContractReward()

	}
	return helpers.EncodeJSON(w, contractReward)
}

type WarMachineQueuePositionRequest struct {
	AssetHash string `json:"assethash"`
}

// PassportWarMachineQueuePositionHandler return the list of user's war machines in the queue
func (pc *PassportWebhookController) WarMachineQueuePositionGet(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &WarMachineQueuePositionRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	userWarMachinePosition, err := pc.API.BattleArena.WarMachineQueue.GetUserWarMachineQueue(req.FactionID, req.UserID)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	// get in game war machine
	for _, wm := range pc.API.BattleArena.InGameWarMachines() {
		if wm.OwnedByID != req.UserID {
			continue
		}
		userWarMachinePosition = append(userWarMachinePosition, &passport.WarMachineQueuePosition{
			WarMachineMetadata: wm,
			Position:           -1,
		})
	}

	return helpers.EncodeJSON(w, userWarMachinePosition)
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

	err = db.AssetRepairIncompleteGet(r.Context(), pc.Conn, record)
	if err != nil {
		if err == pgx.ErrNoRows {
			return helpers.EncodeJSON(w, &server.AssetRepairRecord{})
		}

		return http.StatusInternalServerError, terror.Error(err)
	}

	return helpers.EncodeJSON(w, record)
}

type AuthRingCheckRequest struct {
	User                *server.User  `json:"user"`
	SessionID           hub.SessionID `json:"sessionID"`
	GameserverSessionID string        `json:"gameserverSessionID"`
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

	client, err := pc.API.RingCheckAuthMap.Check(req.GameserverSessionID)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Hub client not found")
	}

	// set client identifier
	client.SetIdentifier(req.User.ID.String())

	// update user faction if user has a faction
	if !req.User.FactionID.IsNil() {
		pc.API.ViewerLiveCount.Swap(server.FactionID(uuid.Nil), req.User.FactionID)
		req.User.Faction = pc.API.factionMap[req.User.FactionID]
	}

	// register user detail to user map
	pc.API.UserMap.UserRegister(client, req.User)

	// set client multipliers
	pc.API.UserMultiplier.Online(req.User.ID)

	b, err := json.Marshal(&BroadcastPayload{
		Key:     HubKeyUserSubscribe,
		Payload: req.User,
	})
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	go client.Send(b)

	return helpers.EncodeJSON(w, true)
}
