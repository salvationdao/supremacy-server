package battle_arena

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"server"
	"server/db"
	"server/helpers"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx"
	"github.com/ninja-software/terror/v2"
	"github.com/sasha-s/go-deadlock"
)

/**************

Battle arena Events are events that the hub client needs to know about.

As an example...
	ba.Events.Trigger(ctx, EventGameStart, &EventData{EventType: EventGameStart, BattleArena: ba.battle})
This calls the passed the EventGameStart event with the battle arena data through the channel.

To handle this event you attached an event handler like below
	api.BattleArena.Events.AddEventHandler(battle_arena.EventGameStart, api.BattleStartSignal)

*************/

type Event string
type EventHandler func(ctx context.Context, ed *EventData)

const (
	EventGameInit                  Event = "GAME_INIT"
	EventGameStart                 Event = "GAME_START"
	EventGameEnd                   Event = "GAME_END"
	EventWarMachineDestroyed       Event = "WAR_MACHINE_DESTROYED"
	EventFactionViewersGet         Event = "FACTION_VIEWERS_GET"
	EventWarMachinePositionChanged Event = "WAR_MACHINE_POSITION_CHANGED"
	EventAISpawned                 Event = "AI_SPAWNED"
)

type EventData struct {
	BattleArena               *server.Battle
	FactionAbilities          []*server.GameAbility
	WarMachineLocation        []byte `json:"warMachineLocation"`
	WinnerFactionViewers      *WinnerFactionViewer
	BattleRewardList          *BattleRewardList
	WarMachineDestroyedRecord *server.WarMachineDestroyedRecord
	WarMachineQueue           *WarMachineQueueUpdateEvent
	SpawnedAI                 *server.WarMachineMetadata
}

type WarMachineQueueUpdateEvent struct {
	FactionID   server.FactionID
	WarMachines []*server.WarMachineMetadata
}

type WinnerFactionViewer struct {
	WinnerFactionID server.FactionID
	CallbackChannel chan []server.UserID
}

type BattleArenaEvents struct {
	events map[Event][]EventHandler
	deadlock.RWMutex
}

func (ev *BattleArenaEvents) AddEventHandler(event Event, handler EventHandler) {
	ev.Lock()
	ev.events[event] = append(ev.events[event], handler)
	ev.Unlock()
}

func (ev *BattleArenaEvents) Trigger(ctx context.Context, event Event, ed *EventData) {
	go func() {
		ev.RLock()
		for _, fn := range ev.events[event] {
			fn(ctx, ed)
		}
		ev.RUnlock()
	}()
}

func (ev *BattleArenaEvents) TriggerMany(ctx context.Context, event Event, ed *EventData) {
	go func() {
		ev.RLock()
		for _, fn := range ev.events[event] {
			fn(ctx, ed)
		}
		ev.RUnlock()
	}()
}

func (ba *BattleArena) FactionStats(w http.ResponseWriter, r *http.Request) (int, error) {
	ctx := context.Background()
	result, err := db.FactionStatAll(ctx, ba.Conn)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err)
	}
	return helpers.EncodeJSON(w, result)
}
func (ba *BattleArena) UserStats(w http.ResponseWriter, r *http.Request) (int, error) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		return http.StatusBadRequest, errors.New("user_id not provided")
	}
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err)
	}

	ctx := context.Background()
	result, err := db.UserStatGet(ctx, ba.Conn, server.UserID(userID))
	if err != nil {
		return http.StatusBadRequest, terror.Error(err)
	}
	return helpers.EncodeJSON(w, result)
}

func (ba *BattleArena) GetBattleQueue(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Header.Get("X-Authorization") != "9b3c60af-035c-4f64-9450-a7dd7cbe2297" {
		return http.StatusForbidden, errors.New("unauthorised")
	}

	// X-Authorization:
	result := &struct {
		Zaibatsu    []*server.WarMachineMetadata `json:"zaibatsu"`
		RedMountain []*server.WarMachineMetadata `json:"red_mountain"`
		Boston      []*server.WarMachineMetadata `json:"boston"`
	}{
		Zaibatsu:    ba.WarMachineQueue.Zaibatsu.QueuingWarMachines,
		RedMountain: ba.WarMachineQueue.RedMountain.QueuingWarMachines,
		Boston:      ba.WarMachineQueue.Boston.QueuingWarMachines,
	}

	return helpers.EncodeJSON(w, result)
}

func (ba *BattleArena) GetEvents(w http.ResponseWriter, r *http.Request) (int, error) {
	ctx := context.Background()
	sinceStr := r.URL.Query().Get("since")
	if sinceStr == "" {
		events, err := db.GetEvents(ctx, ba.Conn, nil)
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err)
		}
		return helpers.EncodeJSON(w, events)
	}
	since, err := strconv.Atoi(sinceStr)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}
	t := time.Unix(int64(since), 0)
	events, err := db.GetEvents(ctx, ba.Conn, &t)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	return helpers.EncodeJSON(w, events)
}

func (ba *BattleArena) GetAbility(w http.ResponseWriter, r *http.Request) (int, error) {
	ctx := context.Background()
	abilityIDStr := r.URL.Query().Get("id")
	if abilityIDStr == "" {
		return http.StatusBadRequest, errors.New("id not provided")
	}
	abilityID, err := uuid.FromString(abilityIDStr)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err)
	}
	result, err := db.GameAbility(ctx, ba.Conn, server.GameAbilityID(abilityID))
	if err != nil {
		return http.StatusBadRequest, terror.Error(err)
	}
	return helpers.EncodeJSON(w, result)
}

func (ba *BattleArena) GetBlob(w http.ResponseWriter, r *http.Request) (int, error) {
	ctx := context.Background()
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		return http.StatusBadRequest, terror.Error(terror.ErrInvalidInput, "no id provided")
	}
	id, err := uuid.FromString(idStr)
	if err != nil {
		return http.StatusBadRequest, terror.Error(terror.ErrInvalidInput, "invalid id provided")
	}
	blobID := server.BlobID(id)

	att := &server.Blob{}
	err = db.FindBlob(ctx, ba.Conn, att, blobID)
	if errors.Is(err, pgx.ErrNoRows) {
		return http.StatusNotFound, terror.Error(err, "attachment not found")
	}
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "could not get attachment")
	}

	disposition := "attachment"
	isViewDisposition := r.URL.Query().Get("view")
	if isViewDisposition == "true" {
		disposition = "inline"
	}

	// tell the browser the returned content should be downloaded/inline
	if att.MimeType != "" && att.MimeType != "unknown" {
		w.Header().Add("Content-Type", att.MimeType)
	}
	w.Header().Add("Content-Disposition", fmt.Sprintf("%s;filename=%s", disposition, att.FileName))
	return w.Write(att.File)
}
