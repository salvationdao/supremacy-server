package db

import (
	"errors"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func MechsByOwnerID(ownerID uuid.UUID) ([]*server.Mech, error) {
	//// TODO: Vinnie fix this
	//mechs, err := boiler.Mechs(boiler.MechWhere.OwnerID.EQ(ownerID.String())).All(gamedb.StdConn)
	//if err != nil {
	//	return nil, err
	//}
	//result := []*server.Mech{}
	//for _, mech := range mechs {
	//	record, err := Mech(uuid.Must(uuid.FromString(mech.ID)))
	//	if err != nil {
	//		return nil, err
	//	}
	//	result = append(result, record)
	//}
	return nil, nil
}

func MechSetName(mechID uuid.UUID, name string) error {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	mech, err := boiler.FindMech(gamedb.StdConn, mechID.String())
	if err != nil {
		return err
	}
	mech.Name = name
	_, err = mech.Update(tx, boil.Whitelist(boiler.MechColumns.Name))
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func MechSetOwner(mechID uuid.UUID, ownerID uuid.UUID) error {
	// TODO: Vinnie fix this
	//tx, err := gamedb.StdConn.Begin()
	//if err != nil {
	//	return err
	//}
	//defer tx.Rollback()
	//mech, err := boiler.FindMech(tx, mechID.String())
	//if err != nil {
	//	return err
	//}
	//mech.OwnerID = ownerID.String()
	//_, err = mech.Update(tx, boil.Whitelist(boiler.MechColumns.OwnerID))
	//if err != nil {
	//	return err
	//}
	//tx.Commit()
	return nil
}

func TemplatePurchasedCount(templateID uuid.UUID) (int, error) {
	// TODO: Fix this, this is in the gameserver storefront refactor
	return 0, nil
}

func DefaultMechs() ([]*server.Mech, error) {
	idq := `SELECT id FROM mechs WHERE is_default=true`

	result, err := gamedb.StdConn.Query(idq)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to query default mechs")
		return nil, err
	}
	defer result.Close()

	ids := []uuid.UUID{}
	for result.Next() {
		id := ""
		err = result.Scan(&id)
		if err != nil {
			return nil, err
		}
		uid, err := uuid.FromString(id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, uid)
	}

	return Mechs(ids...)
}

var ErrNotAllMechsReturned = fmt.Errorf("not all mechs returned")

// Mech gets the whole mech object, all the parts but no part collection details. This should only be used when building a mech to pass into gameserver
// If you want to show the user a mech, it should be lazy loaded via various endpoints, not a single endpoint for an entire mech.
func Mech(mechID string) (*server.Mech, error) {
	mc := &server.Mech{
		CollectionDetails: &server.CollectionDetails{},
	}

	query := `
	SELECT
		ci.collection_slug,
		ci.hash,
		ci.token_id,
		ci.owner_id,
		ci.tier,
		m.id,
		m.name,
		m.label,
		m.weapon_hardpoints,
		m.utility_slots,
		m.speed,
		m.max_hitpoints,
		m.is_default,
		m.is_insured,
		m.genesis_token_id,
		m.limited_release_token_id,
		m.power_core_size,
		m.blueprint_id,
	
		m.brand_id,
		to_json(b) as brand,
	
		to_json(p) as owner,
	
		p.faction_id,
		to_json(f) as faction,
	
		m.model_id,
		to_json(mm) as model,
	
		mm.default_chassis_skin_id,
		to_json(dms) as default_chassis_skin,
	
		m.chassis_skin_id,
		to_json(ms) as chassis_skin,
	
		m.intro_animation_id,
		to_json(ma2) as intro_animation,
	
		m.outro_animation_id,
		to_json(ma1) as outro_animation,
	
		m.power_core_id,
		to_json(ec) as power_core,
	
		w.weapons,
		u.utility
	FROM mechs m
	INNER JOIN collection_items ci on ci.item_id = m.id
	INNER JOIN players p ON p.id = ci.owner_id
	INNER JOIN factions f on p.faction_id = f.id
	LEFT OUTER JOIN power_cores ec ON ec.id = m.power_core_id
	LEFT OUTER JOIN brands b ON b.id = m.brand_id
	LEFT OUTER JOIN mech_model mm ON m.model_id = mm.id
	LEFT OUTER JOIN (
		SELECT _ms.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id
		FROM mech_skin _ms
		INNER JOIN collection_items _ci on _ci.item_id = _ms.id
	) ms ON m.chassis_skin_id = ms.id
			 LEFT OUTER JOIN blueprint_mech_skin dms ON mm.default_chassis_skin_id = dms.id
		-- LEFT OUTER JOIN mech_animation ma1 on ma1.id = m.outro_animation_id
			 LEFT OUTER JOIN (
		SELECT _ma.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id
		FROM mech_animation _ma
				 INNER JOIN collection_items _ci on _ci.item_id = _ma.id
	) ma1 on ma1.id = m.outro_animation_id
		-- LEFT OUTER JOIN mech_animation ma2 on ma2.id = m.intro_animation_id
			 LEFT OUTER JOIN (
		SELECT _ma.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id
		FROM mech_animation _ma
				 INNER JOIN collection_items _ci on _ci.item_id = _ma.id
	) ma2 on ma2.id = m.intro_animation_id
			 LEFT OUTER JOIN (
		SELECT mw.chassis_id, json_agg(w2) as weapons
		FROM mech_weapons mw
				 INNER JOIN
			 (
				 SELECT _w.*, _ci.hash, _ci.token_id, _ci.tier, _ci.owner_id
				 FROM weapons _w
						  INNER JOIN collection_items _ci on _ci.item_id = _w.id
			 ) w2 ON mw.weapon_id = w2.id
		GROUP BY mw.chassis_id
	) w on w.chassis_id = m.id
			 LEFT OUTER JOIN (
		SELECT mw.chassis_id, json_agg(_u) as utility
		FROM mech_utility mw
				 INNER JOIN (
			SELECT
				_u.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id,
				to_json(_us) as shield,
				to_json(_ua) as accelerator,
				to_json(_uam) as attack_drone,
				to_json(_uad) as anti_missile,
				to_json(_urd) as repair_drone
			FROM utility _u
					 INNER JOIN collection_items _ci on _ci.item_id = _u.id
					 LEFT OUTER JOIN utility_shield _us ON _us.utility_id = _u.id
					 LEFT OUTER JOIN utility_accelerator _ua ON _ua.utility_id = _u.id
					 LEFT OUTER JOIN utility_anti_missile _uam ON _uam.utility_id = _u.id
					 LEFT OUTER JOIN utility_attack_drone _uad ON _uad.utility_id = _u.id
					 LEFT OUTER JOIN utility_repair_drone _urd ON _urd.utility_id = _u.id
		) _u ON mw.utility_id = _u.id
		GROUP BY mw.chassis_id
	) u on u.chassis_id = m.id
		WHERE m.id = $1
		`

	result, err := gamedb.StdConn.Query(query, mechID)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	for result.Next() {
		err = result.Scan(
			&mc.CollectionDetails.CollectionSlug,
			&mc.CollectionDetails.Hash,
			&mc.CollectionDetails.TokenID,
			&mc.CollectionDetails.OwnerID,
			&mc.CollectionDetails.Tier,
			&mc.ID,
			&mc.Name,
			&mc.Label,
			&mc.WeaponHardpoints,
			&mc.UtilitySlots,
			&mc.Speed,
			&mc.MaxHitpoints,
			&mc.IsDefault,
			&mc.IsInsured,
			&mc.GenesisTokenID,
			&mc.LimitedReleaseTokenID,
			&mc.PowerCoreSize,
			&mc.BlueprintID,
			&mc.BrandID,
			&mc.Brand,
			&mc.Owner,
			&mc.FactionID,
			&mc.Faction,
			&mc.ModelID,
			&mc.Model,
			&mc.DefaultChassisSkinID,
			&mc.DefaultChassisSkin,
			&mc.ChassisSkinID,
			&mc.ChassisSkin,
			&mc.IntroAnimationID,
			&mc.IntroAnimation,
			&mc.OutroAnimationID,
			&mc.OutroAnimation,
			&mc.PowerCoreID,
			&mc.PowerCore,
			&mc.Weapons,
			&mc.Utility,
		)
		if err != nil {
			return nil, err
		}
	}

	if mc.ID == "" {
		return nil, fmt.Errorf("unable to find mech with id %s", mechID)
	}

	return mc, err
}

func Mechs(mechIDs ...uuid.UUID) ([]*server.Mech, error) {
	if len(mechIDs) == 0 {
		return nil, errors.New("no mech ids provided")
	}

	mcs := make([]*server.Mech, len(mechIDs))

	mechids := make([]interface{}, len(mechIDs))
	var paramrefs string
	for i, id := range mechIDs {
		paramrefs += `$` + strconv.Itoa(i+1) + `,`
		mechids[i] = id.String()
	}
	paramrefs = paramrefs[:len(paramrefs)-1]

	query := `
		SELECT
			ci.collection_slug,
			ci.hash,
			ci.token_id,
			ci.owner_id,
			ci.tier,
			m.id,
			m.name,
			m.label,
			m.weapon_hardpoints,
			m.utility_slots,
			m.speed,
			m.max_hitpoints,
			m.is_default,
			m.is_insured,
			m.genesis_token_id,
			m.limited_release_token_id,
			m.power_core_size,
			m.blueprint_id,
		
			m.brand_id,
			to_json(b) as brand,
		
			to_json(p) as owner,
		
			p.faction_id,
			to_json(f) as faction,
		
			m.model_id,
			to_json(mm) as model,
		
			mm.default_chassis_skin_id,
			to_json(dms) as default_chassis_skin,
		
			m.chassis_skin_id,
			to_json(ms) as chassis_skin,
		
			m.intro_animation_id,
			to_json(ma2) as intro_animation,
		
			m.outro_animation_id,
			to_json(ma1) as outro_animation,
		
			m.power_core_id,
			to_json(ec) as power_core,
		
			w.weapons,
			u.utility
		FROM mechs m
		INNER JOIN collection_items ci on ci.item_id = m.id
		INNER JOIN players p ON p.id = ci.owner_id
		INNER JOIN factions f on p.faction_id = f.id
		LEFT OUTER JOIN power_cores ec ON ec.id = m.power_core_id
		LEFT OUTER JOIN brands b ON b.id = m.brand_id
		LEFT OUTER JOIN mech_model mm ON m.model_id = mm.id
		LEFT OUTER JOIN (
			SELECT _ms.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id
			FROM mech_skin _ms
			INNER JOIN collection_items _ci on _ci.item_id = _ms.id
		) ms ON m.chassis_skin_id = ms.id
				 LEFT OUTER JOIN blueprint_mech_skin dms ON mm.default_chassis_skin_id = dms.id
			-- LEFT OUTER JOIN mech_animation ma1 on ma1.id = m.outro_animation_id
				 LEFT OUTER JOIN (
			SELECT _ma.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id
			FROM mech_animation _ma
					 INNER JOIN collection_items _ci on _ci.item_id = _ma.id
		) ma1 on ma1.id = m.outro_animation_id
			-- LEFT OUTER JOIN mech_animation ma2 on ma2.id = m.intro_animation_id
				 LEFT OUTER JOIN (
			SELECT _ma.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id
			FROM mech_animation _ma
					 INNER JOIN collection_items _ci on _ci.item_id = _ma.id
		) ma2 on ma2.id = m.intro_animation_id
				 LEFT OUTER JOIN (
			SELECT mw.chassis_id, json_agg(w2) as weapons
			FROM mech_weapons mw
					 INNER JOIN
				 (
					 SELECT _w.*, _ci.hash, _ci.token_id, _ci.tier, _ci.owner_id
					 FROM weapons _w
							  INNER JOIN collection_items _ci on _ci.item_id = _w.id
				 ) w2 ON mw.weapon_id = w2.id
			GROUP BY mw.chassis_id
		) w on w.chassis_id = m.id
				 LEFT OUTER JOIN (
			SELECT mw.chassis_id, json_agg(_u) as utility
			FROM mech_utility mw
					 INNER JOIN (
				SELECT
					_u.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id,
					to_json(_us) as shield,
					to_json(_ua) as accelerator,
					to_json(_uam) as attack_drone,
					to_json(_uad) as anti_missile,
					to_json(_urd) as repair_drone
				FROM utility _u
						 INNER JOIN collection_items _ci on _ci.item_id = _u.id
						 LEFT OUTER JOIN utility_shield _us ON _us.utility_id = _u.id
						 LEFT OUTER JOIN utility_accelerator _ua ON _ua.utility_id = _u.id
						 LEFT OUTER JOIN utility_anti_missile _uam ON _uam.utility_id = _u.id
						 LEFT OUTER JOIN utility_attack_drone _uad ON _uad.utility_id = _u.id
						 LEFT OUTER JOIN utility_repair_drone _urd ON _urd.utility_id = _u.id
			) _u ON mw.utility_id = _u.id
			GROUP BY mw.chassis_id
		) u on u.chassis_id = m.id
		WHERE m.id IN (` + paramrefs + `)
		ORDER BY p.faction_id 
		`

	result, err := gamedb.StdConn.Query(query, mechids...)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	i := 0
	for result.Next() {
		mc := &server.Mech{
			CollectionDetails: &server.CollectionDetails{},
		}
		err = result.Scan(
			&mc.CollectionDetails.CollectionSlug,
			&mc.CollectionDetails.Hash,
			&mc.CollectionDetails.TokenID,
			&mc.CollectionDetails.OwnerID,
			&mc.CollectionDetails.Tier,
			&mc.ID,
			&mc.Name,
			&mc.Label,
			&mc.WeaponHardpoints,
			&mc.UtilitySlots,
			&mc.Speed,
			&mc.MaxHitpoints,
			&mc.IsDefault,
			&mc.IsInsured,
			&mc.GenesisTokenID,
			&mc.LimitedReleaseTokenID,
			&mc.PowerCoreSize,
			&mc.BlueprintID,
			&mc.BrandID,
			&mc.Brand,
			&mc.Owner,
			&mc.FactionID,
			&mc.Faction,
			&mc.ModelID,
			&mc.Model,
			&mc.DefaultChassisSkinID,
			&mc.DefaultChassisSkin,
			&mc.ChassisSkinID,
			&mc.ChassisSkin,
			&mc.IntroAnimationID,
			&mc.IntroAnimation,
			&mc.OutroAnimationID,
			&mc.OutroAnimation,
			&mc.PowerCoreID,
			&mc.PowerCore,
			&mc.Weapons,
			&mc.Utility,
		)
		if err != nil {
			return nil, err
		}
		mcs[i] = mc
		i++
	}

	if i < len(mechIDs) {
		mcs = mcs[:len(mcs)-i]
		return mcs, ErrNotAllMechsReturned
	}

	return mcs, err
}

// MechIDFromHash retrieve a mech ID from a hash
func MechIDFromHash(hash string) (uuid.UUID, error) {
	q := `SELECT item_id FROM collection_items WHERE hash = $1`
	var id string
	err := gamedb.StdConn.QueryRow(q, hash).
		Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}
	uid, err := uuid.FromString(id)

	if err != nil {
		return uuid.Nil, err
	}

	return uid, err
}

