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

const SaleAbilityPriceTickerIntervalSeconds KVKey = "sale_ability_price_ticker_interval_seconds"
const SaleAbilityReductionPercentage KVKey = "sale_ability_reduction_percentage"
const SaleAbilityFloorPrice KVKey = "sale_ability_floor_price"
const SaleAbilityLimit KVKey = "sale_ability_limit"
const SaleAbilityTimeBetweenRefreshSeconds KVKey = "sale_ability_time_between_refresh_seconds"
const SaleAbilityInflationPercentage KVKey = "sale_ability_inflation_percentage"
const QueueFeeFloor KVKey = "queue_fee_floor"
const RewardFeeFloor KVKey = "reward_fee_floor"

type KVKey string

// Default contributor formula https://www.desmos.com/calculator/vbfa5llasg
// KeyContributorMaxMultiplier = 3
// KeyContributorMinMultiplier = 0.5
// KeyContributorDecayMultiplier = 2
// KeyContributorSharpnessMultiplier = 0.02
const KeyContributorMaxMultiplier KVKey = "contributor_max_multiplier"
const KeyContributorMinMultiplier KVKey = "contributor_min_multiplier"
const KeyContributorDecayMultiplier KVKey = "contributor_decay_multiplier"
const KeyContributorSharpnessMultiplier KVKey = "contributor_sharpness_multiplier"
const KeyAbilityFloorPrice KVKey = "ability_floor_price"
const KeyBattleAbilityPriceDropRate KVKey = "battle_ability_price_drop_rate"
const KeyFactionAbilityFloorPrice KVKey = "faction_ability_floor_price"
const KeyFactionAbilityPriceDropRate KVKey = "faction_ability_price_drop_rate"

const KeyMarketplaceListingFee KVKey = "marketplace_listing_fee"
const KeyMarketplaceListingBuyoutFee KVKey = "marketplace_listing_buyout_fee"
const KeyMarketplaceListingAuctionReserveFee KVKey = "marketplace_listing_auction_reserve_fee"
const KeyMarketplaceSaleCutPercentageFee KVKey = "marketplace_sale_cut_percentage_fee"

const KeyFirstAbilityCooldown KVKey = "first_ability_cooldown"
const KeyBattleAbilityBribeDuration KVKey = "battle_ability_bribe_duration"
const KeyBattleAbilityLocationSelectDuration KVKey = "battle_ability_location_select_duration"
const KeyAbilityBroadcastRateMilliseconds KVKey = "ability_broadcast_rate_milliseconds"
const KeyPunishVoteCooldownHour KVKey = "punish_vote_cooldown_hour"

const KeyLastTransferEventID KVKey = "last_transfer_event_id"

const KeyInstantPassRequiredAmount KVKey = "instant_pass_required_amount"

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
