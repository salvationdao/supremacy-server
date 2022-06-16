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
	Type           string `db:"type" json:"type"`
	OwnershipID    string `db:"ownership_id" json:"ownership_id"`
	MechID         string `db:"mech_id" json:"mech_id,omitempty"`
	SkinID         string `db:"skin_id" json:"skin_id,omitempty"`
	MysteryCrateID string `db:"mystery_crate_id" json:"mystery_crate_id"`
	CanOpenOn      string `db:"can_open_on" json:"can_open_on,omitempty"`
}

func GetUserMechHangarItems(userID string) ([]*SiloType, error) {
	q := `
	SELECT 	ci.item_type    as type,
			ci.id           as ownership_id,
       		m.blueprint_id  as mech_id,
       		ms.blueprint_id as skin_id
	FROM collection_items ci
         	INNER JOIN mechs m on
    			m.id = ci.item_id
         	INNER JOIN mech_skin ms on
        		ms.id = coalesce(
            			m.chassis_skin_id,
            			(select default_chassis_skin_id from mech_models mm where mm.id = m.model_id)
        				)
	WHERE ci.owner_id = $1
  	AND (ci.item_type = 'mech');
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
	SELECT 	ci.item_type		 	as type,
			smc.id    				as ownership_id,
			smc.id 					as mystery_crate_id,
			mc.locked_until        	as can_open_on
	FROM 	collection_items ci
         	INNER JOIN mystery_crate mc on
    			mc.id = ci.item_id
         	INNER JOIN storefront_mystery_crates smc on
            	smc.mystery_crate_type = mc."type"
        	AND smc.faction_id = mc.faction_id
	WHERE ci.owner_id = $1
  			AND ci.item_type = 'mystery_crate';
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
		err := rows.Scan(&mst.Type, &mst.OwnershipID, &mst.MysteryCrateID, &canOpenOn)
		if err != nil {
			return nil, terror.Error(err, "failed to scan rows")
		}

		mst.CanOpenOn = canOpenOn.Format(time.UnixDate)

		mechSiloType = append(mechSiloType, mst)
	}

	return mechSiloType, nil
}
