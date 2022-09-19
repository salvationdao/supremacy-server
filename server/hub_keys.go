package server

// player_abilities
const HubKeyPlayerAbilitiesList = "PLAYER:ABILITIES:LIST:SUBSCRIBE"
const HubKeyMechMoveCommandSubscribe = "MECH:MOVE:COMMAND:SUBSCRIBE"

const HubKeySaleAbilitiesPriceSubscribe = "SALE:ABILITIES:PRICE:SUBSCRIBE"
const HubKeySaleAbilitiesListSubscribe = "SALE:ABILITIES:LIST:SUBSCRIBE"
const HubKeySaleAbilitiesList = "SALE:ABILITIES:LIST"
const HubKeySaleAbilityClaim = "SALE:ABILITY:CLAIM"
const HubKeySaleAbilityPurchase = "SALE:ABILITY:PURCHASE"

const HubKeyGlobalAnnouncementSubscribe = "GLOBAL_ANNOUNCEMENT:SUBSCRIBE"

const WarMachineRepairStatusSubscribe = "WAR:MACHINE:REPAIR:STATUS:SUBSCRIBE"

const HubKeySyndicateGeneralDetailSubscribe = "SYNDICATE:GENERAL:DETAIL:SUBSCRIBE"
const HubKeySyndicateDirectorsSubscribe = "SYNDICATE:DIRECTORS:SUBSCRIBE"
const HubKeySyndicateCommitteesSubscribe = "SYNDICATE:COMMITTEES:SUBSCRIBE"
const HubKeySyndicateRulesSubscribe = "SYNDICATE:RULES:SUBSCRIBE"
const HubKeySyndicateOngoingMotionSubscribe = "SYNDICATE:ONGOING:MOTION:SUBSCRIBE"
const HubKeySyndicateOngoingElectionSubscribe = "SYNDICATE:ONGOING:ELECTION:SUBSCRIBE"

const HubKeyPlayerRankGet = "PLAYER:RANK:GET"
const HubKeyUserStatSubscribe = "USER:STAT:SUBSCRIBE"
const HubKeyUserSubscribe = "USER:SUBSCRIBE"
const HubKeySyndicateJoinApplicationUpdate = "SYNDICATE:JOIN:APPLICATION:UPDATE"

const HubKeySystemMessageList = "SYSTEM:MESSAGE:LIST"
const HubKeySystemMessageDismiss = "SYSTEM:MESSAGE:DISMISS"
const HubKeySystemMessageListUpdatedSubscribe = "SYSTEM:MESSAGE:LIST:UPDATED"

// repair job

const HubKeyRepairOfferUpdateSubscribe = "MECH:REPAIR:OFFER:LIST:UPDATE"
const HubKeyRepairOfferSubscribe = "MECH:REPAIR:OFFER"
const HubKeyRepairOfferIssue = "MECH:REPAIR:OFFER:ISSUE"
const HubKeyRepairOfferClose = "MECH:REPAIR:OFFER:CLOSE"
const HubKeyRepairAgentRegister = "REPAIR:AGENT:REGISTER"
const HubKeyRepairAgentRecord = "REPAIR:AGENT:RECORD"
const HubKeyRepairAgentComplete = "REPAIR:AGENT:COMPLETE"
const HubKeyRepairAgentAbandon = "REPAIR:AGENT:ABANDON"
const HubKeyMechRepairCase = "MECH:REPAIR:CASE"
const HubKeyMechActiveRepairOffer = "MECH:ACTIVE:REPAIR:OFFER"

// repair bay

const HubKeyMechRepairSlotInsert = "MECH:REPAIR:SLOT:INSERT"
const HubKeyMechRepairSlotRemove = "MECH:REPAIR:SLOT:REMOVE"
const HubKeyMechRepairSlotSwap = "MECH:REPAIR:SLOT:SWAP"
const HubKeyMechRepairSlots = "MECH:REPAIR:SLOTS"

const HubKeyTelegramShortcodeRegistered = "USER:TELEGRAM_SHORTCODE_REGISTERED"
const HubKeySystemMessageSend = "SYSTEM:MESSAGE:SEND"

const HubKeyPlayerQuestStats = "PLAYER:QUEST:STAT"
const HubKeyPlayerQuestProgressions = "PLAYER:QUEST:PROGRESSIONS"

const HubKeyVoiceStreams = "PLAYER:VOICE:STREAM"

// battle arena

const HubKeyArenaStatusSubscribe = "ARENA:STATUS:UPDATED"

const HubKeyChallengeFundSubscribe = "CHALLENGE:FUND"

const HubKeyBattleArenaListSubscribe = "BATTLE:ARENA:LIST"
const HubKeyBattleArenaClosedSubscribe = "BATTLE:ARENA:CLOSED"

// battle abilities

const HubKeyMiniMapAbilityDisplayList = "MINI:MAP:ABILITY:DISPLAY:LIST"
