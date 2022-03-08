package battle

import "server"

type WarMachine struct {
	ID                 string          `json:"id"`
	Hash               string          `json:"Hash"`
	ParticipantID      byte            `json:"participantID"`
	FactionID          string          `json:"factionID"`
	MaxHealth          uint32          `json:"maxHealth"`
	Health             uint32          `json:"health"`
	MaxShield          uint32          `json:"maxShield"`
	Shield             uint32          `json:"shield"`
	Stat               *Stat           `json:"stat"`
	Position           *server.Vector3 `json:"position"`
	Rotation           int             `json:"rotation"`
	OwnedByID          string          `json:"ownedByID"`
	Name               string          `json:"name"`
	Description        *string         `json:"description"`
	ExternalUrl        string          `json:"externalUrl"`
	Image              string          `json:"image"`
	Skin               string          `json:"skin"`
	ShieldRechargeRate float64         `json:"shieldRechargeRate"`
	Speed              int             `json:"speed"`
	Durability         int             `json:"durability"`
	PowerGrid          int             `json:"powerGrid"`
	CPU                int             `json:"cpu"`
	WeaponHardpoint    int             `json:"weaponHardpoint"`
	TurretHardpoint    int             `json:"turretHardpoint"`
	UtilitySlots       int             `json:"utilitySlots"`
	Faction            *Faction        `json:"faction"`
	WeaponNames        []string        `json:"weaponNames"`
}

type Faction struct {
	ID    string        `json:"id"`
	Label string        `json:"label"`
	Theme *FactionTheme `json:"theme"`
}

type FactionTheme struct {
	Primary    string `json:"primary"`
	Secondary  string `json:"secondary"`
	Background string `json:"background"`
}

type Stat struct {
	X        uint32 `json:"x"`
	Y        uint32 `json:"y"`
	Rotation uint32 `json:"rotation"`
}

type DamageRecord struct {
	Amount                 int    `json:"amount"` // The total amount of damage taken from this source
	CausedByWarMachineHash string `json:"caused_by_war_machine,omitempty"`
	SourceName             string `json:"source_name,omitempty"` // The name of the weapon / damage causer (in-case of now TokenID)
}

type WMDestroyedRecord struct {
	DestroyedWarMachine *WarMachine     `json:"destroyed_war_machine"`
	KilledByWarMachine  *WarMachine     `json:"killed_by_war_machine,omitempty"`
	KilledBy            string          `json:"killed_by"`
	DamageRecords       []*DamageRecord `json:"damage_records"`
}

type DamageHistory struct {
	Amount         int    `json:"amount"`         // The total amount of damage taken from this source
	InstigatorHash string `json:"instigatorHash"` // The Hash of the WarMachine that caused the damage (0 if none, ie: an Airstrike)
	SourceHash     string `json:"sourceHash"`     // The Hash of the weapon
	SourceName     string `json:"sourceName"`     // The name of the weapon / damage causer (in-case of now Hash)
}

type WarMachineBrief struct {
	Hash        string        `json:"hash"`
	ImageUrl    string        `json:"image_url"`
	ImageAvatar string        `json:"image_avatar"`
	Name        string        `json:"name"`
	Faction     *FactionBrief `json:"faction"`
}

type FactionBrief struct {
	ID         string        `json:"id"`
	Label      string        `json:"label"`
	LogoBlobID string        `json:"logo_blob_id,omitempty"`
	Theme      *FactionTheme `json:"theme"`
}
