package server

import "server/db/boiler"

type Telegram interface {
	Notify(mechID string, message string) error
	NotificationCreate(mechID string) (*boiler.TelegramNotification, error)
}
