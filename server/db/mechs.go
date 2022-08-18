package db

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/benchmark"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"strconv"
)

type MechColumns string

func (c MechColumns) IsValid() error {
	switch string(c) {
	case boiler.MechColumns.Name:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid mech column"))
}

const outroTableName = "outro"
const introTableName = "intro"
const weaponsTableName = "weapons"
const utilityTableName = "utility"

func getDefaultMechQueryMods() []qm.QueryMod {
	return []qm.QueryMod{
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
			qm.Rels(boiler.TableNames.MechModelSkinCompatibilities, boiler.MechModelSkinCompatibilityColumns.ImageURL),
			qm.Rels(boiler.TableNames.MechModelSkinCompatibilities, boiler.MechModelSkinCompatibilityColumns.CardAnimationURL),
			qm.Rels(boiler.TableNames.MechModelSkinCompatibilities, boiler.MechModelSkinCompatibilityColumns.AvatarURL),
			qm.Rels(boiler.TableNames.MechModelSkinCompatibilities, boiler.MechModelSkinCompatibilityColumns.LargeImageURL),
			qm.Rels(boiler.TableNames.MechModelSkinCompatibilities, boiler.MechModelSkinCompatibilityColumns.BackgroundColor),
			qm.Rels(boiler.TableNames.MechModelSkinCompatibilities, boiler.MechModelSkinCompatibilityColumns.AnimationURL),
			qm.Rels(boiler.TableNames.MechModelSkinCompatibilities, boiler.MechModelSkinCompatibilityColumns.YoutubeURL),
			fmt.Sprintf(`COALESCE(%s, '')`, qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.Username)),
			fmt.Sprintf(`COALESCE(%s, 0)`, qm.Rels(boiler.TableNames.MechStats, boiler.MechStatColumns.TotalWins)),
			fmt.Sprintf(`COALESCE(%s, 0)`, qm.Rels(boiler.TableNames.MechStats, boiler.MechStatColumns.TotalDeaths)),
			fmt.Sprintf(`COALESCE(%s, 0)`, qm.Rels(boiler.TableNames.MechStats, boiler.MechStatColumns.TotalKills)),
			fmt.Sprintf(`COALESCE(%s, 0)`, qm.Rels(boiler.TableNames.MechStats, boiler.MechStatColumns.BattlesSurvived)),
			fmt.Sprintf(`COALESCE(%s, 0)`, qm.Rels(boiler.TableNames.MechStats, boiler.MechStatColumns.TotalLosses)),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Name),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.Label),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.WeaponHardpoints),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.UtilitySlots),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.Speed),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.MaxHitpoints),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.IsDefault),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.IsInsured),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.GenesisTokenID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.LimitedReleaseTokenID),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.PowerCoreSize),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.BlueprintID),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.BrandID),
			fmt.Sprintf("to_json(%s) as brand", boiler.TableNames.Brands),
			fmt.Sprintf("to_json(%s) as owner", boiler.TableNames.Players),
			qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.FactionID),
			fmt.Sprintf("to_json(%s) as faction", boiler.TableNames.Factions),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ChassisSkinID),
			fmt.Sprintf("to_json(%s) as chassis_skin", boiler.TableNames.MechSkin),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.IntroAnimationID),
			fmt.Sprintf("to_json(%s) as intro_animation", introTableName),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.OutroAnimationID),
			fmt.Sprintf("to_json(%s) as outro_animation", outroTableName),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.PowerCoreID),
			fmt.Sprintf("to_json(%s) as power_core", boiler.TableNames.PowerCores),
			weaponsTableName,
			utilityTableName,
			fmt.Sprintf(` 
					(
						SELECT %s
						FROM %s _i
						WHERE %s = %s
							AND %s IS NULL
							AND %s IS NULL
							AND %s > NOW()
						LIMIT 1
					) AS item_sale_id`,
				qm.Rels("_i", boiler.ItemSaleColumns.ID),
				boiler.TableNames.ItemSales,
				qm.Rels("_i", boiler.ItemSaleColumns.CollectionItemID),
				qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ID),
				qm.Rels("_i", boiler.ItemSaleColumns.SoldAt),
				qm.Rels("_i", boiler.ItemSaleColumns.DeletedAt),
				qm.Rels("_i", boiler.ItemSaleColumns.EndAt),
			),
			fmt.Sprintf(`
					(
						SELECT (%s IS NULL OR %s <= NOW())
						FROM %s _bm 
							LEFT JOIN %s _a ON %s = %s
						WHERE %s = %s
						LIMIT 1
					) AS battle_ready`,
				qm.Rels("_bm", boiler.BlueprintMechColumns.AvailabilityID),
				qm.Rels("_a", boiler.AvailabilityColumns.AvailableAt),
				boiler.TableNames.BlueprintMechs,
				boiler.TableNames.Availabilities,
				qm.Rels("_a", boiler.AvailabilityColumns.ID),
				qm.Rels("_bm", boiler.BlueprintMechColumns.AvailabilityID),
				qm.Rels("_bm", boiler.BlueprintMechColumns.ID),
				qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.BlueprintID),
			),
		),
		qm.From(boiler.TableNames.CollectionItems),
		// inner join mechs
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.Mechs,
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
		)),
		// inner join players
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.Players,
			qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.OwnerID),
		)),
		// outer join mech stats
		qm.LeftOuterJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.MechStats,
			qm.Rels(boiler.TableNames.MechStats, boiler.MechStatColumns.MechID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
		)),
		// outer join player faction
		qm.LeftOuterJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.Factions,
			qm.Rels(boiler.TableNames.Factions, boiler.FactionColumns.ID),
			qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.FactionID),
		)),
		// outer join power cores
		qm.LeftOuterJoin(fmt.Sprintf(`(
					SELECT _pc.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id, _bppc.image_url as image_url, _bppc.avatar_url as avatar_url, _bppc.card_animation_url as card_animation_url, _bppc.animation_url as animation_url
					FROM power_cores _pc
					INNER JOIN collection_items _ci on _ci.item_id = _pc.id
					INNER JOIN blueprint_power_cores _bppc on _pc.blueprint_id = _bppc.id
					) %s ON %s = %s`, // TODO: make this boiler/typesafe
			boiler.TableNames.PowerCores,
			qm.Rels(boiler.TableNames.PowerCores, boiler.PowerCoreColumns.ID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.PowerCoreID),
		)),
		// inner join mech blueprint
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.BlueprintMechs,
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.BlueprintID),
		)),
		// inner join brand
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.Brands,
			qm.Rels(boiler.TableNames.Brands, boiler.BrandColumns.ID),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.BrandID),
		)),
		// inner join skin
		qm.InnerJoin(fmt.Sprintf(`(
					SELECT _ms.*, _ci.hash, _ci.token_id, _ci.tier, _ci.owner_id, _bpms.label
					FROM mech_skin _ms
					INNER JOIN collection_items _ci on _ci.item_id = _ms.id
					INNER JOIN blueprint_mech_skin _bpms on _bpms.id = _ms.blueprint_id
				 )%s ON %s = %s`, // TODO: make this boiler/typesafe
			boiler.TableNames.MechSkin,
			qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.ID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ChassisSkinID),
		)),
		// inner join mech skin compatability table (to get images)
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s AND %s = %s",
			boiler.TableNames.MechModelSkinCompatibilities,
			qm.Rels(boiler.TableNames.MechModelSkinCompatibilities, boiler.MechModelSkinCompatibilityColumns.BlueprintMechSkinID),
			qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.BlueprintID),
			qm.Rels(boiler.TableNames.MechModelSkinCompatibilities, boiler.MechModelSkinCompatibilityColumns.MechModelID),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ID),
		)),
		// left join outro
		qm.LeftOuterJoin(fmt.Sprintf("%s AS %s ON %s = %s",
			boiler.TableNames.MechAnimation,
			outroTableName,
			qm.Rels(outroTableName, boiler.MechAnimationColumns.ID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.OutroAnimationID),
		)),
		// left join intro
		qm.LeftOuterJoin(fmt.Sprintf("%s AS %s ON %s = %s",
			boiler.TableNames.MechAnimation,
			introTableName,
			qm.Rels(introTableName, boiler.MechAnimationColumns.ID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.IntroAnimationID),
		)),
		// left join weapons
		qm.LeftOuterJoin(
			// TODO: make this boiler/typesafe
			fmt.Sprintf(`
					(
						SELECT mw.chassis_id, json_agg(w2) as weapons
						FROM mech_weapons mw
						INNER JOIN
							(
								SELECT _w.*, _ci.hash, _ci.token_id, _ci.tier, _ci.owner_id, to_json(_ws) as weapon_skin, _bpw.label, _wmsc.image_url as image_url, _wmsc.avatar_url as avatar_url, _wmsc.card_animation_url as card_animation_url, _wmsc.animation_url as animation_url
								FROM weapons _w
								INNER JOIN collection_items _ci on _ci.item_id = _w.id
								INNER JOIN blueprint_weapons _bpw on _bpw.id = _w.blueprint_id
								INNER JOIN (
										SELECT __ws.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id
										FROM weapon_skin __ws
										INNER JOIN collection_items _ci on _ci.item_id = __ws.id
								) _ws ON _ws.equipped_on = _w.id
								INNER JOIN weapon_model_skin_compatibilities _wmsc on _wmsc.blueprint_weapon_skin_id = _ws.blueprint_id and _wmsc.weapon_model_id = _bpw.weapon_model_id
							) w2 ON mw.weapon_id = w2.id
						GROUP BY mw.chassis_id
				) %s on %s = %s `,
				weaponsTableName,
				qm.Rels(weaponsTableName, boiler.MechWeaponColumns.ChassisID),
				qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
			)),
		// left join utility
		qm.LeftOuterJoin(
			// TODO: make this boiler/typesafe
			fmt.Sprintf(`
				(
					SELECT mw.chassis_id, json_agg(_u) as utility
					FROM mech_utility mw
					INNER JOIN (
						SELECT
							_u.*,_ci.hash, _ci.token_id, _ci.tier, _ci.owner_id, _bpu.image_url as image_url, _bpu.avatar_url as avatar_url, _bpu.card_animation_url as card_animation_url, _bpu.animation_url as animation_url,
							to_json(_us) as shield
						--	to_json(_ua) as accelerator,
						--	to_json(_uam) as attack_drone,
						--	to_json(_uad) as anti_missile,
						--	to_json(_urd) as repair_drone
						FROM utility _u
						INNER JOIN collection_items _ci on _ci.item_id = _u.id
						INNER JOIN blueprint_utility _bpu on _bpu.id = _u.blueprint_id
						LEFT OUTER JOIN utility_shield _us ON _us.utility_id = _u.id
						--LEFT OUTER JOIN utility_accelerator _ua ON _ua.utility_id = _u.id
						--LEFT OUTER JOIN utility_anti_missile _uam ON _uam.utility_id = _u.id
						--LEFT OUTER JOIN utility_attack_drone _uad ON _uad.utility_id = _u.id
						--LEFT OUTER JOIN utility_repair_drone _urd ON _urd.utility_id = _u.id
					) _u ON mw.utility_id = _u.id
					GROUP BY mw.chassis_id
				) %s on %s = %s `,
				utilityTableName,
				qm.Rels(utilityTableName, boiler.MechUtilityColumns.ChassisID),
				qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
			),
		),
	}
}

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

