package supermigrate

import (
	"errors"
	"fmt"
	"server/gamedb"
	"server/gamelog"
)

var BrandMap = map[string]string{
	"Red Mountain":       "Red Mountain Offworld Mining Corporation",
	"Boston Cybernetics": "Boston Cybernetics",
	"Zaibatsu":           "Zaibatsu Heavy Industries",
}

func getMetadataFromHash(hash string, metadata []*MetadataPayload) (*MetadataPayload, error) {
	for _, datum := range metadata {
		if datum.Hash != hash {
			continue
		}
		return datum, nil
	}
	return nil, errors.New("can not find matching metadata")
}

func MigrateAssets(
	metadataPayload []*MetadataPayload,
	assetPayload []*AssetPayload,
	storePayload []*StorePayload,
	factionPayload []*FactionPayload,
	userPayload []*UserPayload,
) error {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	success := 0
	skipped := 0
	updated := 0
	for _, asset := range assetPayload {
		metadata, err := getMetadataFromHash(asset.MetadataHash, metadataPayload)
		if err != nil {
			return fmt.Errorf("get metadata: %w", err)
		}
		wasSkipped, wasUpdated, err := ProcessMech(tx, asset, metadata)
		if wasSkipped {
			skipped++
			continue
		}
		if wasUpdated {
			updated++
			continue
		}
		if err != nil {
			return fmt.Errorf("process template: %w", err)
		}
		success++

	}
	gamelog.L.Info().Int("success", success).Int("skipped", skipped).Int("updated", updated).Msg("finished asset migration")

	tx.Commit()
	return nil
}

func MigrateUsers(
	metadataPayload []*MetadataPayload,
	assetPayload []*AssetPayload,
	storePayload []*StorePayload,
	factionPayload []*FactionPayload,
	userPayload []*UserPayload,
) error {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return err
	}
	success := 0
	skipped := 0
	updated := 0
	defer tx.Rollback()
	for _, user := range userPayload {
		wasSkipped, wasUpdated, err := ProcessUser(tx, user)
		if wasUpdated {
			updated++
			continue
		}
		if wasSkipped {
			skipped++
			continue
		}
		if err != nil {
			return err
		}
		success++
	}
	tx.Commit()
	gamelog.L.Info().
		Int("success", success).
		Int("updated", updated).
		Int("skipped", skipped).
		Msg("finished user migration")
	return nil
}
