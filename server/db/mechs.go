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
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const CompleteMechQuery = `
SELECT
	collection_items.collection_slug,
	collection_items.hash,
	collection_items.token_id,
	collection_items.owner_id,
	collection_items.tier,
	mechs.id,
	mechs.name,
	mechs.label,
	mechs.weapon_hardpoints,
	mechs.utility_slots,
	mechs.speed,
	mechs.max_hitpoints,
	mechs.is_default,
	mechs.is_insured,
	mechs.genesis_token_id,
	mechs.limited_release_token_id,
	mechs.power_core_size,
	mechs.blueprint_id,
	mechs.brand_id,
	to_json(b) as brand,
	to_json(p) as owner,
	p.faction_id,
	to_json(f) as faction,
	mechs.model_id,
	to_json(mm) as model,
	mm.default_chassis_skin_id,
	to_json(dms) as default_chassis_skin,
	mechs.chassis_skin_id,
	to_json(ms) as chassis_skin,
	mechs.intro_animation_id,
	to_json(ma2) as intro_animation,
	mechs.outro_animation_id,
	to_json(ma1) as outro_animation,
	mechs.power_core_id,
	to_json(ec) as power_core,
	w.weapons,
	u.utility
FROM collection_items 
INNER JOIN mechs on collection_items.item_id = mechs.id
INNER JOIN players p ON p.id = collection_items.owner_id
INNER JOIN factions f on p.faction_id = f.id
LEFT OUTER JOIN power_cores ec ON ec.id = mechs.power_core_id
LEFT OUTER JOIN brands b ON b.id = mechs.brand_id
LEFT OUTER JOIN mech_model mm ON mechs.model_id = mm.id
LEFT OUTER JOIN (
	SELECT _ms.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id
	FROM mech_skin _ms
	INNER JOIN collection_items _ci on _ci.item_id = _ms.id
) ms ON mechs.chassis_skin_id = ms.id
LEFT OUTER JOIN blueprint_mech_skin dms ON mm.default_chassis_skin_id = dms.id
LEFT OUTER JOIN (
	SELECT _ma.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id
	FROM mech_animation _ma
	INNER JOIN collection_items _ci on _ci.item_id = _ma.id
) ma1 on ma1.id = mechs.outro_animation_id
LEFT OUTER JOIN (
	SELECT _ma.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id
	FROM mech_animation _ma
	INNER JOIN collection_items _ci on _ci.item_id = _ma.id
) ma2 on ma2.id = mechs.intro_animation_id
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
) w on w.chassis_id = mechs.id
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
) u on u.chassis_id = mechs.id `

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

	var ids []string
	for result.Next() {
		id := ""
		err = result.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
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

	query := fmt.Sprintf(`%s WHERE collection_items.item_id = $1`, CompleteMechQuery)

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

func Mechs(mechIDs ...string) ([]*server.Mech, error) {
	if len(mechIDs) == 0 {
		return nil, errors.New("no mech ids provided")
	}

	mcs := make([]*server.Mech, len(mechIDs))

	mechids := make([]interface{}, len(mechIDs))
	var paramrefs string
	for i, id := range mechIDs {
		paramrefs += `$` + strconv.Itoa(i+1) + `,`
		mechids[i] = id
	}
	paramrefs = paramrefs[:len(paramrefs)-1]

	query := fmt.Sprintf(
		`%s 	
		WHERE mechs.id IN (%s)
		ORDER BY p.faction_id `,
		CompleteMechQuery,
		paramrefs)

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

func IsMechColumn(col string) bool {
	switch col {
	case boiler.MechColumns.ID,
		boiler.MechColumns.BrandID,
		boiler.MechColumns.Label,
		boiler.MechColumns.WeaponHardpoints,
		boiler.MechColumns.UtilitySlots,
		boiler.MechColumns.Speed,
		boiler.MechColumns.MaxHitpoints,
		boiler.MechColumns.DeletedAt,
		boiler.MechColumns.UpdatedAt,
		boiler.MechColumns.CreatedAt,
		boiler.MechColumns.BlueprintID,
		boiler.MechColumns.IsDefault,
		boiler.MechColumns.IsInsured,
		boiler.MechColumns.Name,
		boiler.MechColumns.ModelID,
		boiler.MechColumns.GenesisTokenID,
		boiler.MechColumns.LimitedReleaseTokenID,
		boiler.MechColumns.PowerCoreSize,
		boiler.MechColumns.ChassisSkinID,
		boiler.MechColumns.PowerCoreID,
		boiler.MechColumns.IntroAnimationID,
		boiler.MechColumns.OutroAnimationID:
		return true
	default:
		return false
	}
}

type MechListOpts struct {
	Search   string
	Filter   *ListFilterRequest
	Sort     *ListSortRequest
	PageSize int
	Page     int
	OwnerID  string
}

func MechList(opts *MechListOpts) (int64, []*server.Mech, error) {
	var mechs []*server.Mech

	var queryMods []qm.QueryMod

	// create the where owner id = clause
	queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
		Table:    boiler.TableNames.CollectionItems,
		Column:   boiler.CollectionItemColumns.OwnerID,
		Operator: OperatorValueTypeEquals,
		Value:    opts.OwnerID,
	}, 0, ""),
		GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.CollectionItems,
			Column:   boiler.CollectionItemColumns.ItemType,
			Operator: OperatorValueTypeEquals,
			Value:    boiler.ItemTypeMech,
		}, 0, "and"))

	// Filters
	if opts.Filter != nil {
		// if we have filter
		for i, f := range opts.Filter.Items {
			// validate it is the right table and valid column
			if f.Table == boiler.TableNames.Mechs && IsMechColumn(f.Column) {
				queryMods = append(queryMods, GenerateListFilterQueryMod(*f, i+1, opts.Filter.LinkOperator))
			}

		}
	}
	// Search
	if opts.Search != "" {
		xSearch := ParseQueryText(opts.Search, true)
		if len(xSearch) > 0 {
			queryMods = append(queryMods,
				qm.And(fmt.Sprintf(
					"((to_tsvector('english', %[1]s.%[2]s) @@ to_tsquery(?))",
					boiler.TableNames.Mechs,
					boiler.MechColumns.Label,
				),
					xSearch,
				))
		}
	}

	total, err := boiler.CollectionItems(
		queryMods...,
	).Count(gamedb.StdConn)
	if err != nil {
		return 0, nil, err
	}
	// Sort
	if opts.Sort != nil && opts.Sort.Table == boiler.TableNames.Mechs && IsMechColumn(opts.Sort.Column) && opts.Sort.Direction.IsValid() {
		queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s.%s %s", boiler.TableNames.Mechs, opts.Sort.Column, opts.Sort.Direction)))
	} else {
		queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s.%s desc", boiler.TableNames.Mechs, boiler.MechColumns.Name)))
	}

	// Limit/Offset
	if opts.PageSize > 0 {
		queryMods = append(queryMods, qm.Limit(opts.PageSize))
	}
	if opts.Page > 0 {
		queryMods = append(queryMods, qm.Offset(opts.PageSize*(opts.Page-1)))
	}

	queryMods = append(queryMods,
		qm.Select(
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.CollectionSlug),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.Hash),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.TokenID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.OwnerID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.Tier),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Name),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Label),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.WeaponHardpoints),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.UtilitySlots),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Speed),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.MaxHitpoints),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.IsDefault),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.IsInsured),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.GenesisTokenID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.LimitedReleaseTokenID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.PowerCoreSize),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.PowerCoreID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.BlueprintID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.BrandID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ModelID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ChassisSkinID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.IntroAnimationID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.OutroAnimationID),
		),
		qm.From(boiler.TableNames.CollectionItems),
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.Mechs,
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
		)),
	)

	rows, err := boiler.NewQuery(
		queryMods...,
	).Query(gamedb.StdConn)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		mc := &server.Mech{
			CollectionDetails: &server.CollectionDetails{},
		}
		err = rows.Scan(
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
			&mc.PowerCoreID,
			&mc.BlueprintID,
			&mc.BrandID,
			&mc.ModelID,
			&mc.ChassisSkinID,
			&mc.IntroAnimationID,
			&mc.OutroAnimationID,
			//&mc.DefaultChassisSkinID, // TODO: probably want this? (its  attached to the mech model, could be lazy loaded with the rest)
		)
		if err != nil {
			return total, mechs, err
		}
		mechs = append(mechs, mc)
	}

	return total, mechs, nil
}
