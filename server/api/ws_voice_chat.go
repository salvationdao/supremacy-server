package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/voice_chat"
	"time"

	"github.com/sasha-s/go-deadlock"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func NewVoiceStreamController(api *API) {
	api.SecureUserFactionCommand(server.HubKeyVoiceStreamJoinFactionCommander, api.JoinFactionCommander)
	api.SecureUserFactionCommand(server.HubKeyVoiceStreamLeaveFactionCommander, api.LeaveFactionCommander)
	api.SecureUserFactionCommand(server.HubKeyVoiceStreamVoteKick, api.VoteKickFactionCommander)

	api.SecureUserFactionCommand(server.HubKeyVoiceStreamConnect, api.VoiceChatConnect)
	api.SecureUserFactionCommand(server.HubKeyVoiceStreamDisconnect, api.VoiceChatDisconnect)

}

func (api *API) VoiceStreamSubscribe(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	if !user.FactionID.Valid {
		return fmt.Errorf("faction id not found")
	}

	arenaID := chi.RouteContext(ctx).URLParam("arena_id")
	if arenaID == "" {
		return terror.Error(fmt.Errorf("missing arena id"), "Missing arena id")
	}

	rvs, err := db.GetActiveVoiceChat(user.ID, user.FactionID.String, arenaID)
	if err != nil {
		gamelog.L.Error().Str("user_id", user.ID).Err(err).Msg("failed to get active voice chats")
	}

	reply(rvs)

	return nil
}

func (api *API) VoiceStreamListenersSubscribe(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	if !user.FactionID.Valid {
		return fmt.Errorf("faction id not found")
	}

	arenaID := chi.RouteContext(ctx).URLParam("arena_id")
	if arenaID == "" {
		return terror.Error(fmt.Errorf("missing arena id"), "Missing arena id")
	}

	resp := api.VoiceChatListeners.CurrentVoiceChatListeners()
	reply(resp)

	return nil
}

type VoiceStreamReq struct {
	Payload struct {
		ArenaID string `json:"arena_id"`
	} `json:"payload"`
}

func (api *API) JoinFactionCommander(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &VoiceStreamReq{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	arena, err := api.ArenaManager.GetArena(req.Payload.ArenaID)
	if err != nil {
		return err
	}

	arena.VoiceChannel.Lock()
	defer arena.VoiceChannel.Unlock()

	// check if there is a faction commander
	_, err = boiler.VoiceStreams(
		boiler.VoiceStreamWhere.IsActive.EQ(true),
		boiler.VoiceStreamWhere.SenderType.EQ(boiler.VoiceSenderTypeFACTION_COMMANDER),
		boiler.VoiceStreamWhere.ArenaID.EQ(arena.ID),
		boiler.VoiceStreamWhere.FactionID.EQ(factionID),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to get faction commander voice stream")
	}

	// check if user has been kicked before
	oldExist, err := boiler.VoiceStreams(
		boiler.VoiceStreamWhere.OwnerID.EQ(user.ID),
		boiler.VoiceStreamWhere.KickedAt.IsNotNull(),
		qm.OrderBy(fmt.Sprintf("%s DESC", boiler.VoiceStreamColumns.KickedAt)),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "failed to get faction commander with user id")
	}

	if oldExist != nil {
		if oldExist.KickedAt.Valid {
			banTimeHour := db.GetIntWithDefault(db.KeyVoiceBanTimeHours, 24)
			oldExist.KickedAt.Time.Add(time.Hour * time.Duration(int64(banTimeHour)))

			if oldExist.KickedAt.Time.Before(time.Now()) {
				return terror.Error(fmt.Errorf("you've been voted to be banned for 24 hours"), "You've been voted to be banned for 24 hours")
			}
		}
	}

	// check if already a faction commander
	currentFactionCommander, err := boiler.VoiceStreams(
		boiler.VoiceStreamWhere.KickedAt.IsNull(),
		boiler.VoiceStreamWhere.SenderType.EQ(boiler.VoiceSenderTypeFACTION_COMMANDER),
		boiler.VoiceStreamWhere.IsActive.EQ(true),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "failed to get faction commander")
	}

	if currentFactionCommander != nil {
		return terror.Error(fmt.Errorf("there is already an active faction commmander."), "there is already an active faction commmander.")
	}

	signedURL, err := voice_chat.GetSignedPolicyURL(user.ID)
	if err != nil {
		return terror.Error(err, "Failed to get signed policy url")
	}

	// if user is a mech owner
	mechOwner, err := boiler.VoiceStreams(
		boiler.VoiceStreamWhere.OwnerID.EQ(user.ID),
		boiler.VoiceStreamWhere.KickedAt.IsNull(),
		boiler.VoiceStreamWhere.SenderType.EQ(boiler.VoiceSenderTypeMECH_OWNER),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "failed to get faction commander with user id")
	}

	listenURL := signedURL.ListenURL
	sendURL := signedURL.SendURL
	if mechOwner != nil {

		listenURL = mechOwner.ListenStreamURL
		sendURL = mechOwner.SendStreamURL

	}

	// create one if there is no faction commander
	newFactionCommander := &boiler.VoiceStream{
		ArenaID:         arena.ID,
		OwnerID:         user.ID,
		FactionID:       factionID,
		ListenStreamURL: listenURL,
		SendStreamURL:   sendURL,
		IsActive:        true,
		SenderType:      boiler.VoiceSenderTypeFACTION_COMMANDER,
		SessionExpireAt: signedURL.ExpiredAt,
	}

	err = newFactionCommander.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "failed to create new faction commander")
	}

	err = voice_chat.UpdateFactionVoiceChannel(factionID, arena.ID)
	if err != nil {
		return terror.Error(err, "failed to update faction voice channel")
	}

	reply(true)

	return nil
}

