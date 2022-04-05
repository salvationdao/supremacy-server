package server

import "server/db/boiler"

type Telegram interface {
	Notify(id string, message string) error
	NotificationCreate(mechID string, notification *boiler.BattleQueueNotification) (*boiler.TelegramNotification, error)
}
