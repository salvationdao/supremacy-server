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
	AbilityTypeEnumLANDMINE       = "LANDMINE"
)

// Enum values for BattleEvent
const (
	BattleEventKilled           = "killed"
	BattleEventSpawnedAi        = "spawned_ai"
	BattleEventKill             = "kill"
	BattleEventAbilityTriggered = "ability_triggered"
	BattleEventPickup           = "pickup"
)

// Enum values for WeaponType
const (
	WeaponTypeGrenadeLauncher = "Grenade Launcher"
	WeaponTypeCannon          = "Cannon"
	WeaponTypeMinigun         = "Minigun"
	WeaponTypePlasmaGun       = "Plasma Gun"
	WeaponTypeFlak            = "Flak"
	WeaponTypeMachineGun      = "Machine Gun"
	WeaponTypeFlamethrower    = "Flamethrower"
	WeaponTypeMissileLauncher = "Missile Launcher"
	WeaponTypeLaserBeam       = "Laser Beam"
	WeaponTypeLightningGun    = "Lightning Gun"
	WeaponTypeBFG             = "BFG"
	WeaponTypeRifle           = "Rifle"
	WeaponTypeSniperRifle     = "Sniper Rifle"
	WeaponTypeSword           = "Sword"
)

// Enum values for  are not proper Go identifiers, cannot emit constants
// Enum values for  are not proper Go identifiers, cannot emit constants
// Enum values for  are not proper Go identifiers, cannot emit constants
// Enum values for  are not proper Go identifiers, cannot emit constants
// Enum values for  are not proper Go identifiers, cannot emit constants
// Enum values for  are not proper Go identifiers, cannot emit constants

// Enum values for LocationSelectTypeEnum
const (
	LocationSelectTypeEnumLINE_SELECT     = "LINE_SELECT"
	LocationSelectTypeEnumMECH_SELECT     = "MECH_SELECT"
	LocationSelectTypeEnumLOCATION_SELECT = "LOCATION_SELECT"
	LocationSelectTypeEnumGLOBAL          = "GLOBAL"
)

// Enum values for  are not proper Go identifiers, cannot emit constants

// Enum values for UtilityType
const (
	UtilityTypeSHIELD      = "SHIELD"
	UtilityTypeATTACKDRONE = "ATTACK DRONE"
	UtilityTypeREPAIRDRONE = "REPAIR DRONE"
	UtilityTypeANTIMISSILE = "ANTI MISSILE"
	UtilityTypeACCELERATOR = "ACCELERATOR"
)

// Enum values for  are not proper Go identifiers, cannot emit constants
// Enum values for  are not proper Go identifiers, cannot emit constants

// Enum values for DamageType
const (
	DamageTypeKinetic   = "Kinetic"
	DamageTypeEnergy    = "Energy"
	DamageTypeExplosive = "Explosive"
)

// Enum values for ChatMSGTypeEnum
const (
	ChatMSGTypeEnumTEXT        = "TEXT"
	ChatMSGTypeEnumPUNISH_VOTE = "PUNISH_VOTE"
	ChatMSGTypeEnumSYSTEM_BAN  = "SYSTEM_BAN"
	ChatMSGTypeEnumNEW_BATTLE  = "NEW_BATTLE"
)

// Enum values for  are not proper Go identifiers, cannot emit constants

// Enum values for ItemType
const (
	ItemTypeUtility       = "utility"
	ItemTypeWeapon        = "weapon"
	ItemTypeMech          = "mech"
	ItemTypeMechSkin      = "mech_skin"
	ItemTypeMechAnimation = "mech_animation"
	ItemTypePowerCore     = "power_core"
	ItemTypeMysteryCrate  = "mystery_crate"
	ItemTypeWeaponSkin    = "weapon_skin"
)

// Enum values for CouponItemType
const (
	CouponItemTypeSUPS         = "SUPS"
	CouponItemTypeWEAPON_CRATE = "WEAPON_CRATE"
	CouponItemTypeMECH_CRATE   = "MECH_CRATE"
	CouponItemTypeGENESIS_MECH = "GENESIS_MECH"
)

// Enum values for FeatureName
const (
	FeatureNameMECH_MOVE       = "MECH_MOVE"
	FeatureNamePLAYER_ABILITY  = "PLAYER_ABILITY"
	FeatureNamePUBLIC_PROFILE  = "PUBLIC_PROFILE"
	FeatureNameSYSTEM_MESSAGES = "SYSTEM_MESSAGES"
	FeatureNameCHAT_BAN        = "CHAT_BAN"
	FeatureNamePROFILE_AVATAR  = "PROFILE_AVATAR"
)

// Enum values for AbilityLevel
const (
	AbilityLevelMECH    = "MECH"
	AbilityLevelFACTION = "FACTION"
	AbilityLevelPLAYER  = "PLAYER"
)

// Enum values for MarketplaceEvent
const (
	MarketplaceEventBid       = "bid"
	MarketplaceEventBidRefund = "bid_refund"
	MarketplaceEventPurchase  = "purchase"
	MarketplaceEventCreated   = "created"
	MarketplaceEventSold      = "sold"
	MarketplaceEventCancelled = "cancelled"
)

