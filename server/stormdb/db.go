package stormdb

import (
	"fmt"

	"github.com/asdine/storm/v3"
	"github.com/ninja-software/terror/v2"
)

type StormDB struct {
	db *storm.DB
}

type TelegramNotification struct {
	ID            int `storm:"id,increment"`
	PlayerID      string
	MechID        string
	Shortcode     string
	QueuePosition int
	Registered    bool
}

// Initializes DB file. Opens existing or creates new
func NewStormDB(dbFile string) (*StormDB, error) {
	db, err := storm.Open(dbFile)
	if err != nil {
		return nil, err
	}
	err = db.Init(&TelegramNotification{})
	if err != nil {
		fmt.Println("Fail to initialise TelegramNotifications")
		return nil, err
	}

	database := &StormDB{db: db}

	return database, err
}

func (s *StormDB) TelegramNotificationCreate(playerID string, mechID string, shortCode string) error {

	teleNotif := &TelegramNotification{
		PlayerID:   playerID,
		Shortcode:  shortCode,
		MechID:     mechID,
		Registered: false,
	}

	err := s.db.Save(teleNotif)
	if err != nil {
		return terror.Error(err)
	}

	return nil

}

func (s *StormDB) TelegramNotificationRegister(shortCode string) (*TelegramNotification, error) {

	// find notif based on user id mech id and code
	notif := &TelegramNotification{}
	err := s.db.One("Shortcode", shortCode, notif)
	if err != nil {
		return nil, terror.Error(err)
	}

	notif.Registered = true

	err = s.db.Save(notif)
	if err != nil {
		return nil, terror.Error(err)
	}

	return notif, nil

}
