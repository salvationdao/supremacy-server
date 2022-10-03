package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"server"
	"server/benchmark"
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
			qm.Rels(boiler.TableNames.MechSkin, boiler.BlueprintMechSkinColumns.Tier), // get tier from blueprint mech skin table
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
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.MechType),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.WeaponHardpoints),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.UtilitySlots),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.Speed),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.MaxHitpoints),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ShieldMax),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ShieldRechargeRate),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ShieldRechargeDelay),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ShieldRechargePowerCost),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ShieldTypeID),
			qm.Rels(boiler.TableNames.BlueprintShieldTypes, boiler.BlueprintShieldTypeColumns.Label),
			qm.Rels(boiler.TableNames.BlueprintShieldTypes, boiler.BlueprintShieldTypeColumns.Description),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.RepairBlocks),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.BoostStat),
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
				SELECT
				_pc.*,
				_ci.hash,
				_ci.token_id,
				_ci.tier,
				_ci.owner_id,
				_bppc.label,
				_bppc.size,
				_bppc.capacity,
				_bppc.max_draw_rate,
				_bppc.recharge_rate,
				_bppc.armour,
				_bppc.max_hitpoints,
				_bppc.image_url AS image_url,
				_bppc.avatar_url AS avatar_url,
				_bppc.card_animation_url AS card_animation_url,
				_bppc.animation_url AS animation_url
			FROM
				power_cores _pc
				INNER JOIN collection_items _ci ON _ci.item_id = _pc.id
				INNER JOIN blueprint_power_cores _bppc ON _pc.blueprint_id = _bppc.id
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
		// inner join mech shield type details
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.BlueprintShieldTypes,
			qm.Rels(boiler.TableNames.BlueprintShieldTypes, boiler.BlueprintShieldTypeColumns.ID),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ShieldTypeID),
		)),
		// inner join brand
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.Brands,
			qm.Rels(boiler.TableNames.Brands, boiler.BrandColumns.ID),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.BrandID),
		)),
		// inner join skin
		qm.InnerJoin(fmt.Sprintf(`(
			SELECT
				_ms.*,
				_ci.hash,
				_ci.token_id,
				_ci.tier,
				_ci.owner_id,
				_bpms.label,
				_bpms.default_level,
				json_build_object(
					'image_url', _bpms.image_url,
					'card_animation_url', _bpms.card_animation_url,
					'avatar_url', _bpms.avatar_url,
					'large_image_url', _bpms.large_image_url,
					'background_color', _bpms.background_color,
					'animation_url', _bpms.animation_url,
					'youtube_url', _bpms.youtube_url
				) AS swatch_images
			FROM
				mech_skin _ms
				INNER JOIN collection_items _ci ON _ci.item_id = _ms.id
				INNER JOIN blueprint_mech_skin _bpms ON _bpms.id = _ms.blueprint_id
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
						SELECT 
								mw.%s, 
								json_agg(w2) as weapons
						FROM %s mw
						INNER JOIN
							(
								SELECT 	
										_w.*,
										_ci.hash,
										_ci.token_id,
										_ci.tier,
										_ci.owner_id,
										to_json(_ws) as weapon_skin,
										_bpw.label,
										_bpw.weapon_type,
										_bpw.damage,
										_bpw.default_damage_type,
										_bpw.damage_falloff,
										_bpw.damage_falloff_rate,
										_bpw.spread,
										_bpw.rate_of_fire,
										_bpw.radius,
										_bpw.radius_damage_falloff,
										_bpw.projectile_speed,
										_bpw.power_cost,
										_bpw.max_ammo,
										_bpw.is_melee,
										_bpw.projectile_amount,
										_bpw.dot_tick_damage,
										_bpw.dot_max_ticks,
										_bpw.is_arced,
										_bpw.charge_time_seconds,
										_bpw.burst_rate_of_fire,
										_wmsc.image_url as image_url,
										_wmsc.avatar_url as avatar_url,
										_wmsc.card_animation_url as card_animation_url,
										_wmsc.animation_url as animation_url,
										_mw.slot_number AS slot_number
								FROM weapons _w
								INNER JOIN collection_items _ci on _ci.item_id = _w.id
								INNER JOIN blueprint_weapons _bpw on _bpw.id = _w.blueprint_id
								INNER JOIN mech_weapons _mw ON _mw.weapon_id = _w.id
								INNER JOIN (
										SELECT 
												__ws.*,
												_ci.hash, 
												_ci.token_id, 
												_ci.tier, 
												_ci.owner_id
										FROM weapon_skin __ws
										INNER JOIN collection_items _ci on _ci.item_id = __ws.id
								) _ws ON _ws.equipped_on = _w.id
								INNER JOIN weapon_model_skin_compatibilities _wmsc on _wmsc.blueprint_weapon_skin_id = _ws.blueprint_id and _wmsc.weapon_model_id = _bpw.id
							) w2 ON mw.weapon_id = w2.id
						GROUP BY mw.chassis_id
				) %s on %s = %s `,
				boiler.MechWeaponColumns.ChassisID,
				boiler.TableNames.MechWeapons,
				weaponsTableName,
				qm.Rels(weaponsTableName, boiler.MechWeaponColumns.ChassisID),
				qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
			)),
		// left join utility
		qm.LeftOuterJoin(
			// TODO: make this boiler/typesafe
			fmt.Sprintf(`
				(
					SELECT
						mu.chassis_id,
						json_agg(_u) AS utility
					FROM
						mech_utility mu
						INNER JOIN (
							SELECT
								_u.*,
								_ci.hash,
								_ci.token_id,
								_ci.tier,
								_ci.owner_id,
								_bpu.image_url as image_url,
								_bpu.avatar_url as avatar_url,
								_bpu.card_animation_url as card_animation_url,
								_bpu.animation_url as animation_url,
								_bpu.label as label,
								_mu.slot_number AS slot_number
								-- to_json(_us) as shield
						FROM utility _u
						INNER JOIN collection_items _ci on _ci.item_id = _u.id
						INNER JOIN blueprint_utility _bpu on _bpu.id = _u.blueprint_id
						-- INNER JOIN blueprint_utility_shield _us ON _us.blueprint_utility_id = _u.blueprint_id
						INNER JOIN mech_utility _mu ON _mu.utility_id = _u.id
					) _u ON mu.utility_id = _u.id
					GROUP BY
						mu.chassis_id
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

func GetBlueprintWeaponsIDsWithCompatibleSkinInheritanceFromMechID(conn boil.Executor, mechID string) ([]string, error) {
	// select
	// bw.id
	// from mechs m
	// inner join mech_skin ms on ms.id = m.chassis_skin_id
	// inner join blueprint_mech_skin bms on bms.id = ms.blueprint_id
	// inner join weapon_model_skin_compatibilities wmsc on wmsc.blueprint_weapon_skin_id  = bms.blueprint_weapon_skin_id
	// inner join weapon_models wm on wm.id = wmsc.weapon_model_id
	// inner join blueprint_weapons bw on bw.weapon_model_id = wm.id
	// where m.id = 'cda9cff8-4c03-45f8-b59c-2cefd68e1386';

	var result []struct {
		ID string `boil:"id"`
	}
	err := boiler.NewQuery(
		qm.Select(fmt.Sprintf("%s as id", qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.ID))),
		qm.From(boiler.TableNames.Mechs),
		qm.InnerJoin(fmt.Sprintf("%s on %s = %s",
			boiler.TableNames.MechSkin,
			qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.ID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ChassisSkinID),
		)),
		qm.InnerJoin(fmt.Sprintf("%s on %s = %s",
			boiler.TableNames.BlueprintMechSkin,
			qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.ID),
			qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.BlueprintID),
		)),
		qm.InnerJoin(fmt.Sprintf("%s on %s = %s",
			boiler.TableNames.WeaponModelSkinCompatibilities,
			qm.Rels(boiler.TableNames.WeaponModelSkinCompatibilities, boiler.WeaponModelSkinCompatibilityColumns.BlueprintWeaponSkinID),
			qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.BlueprintWeaponSkinID),
		)),
		qm.InnerJoin(fmt.Sprintf("%s on %s = %s",
			boiler.TableNames.BlueprintWeapons,
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.ID),
			qm.Rels(boiler.TableNames.WeaponModelSkinCompatibilities, boiler.WeaponModelSkinCompatibilityColumns.WeaponModelID),
		)),
		qm.Where(fmt.Sprintf("%s = ?", qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID)),
			mechID,
		),
	).Bind(nil, conn, &result)
	if err != nil {
		return []string{}, err
	}

	ids := []string{}
	for _, i := range result {
		ids = append(ids, i.ID)
	}

	return ids, nil
}

func GetCompatibleBlueprintMechSkinIDsFromMechID(conn boil.Executor, mechID string) ([]string, error) {
	// select
	// mmsc.blueprint_mech_skin_id
	// from mechs m
	// inner join mech_model_skin_compatibilities mmsc on mmsc.mech_model_id = m.blueprint_id
	// where m.id = 'cda9cff8-4c03-45f8-b59c-2cefd68e1386';

	var result []struct {
		ID string `boil:"id"`
	}
	err := boiler.NewQuery(
		qm.Select(fmt.Sprintf("%s as id", boiler.MechModelSkinCompatibilityTableColumns.BlueprintMechSkinID)),
		qm.From(boiler.TableNames.Mechs),
		qm.InnerJoin(fmt.Sprintf("%s on %s = %s",
			boiler.TableNames.MechModelSkinCompatibilities,
			boiler.MechModelSkinCompatibilityTableColumns.MechModelID,
			boiler.MechTableColumns.BlueprintID,
		)),
		qm.Where(fmt.Sprintf("%s = ?", boiler.MechTableColumns.ID),
			mechID),
	).Bind(nil, conn, &result)
	if err != nil {
		return []string{}, err
	}

	ids := []string{}
	for _, i := range result {
		ids = append(ids, i.ID)
	}

	return ids, nil
}

func Mech(conn boil.Executor, mechID string) (*server.Mech, error) {
	bm := benchmark.New()
	bm.Start("db Mech")
	defer func() {
		bm.End("db Mech")
		bm.Alert(150)
	}()

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
			&mc.MechType,
			&mc.WeaponHardpoints,
			&mc.UtilitySlots,
			&mc.Speed,
			&mc.MaxHitpoints,
			&mc.Shield,
			&mc.ShieldRechargeRate,
			&mc.ShieldRechargeDelay,
			&mc.ShieldRechargePowerCost,
			&mc.ShieldTypeID,
			&mc.ShieldTypeLabel,
			&mc.ShieldTypeDescription,
			&mc.RepairBlocks,
			&mc.BoostedStat,
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

	compatibleWeapons, err := GetBlueprintWeaponsIDsWithCompatibleSkinInheritanceFromMechID(gamedb.StdConn, mc.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to find skin inherited weapon models for mech with id %s", mc.ID)
	}
	mc.BlueprintWeaponIDsWithSkinInheritance = compatibleWeapons

	compatibleSkins, err := GetCompatibleBlueprintMechSkinIDsFromMechID(gamedb.StdConn, mc.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to find compatible skins for mech with id %s", mc.ID)
	}
	mc.CompatibleBlueprintMechSkinIDs = compatibleSkins

	if mc.ChassisSkin.Images == nil {
		mc.ChassisSkin.Images = mc.Images
	}

	err = mc.SetBoostedStats()
	if err != nil {
		gamelog.L.Error().Err(err).Interface("mech", mc).Msg("failed to set boosted stats")
		return nil, fmt.Errorf("failed to set stats for mech")
	}

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
			&mc.MechType,
			&mc.Label,
			&mc.WeaponHardpoints,
			&mc.UtilitySlots,
			&mc.Speed,
			&mc.MaxHitpoints,
			&mc.Shield,
			&mc.ShieldRechargeRate,
			&mc.ShieldRechargeDelay,
			&mc.ShieldRechargePowerCost,
			&mc.ShieldTypeID,
			&mc.ShieldTypeLabel,
			&mc.ShieldTypeDescription,
			&mc.RepairBlocks,
			&mc.BoostedStat,
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

	for _, mc := range mcs {
		err = mc.SetBoostedStats()
		if err != nil {
			gamelog.L.Error().Err(err).Interface("mech", mc).Msg("failed to set boosted stats")
			return nil, fmt.Errorf("failed to set stats for mech")
		}
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
		// inner join mech shield type details
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.BlueprintShieldTypes,
			qm.Rels(boiler.TableNames.BlueprintShieldTypes, boiler.BlueprintShieldTypeColumns.ID),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ShieldTypeID),
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
				`NOT EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s ISNULL AND %s * 2 > %s)`,
				boiler.TableNames.RepairCases,
				qm.Rels(boiler.TableNames.RepairCases, boiler.RepairCaseColumns.MechID),
				qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
				qm.Rels(boiler.TableNames.RepairCases, boiler.RepairCaseColumns.CompletedAt),
				qm.Rels(boiler.TableNames.RepairCases, boiler.RepairCaseColumns.BlocksRequiredRepair),
				qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.RepairBlocks),
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
		queryMods = append(queryMods, qm.AndIn(fmt.Sprintf("%s IN ?", qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.Tier)), vals...))
	}
	if len(opts.FilterStatuses) > 0 {
		hasIdleToggled := false
		hasInBattleToggled := false
		hasMarketplaceToggled := false
		hasInQueueToggled := false
		hasBattleReadyToggled := false

		statusFilters := []qm.QueryMod{}

		for _, s := range opts.FilterStatuses {
			switch s {
			case "IDLE":
				if hasIdleToggled {
					continue
				}
				hasIdleToggled = true
				statusFilters = append(statusFilters, qm.Or(fmt.Sprintf(
					`NOT EXISTS (
						SELECT 1
						FROM %s _bq
						WHERE _bq.%s = %s 
							AND _bq.%s ISNULL 
							AND _bq.%s ISNULL 
							AND _bq.%s ISNULL
						LIMIT 1
					)`,
					boiler.TableNames.BattleLobbiesMechs,
					boiler.BattleLobbiesMechColumns.MechID,
					boiler.MechTableColumns.ID,
					boiler.BattleLobbiesMechColumns.EndedAt,
					boiler.BattleLobbiesMechColumns.RefundTXID,
					boiler.BattleLobbiesMechColumns.DeletedAt,
				)))
			case "BATTLE":
				if hasInBattleToggled {
					continue
				}
				hasInBattleToggled = true
				statusFilters = append(statusFilters, qm.Or(fmt.Sprintf(
					`EXISTS (
						SELECT 1
						FROM %s _bq
						WHERE _bq.%s = %s
							AND _bq.%s IS NULL
							AND _bq.%s IS NOT NULL
							AND _bq.%s IS NULL
							AND _bq.%s IS NULL
						LIMIT 1
					)`,
					boiler.TableNames.BattleLobbiesMechs,
					boiler.BattleLobbiesMechColumns.MechID,
					boiler.MechTableColumns.ID,
					boiler.BattleLobbiesMechColumns.EndedAt,
					boiler.BattleLobbiesMechColumns.AssignedToBattleID,
					boiler.BattleLobbiesMechColumns.RefundTXID,
					boiler.BattleLobbiesMechColumns.DeletedAt,
				)))
			case "MARKET":
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
			case "QUEUE":
				if hasInQueueToggled {
					continue
				}
				hasInQueueToggled = true
				statusFilters = append(statusFilters, qm.Or(fmt.Sprintf(
					`EXISTS (
						SELECT 1
						FROM %s _bq
						WHERE _bq.%s = %s
							AND _bq.%s IS NULL
							AND _bq.%s IS NULL
							AND _bq.%s IS NULL
							AND _bq.%s IS NULL
						LIMIT 1
					)`,
					boiler.TableNames.BattleLobbiesMechs,
					boiler.BattleLobbiesMechColumns.MechID,
					boiler.MechTableColumns.ID,
					boiler.BattleLobbiesMechColumns.EndedAt,
					boiler.BattleLobbiesMechColumns.AssignedToBattleID,
					boiler.BattleLobbiesMechColumns.RefundTXID,
					boiler.BattleLobbiesMechColumns.DeletedAt,
				)))
			case "BATTLE_READY":
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
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ShieldMax),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ShieldRechargeRate),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ShieldRechargeDelay),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ShieldRechargePowerCost),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ShieldTypeID),
			qm.Rels(boiler.TableNames.BlueprintShieldTypes, boiler.BlueprintShieldTypeColumns.Label),
			qm.Rels(boiler.TableNames.BlueprintShieldTypes, boiler.BlueprintShieldTypeColumns.Description),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.RepairBlocks),
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.BoostStat),
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
			qm.LeftOuterJoin(fmt.Sprintf(`%s ON %s = %s AND %s ISNULL AND %s ISNULL AND %s ISNULL`,
				boiler.TableNames.BattleLobbiesMechs,
				boiler.BattleLobbiesMechTableColumns.MechID,
				boiler.MechTableColumns.ID,
				boiler.BattleLobbiesMechTableColumns.EndedAt,
				boiler.BattleLobbiesMechTableColumns.RefundTXID,
				boiler.BattleLobbiesMechTableColumns.DeletedAt,
			)),
			qm.OrderBy(fmt.Sprintf("%s %s NULLS LAST, %s, %s",
				boiler.BattleLobbiesMechTableColumns.LockedAt,
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
			queryMods = append(queryMods, GenerateTierSort(qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.Tier), opts.SortDir))
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
			&mc.Shield,
			&mc.ShieldRechargeRate,
			&mc.ShieldRechargeDelay,
			&mc.ShieldRechargePowerCost,
			&mc.ShieldTypeID,
			&mc.ShieldTypeLabel,
			&mc.ShieldTypeDescription,
			&mc.RepairBlocks,
			&mc.BoostedStat,
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

type MechBrief struct {
	ID                  string      `json:"id" db:"id"`
	OwnerID             string      `json:"owner_id" db:"owner_id"`
	MarketLocked        bool        `json:"market_locked" db:"market_locked"`
	XsynLocked          bool        `json:"xsyn_locked" db:"xsyn_locked"`
	LockedToMarketplace bool        `json:"locked_to_marketplace" db:"locked_to_marketplace"`
	Name                string      `json:"name" db:"name"`
	ChassisSkinID       string      `json:"chassis_skin_id" db:"chassis_skin_id"`
	PowerCoreID         null.String `json:"power_core_id,omitempty" db:"power_core_id,omitempty"`
	IntroAnimationID    null.String `json:"intro_animation_id,omitempty" db:"intro_animation_id,omitempty"`
	OutroAnimationID    null.String `json:"outro_animation_id,omitempty" db:"outro_animation_id,omitempty"`
	Label               string      `json:"label" db:"label"`
	RepairBlocks        int         `json:"repair_blocks" db:"repair_blocks"`
	Tier                string      `json:"tier" db:"tier"`
	ImageUrl            string      `json:"image_url" db:"image_url"`
	AvatarUrl           string      `json:"avatar_url" db:"avatar_url"`

	MechType         string      `json:"mech_type" db:"mech_type"`
	BoostStat        null.String `json:"boost_stat,omitempty" db:"boost_stat,omitempty"`
	WeaponHardpoints int         `json:"weapon_hardpoints" db:"weapon_hardpoints"`
	PowerCoreSize    string      `json:"power_core_size" db:"power_core_size"`
	UtilitySlots     int         `json:"utility_slots" db:"utility_slots"`
	Speed            int64       `json:"speed" db:"speed"`
	BoostedSpeed     int64       `json:"boosted_speed"`

	MaxHitpoints        int64 `json:"max_hitpoints" db:"max_hitpoints"`
	BoostedMaxHitpoints int64 `json:"boosted_max_hitpoints"`

	Collection                string `json:"collection" db:"collection"`
	ShieldMax                 int    `json:"shield" db:"shield_max"`
	ShieldRechargeRate        int64  `json:"shield_recharge_rate" db:"shield_recharge_rate"`
	BoostedShieldRechargeRate int64  `json:"boosted_shield_recharge_rate"`

	ShieldRechargePowerCost int   `json:"shield_recharge_power_cost" db:"shield_recharge_power_cost"`
	MechSkinLevel           int64 `json:"mech_skin_level" db:"level"`

	LobbyLockedAt         null.Time           `json:"lobby_locked_at,omitempty" db:"lobby_locked_at"`
	AssignedToBattleID    null.String         `json:"assigned_to_battle_id,omitempty" db:"assigned_to_battle_id"`
	LobbyNumber           null.Int            `json:"lobby_number,omitempty" db:"lobby_number"`
	PowerCoreCapacity     decimal.NullDecimal `json:"power_core_capacity,omitempty" db:"capacity,omitempty"`
	PowerCoreRechargeRate decimal.NullDecimal `json:"power_core_recharge_rate,omitempty" db:"recharge_rate,omitempty"`

	DamagedBlocks int  `json:"damaged_blocks" db:"damaged_blocks"`
	IsBattleReady bool `json:"is_battle_ready" db:"is_battle_ready"`
	InMarketPlace bool `json:"in_market_place" db:"in_market_place"`

	EquippedWeaponCount  int `json:"equipped_weapon_count" db:"equipped_weapon_count"`
	EquippedUtilityCount int `json:"equipped_utility_count" db:"equipped_utility_count"`

	IsStaked bool `json:"is_staked" db:"is_staked"`

	Status    server.MechArenaStatus `json:"status"` // "QUEUE" | "BATTLE" | "MARKET" | "IDLE"
	CanDeploy bool                   `json:"can_deploy"`
}

// OwnedMechsBrief return list for mech for quick deploy
func OwnedMechsBrief(playerID string, mechIDs ...string) ([]*MechBrief, error) {
	// prevent sql injection, check player id is in uuid format
	_, err := uuid.FromString(playerID)
	if err != nil {
		return nil, terror.Error(err, "The player id is not in uuid format.")
	}

	mechIDInQuery := ""
	if len(mechIDs) > 0 {
		mechIDInQuery += fmt.Sprintf("AND %s IN(", boiler.CollectionItemColumns.ItemID)

		for i, mechID := range mechIDs {
			// prevent sql injection, check each mech id is in uuid format
			_, err := uuid.FromString(mechID)
			if err != nil {
				return nil, terror.Error(err, "The mech id is not in uuid format.")
			}

			mechIDInQuery += "'" + mechID + "'"
			if i < len(mechIDs)-1 {
				mechIDInQuery += ","
				continue
			}

			mechIDInQuery += ")"
		}
	}

	queries := []qm.QueryMod{
		qm.Select(
			fmt.Sprintf("_m.%s", boiler.MechColumns.ID),
			fmt.Sprintf("_ci.%s", boiler.CollectionItemColumns.OwnerID),
			fmt.Sprintf("_ci.%s", boiler.CollectionItemColumns.MarketLocked),
			fmt.Sprintf("_ci.%s", boiler.CollectionItemColumns.XsynLocked),
			fmt.Sprintf("_ci.%s", boiler.CollectionItemColumns.LockedToMarketplace),
			fmt.Sprintf("_m.%s", boiler.MechColumns.Name),
			fmt.Sprintf("_m.%s", boiler.MechColumns.ChassisSkinID),
			fmt.Sprintf("_m.%s", boiler.MechColumns.PowerCoreID),
			fmt.Sprintf("_m.%s", boiler.MechColumns.IntroAnimationID),
			fmt.Sprintf("_m.%s", boiler.MechColumns.OutroAnimationID),
			fmt.Sprintf("_bm.%s", boiler.BlueprintMechColumns.Label),
			fmt.Sprintf("_bm.%s", boiler.BlueprintMechColumns.RepairBlocks),
			fmt.Sprintf("_bm.%s", boiler.BlueprintMechColumns.MechType),
			fmt.Sprintf("_bm.%s", boiler.BlueprintMechColumns.BoostStat),
			fmt.Sprintf("_bm.%s", boiler.BlueprintMechColumns.WeaponHardpoints),
			fmt.Sprintf("_bm.%s", boiler.BlueprintMechColumns.PowerCoreSize),
			fmt.Sprintf("_bm.%s", boiler.BlueprintMechColumns.UtilitySlots),
			fmt.Sprintf("_bm.%s", boiler.BlueprintMechColumns.Speed),
			fmt.Sprintf("_bm.%s", boiler.BlueprintMechColumns.MaxHitpoints),
			fmt.Sprintf("_bm.%s", boiler.BlueprintMechColumns.Collection),
			fmt.Sprintf("_bm.%s", boiler.BlueprintMechColumns.ShieldMax),
			fmt.Sprintf("_bm.%s", boiler.BlueprintMechColumns.ShieldRechargeRate),
			fmt.Sprintf("_bm.%s", boiler.BlueprintMechColumns.ShieldRechargePowerCost),
			fmt.Sprintf("_ms.%s", boiler.MechSkinColumns.Level),
			fmt.Sprintf("_bms.%s", boiler.BlueprintMechSkinColumns.Tier),
			fmt.Sprintf("_mmsc.%s", boiler.MechModelSkinCompatibilityColumns.ImageURL),
			fmt.Sprintf("_mmsc.%s", boiler.MechModelSkinCompatibilityColumns.AvatarURL),
			fmt.Sprintf("_blm.%s", boiler.BattleLobbiesMechColumns.LockedAt),
			fmt.Sprintf("_blm.%s", boiler.BattleLobbiesMechColumns.AssignedToBattleID),
			fmt.Sprintf("_blm.%s", boiler.BattleLobbyColumns.Number),
			fmt.Sprintf("_pc.%s", boiler.BlueprintPowerCoreColumns.Capacity),
			fmt.Sprintf("_pc.%s", boiler.BlueprintPowerCoreColumns.RechargeRate),
			fmt.Sprintf(
				"COALESCE((SELECT _rc.%s - _rc.%s FROM %s _rc WHERE _rc.%s = _ci.%s AND _rc.%s ISNULL AND _rc.%s ISNULL LIMIT 1), 0) AS damaged_blocks",
				boiler.RepairCaseColumns.BlocksRequiredRepair,
				boiler.RepairCaseColumns.BlocksRepaired,
				boiler.TableNames.RepairCases,
				boiler.RepairCaseColumns.MechID,
				boiler.CollectionItemColumns.ItemID,
				boiler.RepairCaseColumns.CompletedAt,
				boiler.RepairCaseColumns.DeletedAt,
			),
			fmt.Sprintf(
				"COALESCE((SELECT _a.%s <= now() FROM %s _a WHERE _a.%s = _bm.%s), TRUE) AS is_battle_ready",
				boiler.AvailabilityColumns.AvailableAt,
				boiler.TableNames.Availabilities,
				boiler.AvailabilityColumns.ID,
				boiler.BlueprintMechColumns.AvailabilityID,
			),
			fmt.Sprintf(
				"COALESCE((SELECT TRUE FROM %s _is WHERE _is.%s = _ci.%s AND _is.%s > NOW() AND _is.%s ISNULL AND _is.%s ISNULL), FALSE) AS in_market_place",
				boiler.TableNames.ItemSales,
				boiler.ItemSaleColumns.CollectionItemID,
				boiler.CollectionItemColumns.ID,
				boiler.ItemSaleColumns.EndAt,
				boiler.ItemSaleColumns.SoldAt,
				boiler.ItemSaleColumns.DeletedAt,
			),
			fmt.Sprintf(
				"COALESCE((SELECT COUNT(_w.%s) FROM %s _w WHERE _w.%s = _m.%s), 0) AS equipped_weapon_count",
				boiler.WeaponColumns.ID,
				boiler.TableNames.Weapons,
				boiler.WeaponColumns.EquippedOn,
				boiler.MechColumns.ID,
			),
			fmt.Sprintf(
				"COALESCE((SELECT COUNT(_u.%s) FROM %s _u WHERE _u.%s = _m.%s), 0) AS equipped_utility_count",
				boiler.UtilityColumns.ID,
				boiler.TableNames.Utility,
				boiler.UtilityColumns.EquippedOn,
				boiler.MechColumns.ID,
			),
			fmt.Sprintf(
				"COALESCE((SELECT true FROM %s WHERE %s = _m.%s), false) AS is_staked",
				boiler.TableNames.StakedMechs,
				boiler.StakedMechTableColumns.MechID,
				boiler.MechColumns.ID,
			),
		),

		qm.From(fmt.Sprintf(
			"(SELECT %s, %s, %s, %s, %s, %s FROM %s WHERE %s = 'mech' AND %s = '%s' AND %s IS NULL %s) _ci",
			boiler.CollectionItemColumns.ID,
			boiler.CollectionItemColumns.ItemID,
			boiler.CollectionItemColumns.OwnerID,
			boiler.CollectionItemColumns.MarketLocked,
			boiler.CollectionItemColumns.XsynLocked,
			boiler.CollectionItemColumns.LockedToMarketplace,
			boiler.TableNames.CollectionItems,
			boiler.CollectionItemColumns.ItemType,
			boiler.CollectionItemColumns.OwnerID,
			playerID,
			boiler.CollectionItemColumns.DeletedAt,
			mechIDInQuery,
		)),

		qm.InnerJoin(fmt.Sprintf(
			"(SELECT %s, %s, %s, %s, %s, %s, %s FROM %s) _m ON _m.%s = _ci.%s",
			boiler.MechColumns.ID,
			boiler.MechColumns.BlueprintID,
			boiler.MechColumns.ChassisSkinID,
			boiler.MechColumns.Name,
			boiler.MechColumns.PowerCoreID,
			boiler.MechColumns.IntroAnimationID,
			boiler.MechColumns.OutroAnimationID,
			boiler.TableNames.Mechs,
			boiler.MechColumns.ID,
			boiler.CollectionItemColumns.ItemID,
		)),

		qm.InnerJoin(fmt.Sprintf(
			"%s _bm ON _bm.%s = _m.%s",
			boiler.TableNames.BlueprintMechs,
			boiler.BlueprintMechColumns.ID,
			boiler.MechColumns.BlueprintID,
		)),

		qm.InnerJoin(fmt.Sprintf(
			"(SELECT %s, %s, %s FROM %s) _ms ON _ms.%s = _m.%s",
			boiler.MechSkinColumns.ID,
			boiler.MechSkinColumns.BlueprintID,
			boiler.MechSkinColumns.Level,
			boiler.TableNames.MechSkin,
			boiler.MechSkinColumns.ID,
			boiler.MechColumns.ChassisSkinID,
		)),

		qm.InnerJoin(fmt.Sprintf(
			"(SELECT %s, %s FROM %s) _bms ON _bms.%s = _ms.%s",
			boiler.BlueprintMechSkinColumns.ID,
			boiler.BlueprintMechSkinColumns.Tier,
			boiler.TableNames.BlueprintMechSkin,
			boiler.BlueprintMechSkinColumns.ID,
			boiler.MechSkinColumns.BlueprintID,
		)),

		qm.InnerJoin(fmt.Sprintf(
			"(SELECT %s, %s, %s, %s FROM %s) _mmsc ON _mmsc.%s = _m.%s AND _mmsc.%s = _ms.%s",
			boiler.MechModelSkinCompatibilityColumns.MechModelID,
			boiler.MechModelSkinCompatibilityColumns.BlueprintMechSkinID,
			boiler.MechModelSkinCompatibilityColumns.ImageURL,
			boiler.MechModelSkinCompatibilityColumns.AvatarURL,
			boiler.TableNames.MechModelSkinCompatibilities,
			boiler.MechModelSkinCompatibilityColumns.MechModelID,
			boiler.MechColumns.BlueprintID,
			boiler.MechModelSkinCompatibilityColumns.BlueprintMechSkinID,
			boiler.MechSkinColumns.BlueprintID,
		)),

		qm.LeftOuterJoin(fmt.Sprintf(
			`(
				SELECT 
					%s,
					%s,
					%s,
					(SELECT %s FROM %s WHERE %s = %s)
				FROM %s
				WHERE %s ISNULL AND %s ISNULL
			) _blm ON _blm.%s = _ci.%s`,
			boiler.BattleLobbiesMechTableColumns.MechID,
			boiler.BattleLobbiesMechTableColumns.LockedAt,
			boiler.BattleLobbiesMechTableColumns.AssignedToBattleID,
			boiler.BattleLobbyTableColumns.Number,
			boiler.TableNames.BattleLobbies,
			boiler.BattleLobbyTableColumns.ID,
			boiler.BattleLobbiesMechTableColumns.BattleLobbyID,
			boiler.TableNames.BattleLobbiesMechs,
			boiler.BattleLobbiesMechTableColumns.EndedAt,
			boiler.BattleLobbiesMechTableColumns.DeletedAt,
			boiler.BattleLobbiesMechColumns.MechID,
			boiler.CollectionItemColumns.ItemID,
		)),

		// outer join power cores
		qm.LeftOuterJoin(fmt.Sprintf(`(
					SELECT %s,%s, %s
					FROM %s
					INNER JOIN %s ON %s = %s
					) _pc ON _pc.%s = _m.%s`,
			boiler.PowerCoreTableColumns.ID,
			boiler.BlueprintPowerCoreTableColumns.Capacity,
			boiler.BlueprintPowerCoreTableColumns.RechargeRate,
			boiler.TableNames.PowerCores,
			boiler.TableNames.BlueprintPowerCores,
			boiler.BlueprintPowerCoreTableColumns.ID,
			boiler.PowerCoreTableColumns.BlueprintID,
			boiler.PowerCoreColumns.ID,
			boiler.MechColumns.PowerCoreID,
		)),
	}

	rows, err := boiler.NewQuery(queries...).Query(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msg("Failed to load mechs")
		return nil, terror.Error(err, "Failed to load mechs")
	}

	defer rows.Close()

	resp := []*MechBrief{}
	for rows.Next() {
		mb := &MechBrief{}
		err = rows.Scan(
			&mb.ID,
			&mb.OwnerID,
			&mb.MarketLocked,
			&mb.XsynLocked,
			&mb.LockedToMarketplace,
			&mb.Name,
			&mb.ChassisSkinID,
			&mb.PowerCoreID,
			&mb.IntroAnimationID,
			&mb.OutroAnimationID,
			&mb.Label,
			&mb.RepairBlocks,
			&mb.MechType,
			&mb.BoostStat,
			&mb.WeaponHardpoints,
			&mb.PowerCoreSize,
			&mb.UtilitySlots,
			&mb.Speed,
			&mb.MaxHitpoints,
			&mb.Collection,
			&mb.ShieldMax,
			&mb.ShieldRechargeRate,
			&mb.ShieldRechargePowerCost,
			&mb.MechSkinLevel,
			&mb.Tier,
			&mb.ImageUrl,
			&mb.AvatarUrl,
			&mb.LobbyLockedAt,
			&mb.AssignedToBattleID,
			&mb.LobbyNumber,

			&mb.PowerCoreCapacity,
			&mb.PowerCoreRechargeRate,

			&mb.DamagedBlocks,
			&mb.IsBattleReady,
			&mb.InMarketPlace,

			&mb.EquippedWeaponCount,
			&mb.EquippedUtilityCount,
			&mb.IsStaked,
		)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to scan player battle spectated from db.")
			return nil, terror.Error(err, "Failed to load player battles spectated.")
		}

		// parse queue stat
		mb.Status = server.MechArenaStatusIdle
		mb.CanDeploy = true
		if mb.InMarketPlace {
			mb.Status = server.MechArenaStatusMarket
			mb.CanDeploy = false
		} else if mb.AssignedToBattleID.Valid {
			mb.Status = server.MechArenaStatusBattle
			mb.CanDeploy = false
		} else if mb.LobbyNumber.Valid {
			mb.Status = server.MechArenaStatusQueue
			mb.CanDeploy = false
		} else if mb.DamagedBlocks > 0 {
			mb.Status = server.MechArenaStatusDamaged
			mb.CanDeploy = false
			// if repair more than half of the blocks
			if mb.DamagedBlocks*2 < mb.RepairBlocks {
				mb.CanDeploy = true
			}
		}

		mb.BoostedSpeed = mb.Speed
		mb.BoostedMaxHitpoints = mb.MaxHitpoints
		mb.BoostedShieldRechargeRate = mb.ShieldRechargeRate

		if mb.BoostStat.Valid {
			boostPercent := decimal.NewFromInt(mb.MechSkinLevel).Div(decimal.NewFromInt(100)).Add(decimal.NewFromInt(1))
			// mech boosted stat
			switch mb.BoostStat.String {
			case boiler.BoostStatMECH_SPEED:
				mb.BoostedSpeed = decimal.NewFromInt(mb.Speed).Mul(boostPercent).IntPart()
			case boiler.BoostStatMECH_HEALTH:
				mb.BoostedMaxHitpoints = decimal.NewFromInt(mb.MaxHitpoints).Mul(boostPercent).IntPart()
			case boiler.BoostStatSHIELD_REGEN:
				mb.BoostedShieldRechargeRate = decimal.NewFromInt(mb.ShieldRechargeRate).Mul(boostPercent).IntPart()
			}
		}

		resp = append(resp, mb)
	}

	return resp, nil
}