func Mech(conn boil.Executor, mechID string) (*server.Mech, error) {
	bm := benchmark.New()
	bm.Start("db Mech")
	defer bm.Alert(150)

	mc := &server.Mech{
		CollectionItem: &server.CollectionItem{},
		Stats:          &server.Stats{},
		Owner:          &server.User{},
		Images:         &server.Images{},
	}

	queryMods := getDefaultMechQueryMods()

	// Build query
	queryMods = append(queryMods,
		// where mech id in
		boiler.CollectionItemWhere.ItemID.EQ(mechID),
		// order by faction?
		qm.OrderBy(qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.FactionID)),
	)

	rows, err := boiler.NewQuery(
		queryMods...,
	).Query(conn)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
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
			&mc.Images.ImageURL,
			&mc.Images.CardAnimationURL,
			&mc.Images.AvatarURL,
			&mc.Images.LargeImageURL,
			&mc.Images.BackgroundColor,
			&mc.Images.AnimationURL,
			&mc.Images.YoutubeURL,
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
			gamelog.L.Error().Err(err).Msg("failed to get mech")
			return nil, err
		}
	}

	if mc.ID == "" {
		return nil, fmt.Errorf("unable to find mech with id %s", mechID)
	}

	bm.End("db Mech")
	return mc, err
}

