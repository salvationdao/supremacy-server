package db

import (
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/db/boiler"
	"server/gamedb"
)

//GetPlayerFeatures finds all Features for a player
func GetPlayerFeatures(playerID string) (boiler.FeatureSlice, error) {
	features, err := boiler.Features(
		qm.Select("id",
			"type",
			"deleted_at",
			"updated_at",
			"created_at"),
		qm.InnerJoin("players_features pf on pf.feature_id = features.id"),
		qm.InnerJoin("players p on pf.player_id = p.id"),
		qm.Where("p.id = ?", playerID),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return features, nil
}

//CheckFeatures checks if a user has a matching feature for given task
func CheckFeatures(playerID string, featureID string) error {
	features, err := GetPlayerFeatures(playerID)
	if err != nil {
		return err
	}

	for i := range features {
		if features[i].ID == featureID {
			return nil
		}
	}

	return terror.Error(fmt.Errorf("player: %s does not have necessary feature", playerID), "You do not have the necessary feature to perform this action, try again or contact support.")
}
