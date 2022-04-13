package server

import "server/db/boiler"

type Telegram interface {
	ProfileUpdate(playerID string) (*boiler.PlayerProfile, error)
	Notify(id string, message string) error
	Notify2(telegramID int64, message string) error
}