func Mechs(mechIDs ...string) ([]*server.Mech, error) {
	if len(mechIDs) == 0 {
		return nil, errors.New("no mech ids provided")
	}
	mcs := make([]*server.Mech, len(mechIDs))

	queryMods := getDefaultMechQueryMods()

	// Build query
	queryMods = append(queryMods,
		// where mech id in
		boiler.CollectionItemWhere.ItemID.IN(mechIDs),
		// order by faction?
		qm.OrderBy(qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.FactionID)),
	)

	rows, err := boiler.NewQuery(
		queryMods...,
	).Query(gamedb.StdConn)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		mc := &server.Mech{
			CollectionItem: &server.CollectionItem{},
			Stats:          &server.Stats{},
			Owner:          &server.User{},
			Images:         &server.Images{},
		}
		err = rows.Scan(
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
			&mc.Images.ImageURL,
			&mc.Images.CardAnimationURL,
			&mc.Images.AvatarURL,
			&mc.Images.LargeImageURL,
			&mc.Images.BackgroundColor,
			&mc.Images.AnimationURL,
			&mc.Images.YoutubeURL,
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
	MechID        uuid.UUID `db:"mech_id"`
	QueuePosition int64     `db:"queue_position"`
}

func InsertNewMechAndSkin(tx boil.Executor, ownerID uuid.UUID, mechBlueprint *server.BlueprintMech, mechSkinBlueprint *server.BlueprintMechSkin) (*server.Mech, *server.MechSkin, error) {
	L := gamelog.L.With().Str("func", "InsertNewMech").Interface("mechBlueprint", mechBlueprint).Interface("mechSkinBlueprint", mechSkinBlueprint).Str("ownerID", ownerID.String()).Logger()

	// first insert the new skin
	mechSkin, err := InsertNewMechSkin(tx, ownerID, mechSkinBlueprint, &mechBlueprint.ID)
	if err != nil {
		L.Error().Err(err).Msg("failed to insert new mech skin")
		return nil, nil, terror.Error(err)
	}

	// first insert the mech
	newMech := boiler.Mech{
		BlueprintID:           mechBlueprint.ID,
		ChassisSkinID:         mechSkin.ID,
		IsDefault:             false,
		IsInsured:             false,
		Name:                  "",
		GenesisTokenID:        mechBlueprint.GenesisTokenID,
		LimitedReleaseTokenID: mechBlueprint.LimitedReleaseTokenID,
	}

	err = newMech.Insert(tx, boil.Infer())
	if err != nil {
		L.Error().Err(err).Interface("newMech", newMech).Msg("failed to insert new mech")
		return nil, nil, terror.Error(err)
	}

	_, err = InsertNewCollectionItem(tx,
		mechBlueprint.Collection,
		boiler.ItemTypeMech,
		newMech.ID,
		"",
		ownerID.String(),
	)
	if err != nil {
		L.Error().Err(err).Msg("failed to insert col item")
		return nil, nil, terror.Error(err)
	}

	// update skin to say equipped to this mech
	updated, err := boiler.MechSkins(
		boiler.MechSkinWhere.ID.EQ(mechSkin.ID),
	).UpdateAll(tx, boiler.M{
		boiler.MechSkinColumns.EquippedOn: newMech.ID,
	})
	if err != nil {
		L.Error().Err(err).Msg("failed to update mech skin")
		return nil, nil, terror.Error(err)
	}
	if updated != 1 {
		err = fmt.Errorf("updated %d, expected 1", updated)
		L.Error().Err(err).Msg("failed to update mech skin")
		return nil, nil, terror.Error(err)
	}

	mechSkin.EquippedOn = null.StringFrom(newMech.ID)

	mech, err := Mech(tx, newMech.ID)
	if err != nil {
		L.Error().Err(err).Str("newMechID", newMech.ID).Msg("failed to get mech")
		return nil, nil, terror.Error(err)
	}
	return mech, mechSkin, nil
}

