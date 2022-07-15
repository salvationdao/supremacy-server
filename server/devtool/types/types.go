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

type Faction struct {
	ID              string `json:"id"`
	ContractReward  string `json:"contract_reward"`
	VotePrice       string `json:"vote_Price"`
	Label           string `json:"label"`
	GuildID         string `json:"guild_id"`
	DeletedAt       string `json:"deleted_at"`
	UpdatedAt       string `json:"updated_at"`
	CreatedAt       string `json:"created_at"`
	PrimaryColor    string `json:"primary_color"`
	SecondaryColor  string `json:"secondary_color"`
	BackgroundColor string `json:"background_color"`
	LogoURL         string `json:"logo_url"`
	BackgroundURL   string `json:"background_url"`
	Description     string `json:"description"`
}

type MechSkin struct {
	ID               string      `json:"id"`
	Collection       string      `json:"collection"`
	MechModel        string      `json:"mech_model"`
	Label            string      `json:"label"`
	Tier             string      `json:"tier"`
	ImageUrl         null.String `json:"image_url"`
	AnimationUrl     null.String `json:"animation_url"`
	CardAnimationUrl null.String `json:"card_animation_url"`
	LargeImageUrl    null.String `json:"large_image_url"`
	AvatarUrl        null.String `json:"avatar_url"`
	BackgroundColor  string      `json:"background_color"`
	YoutubeURL       string      `json:"youtube_url"`
	StatModifier     string      `json:"stat_modifier"`
	MechType         string      `json:"mech_type"`
}

type MysteryCrate struct {
	ID               string `json:"id"`
	MysteryCrateType string `json:"mystery_crate_type"`
	FactionID        string `json:"faction_id"`
	Label            string `json:"label"`
	Description      string `json:"description"`
	ImageUrl         string `json:"image_url"`
	CardAnimationUrl string `json:"card_animation_url"`
	AvatarUrl        string `json:"avatar_url"`
	LargeImageUrl    string `json:"large_image_url"`
	BackgroundColor  string `json:"background_color"`
	AnimationUrl     string `json:"animation_url"`
	YoutubeUrl       string `json:"youtube_url"`
}

type WeaponModel struct {
	ID            string `json:"id"`
	BrandID       string `json:"brand_id"`
	Label         string `json:"label"`
	WeaponType    string `json:"weapon_type"`
	DefaultSkinID string `json:"default_skin_id"`
	DeletedAt     string `json:"deleted_at"`
	UpdatedAt     string `json:"updated_at"`
}

type WeaponSkin struct {
	ID               string `json:"id"`
	Label            string `json:"label"`
	WeaponType       string `json:"weapon_typep"`
	Tier             string `json:"tier"`
	ImageUrl         string `json:"image_url"`
	CardAnimationUrl string `json:"card_animation_url"`
	AvatarUrl        string `json:"avatar_url"`
	LargeImageUrl    string `json:"large_image_url"`
	BackgroundColor  string `json:"background_color"`
	AnimationUrl     string `json:"animation_url"`
	YoutubeUrl       string `json:"youtube_url"`
	Collection       string `json:"collection"`
	WeaponModelID    string `json:"weapon_model_id"`
	StatModifier     string `json:"stat_modifier"`
}

type BattleAbility struct {
	ID               string `json:"id"`
	Label            string `json:"label"`
	CoolDownDuration string `json:"cool_down_duration"`
	Description      string `json:"description"`
}

type PowerCores struct {
	ID               string `json:"id"`
	Collection       string `json:"collection"`
	Label            string `json:"label"`
	Size             string `json:"size"`
	Capacity         string `json:"capacity"`
	MaxDrawRate      string `json:"max_draw_rate"`
	RechargeRate     string `json:"recharge_rate"`
	Armour           string `json:"armour"`
	MaxHitpoints     string `json:"max_hitpoints"`
	Tier             string `json:"tier"`
	ImageUrl         string `json:"image_url"`
	CardAnimationUrl string `json:"card_animation_url"`
	AvatarUrl        string `json:"avatar_url"`
	LargeImageUrl    string `json:"large_image_url"`
	BackgroundColor  string `json:"background_color"`
	AnimationUrl     string `json:"animation_url"`
	YoutubeUrl       string `json:"youtube_url"`
}
