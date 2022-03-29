package server

import "server/db/boiler"

type Telegram interface {
	Notify(playerID string, mechID string, message string) error
	NotificationCreate(playerID string, mechID string) (*boiler.TelegramNotification, error)

	List()
	Insert()
}
