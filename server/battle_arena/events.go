package battle_arena

import (
	"context"
	"net/http"
	"server"
	"server/db"
	"server/helpers"
	"sync"

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
	WarMachines []*server.WarMachineNFT
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
	events, err := db.GetEvents(ctx, ba.Conn, nil)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	return helpers.EncodeJSON(w, events)
}
