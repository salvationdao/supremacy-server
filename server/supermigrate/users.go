package supermigrate

import (
	"database/sql"
	"fmt"
	"server/db/boiler"

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
		player.SyndicateID = data.FactionID
		_, err = player.Update(tx, boil.Infer())
		if err != nil {
			return false, false, err
		}
		return false, true, nil
	}

	addr := null.NewString("", false)
	if data.PublicAddress != "" {
		addr.String = data.PublicAddress
		addr.Valid = true
	}
	record := &boiler.Player{
		ID:            data.ID,
		SyndicateID:   data.FactionID,
		PublicAddress: addr,
	}
	err = record.Insert(tx, boil.Infer())
	if err != nil {
		return false, false, fmt.Errorf("insert player: %w", err)
	}
	return false, false, nil
}
