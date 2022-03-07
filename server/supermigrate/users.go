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
	if exists {
		player, err := boiler.FindPlayer(tx, data.ID)
		if err != nil {
			return false, false, err
		}
		player.FactionID = data.FactionID
		player.Username = null.NewString(data.Username, data.Username != "")
		_, err = player.Update(tx, boil.Whitelist(boiler.PlayerColumns.FactionID, boiler.PlayerColumns.Username))
		if err != nil {
			return false, false, err
		}
		return false, true, nil
	}

	addr := null.NewString("", false)
	if data.PublicAddress != "" {
		addr.String = common.HexToAddress(data.PublicAddress).Hex()
		addr.Valid = true
	}
	record := &boiler.Player{
		ID:            data.ID,
		FactionID:     data.FactionID,
		Username:      null.NewString(data.Username, data.Username != ""),
		PublicAddress: addr,
	}
	err = record.Insert(tx, boil.Infer())
	if err != nil {
		return false, false, fmt.Errorf("insert player: %w", err)
	}
	return false, false, nil
}
