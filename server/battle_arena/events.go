package battle_arena

import (
	"context"
	"errors"
	"net/http"
	"server"
	"server/db"
	"server/helpers"
	"strconv"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
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
	EventWarMachineQueueUpdated    Event = "WAR_MACHINE_QUEUE_UPDATED"
	EventWarMachineDestroyed       Event = "WAR_MACHINE_DESTROYED"
	EventFactionViewersGet         Event = "FACTION_VIEWERS_GET"
	EventWarMachinePositionChanged Event = "WAR_MACHINE_POSITION_CHANGED"
)

type EventData struct {
	BattleArena               *server.Battle
	FactionAbilities          []*server.GameAbility
	WarMachineLocation        []byte `json:"warMachineLocation"`
	WinnerFactionViewers      *WinnerFactionViewer
	BattleRewardList          *BattleRewardList
	WarMachineDestroyedRecord *server.WarMachineDestroyedRecord
	WarMachineQueue           *WarMachineQueueUpdateEvent
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
	sync.RWMutex
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