func (api *API) LeaveFactionCommander(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &VoiceStreamReq{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	arena, err := api.ArenaManager.GetArena(req.Payload.ArenaID)
	if err != nil {
		return err
	}

	arena.VoiceChannel.Lock()
	defer arena.VoiceChannel.Unlock()

	// check if there is a faction commander
	activeVoiceCommander, err := boiler.VoiceStreams(
		boiler.VoiceStreamWhere.IsActive.EQ(true),
		boiler.VoiceStreamWhere.SenderType.EQ(boiler.VoiceSenderTypeFACTION_COMMANDER),
		boiler.VoiceStreamWhere.ArenaID.EQ(arena.ID),
		boiler.VoiceStreamWhere.FactionID.EQ(factionID),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to get faction commander voice stream")
	}

	activeVoiceCommander.IsActive = false
	_, err = activeVoiceCommander.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "failed to update active voice commander")
	}

	err = voice_chat.UpdateFactionVoiceChannel(factionID, arena.ID)
	if err != nil {
		return terror.Error(err, "failed to update faction voice channel")
	}

	reply(true)

	return nil
}

func (api *API) VoteKickFactionCommander(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &VoiceStreamReq{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	activeUser, err := boiler.VoiceStreams(
		boiler.VoiceStreamWhere.OwnerID.EQ(user.ID),
		boiler.VoiceStreamWhere.IsActive.EQ(true),
		boiler.VoiceStreamWhere.SenderType.EQ(boiler.VoiceSenderTypeMECH_OWNER),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "failed to get active user")
	}

	// if user already voted
	if activeUser != nil && activeUser.HasVoted {
		return terror.Error(fmt.Errorf("cannot vote to kick faction commander multiple times"), "cannot vote to kick faction commander multiple times")
	}

	factionCommander, err := boiler.VoiceStreams(
		boiler.VoiceStreamWhere.IsActive.EQ(true),
		boiler.VoiceStreamWhere.SenderType.EQ(boiler.VoiceSenderTypeFACTION_COMMANDER),
		boiler.VoiceStreamWhere.FactionID.EQ(factionID),
		boiler.VoiceStreamWhere.KickedAt.IsNull(),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "failed to find faction commander")
	}

	// active voice streams excluding faction commander
	count, err := boiler.VoiceStreams(
		boiler.VoiceStreamWhere.IsActive.EQ(true),
		boiler.VoiceStreamWhere.FactionID.EQ(factionID),
		boiler.VoiceStreamWhere.KickedAt.IsNull(),
		boiler.VoiceStreamWhere.SenderType.NEQ(boiler.VoiceSenderTypeFACTION_COMMANDER),
		qm.GroupBy(boiler.VoiceStreamColumns.OwnerID),
	).Count(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "failed to get voice stream count")
	}

	arena, err := api.ArenaManager.GetArena(req.Payload.ArenaID)
	if err != nil {
		return err
	}

	factionCommander.CurrentKickVote += 1
	if factionCommander.CurrentKickVote >= int(count) {
		factionCommander.KickedAt = null.TimeFrom(time.Now())
		factionCommander.IsActive = false

		// reset has voted
		_, err := boiler.VoiceStreams(
			boiler.VoiceStreamWhere.FactionID.EQ(factionID),
			boiler.VoiceStreamWhere.IsActive.EQ(true),
			boiler.VoiceStreamWhere.ArenaID.EQ(arena.ID),
		).UpdateAll(gamedb.StdConn, boiler.M{boiler.VoiceStreamColumns.HasVoted: false})
		if err != nil {
			return terror.Error(err, "Failed to update active voice streams has_voted status")
		}
	}

	_, err = factionCommander.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "failed to update faction commander")
	}

	activeUser.HasVoted = true

	_, err = activeUser.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "failed to update active user")
	}

	err = voice_chat.UpdateFactionVoiceChannel(factionID, arena.ID)
	if err != nil {
		return terror.Error(err, "failed to update faction voice channel")
	}

	reply(true)

	return nil
}

