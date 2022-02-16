package server

type Stream struct {
	ID            StreamID `json:"id" db:"id"`
	Name          string   `json:"name" db:"name"`
	Url           string   `json:"url" db:"url"`
	Region        string   `json:"region" db:"region"`
	Resolution    string   `json:"resolution" db:"resolution"`
	BitRatesKBits int      `json:"bitRatesKBits" db:"bit_rates_k_bits"`
	UserMax       int      `json:"userMax" db:"user_max"`
	UsersNow      int      `json:"usersNow" db:"users_now"`
	Active        bool     `json:"active" db:"active"`
	Status        string   `json:"status" db:"status"`
	Latitude      float32  `json:"latitude" db:"latitude"`
	Longitude     float32  `json:"longitude" db:"longitude"`
}
