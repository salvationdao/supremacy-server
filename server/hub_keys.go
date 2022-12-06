package server

// player

const HubKeyPlayerMarketingPreferencesUpdate = "PLAYER:MARKETING:UPDATE"

// player_abilities

const HubKeyPlayerQueueStatus = "PLAYER:QUEUE:STATUS"
const HubKeyPlayerAbilitiesList = "PLAYER:ABILITIES:LIST:SUBSCRIBE"
const HubKeyPlayerSupportAbilities = "PLAYER:SUPPORT:ABILITIES"
const HubKeyMechMoveCommandSubscribe = "MECH:MOVE:COMMAND:SUBSCRIBE"

const HubKeySaleAbilitiesPriceSubscribe = "SALE:ABILITIES:PRICE:SUBSCRIBE"
const HubKeySaleAbilitiesListSubscribe = "SALE:ABILITIES:LIST:SUBSCRIBE"
const HubKeySaleAbilitiesList = "SALE:ABILITIES:LIST"
const HubKeySaleAbilityPurchase = "SALE:ABILITY:PURCHASE"

const HubKeyGlobalAnnouncementSubscribe = "GLOBAL_ANNOUNCEMENT:SUBSCRIBE"

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
const HubKeyRepairAgentAbandon = "REPAIR:AGENT:ABANDON"
const HubKeyMechRepairCase = "MECH:REPAIR:CASE"
const HubKeyMechActiveRepairOffer = "MECH:ACTIVE:REPAIR:OFFER"

// repair bay

const HubKeyMechRepairSlotInsert = "MECH:REPAIR:SLOT:INSERT"
const HubKeyMechRepairSlotRemove = "MECH:REPAIR:SLOT:REMOVE"
const HubKeyMechRepairSlotSwap = "MECH:REPAIR:SLOT:SWAP"
const HubKeyMechRepairSlots = "MECH:REPAIR:SLOTS"

const HubKeyNextRepairGameBlock = "NEXT:REPAIR:GAME:BLOCK"

const HubKeyTelegramShortcodeRegistered = "USER:TELEGRAM_SHORTCODE_REGISTERED"
const HubKeySystemMessageSend = "SYSTEM:MESSAGE:SEND"

const HubKeyPlayerQuestStats = "PLAYER:QUEST:STAT"
const HubKeyPlayerQuestProgressions = "PLAYER:QUEST:PROGRESSIONS"

// battle arena

const HubKeyArenaStatusSubscribe = "ARENA:STATUS:UPDATED"

const HubKeyChallengeFundSubscribe = "CHALLENGE:FUND"

const HubKeyBattleArenaListSubscribe = "BATTLE:ARENA:LIST"
const HubKeyBattleArenaClosedSubscribe = "BATTLE:ARENA:CLOSED"
const HubKeyNextBattleDetails = "NEXT:BATTLE:DETAILS"
const HubKeyBattleState = "BATTLE:STATE"

// store

const HubKeyMysteryCrateSubscribe = "STORE:MYSTERY:CRATE:SUBSCRIBE"

// fiat

const HubKeyShoppingCartExpired = "FIAT:SHOPPING_CART:EXPIRED"
const HubKeyShoppingCartUpdated = "FIAT:SHOPPING_CART:UPDATED"

const HubKeyMiniMapAbilityContentSubscribe = "MINI:MAP:ABILITY:CONTENT"
const HubKeyMechCommandUpdateSubscribe = "MECH:COMMAND:UPDATE"
const HubKeyFactionMechCommandUpdateSubscribe = "FACTION:MECH:COMMANDS:UPDATE"
const HubKeyMiniMapUpdateSubscribe = "MINIMAP:UPDATES:SUBSCRIBE"

// binary key
const (
	BinaryKeyWarMachineStats byte = 1
	BinaryKeyMiniMapEvents   byte = 2
)

// json binary

type JsonBinaryData struct {
	Key  string      `json:"key"`
	Data interface{} `json:"data"`
}

// voice streams

const HubKeyVoiceStreams = "PLAYER:VOICE:STREAM"
const HubKeyVoiceStreamsListeners = "PLAYER:VOICE:STREAM:LISTENERS"
const HubKeyVoiceStreamJoinFactionCommander = "VOICE:JOIN:FACTION:COMMANDER"
const HubKeyVoiceStreamLeaveFactionCommander = "VOICE:LEAVE:FACTION:COMMANDER"
const HubKeyVoiceStreamVoteKick = "VOICE:VOTE:KICK"
const HubKeyVoiceStreamConnect = "VOICE:STREAM:CONNECT"
const HubKeyVoiceStreamDisconnect = "VOICE:STREAM:DISCONNECT"
const HubKeyVoiceStreamGetListeners = "VOICE:STREAM:GET:LISTENERS"

// battle queue

const HubKeyPlayerBrowserAlert = "PLAYER:BROWSER:ALERT"

const HubKeyPlayerOwnedMechs = "PLAYER:OWNED:MECHS"
const HubKeyPlayerOwnedWeapons = "PLAYER:OWNED:WEAPONS"
const HubKeyPlayerOwnedMechSkins = "PLAYER:OWNED:MECH:SKINS"
const HubKeyPlayerOwnedWeaponSkins = "PLAYER:OWNED:WEAPON:SKINS"
const HubKeyPlayerOwnedMysteryCrates = "PLAYER:OWNED:MYSTERY:CRATES"
const HubKeyPlayerOwnedKeycards = "PLAYER:OWNED:KEYCARDS"
const HubKeyFactionStakedMechs = "FACTION:STAKED:MECHS"

const HubKeyBattleLobbyListUpdate = "BATTLE:LOBBY:LIST:UPDATE"
const HubKeyBattleLobbyUpdate = "BATTLE:LOBBY:UPDATE"
const HubKeyInvolvedBattleLobbyListUpdate = "INVOLVED:BATTLE:LOBBY:LIST:UPDATE"
const HubKeyPrivateBattleLobbyUpdate = "PRIVATE:BATTLE:LOBBY:UPDATE"
const HubKeyPlayerAssetMechQueueSubscribe = "PLAYER:ASSET:MECH:QUEUE:SUBSCRIBE"
const HubKeyBattleETAUpdate = "BATTLE:ETA:UPDATE"

// faction pass

const HubKeyFactionMostPopularStakedMech = "FACTION:MOST:POPULAR:STAKED:MECH"
const HubKeyFactionStakedMechCount = "FACTION:STAKED:MECH:COUNT"
const HubKeyFactionStakedMechInQueueCount = "FACTION:STAKED:MECH:IN:QUEUE:COUNT"
const HubKeyFactionStakedMechDamagedCount = "FACTION:STAKED:MECH:DAMAGED:COUNT"
const HubKeyFactionStakedMechBattleReadyCount = "FACTION:STAKED:MECH:BATTLE:READY:COUNT"
const HubKeyFactionStakedMechInBattleCount = "FACTION:STAKED:MECH:IN:BATTLE:COUNT"
const HubKeyFactionStakedMechBattledCount = "FACTION:STAKED:MECH:BATTLED:COUNT"
const HubKeyFactionStakedMechInRepairBay = "FACTION:STAKED:MECH:IN:REPAIR:BAY"
