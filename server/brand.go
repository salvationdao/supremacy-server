package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/volatiletech/null/v8"
)

type Brand struct {
	ID        string    `json:"id"`
	FactionID string    `json:"faction_id"`
	Label     string    `json:"label"`
	DeletedAt null.Time `json:"deleted_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (b *Brand) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}
