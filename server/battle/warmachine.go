package battle

import "server/db/boiler"

type Stat struct {
	X        uint32 `json:"x"`
	Y        uint32 `json:"y"`
	Rotation uint32 `json:"rotation"`
}

type DamageRecord struct {
	Amount             int              `json:"amount"` // The total amount of damage taken from this source
	CausedByWarMachine *WarMachineBrief `json:"caused_by_war_machine,omitempty"`
	SourceName         string           `json:"source_name,omitempty"` // The name of the weapon / damage causer (in-case of now TokenID)
}

type WMDestroyedRecord struct {
	DestroyedWarMachine *WarMachineBrief `json:"destroyed_war_machine"`
	KilledByWarMachine  *WarMachineBrief `json:"killed_by_war_machine,omitempty"`
	KilledBy            string           `json:"killed_by"`
	DamageRecords       []*DamageRecord  `json:"damage_records"`
}

type DamageHistory struct {
	Amount         int    `json:"amount"`          // The total amount of damage taken from this source
	InstigatorHash string `json:"instigator_hash"` // The Hash of the WarMachine that caused the damage (0 if none, ie: an Airstrike)
	SourceHash     string `json:"source_hash"`     // The Hash of the weapon
	SourceName     string `json:"source_name"`     // The name of the weapon / damage causer (in-case of now Hash)
}

type WarMachineBrief struct {
	ParticipantID byte            `json:"participantID"`
	Hash          string          `json:"hash"`
	ImageUrl      string          `json:"imageUrl"`
	ImageAvatar   string          `json:"imageAvatar"`
	Name          string          `json:"name"`
	FactionID     string          `json:"factionID"`
	Faction       *boiler.Faction `json:"faction"`
}
