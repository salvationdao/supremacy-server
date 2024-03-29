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

// Enum values for AbilityKillingPowerLevel
const (
	AbilityKillingPowerLevelDEADLY = "DEADLY"
	AbilityKillingPowerLevelNORMAL = "NORMAL"
	AbilityKillingPowerLevelNONE   = "NONE"
)

// Enum values for AbilityTriggerType
const (
	AbilityTriggerTypeBATTLE_ABILITY = "BATTLE_ABILITY"
	AbilityTriggerTypeMECH_ABILITY   = "MECH_ABILITY"
	AbilityTriggerTypePLAYER_ABILITY = "PLAYER_ABILITY"
)

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
	BattleEventStunned          = "stunned"
	BattleEventHacked           = "hacked"
)

// Enum values for RecordingStatus
const (
	RecordingStatusRECORDING = "RECORDING"
	RecordingStatusSTOPPED   = "STOPPED"
	RecordingStatusIDLE      = "IDLE"
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
	WeaponTypeRocketPods      = "Rocket Pods"
)

// Enum values for  are not proper Go identifiers, cannot emit constants
// Enum values for  are not proper Go identifiers, cannot emit constants
// Enum values for  are not proper Go identifiers, cannot emit constants
// Enum values for  are not proper Go identifiers, cannot emit constants
// Enum values for  are not proper Go identifiers, cannot emit constants

// Enum values for MechType
const (
	MechTypeHUMANOID = "HUMANOID"
	MechTypePLATFORM = "PLATFORM"
)

// Enum values for BoostStat
const (
	BoostStatMECH_HEALTH                 = "MECH_HEALTH"
	BoostStatMECH_SPEED                  = "MECH_SPEED"
	BoostStatSHIELD_REGEN                = "SHIELD_REGEN"
	BoostStatMECH_MAX_SHIELD             = "MECH_MAX_SHIELD"
	BoostStatMECH_SPRINT_SPREAD_MODIFIER = "MECH_SPRINT_SPREAD_MODIFIER"
	BoostStatMECH_WALK_SPEED_MODIFIER    = "MECH_WALK_SPEED_MODIFIER"
	BoostStatWEAPON_DAMAGE_FALLOFF       = "WEAPON_DAMAGE_FALLOFF"
	BoostStatWEAPON_SPREAD               = "WEAPON_SPREAD"
)

// Enum values for PowercoreSize
const (
	PowercoreSizeSMALL  = "SMALL"
	PowercoreSizeMEDIUM = "MEDIUM"
	PowercoreSizeLARGE  = "LARGE"
	PowercoreSizeTURBO  = "TURBO"
)

// Enum values for  are not proper Go identifiers, cannot emit constants
// Enum values for  are not proper Go identifiers, cannot emit constants

// Enum values for LocationSelectTypeEnum
const (
	LocationSelectTypeEnumLINE_SELECT          = "LINE_SELECT"
	LocationSelectTypeEnumMECH_SELECT          = "MECH_SELECT"
	LocationSelectTypeEnumLOCATION_SELECT      = "LOCATION_SELECT"
	LocationSelectTypeEnumGLOBAL               = "GLOBAL"
	LocationSelectTypeEnumMECH_SELECT_ALLIED   = "MECH_SELECT_ALLIED"
	LocationSelectTypeEnumMECH_SELECT_OPPONENT = "MECH_SELECT_OPPONENT"
)

// Enum values for MiniMapDisplayEffectType
const (
	MiniMapDisplayEffectTypeNONE        = "NONE"
	MiniMapDisplayEffectTypeRANGE       = "RANGE"
	MiniMapDisplayEffectTypeMECH_PULSE  = "MECH_PULSE"
	MiniMapDisplayEffectTypeMECH_BORDER = "MECH_BORDER"
	MiniMapDisplayEffectTypePULSE       = "PULSE"
	MiniMapDisplayEffectTypeBORDER      = "BORDER"
	MiniMapDisplayEffectTypeDROP        = "DROP"
	MiniMapDisplayEffectTypeSHAKE       = "SHAKE"
)

// Enum values for  are not proper Go identifiers, cannot emit constants

// Enum values for QuestEventType
const (
	QuestEventTypeDailyQuest     = "daily_quest"
	QuestEventTypeWeeklyQuest    = "weekly_quest"
	QuestEventTypeMonthlyQuest   = "monthly_quest"
	QuestEventTypeProvingGrounds = "proving_grounds"
)

// Enum values for QuestKey
const (
	QuestKeyAbilityKill                  = "ability_kill"
	QuestKeyMechKill                     = "mech_kill"
	QuestKeyTotalBattleUsedMechCommander = "total_battle_used_mech_commander"
	QuestKeyRepairForOther               = "repair_for_other"
	QuestKeyChatSent                     = "chat_sent"
	QuestKeyMechJoinBattle               = "mech_join_battle"
)

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

// Enum values for  are not proper Go identifiers, cannot emit constants

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

