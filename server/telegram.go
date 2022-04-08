package server

import "server/db/boiler"

type Telegram interface {
	PlayerCreate(player *boiler.Player) (*boiler.TelegramPlayer, error)
	Notify(id string, message string) error
	Notify2(telegramID int64, message string) error
}
