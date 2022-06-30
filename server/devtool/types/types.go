package types

import "github.com/volatiletech/null/v8"

type MechModel struct {
	ID                   string      `json:"id"`
	Label                string      `json:"label"`
	DefaultChassisSkinID string      `json:"default_chassis_skin_id"`
	BrandID              null.String `json:"brand_id"`
	MechType             string      `json:"mech_type"`
}

type Brands struct {
	ID        string `json:"id"`
	FactionID string `json:"faction_id"`
	Label     string `json:"label"`
}

type MechSkin struct {
	ID               string      `json:"id"`
	Collection       string      `json:"collection"`
	Label            string      `json:"label"`
	Tier             string      `json:"tier"`
	ImageUrl         null.String `json:"image_url"`
	AnimationUrl     null.String `json:"animation_url"`
	CardAnimationUrl null.String `json:"card_animation_url"`
	LargeImageUrl    null.String `json:"large_image_url"`
	AvatarUrl        null.String `json:"avatar_url"`
	MechType         string      `json:"mech_type"`
}