// MechIDsFromHash retrieve a slice mech IDs from hash variatic
func MechIDsFromHash(hashes ...string) ([]uuid.UUID, error) {
	var paramrefs string
	idintf := []interface{}{}
	for i, hash := range hashes {
		if hash != "" {
			paramrefs += `$` + strconv.Itoa(i+1) + `,`
			idintf = append(idintf, hash)
		}
	}
	paramrefs = paramrefs[:len(paramrefs)-1]
	q := `	SELECT ci.item_id, ci.hash 
			FROM collection_items ci
			WHERE ci.hash IN (` + paramrefs + `)`

	result, err := gamedb.StdConn.Query(q, idintf...)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	ids := make([]uuid.UUID, len(hashes))
	i := 0
	for result.Next() {
		var idStr string
		var hash string
		err = result.Scan(&idStr, &hash)
		if err != nil {
			return nil, err
		}

		uid, err := uuid.FromString(idStr)
		if err != nil {
			gamelog.L.Error().Str("mechID", idStr).Str("db func", "MechIDsFromHash").Err(err).Msg("unable to convert id to uuid")
		}

		// set id in correct order
		for index, h := range hashes {
			if h == hash {
				ids[index] = uid
				i++
			}
		}
	}

	if i == 0 {
		return nil, errors.New("no ids were scanned from result")
	}

	return ids, err
}

