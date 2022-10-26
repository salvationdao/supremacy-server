package db

import (
	"database/sql"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"strconv"
	"time"

	"github.com/friendsofgo/errors"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type KVKey string

const KeyProdReopeningDate = "prod_reopening_date"

const KeyQueueTickerIntervalSeconds = "queue_ticker_interval_seconds"

const KeySaleAbilityFloorPrice KVKey = "sale_ability_floor_price"
const KeySaleAbilityReductionPercentage KVKey = "sale_ability_reduction_percentage"
const KeySaleAbilityInflationPercentage KVKey = "sale_ability_inflation_percentage"

const KeySaleAbilityTimeBetweenRefreshSeconds KVKey = "sale_ability_time_between_refresh_seconds"
const KeySaleAbilityPurchaseLimit KVKey = "sale_ability_purchase_limit"
const KeySaleAbilityPriceTickerIntervalSeconds KVKey = "sale_ability_price_ticker_interval_seconds"
const KeySaleAbilityLimit KVKey = "sale_ability_limit"

const KeyPlayerQueueLimit KVKey = "player_queue_limit"

const KeyPlayerAbilityMechMoveCommandCooldownSeconds KVKey = "player_ability_mech_move_command_cooldown_seconds"
const KeyPlayerAbilityMiniMechMoveCommandCooldownSeconds KVKey = "player_ability_mini_mech_move_command_cooldown_seconds"
const KeyPlayerAbilityIncognitoDurationSeconds KVKey = "player_ability_ability_incognito_duration_seconds"

const KeyMarketplaceListingFee KVKey = "marketplace_listing_fee"
const KeyMarketplaceListingBuyoutFee KVKey = "marketplace_listing_buyout_fee"
const KeyMarketplaceListingAuctionReserveFee KVKey = "marketplace_listing_auction_reserve_fee"
const KeyMarketplaceSaleCutPercentageFee KVKey = "marketplace_sale_cut_percentage_fee"

const KeyBattleAbilityBribeDuration KVKey = "battle_ability_bribe_duration"
const KeyBattleAbilityLocationSelectDuration KVKey = "battle_ability_location_select_duration"
const KeyAdvanceBattleAbilityShowUpUntilSeconds KVKey = "advance_battle_ability_show_up_until_seconds"
const KeyAdvanceBattleAbilityLabel KVKey = "advance_battle_ability_label"
const KeyFirstBattleAbilityLabel KVKey = "first_battle_ability_label"
const KeyPunishVoteCooldownHour KVKey = "punish_vote_cooldown_hour"
const KeyPreBattleTimeSeconds KVKey = "pre_battle_time_seconds"

const KeyLastTransferEventID KVKey = "last_transfer_event_id"

const KeyInstantPassRequiredAmount KVKey = "instant_pass_required_amount"
const KeyJudgingCountdownSeconds KVKey = "system_ban_judging_countdown_seconds"
const KeySystemBanTeamKillDefaultReason KVKey = "system_ban_team_kill_default_reason"
const KeySystemBanTeamKillBanBaseDurationHours KVKey = "system_ban_team_kill_ban_base_duration_hours"
const KeySystemBanTeamKillBanDurationMultiplier KVKey = "system_ban_team_kill_ban_duration_multiplier"
const KeySystemBanTeamKillPermanentBanBottomLineHours KVKey = "system_ban_team_kill_permanent_ban_bottom_line_hours"
const KeyRepairMiniGameFailedRate KVKey = "repair_mini_game_failed_rate"

const KeyMechAbilityCoolDownSeconds KVKey = "mech_ability_cool_down_seconds"
const KeyRequiredRepairStacks KVKey = "required_repair_stacks"
const KeyBattleQueueFee KVKey = "battle_queue_fee"
const KeyDefaultRepairBlocks KVKey = "default_repair_blocks"
const KeyBattleRewardTaxRatio KVKey = "battle_reward_tax_ratio"
const KeyFirstRankFactionRewardSups KVKey = "first_rank_faction_reward_sups"
const KeySecondRankFactionRewardSups KVKey = "second_rank_faction_reward_sups"
const KeyThirdRankFactionRewardSups KVKey = "third_rank_faction_reward_sups"
const KeyBattleSupsRewardBonus KVKey = "battle_sups_reward_bonus"
const KeyCanDeployDamagedRatio KVKey = "can_deploy_damaged_ratio"

const KeyDecentralisedAutonomousSyndicateTax KVKey = "decentralised_autonomous_syndicate_tax"
const KeyCorporationSyndicateTax KVKey = "corporation_syndicate_tax"

const KeyOvenmediaAPIBaseUrl KVKey = "ovenmedia_api_base_url"
const KeyOvenmediaVoiceStreamURL KVKey = "ovenmedia_stream_voice_base_url"
const KeyOvenmediaStreamURL KVKey = "ovenmedia_stream_base_url"
const KeyCanRecordReplayStatus KVKey = "can_record_replay"
const KeyVoiceExpiryTimeHours KVKey = "voice_expiry_time_hours"
const KeyVoiceBanTimeHours KVKey = "voice_ban_time_hours"

const KeySlackModChannelID KVKey = "slack_mod_channel_id"
const KeySlackRapiChannelID KVKey = "slack_rapid_channel_id"
const KeySlackDevChannelID KVKey = "slack_dev_channel_id"

const KeyAutoRepairSlotCount KVKey = "auto_repair_slot_count"
const KeyAutoRepairDurationSeconds KVKey = "auto_repair_duration_seconds"

const KeyMinimumMechActionCount KVKey = "minimum_mech_action_count"

const KeyFiatToSUPCut KVKey = "fiat_to_sup_cut" // TODO: find better name to describe: "20% cheaper than fiat pricing"

const KeyDefaultPublicLobbyCount KVKey = "default_public_lobby_count"
const KeyScheduledLobbyPrepareDurationSeconds KVKey = "scheduled_lobby_prepare_duration_seconds"
const KeyOpenNewArenaEveryXAmountOfBattleLobbies KVKey = "open_new_arena_after_x_amount_of_battle_lobbies"

const KeyDeductBlockCountFromBomb KVKey = "deduct_block_count_from_bomb"

const KeyAutoFillLobbyAfterDurationSecond KVKey = "auto_fill_lobby_after_duration_second"
const KeyPublicExhibitionLobbyExpireAfterDurationSecond KVKey = "public_exhibition_lobby_expire_after_duration_second"

func get(key KVKey) string {
	kv, err := boiler.KVS(boiler.KVWhere.Key.EQ(string(key))).One(gamedb.StdConn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Err(errors.New("kv does not exist")).Str("key", string(key)).Msg("kv does not exist")
		} else {
			gamelog.L.Err(err).Str("key", string(key)).Msg("could not get kv")
		}
		return ""
	}
	return kv.Value
}

