package db

import (
	"errors"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type KVKey string

func get(key KVKey) string {
	exists, err := boiler.KVS(boiler.KVWhere.Key.EQ(string(key))).Exists(gamedb.StdConn)
	if err != nil {
		gamelog.L.Err(err).Str("key", string(key)).Msg("could not check kv exists")
		return ""
	}
	if !exists {
		gamelog.L.Err(errors.New("kv does not exist")).Str("key", string(key)).Msg("kv does not exist")
		return ""
	}
	kv, err := boiler.KVS(boiler.KVWhere.Key.EQ(string(key))).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Err(err).Str("key", string(key)).Msg("could not get kv")
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

func GetStr(key KVKey) string {
	return get(key)

}
func GetStrWithDefault(key KVKey, defaultValue string) string {
	vStr := get(key)
	if vStr == "" {
		PutStr(key, defaultValue)
		return defaultValue
	}

	return GetStr(key)
}
func PutStr(key KVKey, value string) {
	put(key, value)
}
func GetBool(key KVKey) bool {
	v := get(key)
	b, err := strconv.ParseBool(v)
	if err != nil {
		gamelog.L.Err(err).Str("key", string(key)).Str("val", v).Msg("could not parse boolean")
		return false
	}
	return b
}

func GetBoolWithDefault(key KVKey, defaultValue bool) bool {
	vStr := get(key)
	if vStr == "" {
		PutBool(key, defaultValue)
		return defaultValue
	}

	return GetBool(key)
}
func PutBool(key KVKey, value bool) {
	put(key, strconv.FormatBool(value))
}

func GetInt(key KVKey) int {
	vStr := get(key)
	v, err := strconv.Atoi(vStr)
	if err != nil {
		gamelog.L.Err(err).Str("key", string(key)).Str("val", vStr).Msg("could not parse int")
		return 0
	}
	return v
}

func GetIntWithDefault(key KVKey, defaultValue int) int {
	vStr := get(key)
	if vStr == "" {
		PutInt(key, defaultValue)
		return defaultValue
	}

	return GetInt(key)
}

func PutInt(key KVKey, value int) {
	put(key, strconv.Itoa(value))
}

func GetDecimal(key KVKey) decimal.Decimal {
	vStr := get(key)
	v, err := decimal.NewFromString(vStr)
	if err != nil {
		gamelog.L.Err(err).Str("key", string(key)).Str("val", vStr).Msg("could not parse decimal")
		return decimal.Zero
	}
	return v
}
func GetDecimalWithDefault(key KVKey, defaultValue decimal.Decimal) decimal.Decimal {
	vStr := get(key)

	if vStr == "" {
		PutDecimal(key, defaultValue)
		return defaultValue
	}
	return GetDecimal(key)
}

func PutDecimal(key KVKey, value decimal.Decimal) {
	put(key, value.String())
}
func GetTime(key KVKey) time.Time {
	vStr := get(key)
	t, err := time.Parse(time.RFC3339, vStr)
	if err != nil {
		gamelog.L.Err(err).Str("key", string(key)).Str("val", vStr).Msg("could not parse time")
		return time.Time{}
	}
	return t
}
func GetTimeWithDefault(key KVKey, defaultValue time.Time) time.Time {
	vStr := get(key)
	if vStr == "" {
		PutTime(key, defaultValue)
		return defaultValue
	}

	return GetTime(key)
}
func PutTime(key KVKey, value time.Time) {
	put(key, value.Format(time.RFC3339))
}