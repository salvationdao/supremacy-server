package passport

import (
	"context"
	"sync"
)

/**************

Battle arena Events are events that the hub client needs to know about.

As an example...
	ba.Events.Trigger(ctx, EventGameStart, &EventData{EventType: EventGameStart, Passport: ba.battle})
This calls the passed the EventGameStart event with the battle arena data through the channel.

To handle this event you attached an event handler like below
	api.Passport.Events.AddEventHandler(passport.EventGameStart, api.BattleStartSignal)

*************/

type Event string
type EventHandler func(ctx context.Context, payload interface{})

type Events struct {
	events map[Event][]EventHandler
	sync.RWMutex
}

func (ev *Events) AddEventHandler(event Event, handler EventHandler) {
	ev.Lock()
	ev.events[event] = append(ev.events[event], handler)
	ev.Unlock()
}

func (ev *Events) Trigger(ctx context.Context, event Event, payload interface{}) {
	go func() {
		ev.RLock()
		for _, fn := range ev.events[event] {
			fn(ctx, payload)
		}
		ev.RUnlock()
	}()
}

func (ev *Events) TriggerMany(ctx context.Context, event Event, payload interface{}) {
	go func() {
		ev.RLock()
		for _, fn := range ev.events[event] {
			fn(ctx, payload)
		}
		ev.RUnlock()
	}()
}
