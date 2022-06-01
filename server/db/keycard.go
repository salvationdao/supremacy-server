package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/db/boiler"
	"server/gamedb"
)

func CreateOrGetKeycard(ownerID string, tokenID int) (*boiler.PlayerKeycard, error) {
	keycard, err := boiler.PlayerKeycards(
		boiler.PlayerKeycardWhere.PlayerID.EQ(ownerID),
		qm.InnerJoin(
			fmt.Sprintf(`%s ON %s = %s AND %s = $1`,
				boiler.TableNames.BlueprintKeycards,
				qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.ID),
				qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.BlueprintKeycardID),
				qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.KeycardTokenID),
			),
			tokenID,
		),
	).One(gamedb.StdConn)
	if errors.Is(err, sql.ErrNoRows) {
		blueprint, err := boiler.BlueprintKeycards(boiler.BlueprintKeycardWhere.KeycardTokenID.EQ(tokenID)).One(gamedb.StdConn)
		if err != nil {
			return nil, err
		}

		newKeycard := &boiler.PlayerKeycard{
			PlayerID:           ownerID,
			BlueprintKeycardID: blueprint.ID,
			Count:              0,
		}

		err = newKeycard.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return nil, err
		}

		return newKeycard, nil
	}
	if err != nil {
		return nil, err
	}

	return keycard, nil
}
