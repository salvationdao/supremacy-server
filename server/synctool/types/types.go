package types

import (
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

type MechModel struct {
	ID                      string      `json:"id"`
	Label                   string      `json:"label"`
	DefaultChassisSkinID    string      `json:"default_chassis_skin_id"`
	BrandID                 string      `json:"brand_id"`
	MechType                string      `json:"mech_type"`
	BoostStat               string      `json:"boost_stat"`
	WeaponHardpoints        string      `json:"weapon_hardpoints"`
	UtilitySlots            string      `json:"utility_slots"`
	Speed                   string      `json:"speed"`
	MaxHitpoints            string      `json:"max_hitpoints"`
	PowerCoreSize           string      `json:"power_core_size"`
	Collection              string      `json:"collection"`
	AvailabilityID          null.String `json:"availability_id"`
	ShieldMax               string      `json:"shield_max"`
	ShieldRechargeRate      string      `json:"shield_recharge_rate"`
	ShieldRechargePowerCost string      `json:"shield_recharge_power_cost"`
	ShieldTypeID            string      `json:"shield_type_id"`
	ShieldRechargeDelay     string      `json:"shield_recharge_delay"`
	HeightMeters            string      `json:"height_meter"`
	WalkSpeedModifier       string      `json:"walk_speed_modifier"`
	SprintSpreadModifier    string      `json:"sprint_spread_modifier"`
	IdleDrain               string      `json:"idle_drain"`
	WalkDrain               string      `json:"walk_drain"`
	RunDrain                string      `json:"run_drain"`
}

type Brands struct {
	ID        string `json:"id"`
	FactionID string `json:"faction_id"`
	Label     string `json:"label"`
}

type Faction struct {
	ID             string `json:"id"`
	ContractReward string `json:"contract_reward"`
	VotePrice      string `json:"vote_Price"`
	Label          string `json:"label"`
	GuildID        string `json:"guild_id"`
	DeletedAt      string `json:"deleted_at"`
	UpdatedAt      string `json:"updated_at"`
	CreatedAt      string `json:"created_at"`
	LogoURL        string `json:"logo_url"`
	BackgroundURL  string `json:"background_url"`
	Description    string `json:"description"`
}

type FactionPalette struct {
	FactionID  string `json:"faction_id"`
	Primary    string `json:"primary"`
	Text       string `json:"text"`
	Background string `json:"background"`
	S100       string `json:"s100"`
	S200       string `json:"s200"`
	S300       string `json:"s300"`
	S400       string `json:"s400"`
	S500       string `json:"s500"`
	S600       string `json:"s600"`
	S700       string `json:"s700"`
	S800       string `json:"s800"`
	S900       string `json:"s900"`
}

type FactionPass struct {
	ID          string          `json:"id"`
	Label       string          `json:"label"`
	LastForDays int             `json:"last_for_days"`
	SupsCost    decimal.Decimal `json:"sups_cost"`
	DeletedAt   null.Time       `json:"deleted_at"`
}

type MechSkin struct {
	ID                    string      `json:"id"`
	Collection            string      `json:"collection"`
	Label                 string      `json:"label"`
	Tier                  string      `json:"tier"`
	DefaultLevel          string      `json:"default_level"`
	BlueprintWeaponSkinID null.String `json:"blueprint_weapon_skin_id"`
	ImageUrl              null.String `json:"image_url"`
	AnimationUrl          null.String `json:"animation_url"`
	CardAnimationUrl      null.String `json:"card_animation_url"`
	LargeImageUrl         null.String `json:"large_image_url"`
	AvatarUrl             null.String `json:"avatar_url"`
	BackgroundColor       null.String `json:"background_color"`
	YoutubeUrl            null.String `json:"youtube_url"`
}

type MechModelSkinCompatibility struct {
	MechSkinID       string `json:"mech_skin_id"`
	MechModelID      string `json:"mech_model_id"`
	ImageUrl         string `json:"image_url"`
	AnimationUrl     string `json:"animation_url"`
	CardAnimationUrl string `json:"card_animation_url"`
	LargeImageUrl    string `json:"large_image_url"`
	AvatarUrl        string `json:"avatar_url"`
	BackgroundColor  string `json:"background_color"`
	YoutubeUrl       string `json:"youtube_url"`
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

type Weapon struct {
	ID                  string      `json:"id"`
	BrandID             string      `json:"brand_id"`
	Label               string      `json:"label"`
	WeaponType          string      `json:"weapon_type"`
	DefaultSkinID       string      `json:"default_skin_id"`
	Damage              string      `json:"damage"`
	DamageFallOff       string      `json:"damage_falloff"`
	DamageFalloffRate   string      `json:"damage_falloff_rate"`
	Radius              string      `json:"radius"`
	RadiusDamageFalloff string      `json:"radius_damage_falloff"`
	Spread              string      `json:"spread"`
	RateOfFire          string      `json:"rate_of_fire"`
	ProjectileSpeed     string      `json:"projectile_speed"`
	MaxAmmo             string      `json:"max_ammo"`
	IsMelee             string      `json:"is_melee"`
	PowerCost           string      `json:"power_cost"`
	GameClientWeaponID  null.String `json:"game_client_weapon_id"`
	Collection          string      `json:"collection"`
	DefaultDamageType   string      `json:"default_damage_type"`
	ProjectileAmount    string      `json:"projectile_amount"`
	DotTickDamage       string      `json:"dot_tick_damage"`
	DotMaxTicks         string      `json:"dot_max_ticks"`
	IsArced             string      `json:"is_arced"`
	ChargeTimeSeconds   string      `json:"charge_time_seconds"`
	BurstRateOfFire     string      `json:"burst_rate_of_fire"`
	PowerInstantDrain   string      `json:"power_instant_drain"`
	DotTickDuration     string      `boiler:"dot_tick_duration"`
	ProjectileLifeSpan  string      `boiler:"projectile_life_span"`
	RecoilForce         string      `boiler:"recoil_force"`
	IdlePowerCost       string      `boiler:"idle_power_cost"`
}

type WeaponSkin struct {
	ID               string      `json:"id"`
	Label            string      `json:"label"`
	Tier             string      `json:"tier"`
	CreatedAt        string      `json:"created_at"`
	Collection       string      `json:"collection"`
	StatModifier     string      `json:"stat_modifier"`
	ImageUrl         null.String `json:"image_url"`
	AnimationUrl     null.String `json:"animation_url"`
	CardAnimationUrl null.String `json:"card_animation_url"`
	LargeImageUrl    null.String `json:"large_image_url"`
	AvatarUrl        null.String `json:"avatar_url"`
	BackgroundColor  null.String `json:"background_color"`
	YoutubeUrl       null.String `json:"youtube_url"`
}

type WeaponModelSkinCompatibility struct {
	WeaponSkinID     string `json:"weapon_skin_id"`
	WeaponModelID    string `json:"weapon_model_id"`
	ImageUrl         string `json:"image_url"`
	AnimationUrl     string `json:"animation_url"`
	CardAnimationUrl string `json:"card_animation_url"`
	LargeImageUrl    string `json:"large_image_url"`
	AvatarUrl        string `json:"avatar_url"`
	BackgroundColor  string `json:"background_color"`
	YoutubeUrl       string `json:"youtube_url"`
}

type BattleAbility struct {
	ID               string `json:"id"`
	Label            string `json:"label"`
	CoolDownDuration string `json:"cool_down_duration"`
	Description      string `json:"description"`
}

type GameAbility struct {
	ID                  string  `json:"id"`
	GameClientAbilityID int     `json:"game_client_ability_id"`
	FactionID           string  `json:"faction_id"`
	BattleAbilityID     *string `json:"battle_ability_id"`
	Label               string  `json:"label"`
	Colour              string  `json:"colour"`
	ImageUrl            string  `json:"image_url"`
	SupsCost            string  `json:"sups_cost"`
	Description         string  `json:"description"`
	TextColour          string  `json:"text_colour"`
	CurrentSups         string  `json:"current_sups"`
	Level               string  `json:"level"`
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
	WeaponShare      string `json:"weapon_share"`
	MovementShare    string `json:"movement_share"`
	UtilityShare     string `json:"utility_share"`
}

type BlueprintWeapons struct {
	ID                  string `json:"id"`
	Label               string `json:"label"`
	Slug                string `json:"slug"`
	Damage              string `json:"damage"`
	DeletedAt           string `json:"deleted_at"`
	UpdatedAt           string `json:"updated_at"`
	CreatedAt           string `json:"created_at"`
	GameClientWeaponID  string `json:"game_client_weapon_id"`
	WeaponType          string `json:"weapon_type"`
	Collection          string `json:"collection"`
	DefaultDamageType   string `json:"default_damage_type"`
	DamageFalloff       string `json:"damage_falloff"`
	DamageFalloffRate   string `json:"damage_falloff_rate"`
	Radius              string `json:"radius"`
	RadiusDamageFalloff string `json:"radius_damage_falloff"`
	Spread              string `json:"spread"`
	RateOfFire          string `json:"rate_of_fire"`
	ProjectileSpeed     string `json:"projectile_speed"`
	MaxAmmo             string `json:"max_ammo"`
	IsMelee             string `json:"is_melee"`
	Tier                string `json:"tier"`
	EnergyCost          string `json:"energy_cost"`
	WeaponModelID       string `json:"weapon_model_id"`
}
