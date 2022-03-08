package supermigrate

import (
	"database/sql"
	"fmt"
	"server/db/boiler"

	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func ProcessUser(tx *sql.Tx, data *UserPayload) (bool, bool, error) {
	exists, err := boiler.PlayerExists(tx, data.ID)
	if err != nil {
		return false, false, fmt.Errorf("check player exists: %w", err)
	}
	addr := common.HexToAddress(data.PublicAddress)
	addressMatchCount, err := boiler.Players(boiler.PlayerWhere.PublicAddress.EQ(null.StringFrom(addr.Hex()))).Count(tx)
	if err != nil {
		return false, false, fmt.Errorf("check player exists: %w", err)
	}
	if exists {
		player, err := boiler.FindPlayer(tx, data.ID)
		if err != nil {
			return false, false, fmt.Errorf("find existing player: %w", err)
		}
		player.FactionID = data.FactionID
		player.Username = null.NewString(data.Username, data.Username != "")
		_, err = player.Update(tx, boil.Whitelist(boiler.PlayerColumns.FactionID, boiler.PlayerColumns.Username))
		if err != nil {
			return false, false, fmt.Errorf("update existing player: %w", err)
		}
		return false, true, nil
	}

	if addressMatchCount > 0 {
		player, err := boiler.Players(boiler.PlayerWhere.PublicAddress.EQ(null.StringFrom(addr.Hex()))).One(tx)
		if err != nil {
			return false, false, fmt.Errorf("check player exists: %w", err)
		}
		player.FactionID = data.FactionID
		player.ID = data.ID
		player.Username = null.NewString(data.Username, data.Username != "")
		_, err = player.Update(tx, boil.Whitelist(boiler.PlayerColumns.ID, boiler.PlayerColumns.FactionID, boiler.PlayerColumns.Username))
		if err != nil {
			return false, false, fmt.Errorf("update player with matching address: %w", err)
		}
		return false, true, nil
	}

	addrValue := null.NewString(addr.Hex(), true)
	if addr == common.HexToAddress("") {
		addrValue.Valid = false
	}

	record := &boiler.Player{
		ID:            data.ID,
		FactionID:     data.FactionID,
		Username:      null.NewString(data.Username, data.Username != ""),
		PublicAddress: addrValue,
	}
	err = record.Insert(tx, boil.Infer())
	if err != nil {
		return false, false, fmt.Errorf("insert player %s %s %s: %w", data.ID, data.Username, data.PublicAddress, err)
	}
	return false, false, nil
}
