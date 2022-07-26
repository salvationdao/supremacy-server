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
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type MechColumns string

func (c MechColumns) IsValid() error {
	switch string(c) {
	case boiler.MechColumns.Name:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid mech column"))
}

const CompleteMechQuery = `
SELECT
	collection_items.collection_slug,
	collection_items.hash,
	collection_items.token_id,
	collection_items.owner_id,
	collection_items.tier,
	collection_items.item_type,
	collection_items.market_locked,
	collection_items.xsyn_locked,
	collection_items.locked_to_marketplace,
	collection_items.asset_hidden,
	collection_items.id AS collection_item_id,
	COALESCE(p.username, ''),
	COALESCE(mech_stats.total_wins, 0),
	COALESCE(mech_stats.total_deaths, 0),
	COALESCE(mech_stats.total_kills, 0),
	COALESCE(mech_stats.battles_survived, 0), 
	COALESCE(mech_stats.total_losses, 0),
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
	u.utility,
	(
		SELECT _i.id
		FROM item_sales _i
		WHERE _i.collection_item_id = collection_items.id
			AND _i.sold_at IS NULL
			AND _i.deleted_at IS NULL
			AND _i.end_at > NOW()
		LIMIT 1
	) AS item_sale_id,
	(
		SELECT (_bm.availability_id IS NULL OR _a.available_at <= NOW())
		FROM blueprint_mechs _bm 
			LEFT JOIN availabilities _a ON _a.id = _bm.availability_id
		WHERE _bm.id = mechs.blueprint_id
		LIMIT 1
	) AS battle_ready
FROM collection_items 
INNER JOIN mechs on collection_items.item_id = mechs.id
INNER JOIN players p ON p.id = collection_items.owner_id
LEFT OUTER JOIN mech_stats  ON mech_stats.mech_id = mechs.id
LEFT OUTER JOIN factions f on p.faction_id = f.id
LEFT OUTER JOIN (
	SELECT _pc.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id, _ci.image_url, _ci.avatar_url, _ci.card_animation_url, _ci.animation_url
	FROM power_cores _pc
	INNER JOIN collection_items _ci on _ci.item_id = _pc.id
	) ec ON ec.id = mechs.power_core_id
LEFT OUTER JOIN brands b ON b.id = mechs.brand_id
LEFT OUTER JOIN mech_models mm ON mechs.model_id = mm.id
LEFT OUTER JOIN (
	SELECT _ms.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id
	FROM mech_skin _ms
	INNER JOIN collection_items _ci on _ci.item_id = _ms.id
) ms ON mechs.chassis_skin_id = ms.id
LEFT OUTER JOIN blueprint_mech_skin dms ON mm.default_chassis_skin_id = dms.id
LEFT OUTER JOIN (
	SELECT _ma.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id, _ci.image_url, _ci.avatar_url, _ci.card_animation_url, _ci.animation_url
	FROM mech_animation _ma
	INNER JOIN collection_items _ci on _ci.item_id = _ma.id
) ma1 on ma1.id = mechs.outro_animation_id
LEFT OUTER JOIN (
	SELECT _ma.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id, _ci.image_url, _ci.avatar_url, _ci.card_animation_url, _ci.animation_url
	FROM mech_animation _ma
	INNER JOIN collection_items _ci on _ci.item_id = _ma.id
) ma2 on ma2.id = mechs.intro_animation_id
LEFT OUTER JOIN (
	SELECT mw.chassis_id, json_agg(w2) as weapons
	FROM mech_weapons mw
	INNER JOIN
		(
			SELECT _w.*, _ci.hash, _ci.token_id, _ci.tier, _ci.owner_id, _ci.image_url, _ci.avatar_url, _ci.card_animation_url, _ci.animation_url, to_json(_ws) as weapon_skin
			FROM weapons _w
			INNER JOIN collection_items _ci on _ci.item_id = _w.id
			LEFT OUTER JOIN (
					SELECT __ws.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id, _ci.image_url, _ci.avatar_url, _ci.card_animation_url, _ci.animation_url
					FROM weapon_skin __ws
					INNER JOIN collection_items _ci on _ci.item_id = __ws.id
			) _ws ON _ws.equipped_on = _w.id
		) w2 ON mw.weapon_id = w2.id
	GROUP BY mw.chassis_id
) w on w.chassis_id = mechs.id
LEFT OUTER JOIN (
	SELECT mw.chassis_id, json_agg(_u) as utility
	FROM mech_utility mw
	INNER JOIN (
		SELECT
			_u.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id, _ci.image_url, _ci.avatar_url, _ci.card_animation_url, _ci.animation_url,
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

func DefaultMechs() ([]*server.Mech, error) {
	idq := `SELECT id FROM mechs WHERE is_default=TRUE`

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
func Mech(conn boil.Executor, mechID string) (*server.Mech, error) {
	mc := &server.Mech{
		CollectionItem: &server.CollectionItem{},
		Stats:          &server.Stats{},
		Owner:          &server.User{},
	}

	query := fmt.Sprintf(`%s WHERE collection_items.item_id = $1`, CompleteMechQuery)

	result, err := conn.Query(query, mechID)
	if err != nil {
		fmt.Println("here 11")
		return nil, err
	}
	defer result.Close()

	for result.Next() {
		err = result.Scan(
			&mc.CollectionItem.CollectionSlug,
			&mc.CollectionItem.Hash,
			&mc.CollectionItem.TokenID,
			&mc.CollectionItem.OwnerID,
			&mc.CollectionItem.Tier,
			&mc.CollectionItem.ItemType,
			&mc.CollectionItem.MarketLocked,
			&mc.CollectionItem.XsynLocked,
			&mc.CollectionItem.LockedToMarketplace,
			&mc.CollectionItem.AssetHidden,
			&mc.CollectionItemID,
			&mc.Owner.Username,
			&mc.Stats.TotalWins,
			&mc.Stats.TotalDeaths,
			&mc.Stats.TotalKills,
			&mc.Stats.BattlesSurvived,
			&mc.Stats.TotalLosses,
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
			&mc.ItemSaleID,
			&mc.BattleReady,
		)
		if err != nil {
			fmt.Println("here 22")
			gamelog.L.Error().Err(err).Msg("failed to get mech")
			return nil, err
		}
	}

	if mc.ID == "" {
		fmt.Println("here 33")
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
			CollectionItem: &server.CollectionItem{},
			Stats:          &server.Stats{},
			Owner:          &server.User{},
		}
		err = result.Scan(
			&mc.CollectionItem.CollectionSlug,
			&mc.CollectionItem.Hash,
			&mc.CollectionItem.TokenID,
			&mc.CollectionItem.OwnerID,
			&mc.CollectionItem.Tier,
			&mc.CollectionItem.ItemType,
			&mc.CollectionItem.MarketLocked,
			&mc.CollectionItem.XsynLocked,
			&mc.CollectionItem.LockedToMarketplace,
			&mc.CollectionItem.AssetHidden,
			&mc.CollectionItemID,
			&mc.Owner.Username,
			&mc.Stats.TotalWins,
			&mc.Stats.TotalDeaths,
			&mc.Stats.TotalKills,
			&mc.Stats.BattlesSurvived,
			&mc.Stats.TotalLosses,
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
			&mc.ItemSaleID,
			&mc.BattleReady,
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
	MechID           uuid.UUID   `db:"mech_id"`
	QueuePosition    int64       `db:"queue_position"`
	BattleContractID null.String `db:"battle_contract_id"`
}

// TODO: I want InsertNewMech tested.

func InsertNewMech(tx boil.Executor, ownerID uuid.UUID, mechBlueprint *server.BlueprintMech) (*server.Mech, error) {
	mechModel, err := boiler.MechModels(
		boiler.MechModelWhere.ID.EQ(mechBlueprint.ModelID),
		qm.Load(boiler.MechModelRels.DefaultChassisSkin),
	).One(tx)
	if err != nil {
		return nil, terror.Error(err)
	}

	if mechModel.R == nil || mechModel.R.DefaultChassisSkin == nil {
		return nil, terror.Error(fmt.Errorf("could not find default skin relationship to mech"), "Could not find mech default skin relationship, try again or contact support")
	}

	//bpms := mechModel.R.DefaultChassisSkin

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

	_, err = InsertNewCollectionItem(tx,
		mechBlueprint.Collection,
		boiler.ItemTypeMech,
		newMech.ID,
		mechBlueprint.Tier,
		ownerID.String(),
	)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to insert col item")
		return nil, terror.Error(err)
	}

	mech, err := Mech(tx, newMech.ID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to get mech")
		return nil, terror.Error(err)
	}
	return mech, nil
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
	Search              string
	Filter              *ListFilterRequest
	Sort                *ListSortRequest
	PageSize            int
	Page                int
	OwnerID             string
	QueueSort           *MechListQueueSortOpts
	DisplayXsynMechs    bool
	ExcludeMarketLocked bool
	IncludeMarketListed bool
	FilterRarities      []string `json:"rarities"`
	FilterStatuses      []string `json:"statuses"`
}

type MechListQueueSortOpts struct {
	FactionID string
	SortDir   SortByDir
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
		}, 0, "and"),
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.Mechs,
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
		)),
	)

	if !opts.DisplayXsynMechs || !opts.IncludeMarketListed {
		queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.CollectionItems,
			Column:   boiler.CollectionItemColumns.XsynLocked,
			Operator: OperatorValueTypeIsFalse,
		}, 0, ""))
	}
	if opts.ExcludeMarketLocked {
		queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.CollectionItems,
			Column:   boiler.CollectionItemColumns.MarketLocked,
			Operator: OperatorValueTypeIsFalse,
		}, 0, ""))
	}
	if !opts.IncludeMarketListed {
		queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.CollectionItems,
			Column:   boiler.CollectionItemColumns.LockedToMarketplace,
			Operator: OperatorValueTypeIsFalse,
		}, 0, ""))
	}

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
	if len(opts.FilterRarities) > 0 {
		vals := []interface{}{}
		for _, r := range opts.FilterRarities {
			vals = append(vals, r)
		}
		queryMods = append(queryMods,
			qm.LeftOuterJoin(fmt.Sprintf(
				"%s msc ON msc.%s = %s AND msc.%s = ?",
				boiler.TableNames.CollectionItems,
				boiler.CollectionItemColumns.ItemID,
				qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ChassisSkinID),
				boiler.CollectionItemColumns.ItemType,
			), boiler.ItemTypeMechSkin),
			qm.AndIn("msc.tier IN ?", vals...),
		)
	}
	if len(opts.FilterStatuses) > 0 {
		hasIdleToggled := false
		hasInBattleToggled := false
		hasMarketplaceToggled := false
		hasInQueueToggled := false
		hasBattleReadyToggled := false

		statusFilters := []qm.QueryMod{}

		for _, s := range opts.FilterStatuses {
			if s == "IDLE" {
				if hasIdleToggled {
					continue
				}
				hasIdleToggled = true
				statusFilters = append(statusFilters, qm.Or(fmt.Sprintf(
					`NOT EXISTS (
						SELECT _bq.%s
						FROM %s _bq
						WHERE _bq.%s = %s
						LIMIT 1
					)`,
					boiler.BattleQueueColumns.ID,
					boiler.TableNames.BattleQueue,
					boiler.BattleQueueColumns.MechID,
					qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
				)))
			} else if s == "BATTLE" {
				if hasInBattleToggled {
					continue
				}
				hasInBattleToggled = true
				statusFilters = append(statusFilters, qm.Or(fmt.Sprintf(
					`EXISTS (
						SELECT _bq.%s
						FROM %s _bq
						WHERE _bq.%s = %s
							AND _bq.%s IS NOT NULL
						LIMIT 1
					)`,
					boiler.BattleQueueColumns.ID,
					boiler.TableNames.BattleQueue,
					boiler.BattleQueueColumns.MechID,
					qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
					boiler.BattleQueueColumns.BattleID,
				)))
			} else if s == "MARKET" {
				if hasMarketplaceToggled {
					continue
				}
				hasMarketplaceToggled = true
				statusFilters = append(statusFilters, qm.Or(fmt.Sprintf(
					`EXISTS (
						SELECT _i.%s
						FROM %s _i
						WHERE _i.%s = %s
							AND _i.%s IS NULL
							AND _i.%s IS NULL
							AND _i.%s > NOW()
						LIMIT 1
					)`,
					boiler.ItemSaleColumns.ID,
					boiler.TableNames.ItemSales,
					boiler.ItemSaleColumns.CollectionItemID,
					qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ID),
					boiler.ItemSaleColumns.SoldAt,
					boiler.ItemSaleColumns.DeletedAt,
					boiler.ItemSaleColumns.EndAt,
				)))
			} else if s == "QUEUE" {
				if hasInQueueToggled {
					continue
				}
				hasInQueueToggled = true
				statusFilters = append(statusFilters, qm.Or(fmt.Sprintf(
					`EXISTS (
						SELECT _bq.%s
						FROM %s _bq
						WHERE _bq.%s = %s
							AND _bq.%s IS NULL
						LIMIT 1
					)`,
					boiler.BattleQueueColumns.ID,
					boiler.TableNames.BattleQueue,
					boiler.BattleQueueColumns.MechID,
					qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
					boiler.BattleQueueColumns.BattleID,
				)))
			} else if s == "BATTLE_READY" {
				if hasBattleReadyToggled {
					continue
				}
				hasBattleReadyToggled = true
				statusFilters = append(statusFilters, qm.Or(fmt.Sprintf(
					`EXISTS (
						SELECT 1 
						FROM %s _bm
							LEFT JOIN %s _a ON _a.%s = _bm.%s
						WHERE _bm.%s = %s 
							AND (
								_a.%s IS NULL
								OR _a.%s <= NOW()
							)
					)`,
					boiler.TableNames.BlueprintMechs,
					boiler.TableNames.Availabilities,
					boiler.AvailabilityColumns.ID,
					boiler.BlueprintMechColumns.AvailabilityID,
					boiler.BlueprintMechColumns.ID,
					qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.BlueprintID),
					boiler.AvailabilityColumns.ID,
					boiler.AvailabilityColumns.AvailableAt,
				)))
			}
			if hasIdleToggled && hasInBattleToggled && hasMarketplaceToggled && hasInQueueToggled && hasBattleReadyToggled {
				break
			}
		}

		if len(statusFilters) > 0 {
			queryMods = append(queryMods, qm.Expr(statusFilters...))
		}
	}

	// Search
	if opts.Search != "" {
		xSearch := ParseQueryText(opts.Search, true)
		if len(xSearch) > 0 {
			queryMods = append(queryMods,
				qm.And(fmt.Sprintf(
					"(to_tsvector('english', %s) @@ to_tsquery(?))",
					qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Label),
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

	// Limit/Offset
	if opts.PageSize > 0 {
		queryMods = append(queryMods, qm.Limit(opts.PageSize))
	}
	if opts.Page > 0 {
		queryMods = append(queryMods, qm.Offset(opts.PageSize*(opts.Page-1)))
	}

	// Build query
	queryMods = append(queryMods,
		qm.Select(
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.CollectionSlug),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.Hash),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.TokenID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.OwnerID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.Tier),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemType),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.MarketLocked),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.XsynLocked),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.LockedToMarketplace),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.AssetHidden),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ID),
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
	)

	// Sort
	if opts.QueueSort != nil {
		queryMods = append(queryMods,
			qm.Select("_bq.queue_position AS queue_position"),
			qm.LeftOuterJoin(
				fmt.Sprintf(`(
					SELECT  _bq.mech_id, _bq.battle_contract_id, row_number () OVER (ORDER BY _bq.queued_at) AS queue_position
						from battle_queue _bq
						where _bq.faction_id = ?
							AND _bq.battle_id IS NULL
					) _bq ON _bq.mech_id = %s`,
					qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
				),
				opts.QueueSort.FactionID,
			),
			qm.OrderBy(fmt.Sprintf("queue_position %s NULLS LAST, %s, %s",
				opts.QueueSort.SortDir,
				qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Name),
				qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
			)),
		)
	} else {
		if opts.Sort != nil && opts.Sort.Table == boiler.TableNames.Mechs && IsMechColumn(opts.Sort.Column) && opts.Sort.Direction.IsValid() {
			queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s.%s %s", boiler.TableNames.Mechs, opts.Sort.Column, opts.Sort.Direction)))
		} else {
			queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s.%s desc", boiler.TableNames.Mechs, boiler.MechColumns.Name)))
		}
	}
	rows, err := boiler.NewQuery(
		queryMods...,
	).Query(gamedb.StdConn)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		mc := &server.Mech{
			CollectionItem: &server.CollectionItem{},
		}

		scanArgs := []interface{}{
			&mc.CollectionItem.CollectionSlug,
			&mc.CollectionItem.Hash,
			&mc.CollectionItem.TokenID,
			&mc.CollectionItem.OwnerID,
			&mc.CollectionItem.Tier,
			&mc.CollectionItem.ItemType,
			&mc.CollectionItem.MarketLocked,
			&mc.CollectionItem.XsynLocked,
			&mc.CollectionItem.LockedToMarketplace,
			&mc.CollectionItem.AssetHidden,
			&mc.CollectionItemID,
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
		}
		if opts.QueueSort != nil {
			scanArgs = append(scanArgs, &mc.QueuePosition)
		}
		err = rows.Scan(scanArgs...)
		if err != nil {
			return total, mechs, err
		}
		mechs = append(mechs, mc)
	}

	return total, mechs, nil
}

func MechRename(mechID string, ownerID string, name string) (string, error) {

	// get mech
	mech, err := boiler.FindMech(gamedb.StdConn, mechID)
	if err != nil {
		return "", terror.Error(err)
	}

	item, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(mech.ID)).One(gamedb.StdConn)
	if err != nil {
		return "", terror.Error(err)
	}

	// check owner
	if item.OwnerID != ownerID {
		err = fmt.Errorf("failed to update mech name, must be the owner of the mech")
		return "", terror.Error(err)
	}

	// update mech name
	mech.Name = name
	_, err = mech.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		return "", terror.Error(err)
	}

	return name, nil

}

func MechEquippedOnDetails(trx boil.Executor, equippedOnID string) (*server.EquippedOnDetails, error) {
	tx := trx
	if trx == nil {
		tx = gamedb.StdConn
	}

	eid := &server.EquippedOnDetails{}

	err := boiler.NewQuery(
		qm.Select(
			boiler.CollectionItemColumns.ItemID,
			boiler.CollectionItemColumns.Hash,
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Name),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Label),
		),
		qm.From(boiler.TableNames.CollectionItems),
		qm.InnerJoin(fmt.Sprintf(
			"%s on %s = %s",
			boiler.TableNames.Mechs,
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
		)),
		qm.Where(fmt.Sprintf("%s = ?", boiler.CollectionItemColumns.ItemID), equippedOnID),
	).QueryRow(tx).Scan(
		&eid.ID,
		&eid.Hash,
		&eid.Name,
		&eid.Label,
	)
	if err != nil {
		return nil, err
	}

	return eid, nil
}

// MechSetAllEquippedAssetsAsHidden marks all the attached items with the given asset_hidden reason
// passing in a null reason will update it to be unhidden
func MechSetAllEquippedAssetsAsHidden(trx boil.Executor, mechID string, reason null.String) error {
	tx := trx
	if trx == nil {
		tx = gamedb.StdConn
	}

	itemIDsToUpdate := []string{}

	// get equipped mech skin
	mSkins, err := boiler.MechSkins(
		boiler.MechSkinWhere.EquippedOn.EQ(null.StringFrom(mechID)),
	).All(tx)
	if err != nil {
		return err
	}
	for _, itm := range mSkins {
		itemIDsToUpdate = append(itemIDsToUpdate, itm.ID)
	}

	// get equipped mech animations
	mAnim, err := boiler.MechAnimations(
		boiler.MechAnimationWhere.EquippedOn.EQ(null.StringFrom(mechID)),
	).All(tx)
	if err != nil {
		return err
	}
	for _, itm := range mAnim {
		itemIDsToUpdate = append(itemIDsToUpdate, itm.ID)
	}

	// get equipped mech weapons
	mWpn, err := boiler.Weapons(
		boiler.WeaponWhere.EquippedOn.EQ(null.StringFrom(mechID)),
	).All(tx)
	if err != nil {
		return err
	}
	for _, itm := range mWpn {
		itemIDsToUpdate = append(itemIDsToUpdate, itm.ID)
		// get equipped mech weapon skins
		mWpnSkin, err := boiler.WeaponSkins(
			boiler.WeaponSkinWhere.EquippedOn.EQ(null.StringFrom(itm.ID)),
		).All(tx)
		if err != nil {
			return err
		}
		for _, itm := range mWpnSkin {
			itemIDsToUpdate = append(itemIDsToUpdate, itm.ID)
		}
	}

	// get equipped mech utilities
	mUtil, err := boiler.Utilities(
		boiler.UtilityWhere.EquippedOn.EQ(null.StringFrom(mechID)),
	).All(tx)
	if err != nil {
		return err
	}
	for _, itm := range mUtil {
		itemIDsToUpdate = append(itemIDsToUpdate, itm.ID)
	}

	// update!
	_, err = boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.IN(itemIDsToUpdate),
	).UpdateAll(tx, boiler.M{
		"asset_hidden": reason,
	})
	if err != nil {
		return err
	}

	return nil
}