// Enum values for PaymentMethods
const (
	PaymentMethodsSups   = "sups"
	PaymentMethodsStripe = "stripe"
	PaymentMethodsEth    = "eth"
	PaymentMethodsUsd    = "usd"
)

// Enum values for FeatureName
const (
	FeatureNameMECH_MOVE       = "MECH_MOVE"
	FeatureNamePLAYER_ABILITY  = "PLAYER_ABILITY"
	FeatureNamePUBLIC_PROFILE  = "PUBLIC_PROFILE"
	FeatureNameSYSTEM_MESSAGES = "SYSTEM_MESSAGES"
	FeatureNameCHAT_BAN        = "CHAT_BAN"
	FeatureNamePROFILE_AVATAR  = "PROFILE_AVATAR"
	FeatureNameVOICE_CHAT      = "VOICE_CHAT"
)

// Enum values for FiatProductItemTypes
const (
	FiatProductItemTypesMechPackage   = "mech_package"
	FiatProductItemTypesWeaponPackage = "weapon_package"
	FiatProductItemTypesSingleItem    = "single_item"
)

// Enum values for FiatProductTypes
const (
	FiatProductTypesStarterPackage = "starter_package"
	FiatProductTypesMysteryCrate   = "mystery_crate"
	FiatProductTypesMechSkin       = "mech_skin"
	FiatProductTypesWeaponSkin     = "weapon_skin"
	FiatProductTypesMechAnimation  = "mech_animation"
	FiatProductTypesMech           = "mech"
	FiatProductTypesWeapon         = "weapon"
)

// Enum values for AbilityLevel
const (
	AbilityLevelMECH    = "MECH"
	AbilityLevelFACTION = "FACTION"
	AbilityLevelPLAYER  = "PLAYER"
)

// Enum values for AvatarLayer
const (
	AvatarLayerHAIR      = "HAIR"
	AvatarLayerFACE      = "FACE"
	AvatarLayerBODY      = "BODY"
	AvatarLayerACCESSORY = "ACCESSORY"
	AvatarLayerEYEWEAR   = "EYEWEAR"
	AvatarLayerHELMET    = "HELMET"
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

// Enum values for ModActionType
const (
	ModActionTypeBAN         = "BAN"
	ModActionTypeUNBAN       = "UNBAN"
	ModActionTypeRESTART     = "RESTART"
	ModActionTypeLOOKUP      = "LOOKUP"
	ModActionTypeMECH_RENAME = "MECH_RENAME"
	ModActionTypeUSER_RENAME = "USER_RENAME"
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

// Enum values for OrderStatuses
const (
	OrderStatusesPending   = "pending"
	OrderStatusesCompleted = "completed"
	OrderStatusesRefunded  = "refunded"
)

// Enum values for BanFromType
const (
	BanFromTypeSYSTEM = "SYSTEM"
	BanFromTypeADMIN  = "ADMIN"
	BanFromTypePLAYER = "PLAYER"
)

// Enum values for RepairSlotStatus
const (
	RepairSlotStatusREPAIRING = "REPAIRING"
	RepairSlotStatusPENDING   = "PENDING"
	RepairSlotStatusDONE      = "DONE"
)

// Enum values for PlayerRankEnum
const (
	PlayerRankEnumGENERAL     = "GENERAL"
	PlayerRankEnumCORPORAL    = "CORPORAL"
	PlayerRankEnumPRIVATE     = "PRIVATE"
	PlayerRankEnumNEW_RECRUIT = "NEW_RECRUIT"
)

// Enum values for QuestEventDurationType
const (
	QuestEventDurationTypeDaily   = "daily"
	QuestEventDurationTypeWeekly  = "weekly"
	QuestEventDurationTypeMonthly = "monthly"
	QuestEventDurationTypeCustom  = "custom"
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

// Enum values for RepairGameBlockType
const (
	RepairGameBlockTypeNORMAL = "NORMAL"
	RepairGameBlockTypeSHRINK = "SHRINK"
	RepairGameBlockTypeFAST   = "FAST"
	RepairGameBlockTypeBOMB   = "BOMB"
	RepairGameBlockTypeEND    = "END"
)

// Enum values for RepairGameBlockTriggerKey
const (
	RepairGameBlockTriggerKeyM        = "M"
	RepairGameBlockTriggerKeyN        = "N"
	RepairGameBlockTriggerKeySPACEBAR = "SPACEBAR"
)

// Enum values for RepairFinishReason
const (
	RepairFinishReasonEXPIRED   = "EXPIRED"
	RepairFinishReasonSTOPPED   = "STOPPED"
	RepairFinishReasonSUCCEEDED = "SUCCEEDED"
)

// Enum values for RoleName
const (
	RoleNamePLAYER    = "PLAYER"
	RoleNameMODERATOR = "MODERATOR"
	RoleNameADMIN     = "ADMIN"
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

// Enum values for VoiceSenderType
const (
	VoiceSenderTypeMECH_OWNER        = "MECH_OWNER"
	VoiceSenderTypeFACTION_COMMANDER = "FACTION_COMMANDER"
)
