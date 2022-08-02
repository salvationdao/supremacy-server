package server

import (
	"server/db/boiler"
	"time"

	"github.com/volatiletech/null/v8"
)

type Feature struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	GloballyEnabled bool      `json:"globally_enabled"`
	DeletedAt       null.Time `json:"deleted_at,omitempty"`
	UpdatedAt       time.Time `json:"updated_at"`
	CreatedAt       time.Time `json:"created_at"`
}

func FeaturesFromBoiler(features boiler.FeatureSlice) []*Feature {
	var serverFeatures []*Feature
	for _, feature := range features {
		serverFeature := &Feature{
			Name:            feature.Name,
			GloballyEnabled: feature.GloballyEnabled,
			DeletedAt:       feature.DeletedAt,
			UpdatedAt:       feature.UpdatedAt,
			CreatedAt:       feature.CreatedAt,
		}

		serverFeatures = append(serverFeatures, serverFeature)
	}
	return serverFeatures
}