type BattleQueuePosition struct {
	MechID           uuid.UUID `db:"mech_id"`
	QueuePosition    int64     `db:"queue_position"`
	BattleContractID string    `db:"battle_contract_id"`
}

// MechQueuePosition return a list of mech queue position of the player (exclude in battle)
func MechQueuePosition(factionID string, ownerID string) ([]*BattleQueuePosition, error) {
	q := `
		SELECT
			x.mech_id,
			x.queue_position,
		    x.battle_contract_id
		FROM
			(
				SELECT
					bq.mech_id,
				    bq.owner_id,
				    bq.battle_contract_id,
					row_number () over (ORDER BY bq.queued_at) AS queue_position
				FROM
					battle_queue bq
				WHERE 
					bq.faction_id = $1 AND bq.battle_id isnull
			) x
		WHERE
			x.owner_id = $2
		ORDER BY
			x.queue_position
	`

	result, err := gamedb.StdConn.Query(q, factionID, ownerID)
	if err != nil {
		return nil, err
	}

	mqp := []*BattleQueuePosition{}
	for result.Next() {
		qp := &BattleQueuePosition{}
		err = result.Scan(&qp.MechID, &qp.QueuePosition, &qp.BattleContractID)
		if err != nil {
			return nil, err
		}

		mqp = append(mqp, qp)
	}

	return mqp, nil
}

