package server

import "server/db/boiler"

type Telegram interface {
	PreferencesUpdate(playerID string) (*boiler.PlayerSettingsPreference, error)
	Notify(id string, message string) error
	Notify2(telegramID int64, message string) error
}
