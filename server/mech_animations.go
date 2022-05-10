package server

import (
	"server/db/boiler"
	"time"

	"github.com/volatiletech/null/v8"
)

type MechAnimation struct {
	*CollectionDetails
	ID               string      `json:"id"`
	BlueprintID      string      `json:"blueprint_id"`
	CollectionItemID string      `json:"collection_item_id"`
	Label            string      `json:"label"`
	OwnerID          string      `json:"owner_id"`
	ChassisModel     string      `json:"chassis_model"`
	EquippedOn       null.String `json:"equipped_on,omitempty"`
	Tier             null.String `json:"tier,omitempty"`
	IntroAnimation   null.Bool   `json:"intro_animation,omitempty"`
	OutroAnimation   null.Bool   `json:"outro_animation,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
}

type BlueprintMechAnimation struct {
	ID             string      `json:"id"`
	Collection     string      `json:"collection"`
	Label          string      `json:"label"`
	ChassisModel   string      `json:"chassis_model"`
	EquippedOn     null.String `json:"equipped_on,omitempty"`
	Tier           null.String `json:"tier,omitempty"`
	IntroAnimation null.Bool   `json:"intro_animation,omitempty"`
	OutroAnimation null.Bool   `json:"outro_animation,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
}

func BlueprintMechAnimationFromBoiler(animation *boiler.BlueprintMechAnimation) *BlueprintMechAnimation {
	return &BlueprintMechAnimation{
		ID:             animation.ID,
		Collection:     animation.Collection,
		Label:          animation.Label,
		ChassisModel:   animation.ChassisModel,
		EquippedOn:     animation.EquippedOn,
		Tier:           animation.Tier,
		IntroAnimation: animation.IntroAnimation,
		OutroAnimation: animation.OutroAnimation,
		CreatedAt:      animation.CreatedAt,
	}
}
