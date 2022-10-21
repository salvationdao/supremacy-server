package battle

import (
	"fmt"
	"github.com/ninja-software/terror/v2"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
)

func stackedAIMechsCheck() error {
	l := gamelog.L.With().Str("func", "stackedAIMechsCheck").Logger()
	// get AI mechs
	q := fmt.Sprintf(`
		INSERT INTO %s (%s, %s, %s)
		SELECT %s, %s, %s
		FROM %s
		INNER JOIN %s ON %s = %s AND %s = TRUE AND %s NOTNULL
		WHERE %s = $1
		ON CONFLICT DO NOTHING;
	`,
		boiler.TableNames.StakedMechs,
		boiler.StakedMechColumns.MechID,
		boiler.StakedMechColumns.OwnerID,
		boiler.StakedMechColumns.FactionID,
		boiler.CollectionItemTableColumns.ItemID,
		boiler.CollectionItemTableColumns.OwnerID,
		boiler.PlayerTableColumns.FactionID,
		boiler.TableNames.CollectionItems,
		boiler.TableNames.Players,
		boiler.PlayerTableColumns.ID,
		boiler.CollectionItemTableColumns.OwnerID,
		boiler.PlayerTableColumns.IsAi,
		boiler.PlayerTableColumns.FactionID,
		boiler.CollectionItemTableColumns.ItemType,
	)

	_, err := gamedb.StdConn.Exec(q, boiler.ItemTypeMech)
	if err != nil {
		l.Error().Err(err).Msg("Failed to upsert AI mechs into stake pool")
		return terror.Error(err, "Failed to upsert AI mechs into stake pool")
	}

	return nil
}