func IsMechColumn(col string) bool {
	switch col {
	case boiler.MechColumns.ID,
		boiler.MechColumns.DeletedAt,
		boiler.MechColumns.UpdatedAt,
		boiler.MechColumns.CreatedAt,
		boiler.MechColumns.BlueprintID,
		boiler.MechColumns.IsDefault,
		boiler.MechColumns.IsInsured,
		boiler.MechColumns.Name,
		boiler.MechColumns.GenesisTokenID,
		boiler.MechColumns.LimitedReleaseTokenID,
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
	SortBy              string
	SortDir             SortByDir
	PageSize            int
	Page                int
	OwnerID             string
	QueueSort           *MechListQueueSortOpts
	DisplayXsynMechs    bool
	ExcludeMarketLocked bool
	IncludeMarketListed bool
	ExcludeDamagedMech  bool
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

	queryMods = append(queryMods,
		// where owner id = ?
		GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.CollectionItems,
			Column:   boiler.CollectionItemColumns.OwnerID,
			Operator: OperatorValueTypeEquals,
			Value:    opts.OwnerID,
		}, 0, ""),
		// and item type = mech
		GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.CollectionItems,
			Column:   boiler.CollectionItemColumns.ItemType,
			Operator: OperatorValueTypeEquals,
			Value:    boiler.ItemTypeMech,
		}, 0, "and"),
		// inner join mechs
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.Mechs,
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
		)),
		// inner join mechs skin
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.MechSkin,
			qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.ID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ChassisSkinID),
		)),
		// inner join mechs skin blueprint
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.BlueprintMechSkin,
			qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.ID),
			qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.BlueprintID),
		)),
		// inner join mech blueprint
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.BlueprintMechs,
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.BlueprintID),
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
	if opts.ExcludeDamagedMech {
		queryMods = append(queryMods, qm.Where(
			fmt.Sprintf(
				"NOT EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s ISNULL AND %s * 2 < %s)",
				boiler.TableNames.RepairCases,
				qm.Rels(boiler.TableNames.RepairCases, boiler.RepairCaseColumns.MechID),
				qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
				qm.Rels(boiler.TableNames.RepairCases, boiler.RepairCaseColumns.CompletedAt),
				qm.Rels(boiler.TableNames.RepairCases, boiler.RepairCaseColumns.BlocksRepaired),
				qm.Rels(boiler.TableNames.RepairCases, boiler.RepairCaseColumns.BlocksRequiredRepair),
			),
		))
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
		queryMods = append(queryMods, qm.AndIn(fmt.Sprintf("%s IN ?", qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.ID)), vals...))
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

	//Search
	if opts.Search != "" {
		xSearch := ParseQueryText(opts.Search, true)
		if len(xSearch) > 0 {
			queryMods = append(queryMods,
				qm.And(fmt.Sprintf(
					"(to_tsvector('english', %s) @@ to_tsquery(?))",
					qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.Label),
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
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.Label),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.WeaponHardpoints),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.UtilitySlots),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.Speed),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.MaxHitpoints),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.IsDefault),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.IsInsured),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.GenesisTokenID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.LimitedReleaseTokenID),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.PowerCoreSize),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.PowerCoreID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.BlueprintID),
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
					SELECT  _bq.mech_id, row_number () OVER (ORDER BY _bq.queued_at) AS queue_position
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
	} else if opts.Sort != nil && opts.Sort.Table == boiler.TableNames.Mechs && IsMechColumn(opts.Sort.Column) && opts.Sort.Direction.IsValid() {
		queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s.%s %s", boiler.TableNames.Mechs, opts.Sort.Column, opts.Sort.Direction)))
	} else if opts.SortBy != "" && opts.SortDir.IsValid() {
		if opts.SortBy == "alphabetical" {
			queryMods = append(queryMods,
				qm.OrderBy(
					fmt.Sprintf("(CASE WHEN %[1]s IS NOT NULL AND %[1]s != '' THEN %[1]s ELSE %[2]s END) %[3]s",
						qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Name),
						qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.Label),
						opts.SortDir,
					)))
		} else if opts.SortBy == "rarity" {
			queryMods = append(queryMods, GenerateTierSort(qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.ID), opts.SortDir))
		}
	} else {
		queryMods = append(queryMods,
			qm.OrderBy(
				fmt.Sprintf("(CASE WHEN %[1]s IS NOT NULL AND %[1]s != '' THEN %[1]s ELSE %[2]s END) ASC",
					qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Name),
					qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.Label),
				)))
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

