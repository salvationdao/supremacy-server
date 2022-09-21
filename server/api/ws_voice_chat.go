package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/voice_chat"
	"time"

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

	signedURL, err := voice_chat.GetSignedPolicyURL(user.ID)
	if err != nil {
		return terror.Error(err, "Failed to get signed policy url")
	}

	// create one if there is no faction commander
	newFactionCommander := &boiler.VoiceStream{
		ArenaID:         arena.ID,
		OwnerID:         user.ID,
		FactionID:       factionID,
		ListenStreamURL: signedURL.ListenURL,
		SendStreamURL:   signedURL.SendURL,
		IsActive:        false,
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
		boiler.VoiceStreamWhere.HasVoted.EQ(false),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "failed to get active user")
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

	count, err := boiler.VoiceStreams(
		boiler.VoiceStreamWhere.IsActive.EQ(true),
		boiler.VoiceStreamWhere.FactionID.EQ(factionID),
		boiler.VoiceStreamWhere.KickedAt.IsNull(),
	).Count(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "failed to get voice stream count")
	}
	factionCommander.CurrentKickVote += 1
	if factionCommander.CurrentKickVote >= int(count) {
		factionCommander.KickedAt = null.TimeFrom(time.Now())
		factionCommander.IsActive = false
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

	arena, err := api.ArenaManager.GetArena(req.Payload.ArenaID)
	if err != nil {
		return err
	}

	err = voice_chat.UpdateFactionVoiceChannel(factionID, arena.ID)
	if err != nil {
		return terror.Error(err, "failed to update faction voice channel")
	}

	reply(true)

	return nil
}
