package server

import "time"

type Stream struct {
	Host          string  `json:"host" db:"host"`
	Name          string  `json:"name" db:"name"`
	StreamID      string  `json:"streamID" db:"stream_id"`
	URL           string  `json:"url" db:"url"`
	Region        string  `json:"region" db:"region"`
	Resolution    string  `json:"resolution" db:"resolution"`
	BitRatesKBits int     `json:"bitRatesKBits" db:"bit_rates_k_bits"`
	UserMax       int     `json:"userMax" db:"user_max"`
	UsersNow      int     `json:"usersNow" db:"users_now"`
	Active        bool    `json:"active" db:"active"`
	Status        string  `json:"status" db:"status"`
	Latitude      float32 `json:"latitude" db:"latitude"`
	Longitude     float32 `json:"longitude" db:"longitude"`
}

type GamesToCloseStream struct {
	GamesToClose int `json:"gamesToClose"`
}

type GlobalAnnouncement struct {
	Title      string     `json:"title"`
	Message    string     `json:"message"`
	GamesUntil *int       `json:"gamesUntil,omitempty"`
	ShowUntil  *time.Time `json:"showUntil,omitempty"`
}
