package stormdb

import (
	"fmt"
	"math/rand"

	"github.com/asdine/storm/v3"
	"github.com/ninja-software/terror/v2"
)

type StormDB struct {
	db *storm.DB
}

type TelegramNotification struct {
	ID            int `storm:"id,increment"`
	PlayerID      string
	TelegramID    int
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
		TelegramID: 0,
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

// TelegramNotificationRegister if code entered correctly sets notification registered status to true and sets notification's telegram id
func (s *StormDB) TelegramNotificationRegister(shortCode string, mechID string, telegramID int) (*TelegramNotification, error) {

	// find notification based on user id mech id and code
	notif := &TelegramNotification{}
	err := s.db.One("Shortcode", shortCode, notif)
	if err != nil {
		return nil, terror.Error(err)
	}

	if notif == nil {
		// TODO: send error message to telegram
		return nil, terror.Error(err, "shortcode not found")
	}

	notif.Registered = true
	notif.TelegramID = telegramID

	err = s.db.Save(notif)
	if err != nil {
		return nil, terror.Error(err)
	}

	return notif, nil

}

func (s *StormDB) TelegramNotificationsList() ([]*TelegramNotification, error) {

	// find notif based on user id mech id and code
	notifs := []*TelegramNotification{}
	err := s.db.All(&notifs)
	if err != nil {
		return nil, terror.Error(err)
	}

	if notifs == nil {
		return nil, terror.Error(err, "no notifications found")
	}

	return notifs, nil

}

func genCode() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz")
	n := 5
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (s *StormDB) TelegramGenCodeTest() (*TelegramNotification, error) {

	// gen shortcode
	code := genCode()

	teleNotif := &TelegramNotification{
		PlayerID:   "1a657a32-778e-4612-8cc1-14e360665f2b",
		Shortcode:  code,
		MechID:     "fc43fa34-b23f-40f4-afaa-465f4880ef59",
		Registered: false,
	}

	err := s.db.Save(teleNotif)
	if err != nil {
		return nil, terror.Error(err)
	}

	return teleNotif, nil

}

func (s *StormDB) TelegramGetNotificationByPlayerCode(playerID string, mechID string) (*TelegramNotification, error) {
	// find notif based on user id mech id
	notif := &TelegramNotification{}
	err := s.db.One("MechID", mechID, notif)
	if err != nil {
		return nil, terror.Error(err)
	}

	err = s.db.One("PlayerID", playerID, notif)
	if err != nil {
		return nil, terror.Error(err)
	}

	return notif, nil

}

// todo endpoint to generate shortcode
// handle in front end

// deploy mech
// - gen shortcode
// - save that in storm (telegram_notifs)
// - set registered false
// - with users + mech details

// go to telegram
// - user send tele: register
// - reply tele: type code
// - query storm select * from telegram_notifs where shortcode = s
// - set registered true
// - reply tele: notification set for {mech} {queue number}
// - save that in storm

// on notify
// get user info from storm
// send tele mesage

// telegram_notifs
// shortcode
// payer_id
// mech_id
// message
// registered
