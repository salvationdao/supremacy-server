package server

import "server/db/boiler"

type Telegram interface {
	PreferencesUpdate(playerID string) (*boiler.PlayerSettingsPreference, error)
	NotifyDEPRECATED(id string, message string) error
	Notify(telegramID int64, message string) error
}
