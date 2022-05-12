// Code generated by SQLBoiler 4.8.6 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package boiler

import (
	"strconv"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/strmangle"
)

// M type is for providing columns and column values to UpdateAll.
type M map[string]interface{}

// ErrSyncFail occurs during insert when the record could not be retrieved in
// order to populate default value information. This usually happens when LastInsertId
// fails or there was a primary key configuration that was not resolvable.
var ErrSyncFail = errors.New("boiler: failed to synchronize data after insert")

type insertCache struct {
	query        string
	retQuery     string
	valueMapping []uint64
	retMapping   []uint64
}

type updateCache struct {
	query        string
	valueMapping []uint64
}

func makeCacheKey(cols boil.Columns, nzDefaults []string) string {
	buf := strmangle.GetBuffer()

	buf.WriteString(strconv.Itoa(cols.Kind))
	for _, w := range cols.Cols {
		buf.WriteString(w)
	}

	if len(nzDefaults) != 0 {
		buf.WriteByte('.')
	}
	for _, nz := range nzDefaults {
		buf.WriteString(nz)
	}

	str := buf.String()
	strmangle.PutBuffer(buf)
	return str
}

// Enum values for AbilityTypeEnum
const (
	AbilityTypeEnumAIRSTRIKE      = "AIRSTRIKE"
	AbilityTypeEnumNUKE           = "NUKE"
	AbilityTypeEnumREPAIR         = "REPAIR"
	AbilityTypeEnumROB            = "ROB"
	AbilityTypeEnumREINFORCEMENTS = "REINFORCEMENTS"
	AbilityTypeEnumROBOTDOGS      = "ROBOT DOGS"
	AbilityTypeEnumOVERCHARGE     = "OVERCHARGE"
	AbilityTypeEnumFIREWORKS      = "FIREWORKS"
)

// Enum values for BattleEvent
const (
	BattleEventKilled           = "killed"
	BattleEventSpawnedAi        = "spawned_ai"
	BattleEventKill             = "kill"
	BattleEventAbilityTriggered = "ability_triggered"
	BattleEventPickup           = "pickup"
)

// Enum values for LocationSelectTypeEnum
const (
	LocationSelectTypeEnumLINE_SELECT     = "LINE_SELECT"
	LocationSelectTypeEnumMECH_SELECT     = "MECH_SELECT"
	LocationSelectTypeEnumLOCATION_SELECT = "LOCATION_SELECT"
	LocationSelectTypeEnumGLOBAL          = "GLOBAL"
)

// Enum values for ChatMSGTypeEnum
const (
	ChatMSGTypeEnumTEXT        = "TEXT"
	ChatMSGTypeEnumPUNISH_VOTE = "PUNISH_VOTE"
)

// Enum values for AbilityLevel
const (
	AbilityLevelMECH    = "MECH"
	AbilityLevelFACTION = "FACTION"
	AbilityLevelPLAYER  = "PLAYER"
)

// Enum values for MultiplierTypeEnum
const (
	MultiplierTypeEnumSpendAverage = "spend_average"
	MultiplierTypeEnumMostSupsLost = "most_sups_lost"
	MultiplierTypeEnumGabAbility   = "gab_ability"
	MultiplierTypeEnumComboBreaker = "combo_breaker"
	MultiplierTypeEnumPlayerMech   = "player_mech"
	MultiplierTypeEnumHoursOnline  = "hours_online"
	MultiplierTypeEnumSyndicateWin = "syndicate_win"
	MultiplierTypeEnumContribute   = "contribute"
)

// Enum values for PlayerRankEnum
const (
	PlayerRankEnumGENERAL     = "GENERAL"
	PlayerRankEnumCORPORAL    = "CORPORAL"
	PlayerRankEnumPRIVATE     = "PRIVATE"
	PlayerRankEnumNEW_RECRUIT = "NEW_RECRUIT"
)
