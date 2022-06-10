package db

import (
	"database/sql"
	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/db/boiler"
	"server/gamedb"
	"time"
)

type SiloType struct {
	Type        string `db:"type" json:"type"`
	OwnershipID string `db:"ownership_id" json:"ownership_id"`
	MechID      string `db:"mech_id" json:"mech_id,omitempty"`
	SkinID      string `db:"skin_id" json:"skin_id,omitempty"`
	CanOpenOn   string `db:"can_open_on" json:"can_open_on,omitempty"`
}

func GetUserMechHangarItems(userID string) ([]*SiloType, error) {
	q := `
	SELECT 	
			ci.item_type as type,
			ci.id as ownership_id,
			m.model_id as mech_id,
			COALESCE(
				m.chassis_skin_id::TEXT,
				(SELECT mm.default_chassis_skin_id::TEXT FROM mech_models mm WHERE mm.id = m.model_id)
			) as skin_id
		FROM collection_items ci
		INNER JOIN mechs m ON m.id = ci.item_id
		WHERE ci.owner_id = $1 AND (ci.item_type = 'mech')
	`
	rows, err := boiler.NewQuery(qm.SQL(q, userID)).Query(gamedb.StdConn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*SiloType{}, nil
		}
		return nil, terror.Error(err, "failed to query for finding silos")
	}

	mechSiloType := make([]*SiloType, 0)
	defer rows.Close()
	for rows.Next() {
		mst := &SiloType{}

		err := rows.Scan(&mst.Type, &mst.OwnershipID, &mst.MechID, &mst.SkinID)
		if err != nil {
			return nil, terror.Error(err, "failed to scan rows")
		}

		mechSiloType = append(mechSiloType, mst)
	}

	return mechSiloType, nil
}

func GetUserMysteryCrateHangarItems(userID string) ([]*SiloType, error) {
	q := `
	SELECT 	ci.id as ownership_id,
            ci.item_type as type,
			mc.locked_until as can_open_on
	FROM collection_items ci
	INNER JOIN mystery_crate mc ON mc.id = ci.item_id
	WHERE ci.owner_id = $1 AND (ci.item_type = 'mystery_crate')
	`
	rows, err := boiler.NewQuery(qm.SQL(q, userID)).Query(gamedb.StdConn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*SiloType{}, nil
		}
		return nil, terror.Error(err, "failed to query for finding silos")
	}

	mechSiloType := make([]*SiloType, 0)
	defer rows.Close()
	for rows.Next() {
		mst := &SiloType{}
		var canOpenOn time.Time
		err := rows.Scan(&mst.OwnershipID, &mst.Type, &canOpenOn)
		if err != nil {
			return nil, terror.Error(err, "failed to scan rows")
		}

		mst.CanOpenOn = canOpenOn.Format(time.UnixDate)

		mechSiloType = append(mechSiloType, mst)
	}

	return mechSiloType, nil
}
