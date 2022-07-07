package db

import (
	"database/sql"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"time"
)

// GetAllFeatures gets all features in the features table
func GetAllFeatures() ([]*server.Feature, error) {
	features, err := boiler.Features().All(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err, "Could not get all features. try again or contact support.")
	}

	serverFeatures := server.FeaturesFromBoiler(features)
	return serverFeatures, nil
}

//GetPlayerFeaturesByID finds all Features for a player
func GetPlayerFeaturesByID(playerID string) (boiler.FeatureSlice, error) {
	features, err := boiler.Features(
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.PlayersFeatures,
			qm.Rels(boiler.TableNames.Features, boiler.FeatureColumns.Type),
			qm.Rels(boiler.TableNames.PlayersFeatures, boiler.PlayersFeatureColumns.FeatureType),
		)),
		qm.Where(fmt.Sprintf("%s = ?",
			qm.Rels(boiler.TableNames.PlayersFeatures, boiler.PlayersFeatureColumns.PlayerID),
		), playerID),
		qm.And(fmt.Sprintf("%s IS NULL",
			qm.Rels(boiler.TableNames.PlayersFeatures, boiler.PlayersFeatureColumns.DeletedAt),
		)),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return features, nil
}

func AddFeatureToPlayerIDs(featureType string, ids []string) error {
	featureExists, err := boiler.FeatureExists(gamedb.StdConn, featureType)
	if err != nil {
		return terror.Error(err, "Error finding if feature exists")
	}
	if !featureExists {
		return terror.Error(fmt.Errorf("feature: %s does not exist", featureType), fmt.Sprintf("feature: %s does not exist", featureType))

	}

	for _, id := range ids {
		playerExists, err := boiler.PlayerExists(gamedb.StdConn, id)
		if err != nil {
			return terror.Error(err, "Error finding if player exists")
		}

		if !playerExists {
			return terror.Error(fmt.Errorf("player: %s does not exist", id), fmt.Sprintf("player: %s does not exist", id))
		}

		playerFeatureExists, err := boiler.PlayersFeatures(
			boiler.PlayersFeatureWhere.PlayerID.EQ(id),
			boiler.PlayersFeatureWhere.FeatureType.EQ(featureType),
			boiler.PlayersFeatureWhere.DeletedAt.IsNull(),
		).Exists(gamedb.StdConn)
		if err != nil {
			return terror.Error(err, "Error finding if player feature exists")
		}
		if playerFeatureExists {
			break
		}

		pf := &boiler.PlayersFeature{
			PlayerID:    id,
			FeatureType: featureType,
		}

		err = pf.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return terror.Error(err, "Could not insert into player features")
		}
	}

	return nil
}

func AddFeatureToPublicAddresses(featureType string, addresses []string) error {
	featureExists, err := boiler.FeatureExists(gamedb.StdConn, featureType)
	if err != nil {
		return terror.Error(err, "Error finding if feature exists")
	}
	if !featureExists {
		return terror.Error(fmt.Errorf("feature: %s does not exist", featureType), fmt.Sprintf("feature: %s does not exist", featureType))

	}

	for _, address := range addresses {
		if address == "" {
			break
		}

		player, err := boiler.Players(
			boiler.PlayerWhere.PublicAddress.EQ(null.StringFrom(address)),
		).One(gamedb.StdConn)
		if err != nil {
			return terror.Error(err, "Error finding player")
		}

		playerFeatureExists, err := boiler.PlayersFeatures(
			boiler.PlayersFeatureWhere.PlayerID.EQ(player.ID),
			boiler.PlayersFeatureWhere.FeatureType.EQ(featureType),
			boiler.PlayersFeatureWhere.DeletedAt.IsNull(),
		).Exists(gamedb.StdConn)
		if err != nil {
			return terror.Error(err, "Error finding if player feature exists")
		}
		if playerFeatureExists {
			break
		}

		pf := &boiler.PlayersFeature{
			PlayerID:    player.ID,
			FeatureType: featureType,
		}

		err = pf.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return terror.Error(err, "Could not insert into player features")
		}
	}
	return nil
}

func RemoveFeatureFromPlayerIDs(featureType string, ids []string) error {
	featureExists, err := boiler.FeatureExists(gamedb.StdConn, featureType)
	if err != nil {
		return terror.Error(err, "Error finding if feature exists")
	}
	if !featureExists {
		return terror.Error(fmt.Errorf("feature: %s does not exist", featureType), fmt.Sprintf("feature: %s does not exist", featureType))

	}

	for _, id := range ids {
		pf, err := boiler.PlayersFeatures(
			boiler.PlayersFeatureWhere.PlayerID.EQ(id),
			boiler.PlayersFeatureWhere.FeatureType.EQ(featureType),
			boiler.PlayersFeatureWhere.DeletedAt.IsNull(),
		).One(gamedb.StdConn)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			return terror.Error(err, "Error finding player feature")
		}

		pf.DeletedAt = null.TimeFrom(time.Now())
		_, err = pf.Update(gamedb.StdConn, boil.Infer())
		if err != nil {
			return terror.Error(err, "Could not UPDATE deletedAt column in player features")
		}
	}

	return nil
}

func RemoveFeatureFromPublicAddresses(featureType string, addresses []string) error {
	featureExists, err := boiler.FeatureExists(gamedb.StdConn, featureType)
	if err != nil {
		return terror.Error(err, "Error finding if feature exists")
	}
	if !featureExists {
		return terror.Error(fmt.Errorf("feature: %s does not exist", featureType), fmt.Sprintf("feature: %s does not exist", featureType))

	}

	for _, address := range addresses {
		if address == "" {
			break
		}

		player, err := boiler.Players(
			boiler.PlayerWhere.PublicAddress.EQ(null.StringFrom(address)),
		).One(gamedb.StdConn)
		if err != nil {
			return terror.Error(err, "Error finding player")
		}

		pf, err := boiler.PlayersFeatures(
			boiler.PlayersFeatureWhere.PlayerID.EQ(player.ID),
			boiler.PlayersFeatureWhere.FeatureType.EQ(featureType),
			boiler.PlayersFeatureWhere.DeletedAt.IsNull(),
		).One(gamedb.StdConn)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			return terror.Error(err, "Error finding player feature")
		}

		pf.DeletedAt = null.TimeFrom(time.Now())
		_, err = pf.Update(gamedb.StdConn, boil.Infer())
		if err != nil {
			return terror.Error(err, "Could not UPDATE deletedAt column in player features")
		}
	}

	return nil
}