func MechEquippedOnDetails(tx boil.Executor, equippedOnID string) (*server.EquippedOnDetails, error) {
	eid := &server.EquippedOnDetails{}

	err := boiler.NewQuery(
		qm.Select(
			boiler.CollectionItemColumns.ItemID,
			boiler.CollectionItemColumns.Hash,
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Name),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.Label),
		),
		qm.From(boiler.TableNames.CollectionItems),
		qm.InnerJoin(fmt.Sprintf(
			"%s on %s = %s",
			boiler.TableNames.Mechs,
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
		)),
		qm.InnerJoin(fmt.Sprintf(
			"%s on %s = %s",
			boiler.TableNames.BlueprintMechs,
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.BlueprintID),
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

func MechBattleReady(mechID string) (bool, error) {
	q := `
		SELECT (_bm.availability_id IS NULL OR _a.available_at <= NOW())
		FROM blueprint_mechs _bm 
			LEFT JOIN availabilities _a ON _a.id = _bm.availability_id
		WHERE _bm.id = (SELECT m.blueprint_id FROM mechs m WHERE m.id = $1)
		LIMIT 1
	`

	battleReady := false

	err := gamedb.StdConn.QueryRow(q, mechID).Scan(&battleReady)
	if err != nil {
		return false, terror.Error(err, "Failed to load battle ready status")
	}

	return battleReady, nil
}