func put(key KVKey, value string) {
	kv := boiler.KV{
		Key:   string(key),
		Value: value,
	}
	err := kv.Upsert(gamedb.StdConn, true, []string{boiler.KVColumns.Key}, boil.Whitelist(boiler.KVColumns.Value), boil.Infer())
	if err != nil {
		gamelog.L.Err(err).Msg("could not put kv")
		return
	}
}

func GetStrWithDefault(key KVKey, defaultValue string) string {
	vStr := get(key)
	if vStr == "" {
		PutStr(key, defaultValue)
		return defaultValue
	}

	return vStr
}
func PutStr(key KVKey, value string) {
	put(key, value)
}

func GetBoolWithDefault(key KVKey, defaultValue bool) bool {
	vStr := get(key)
	if vStr == "" {
		PutBool(key, defaultValue)
		return defaultValue
	}

	b, err := strconv.ParseBool(vStr)
	if err != nil {
		gamelog.L.Err(err).Str("key", string(key)).Str("val", vStr).Msg("could not parse boolean")
		return false
	}
	return b
}
func PutBool(key KVKey, value bool) {
	put(key, strconv.FormatBool(value))
}

func GetIntWithDefault(key KVKey, defaultValue int) int {
	vStr := get(key)
	if vStr == "" {
		PutInt(key, defaultValue)
		return defaultValue
	}

	v, err := strconv.Atoi(vStr)
	if err != nil {
		gamelog.L.Err(err).Str("key", string(key)).Str("val", vStr).Msg("could not parse int")
		return 0
	}

	return v
}

func PutInt(key KVKey, value int) {
	put(key, strconv.Itoa(value))
}

func GetDecimalWithDefault(key KVKey, defaultValue decimal.Decimal) decimal.Decimal {
	vStr := get(key)

	if vStr == "" {
		PutDecimal(key, defaultValue)
		return defaultValue
	}

	v, err := decimal.NewFromString(vStr)
	if err != nil {
		gamelog.L.Err(err).Str("key", string(key)).Str("val", vStr).Msg("could not parse decimal")
		return decimal.Zero
	}
	return v
}

func PutDecimal(key KVKey, value decimal.Decimal) {
	put(key, value.String())
}

func GetTimeWithDefault(key KVKey, defaultValue time.Time) time.Time {
	vStr := get(key)
	if vStr == "" {
		PutTime(key, defaultValue)
		return defaultValue
	}

	t, err := time.Parse(time.RFC3339, vStr)
	if err != nil {
		gamelog.L.Err(err).Str("key", string(key)).Str("val", vStr).Msg("could not parse time")
		return time.Time{}
	}
	return t
}
func PutTime(key KVKey, value time.Time) {
	put(key, value.Format(time.RFC3339))
}
