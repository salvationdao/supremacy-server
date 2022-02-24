package passport

import (
	"context"
	"sync"
)

/**************

Battle arena Events are events that the hub client needs to know about.

As an example...
	pp.Events.Trigger(ctx, EventGameStart, &EventData{EventType: EventGameStart, Passport: ba.battle})
This calls the passed the EventGameStart event with the battle arena data through the channel.

To handle this event you attached an event handler like below
	api.Passport.Events.AddEventHandler(passport.EventGameStart, api.BattleStartSignal)

*************/

type Event string

const (
	EventAuthed                     Event = "PASSPORT:AUTHED"
	EventUserOnlineStatus           Event = "PASSPORT:USER:ONLINE_STATUS"
	EventUserUpdated                Event = "PASSPORT:USER:UPDATED"
	EventUserEnlistFaction          Event = "PASSPORT:USER:ENLIST:FACTION"
	EventBattleQueueJoin            Event = "PASSPORT:ASSET:QUEUE:JOIN"
	EventBattleQueueLeave           Event = "PASSPORT:ASSET:QUEUE:LEAVE"
	EventWarMachineQueuePositionGet Event = "PASSPORT:WAR:MACHINE:QUEUE:POSITION:GET"
	EventAssetInsurancePay          Event = "PASSPORT:ASSET:INSURANCE:PAY"
	EventFactionStatGet             Event = "PASSPORT:FACTION:STAT:GET"
	EventAuthRingCheck              Event = "PASSPORT:AUTH:RING:CHECK"
	EventUserSupsMultiplierGet      Event = "PASSPORT:USER:SUPS:MULTIPLIER:GET"
	EventUserStatGet                Event = "PASSPORT:USER:STAT:GET"
)

type EventHandler func(ctx context.Context, payload []byte)

type Events struct {
	events map[Event][]EventHandler
	sync.RWMutex
}

func (ev *Events) AddEventHandler(event Event, handler EventHandler) {
	ev.Lock()
	ev.events[event] = append(ev.events[event], handler)
	ev.Unlock()
}

func (ev *Events) Trigger(event Event, payload []byte) {
	go func() {
		ev.RLock()
		for _, fn := range ev.events[event] {
			fn(context.Background(), payload)
		}
		ev.RUnlock()
	}()
}

func (ev *Events) TriggerMany(ctx context.Context, event Event, payload []byte) {
	go func() {
		ev.RLock()
		for _, fn := range ev.events[event] {
			fn(ctx, payload)
		}
		ev.RUnlock()
	}()
}
