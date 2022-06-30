package server

import (
	"server/db/boiler"
	"time"

	"github.com/volatiletech/null/v8"
)

type Feature struct {
	ID        string    `json:"id"`
	Type      string    `json:"Type"`
	DeletedAt null.Time `json:"deleted_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

func FeaturesFromBoiler(features boiler.FeatureSlice) []*Feature {
	var serverFeatures []*Feature
	for _, feature := range features {
		serverFeature := &Feature{
			ID:        feature.ID,
			Type:      feature.Type,
			DeletedAt: feature.DeletedAt,
			UpdatedAt: feature.UpdatedAt,
			CreatedAt: feature.CreatedAt,
		}

		serverFeatures = append(serverFeatures, serverFeature)
	}
	return serverFeatures
}