// TODO: I want InsertNewMech tested.

func InsertNewMech(ownerID uuid.UUID, mechBlueprint *server.BlueprintMech) (*server.Mech, error) {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return nil, terror.Error(err)
	}

	// first insert the mech
	newMech := boiler.Mech{
		BlueprintID:           mechBlueprint.ID,
		BrandID:               mechBlueprint.BrandID,
		Label:                 mechBlueprint.Label,
		WeaponHardpoints:      mechBlueprint.WeaponHardpoints,
		UtilitySlots:          mechBlueprint.UtilitySlots,
		Speed:                 mechBlueprint.Speed,
		MaxHitpoints:          mechBlueprint.MaxHitpoints,
		IsDefault:             false,
		IsInsured:             false,
		Name:                  "",
		ModelID:               mechBlueprint.ModelID,
		PowerCoreSize:         mechBlueprint.PowerCoreSize,
		GenesisTokenID:        mechBlueprint.GenesisTokenID,
		LimitedReleaseTokenID: mechBlueprint.LimitedReleaseTokenID,
	}

	err = newMech.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	err = InsertNewCollectionItem(tx, mechBlueprint.Collection, boiler.ItemTypeMech, newMech.ID, mechBlueprint.Tier, ownerID.String())
	if err != nil {
		return nil, terror.Error(err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, terror.Error(err)
	}

	return Mech(newMech.ID)
}
