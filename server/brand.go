package server

import (
	"encoding/json"
	"fmt"
	"server/db/boiler"
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
	LogoURL   string    `json:"logo_url"`
}

func (b *Brand) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

func BrandFromBoiler(brand *boiler.Brand) *Brand {
	return &Brand{
		ID:        brand.ID,
		FactionID: brand.FactionID,
		Label:     brand.Label,
		DeletedAt: brand.DeletedAt,
		UpdatedAt: brand.UpdatedAt,
		CreatedAt: brand.CreatedAt,
		LogoURL:   brand.LogoURL,
	}
}
