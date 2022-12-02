package oven_stream

import (
	"fmt"
	"server"
	"server/db"
)

type OvenStream struct {
	Name                 string   `json:"name"`
	BaseUrl              string   `json:"base_url"`
	AvailableResolutions []string `json:"available_resolutions"`
	DefaultResolution    string   `json:"default_resolution"`
}

func GetStreamDetails(name, arenaID string) *OvenStream {
	env := "staging"
	if server.IsProductionEnv() {
		env = "production"
	}

	baseURL := fmt.Sprintf("%s/%s-%s", db.GetStrWithDefault(db.KeyOvenmediaStreamURL, "wss://stream2.supremacy.game:3334/app"), env, arenaID)
	availableResolution := []string{"240", "360", "480", "720", "1080"}

	ovenStream := &OvenStream{
		Name:                 name,
		BaseUrl:              baseURL,
		AvailableResolutions: availableResolution,
		DefaultResolution:    "1080",
	}
	return ovenStream
}
