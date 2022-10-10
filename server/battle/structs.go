package battle

import (
	"github.com/sasha-s/go-deadlock"
	"time"

	"github.com/gofrs/uuid"
)

type Started struct {
	BattleID           string        `json:"battleID"`
	WarMachines        []*WarMachine `json:"warMachines"`
	WarMachineLocation []byte        `json:"warMachineLocation"`
}

type BattleUser struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FactionID string    `json:"faction_id"`
	deadlock.RWMutex
}

type BattleEndDetail struct {
	BattleID                     string        `json:"battle_id"`
	BattleIdentifier             int           `json:"battle_identifier"`
	StartedAt                    time.Time     `json:"started_at"`
	EndedAt                      time.Time     `json:"ended_at"`
	WinningCondition             string        `json:"winning_condition"`
	WinningFaction               *Faction      `json:"winning_faction"`
	WinningFactionIDOrder        []string      `json:"winning_faction_id_order"`
	WinningWarMachines           []*WarMachine `json:"winning_war_machines"`
	MostFrequentAbilityExecutors []*BattleUser `json:"most_frequent_ability_executors"`

	MechRewards []*MechReward `json:"mech_rewards"` // reward for mech owners
}

type Faction struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Theme *Theme `json:"theme"`
}

type Theme struct {
	PrimaryColor    string `json:"primary"`
	SecondaryColor  string `json:"secondary"`
	BackgroundColor string `json:"background"`
}

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
	KilledByUser        *UserBrief       `json:"killed_by_user"`
	KilledBy            string           `json:"killed_by"`
	KillerFactionID     string           `json:"killer_faction_id"`
	DamageRecords       []*DamageRecord  `json:"damage_records"`
}

type DamageHistory struct {
	Amount         int    `json:"amount"`          // The total amount of damage taken from this source
	InstigatorHash string `json:"instigator_hash"` // The Hash of the WarMachine that caused the damage (0 if none, ie: an Airstrike)
	SourceHash     string `json:"source_hash"`     // The Hash of the weapon
	SourceName     string `json:"source_name"`     // The name of the weapon / damage causer (in-case of now Hash)
}

type WarMachineBrief struct {
	ParticipantID byte   `json:"participantID"`
	Hash          string `json:"hash"`
	ImageUrl      string `json:"imageUrl"`
	ImageAvatar   string `json:"imageAvatar"`
	Name          string `json:"name"`
	FactionID     string `json:"factionID"`
}

type FactionBrief struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	LogoBlobID string `json:"logo_blob_id,omitempty"`
	Primary    string `json:"primary_color"`
	Secondary  string `json:"secondary_color"`
	Background string `json:"background_color"`
}
