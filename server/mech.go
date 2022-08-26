package server

import (
	"encoding/json"
	"fmt"
	"server/db/boiler"
	"time"

	"github.com/volatiletech/null/v8"
)

/*
	THIS FILE SHOULD CONTAIN ZERO BOILER STRUCTS
*/

type Stats struct {
	TotalWins       int `json:"total_wins"`
	TotalDeaths     int `json:"total_deaths"`
	TotalKills      int `json:"total_kills"`
	BattlesSurvived int `json:"battles_survived"`
	TotalLosses     int `json:"total_losses"`
}

// Mech is the struct that rpc expects for mechs
type Mech struct {
	*CollectionItem
	*Stats
	*Images

	ID                    string     `json:"id"`
	Label                 string     `json:"label"`
	WeaponHardpoints      int        `json:"weapon_hardpoints"`
	UtilitySlots          int        `json:"utility_slots"`
	Speed                 int        `json:"speed"`
	MaxHitpoints          int        `json:"max_hitpoints"`
	IsDefault             bool       `json:"is_default"`
	IsInsured             bool       `json:"is_insured"`
	Name                  string     `json:"name"`
	GenesisTokenID        null.Int64 `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID null.Int64 `json:"limited_release_token_id,omitempty"`
	PowerCoreSize         string     `json:"power_core_size"`
	CollectionItemID      string     `json:"-"`
	QueuePosition         null.Int   `json:"queue_position"`
	BattleReady           bool       `json:"battle_ready"`

	BlueprintID string         `json:"blueprint_id"`
	Blueprint   *BlueprintMech `json:"blueprint_mech,omitempty"`

	BrandID string `json:"brand_id"`
	Brand   *Brand `json:"brand"`

	Owner *User `json:"user"`

	FactionID null.String `json:"faction_id"`
	Faction   *Faction    `json:"faction,omitempty"`

	ModelID string     `json:"model_id"`
	Model   *MechModel `json:"model"`

	// Connected objects
	ChassisSkinID string    `json:"chassis_skin_id,omitempty"`
	ChassisSkin   *MechSkin `json:"chassis_skin,omitempty"`

	IntroAnimationID null.String    `json:"intro_animation_id,omitempty"`
	IntroAnimation   *MechAnimation `json:"intro_animation,omitempty"`

	OutroAnimationID null.String    `json:"outro_animation_id,omitempty"`
	OutroAnimation   *MechAnimation `json:"outro_animation,omitempty"`

	PowerCoreID null.String `json:"power_core_id,omitempty"`
	PowerCore   *PowerCore  `json:"power_core,omitempty"`

	Weapons WeaponSlice  `json:"weapons"`
	Utility UtilitySlice `json:"utility"`

	ItemSaleID null.String `json:"item_sale_id"`

	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

type BlueprintMech struct {
	ID                   string    `json:"id"`
	Label                string    `json:"label"`
	Slug                 string    `json:"slug"`
	Skin                 string    `json:"skin"`
	WeaponHardpoints     int       `json:"weapon_hardpoints"`
	UtilitySlots         int       `json:"utility_slots"`
	Speed                int       `json:"speed"`
	MaxHitpoints         int       `json:"max_hitpoints"`
	UpdatedAt            time.Time `json:"updated_at"`
	CreatedAt            time.Time `json:"created_at"`
	PowerCoreSize        string    `json:"power_core_size,omitempty"`
	Tier                 string    `json:"tier,omitempty"`
	DefaultChassisSkinID string    `json:"default_chassis_skin_id"`
	Collection           string    `json:"collection"`
	ModelID              string    `json:"model_id"`

	// only used on inserting new mechs/items, since we are still giving away some limited released and genesis
	GenesisTokenID        null.Int64 `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID null.Int64 `json:"limited_release_token_id,omitempty"`
}

func (b *BlueprintMech) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

func BlueprintMechFromBoiler(mech *boiler.BlueprintMech) *BlueprintMech {
	return &BlueprintMech{
		ID:               mech.ID,
		Label:            mech.Label,
		Slug:             mech.Slug,
		WeaponHardpoints: mech.WeaponHardpoints,
		UtilitySlots:     mech.UtilitySlots,
		Speed:            mech.Speed,
		MaxHitpoints:     mech.MaxHitpoints,
		UpdatedAt:        mech.UpdatedAt,
		CreatedAt:        mech.CreatedAt,
		PowerCoreSize:    mech.PowerCoreSize,
		Tier:             mech.Tier,
		Collection:       mech.Collection,
		ModelID:          mech.ModelID,
	}
}

