package server

type Stream struct {
	Host          string  `json:"host" db:"host"`
	Name          string  `json:"name" db:"name"`
	StreamID      string  `json:"stream_id" db:"stream_id"`
	URL           string  `json:"url" db:"url"`
	Region        string  `json:"region" db:"region"`
	Resolution    string  `json:"resolution" db:"resolution"`
	BitRatesKBits int     `json:"bit_rates_kbits" db:"bit_rates_k_bits"`
	UserMax       int     `json:"user_max" db:"user_max"`
	UsersNow      int     `json:"users_now" db:"users_now"`
	Active        bool    `json:"active" db:"active"`
	Status        string  `json:"status" db:"status"`
	Latitude      float32 `json:"latitude" db:"latitude"`
	Longitude     float32 `json:"longitude" db:"longitude"`
}

type GamesToCloseStream struct {
	GamesToClose int `json:"games_to_close"`
}

type GlobalAnnouncement struct {
	ID                    string `json:"id"`
	Title                 string `json:"title"`
	Message               string `json:"message"`
	ShowFromBattleNumber  *int   `json:"show_from_battle_number,omitempty"`  // the battle number this announcement wil show
	ShowUntilBattleNumber *int   `json:"show_until_battle_number,omitempty"` // the battle number this announcement will be deleted
}