type VoiceChatListeners struct {
	Listeners []*server.PublicPlayer
	API       *API
	deadlock.RWMutex
}

func NewVoiceChatListeners() *VoiceChatListeners {
	vcl := &VoiceChatListeners{}

	return vcl

}

// CurrentVoiceChatListeners return a copy of current voice stream listeners
func (vcl *VoiceChatListeners) CurrentVoiceChatListeners() []*server.PublicPlayer {
	vcl.RLock()
	defer vcl.RUnlock()

	return vcl.Listeners
}

func (vcl *VoiceChatListeners) AddListener(newListener *server.PublicPlayer) {
	vcl.Lock()
	found := false
	for _, l := range vcl.Listeners {
		if l.ID == newListener.ID {
			found = true
		}
	}

	if !found {
		vcl.Listeners = append(vcl.Listeners, newListener)
	}
	vcl.Unlock()
}

func (vcl *VoiceChatListeners) RemoveListener(listenerID string) {
	vcl.Lock()

	newSlice := []*server.PublicPlayer{}
	for _, v := range vcl.Listeners {
		if v.ID != listenerID {
			newSlice = append(newSlice, v)
		}
	}

	vcl.Listeners = newSlice
	vcl.Unlock()

}

func (api *API) VoiceChatConnect(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &VoiceStreamReq{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	p := &server.PublicPlayer{
		ID:        user.ID,
		Username:  user.Username,
		Gid:       user.Gid,
		FactionID: user.FactionID,
	}

	// add
	api.VoiceChatListeners.AddListener(p)

	listeners := api.VoiceChatListeners.CurrentVoiceChatListeners()
	reply(true)

	err = voice_chat.UpdateFactionVoiceStreamListeners(user.FactionID.String, req.Payload.ArenaID, listeners)
	if err != nil {
		return terror.Error(err, "failed to update voice stream listeners")
	}

	return nil
}

func (api *API) VoiceChatDisconnect(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &VoiceStreamReq{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// remove
	api.VoiceChatListeners.RemoveListener(user.ID)

	listeners := api.VoiceChatListeners.CurrentVoiceChatListeners()

	err = voice_chat.UpdateFactionVoiceStreamListeners(user.FactionID.String, req.Payload.ArenaID, listeners)
	if err != nil {
		return terror.Error(err, "failed to update voice stream listeners")
	}

	reply(true)
	return nil
}