// Enum values for MechType
const (
	MechTypeHUMANOID = "HUMANOID"
	MechTypePLATFORM = "PLATFORM"
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

// Enum values for CrateType
const (
	CrateTypeMECH   = "MECH"
	CrateTypeWEAPON = "WEAPON"
)

// Enum values for TemplateItemType
const (
	TemplateItemTypeMECH           = "MECH"
	TemplateItemTypeMECH_ANIMATION = "MECH_ANIMATION"
	TemplateItemTypeMECH_SKIN      = "MECH_SKIN"
	TemplateItemTypeUTILITY        = "UTILITY"
	TemplateItemTypeWEAPON         = "WEAPON"
	TemplateItemTypeAMMO           = "AMMO"
	TemplateItemTypePOWER_CORE     = "POWER_CORE"
	TemplateItemTypeWEAPON_SKIN    = "WEAPON_SKIN"
	TemplateItemTypePLAYER_ABILITY = "PLAYER_ABILITY"
)

// Enum values for BanFromType
const (
	BanFromTypeSYSTEM = "SYSTEM"
	BanFromTypeADMIN  = "ADMIN"
	BanFromTypePLAYER = "PLAYER"
)

// Enum values for PlayerRankEnum
const (
	PlayerRankEnumGENERAL     = "GENERAL"
	PlayerRankEnumCORPORAL    = "CORPORAL"
	PlayerRankEnumPRIVATE     = "PRIVATE"
	PlayerRankEnumNEW_RECRUIT = "NEW_RECRUIT"
)

// Enum values for RepairTriggerWithType
const (
	RepairTriggerWithTypeSPACE_BAR  = "SPACE_BAR"
	RepairTriggerWithTypeLEFT_CLICK = "LEFT_CLICK"
	RepairTriggerWithTypeTOUCH      = "TOUCH"
	RepairTriggerWithTypeNONE       = "NONE"
)

// Enum values for RepairAgentFinishReason
const (
	RepairAgentFinishReasonABANDONED = "ABANDONED"
	RepairAgentFinishReasonEXPIRED   = "EXPIRED"
	RepairAgentFinishReasonSUCCEEDED = "SUCCEEDED"
)

// Enum values for RepairFinishReason
const (
	RepairFinishReasonEXPIRED   = "EXPIRED"
	RepairFinishReasonSTOPPED   = "STOPPED"
	RepairFinishReasonSUCCEEDED = "SUCCEEDED"
)

// Enum values for SyndicateElectionType
const (
	SyndicateElectionTypeADMIN = "ADMIN"
	SyndicateElectionTypeCEO   = "CEO"
)

// Enum values for SyndicateElectionResult
const (
	SyndicateElectionResultWINNER_APPEAR   = "WINNER_APPEAR"
	SyndicateElectionResultTIE             = "TIE"
	SyndicateElectionResultTIE_SECOND_TIME = "TIE_SECOND_TIME"
	SyndicateElectionResultNO_VOTE         = "NO_VOTE"
	SyndicateElectionResultNO_CANDIDATE    = "NO_CANDIDATE"
	SyndicateElectionResultTERMINATED      = "TERMINATED"
)

// Enum values for SyndicateJoinApplicationResult
const (
	SyndicateJoinApplicationResultACCEPTED   = "ACCEPTED"
	SyndicateJoinApplicationResultREJECTED   = "REJECTED"
	SyndicateJoinApplicationResultTERMINATED = "TERMINATED"
)

// Enum values for SyndicateMotionType
const (
	SyndicateMotionTypeCHANGE_GENERAL_DETAIL = "CHANGE_GENERAL_DETAIL"
	SyndicateMotionTypeCHANGE_ENTRY_FEE      = "CHANGE_ENTRY_FEE"
	SyndicateMotionTypeCHANGE_MONTHLY_DUES   = "CHANGE_MONTHLY_DUES"
	SyndicateMotionTypeCHANGE_BATTLE_WIN_CUT = "CHANGE_BATTLE_WIN_CUT"
	SyndicateMotionTypeADD_RULE              = "ADD_RULE"
	SyndicateMotionTypeREMOVE_RULE           = "REMOVE_RULE"
	SyndicateMotionTypeCHANGE_RULE           = "CHANGE_RULE"
	SyndicateMotionTypeREMOVE_MEMBER         = "REMOVE_MEMBER"
	SyndicateMotionTypeAPPOINT_COMMITTEE     = "APPOINT_COMMITTEE"
	SyndicateMotionTypeREMOVE_COMMITTEE      = "REMOVE_COMMITTEE"
	SyndicateMotionTypeDEPOSE_ADMIN          = "DEPOSE_ADMIN"
	SyndicateMotionTypeAPPOINT_DIRECTOR      = "APPOINT_DIRECTOR"
	SyndicateMotionTypeREMOVE_DIRECTOR       = "REMOVE_DIRECTOR"
	SyndicateMotionTypeDEPOSE_CEO            = "DEPOSE_CEO"
)

// Enum values for SyndicateMotionResult
const (
	SyndicateMotionResultPASSED          = "PASSED"
	SyndicateMotionResultFAILED          = "FAILED"
	SyndicateMotionResultTERMINATED      = "TERMINATED"
	SyndicateMotionResultLEADER_ACCEPTED = "LEADER_ACCEPTED"
	SyndicateMotionResultLEADER_REJECTED = "LEADER_REJECTED"
)

// Enum values for QuestionnaireUsage
const (
	QuestionnaireUsageJOIN_REQUEST = "JOIN_REQUEST"
)

// Enum values for QuestionnaireType
const (
	QuestionnaireTypeTEXT          = "TEXT"
	QuestionnaireTypeSINGLE_SELECT = "SINGLE_SELECT"
	QuestionnaireTypeMULTI_SELECT  = "MULTI_SELECT"
)

// Enum values for SyndicateType
const (
	SyndicateTypeCORPORATION   = "CORPORATION"
	SyndicateTypeDECENTRALISED = "DECENTRALISED"
)