type BlueprintUtility struct {
	ID               string      `json:"id"`
	BrandID          null.String `json:"brand_id,omitempty"`
	Label            string      `json:"label"`
	UpdatedAt        time.Time   `json:"updated_at"`
	CreatedAt        time.Time   `json:"created_at"`
	Type             string      `json:"type"`
	Collection       string      `json:"collection"`
	Tier             string      `json:"tier,omitempty"`
	ImageURL         null.String `json:"image_url,omitempty"`
	CardAnimationURL null.String `json:"card_animation_url,omitempty"`
	AvatarURL        null.String `json:"avatar_url,omitempty"`
	LargeImageURL    null.String `json:"large_image_url,omitempty"`
	BackgroundColor  null.String `json:"background_color,omitempty"`
	AnimationURL     null.String `json:"animation_url,omitempty"`
	YoutubeURL       null.String `json:"youtube_url,omitempty"`

	ShieldBlueprint      *BlueprintUtilityShield      `json:"shield_blueprint,omitempty"`
	AttackDroneBlueprint *BlueprintUtilityAttackDrone `json:"attack_drone_blueprint,omitempty"`
	RepairDroneBlueprint *BlueprintUtilityRepairDrone `json:"repair_drone_blueprint,omitempty"`
	AcceleratorBlueprint *BlueprintUtilityAccelerator `json:"accelerator_blueprint,omitempty"`
	AntiMissileBlueprint *BlueprintUtilityAntiMissile `json:"anti_missile_blueprint,omitempty"`

	// only used on inserting new mechs/items, since we are still giving away some limited released and genesis
	GenesisTokenID        null.Int64 `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID null.Int64 `json:"limited_release_token_id,omitempty"`
}

type MechModel struct {
	ID                   string      `json:"id"`
	Label                string      `json:"label"`
	CreatedAt            time.Time   `json:"created_at"`
	DefaultChassisSkinID null.String `json:"default_chassis_skin_id,omitempty"`
	BrandID              null.String `json:"brand_id"`
	MechType             null.String `json:"mech_type"`
	RepairBlocks         int         `json:"repair_blocks"`
}

func (b *MechModel) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

// IsBattleReady checks if a mech has the minimum it needs for battle
func (m *Mech) IsBattleReady() bool {
	if !m.PowerCoreID.Valid {
		return false
	}
	if len(m.Weapons) <= 0 {
		return false
	}
	return true
}

func (m *Mech) CheckAndSetAsGenesisOrLimited() (genesisID null.Int64, limitedID null.Int64) {
	if !m.GenesisTokenID.Valid && !m.LimitedReleaseTokenID.Valid {
		return
	}
	if m.GenesisTokenID.Valid && m.IsCompleteGenesis() {
		genesisID = m.GenesisTokenID
		m.TokenID = m.GenesisTokenID.Int64
		m.CollectionSlug = "supremacy-genesis"
		return
	}
	if m.LimitedReleaseTokenID.Valid && m.IsCompleteLimited() {
		limitedID = m.LimitedReleaseTokenID
		m.TokenID = m.LimitedReleaseTokenID.Int64
		m.CollectionSlug = "supremacy-limited-release"
		return
	}
	return
}

// IsCompleteGenesis returns true if all parts of this mech are genesis with matching genesis token IDs
func (m *Mech) IsCompleteGenesis() bool {
	if !m.GenesisTokenID.Valid {
		return false
	}
	// this checks if mech is complete genesis
	// the shield and skins are locked to genesis, so they are true
	// we just need to check the first 2 weapons, since rocket pods are also locked
	if m.Weapons[0] == nil || !m.Weapons[0].GenesisTokenID.Valid ||
		m.Weapons[0].GenesisTokenID.Int64 != m.GenesisTokenID.Int64 {
		return false
	}
	if m.Weapons[1] == nil || !m.Weapons[1].GenesisTokenID.Valid ||
		m.Weapons[1].GenesisTokenID.Int64 != m.GenesisTokenID.Int64 {
		return false
	}
	if m.Weapons[2] == nil || !m.Weapons[2].GenesisTokenID.Valid ||
		m.Weapons[2].GenesisTokenID.Int64 != m.GenesisTokenID.Int64 {
		return false
	}
	return true
}

// IsCompleteLimited returns true if all parts of this mech are limited with matching limited token IDs
func (m *Mech) IsCompleteLimited() bool {
	if !m.LimitedReleaseTokenID.Valid {
		return false
	}
	// this checks if mech is complete genesis
	// the shield and skins are locked to genesis, so they are true
	// we just need to check the first 2 weapons, since rocket pods are also locked
	if m.Weapons[0] == nil || !m.Weapons[0].LimitedReleaseTokenID.Valid ||
		m.Weapons[0].LimitedReleaseTokenID.Int64 != m.LimitedReleaseTokenID.Int64 {
		return false
	}
	if m.Weapons[1] == nil || !m.Weapons[1].LimitedReleaseTokenID.Valid ||
		m.Weapons[1].LimitedReleaseTokenID.Int64 != m.LimitedReleaseTokenID.Int64 {
		return false
	}
	if m.Weapons[2] == nil || !m.Weapons[2].LimitedReleaseTokenID.Valid ||
		m.Weapons[2].LimitedReleaseTokenID.Int64 != m.LimitedReleaseTokenID.Int64 {
		return false
	}
	return true
}

func MechToGenesisOrLimited() {

}

func MechModelsFromBoiler(mechModel []*boiler.MechModel) []*MechModel {
	var smm []*MechModel

	for _, mm := range mechModel {
		smm = append(smm, &MechModel{
			ID:                   mm.ID,
			Label:                mm.Label,
			CreatedAt:            mm.CreatedAt,
			DefaultChassisSkinID: null.StringFrom(mm.DefaultChassisSkinID),
			BrandID:              mm.BrandID,
			MechType:             mm.MechType,
			RepairBlocks:         mm.RepairBlocks,
		})
	}
	return smm
}
