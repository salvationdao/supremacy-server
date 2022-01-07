package battle_arena

import (
	"context"
	"server"
	"sync"
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
	EventAnamationEnd              Event = "ANIMATION_END"
	EventWarMachinePositionChanged Event = "WAR_MACHINE_POSITION_CHANGED"
	EventWarMachineDestroyed       Event = "WAR_MACHINE_DESTROYED"
)

type EventData struct {
	BattleArena         *server.Battle
	WarMachineDestroyed *server.WarMachineDestroyed
	FactionActions      []*server.FactionAction
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
