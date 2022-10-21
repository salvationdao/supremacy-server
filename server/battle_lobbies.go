package server

import (
	"database/sql"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
)

type BattleLobbyStageOrder int

const (
	BattleLobbyStageOrderPending BattleLobbyStageOrder = 3
	BattleLobbyStageOrderReady   BattleLobbyStageOrder = 2
	BattleLobbyStageOrderBattle  BattleLobbyStageOrder = 1
	BattleLobbyStageOrderEnd     BattleLobbyStageOrder = 0
)

type BattleLobby struct {
	*boiler.BattleLobby
	HostBy                     *boiler.Player          `json:"host_by"`
	GameMap                    *boiler.GameMap         `json:"game_map"`
	BattleLobbiesMechs         []*BattleLobbiesMech    `json:"battle_lobbies_mechs"`
	OptedInRedMountSupporters  []*BattleLobbySupporter `json:"opted_in_rm_supporters"`
	OptedInZaiSupporters       []*BattleLobbySupporter `json:"opted_in_zai_supporters"`
	OptedInBostonSupporters    []*BattleLobbySupporter `json:"opted_in_bc_supporters"`
	SelectedRedMountSupporters []*BattleLobbySupporter `json:"selected_rm_supporters"`
	SelectedZaiSupporters      []*BattleLobbySupporter `json:"selected_zai_supporters"`
	SelectedBostonSupporters   []*BattleLobbySupporter `json:"selected_bc_supporters"`
	IsPrivate                  bool                    `json:"is_private"`
	StageOrder                 BattleLobbyStageOrder   `json:"stage_order"`
	SupsPool                   decimal.Decimal         `json:"sups_pool"`
}

type BattleLobbiesMech struct {
	*BlueprintMech

	MechID        string `json:"mech_id" db:"mech_id"`
	BattleLobbyID string `json:"battle_lobby_id" db:"battle_lobby_id"`
	AvatarURL     string `json:"avatar_url" db:"avatar_url"`
	Name          string `json:"name" db:"name"`
	Tier          string `json:"tier" db:"tier"`

	IsDestroyed bool           `json:"is_destroyed"`
	Owner       *boiler.Player `json:"owner"`
	FactionID   null.String    `json:"faction_id"`

	// player might queue other players' mechs from the mech staking list
	QueuedBy *PublicPlayer `json:"queued_by"`

	WeaponSlots []*WeaponSlot `json:"weapon_slots"`
}

type WeaponSlot struct {
	MechID          string      `json:"mech_id"`
	WeaponID        null.String `json:"weapon_id"`
	SlotNumber      int         `json:"slot_number"`
	AllowMelee      bool        `json:"allow_melee"`
	IsSkinInherited bool        `json:"is_skin_inherited"`
	Weapon          *Weapon     `json:"weapon"`
}

type BattleLobbySupporter struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	FactionID      string `json:"faction_id"`
	AvatarURL      string `json:"avatar_url,omitempty"`
	CustomAvatarID string `json:"custom_avatar_id,omitempty"`
}

func BattleLobbiesFromBoiler(bls []*boiler.BattleLobby) ([]*BattleLobby, error) {
	resp := []*BattleLobby{}

	if bls == nil || len(bls) == 0 {
		return resp, nil
	}

	battleIDs := []string{}
	for _, bl := range bls {
		copiedBattleLobby := *bl
		sbl := &BattleLobby{
			BattleLobby:                &copiedBattleLobby,
			IsPrivate:                  copiedBattleLobby.AccessCode.Valid,
			BattleLobbiesMechs:         []*BattleLobbiesMech{},
			OptedInRedMountSupporters:  []*BattleLobbySupporter{},
			OptedInZaiSupporters:       []*BattleLobbySupporter{},
			OptedInBostonSupporters:    []*BattleLobbySupporter{},
			SelectedRedMountSupporters: []*BattleLobbySupporter{},
			SelectedZaiSupporters:      []*BattleLobbySupporter{},
			SelectedBostonSupporters:   []*BattleLobbySupporter{},
			SupsPool:                   decimal.Zero,
		}

		if sbl.EndedAt.Valid {
			sbl.StageOrder = BattleLobbyStageOrderEnd
		} else if sbl.AssignedToArenaID.Valid {
			sbl.StageOrder = BattleLobbyStageOrderBattle
		} else if sbl.ReadyAt.Valid {
			sbl.StageOrder = BattleLobbyStageOrderReady
		} else {
			sbl.StageOrder = BattleLobbyStageOrderPending
		}

		if bl.R != nil {
			for _, reward := range bl.R.BattleLobbyExtraSupsRewards {
				sbl.SupsPool = sbl.SupsPool.Add(reward.Amount)
			}

			if bl.R.HostBy != nil {
				host := bl.R.HostBy
				// trim info
				sbl.HostBy = &boiler.Player{
					ID:        host.ID,
					Username:  host.Username,
					FactionID: host.FactionID,
					Gid:       host.Gid,
					Rank:      host.Rank,
				}
			}

			if bl.R.GameMap != nil {
				sbl.GameMap = bl.R.GameMap
			}

			// opted in peeps
			if bl.R.BattleLobbySupporterOptIns != nil && len(bl.R.BattleLobbySupporterOptIns) > 0 {
				for _, sup := range bl.R.BattleLobbySupporterOptIns {
					switch sup.FactionID {
					case RedMountainFactionID:
						supper := &BattleLobbySupporter{
							ID:             sup.SupporterID,
							Username:       sup.R.Supporter.Username.String,
							FactionID:      sup.R.Supporter.FactionID.String,
							CustomAvatarID: sup.R.Supporter.CustomAvatarID.String,
						}
						if sup.R.Supporter.R != nil && sup.R.Supporter.R.ProfileAvatar != nil {
							supper.AvatarURL = sup.R.Supporter.R.ProfileAvatar.AvatarURL
						}

						sbl.OptedInRedMountSupporters = append(sbl.OptedInRedMountSupporters, supper)
					case BostonCyberneticsFactionID:
						supper := &BattleLobbySupporter{
							ID:             sup.SupporterID,
							Username:       sup.R.Supporter.Username.String,
							FactionID:      sup.R.Supporter.FactionID.String,
							CustomAvatarID: sup.R.Supporter.CustomAvatarID.String,
						}
						if sup.R.Supporter.R != nil && sup.R.Supporter.R.ProfileAvatar != nil {
							supper.AvatarURL = sup.R.Supporter.R.ProfileAvatar.AvatarURL
						}

						sbl.OptedInBostonSupporters = append(sbl.OptedInBostonSupporters, supper)
					case ZaibatsuFactionID:
						supper := &BattleLobbySupporter{
							ID:             sup.SupporterID,
							Username:       sup.R.Supporter.Username.String,
							FactionID:      sup.R.Supporter.FactionID.String,
							CustomAvatarID: sup.R.Supporter.CustomAvatarID.String,
						}
						if sup.R.Supporter.R != nil && sup.R.Supporter.R.ProfileAvatar != nil {
							supper.AvatarURL = sup.R.Supporter.R.ProfileAvatar.AvatarURL
						}

						sbl.OptedInZaiSupporters = append(sbl.OptedInZaiSupporters, supper)
					}
				}
			}

			// selected peeps
			if bl.R.BattleLobbySupporters != nil && len(bl.R.BattleLobbySupporters) > 0 {
				for _, sup := range bl.R.BattleLobbySupporters {
					switch sup.FactionID {
					case RedMountainFactionID:
						supper := &BattleLobbySupporter{
							ID:             sup.SupporterID,
							Username:       sup.R.Supporter.Username.String,
							FactionID:      sup.R.Supporter.FactionID.String,
							CustomAvatarID: sup.R.Supporter.CustomAvatarID.String,
						}
						if sup.R.Supporter.R != nil && sup.R.Supporter.R.ProfileAvatar != nil {
							supper.AvatarURL = sup.R.Supporter.R.ProfileAvatar.AvatarURL
						}

						sbl.SelectedRedMountSupporters = append(sbl.SelectedRedMountSupporters, supper)
					case BostonCyberneticsFactionID:
						supper := &BattleLobbySupporter{
							ID:             sup.SupporterID,
							Username:       sup.R.Supporter.Username.String,
							FactionID:      sup.R.Supporter.FactionID.String,
							CustomAvatarID: sup.R.Supporter.CustomAvatarID.String,
						}
						if sup.R.Supporter.R != nil && sup.R.Supporter.R.ProfileAvatar != nil {
							supper.AvatarURL = sup.R.Supporter.R.ProfileAvatar.AvatarURL
						}

						sbl.SelectedBostonSupporters = append(sbl.SelectedBostonSupporters, supper)
					case ZaibatsuFactionID:
						supper := &BattleLobbySupporter{
							ID:             sup.SupporterID,
							Username:       sup.R.Supporter.Username.String,
							FactionID:      sup.R.Supporter.FactionID.String,
							CustomAvatarID: sup.R.Supporter.CustomAvatarID.String,
						}
						if sup.R.Supporter.R != nil && sup.R.Supporter.R.ProfileAvatar != nil {
							supper.AvatarURL = sup.R.Supporter.R.ProfileAvatar.AvatarURL
						}

						sbl.SelectedZaiSupporters = append(sbl.SelectedZaiSupporters, supper)
					}
				}
			}
		}

		resp = append(resp, sbl)

		if bl.AssignedToBattleID.Valid {
			battleIDs = append(battleIDs, bl.AssignedToBattleID.String)
		}
	}

	destroyedMechIDs := []string{}
	if len(battleIDs) > 0 {
		bhs, err := boiler.BattleHistories(
			boiler.BattleHistoryWhere.BattleID.IN(battleIDs),
			boiler.BattleHistoryWhere.EventType.EQ(boiler.BattleEventKilled),
		).All(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Strs("battle id list", battleIDs).Msg("Failed to load killed battle histories.")
		}

		for _, bh := range bhs {
			destroyedMechIDs = append(destroyedMechIDs, bh.WarMachineOneID)
		}
	}

	// get all the related mechs
	battleLobbyIDInClause := " IN ("
	for i, bl := range bls {
		battleLobbyIDInClause += "'" + bl.ID + "'"

		if i < len(bls)-1 {
			battleLobbyIDInClause += ","
			continue
		}

		battleLobbyIDInClause += ")"
	}

	// get all the mech brief details
	queries := []qm.QueryMod{
		qm.Select(
			// mech info
			fmt.Sprintf("_ci.%s AS mech_id", boiler.CollectionItemColumns.ItemID),
			fmt.Sprintf("_ci.%s", boiler.BattleLobbiesMechColumns.BattleLobbyID),
			boiler.MechTableColumns.Name,
			fmt.Sprintf("TO_JSON( %s.* ) AS bm", boiler.TableNames.BlueprintMechs),
			boiler.MechSkinTableColumns.Level,
			boiler.BlueprintMechSkinTableColumns.Tier,
			fmt.Sprintf(
				"(SELECT %s FROM %s WHERE %s = %s AND %s = %s) AS %s",
				boiler.MechModelSkinCompatibilityTableColumns.AvatarURL,
				boiler.TableNames.MechModelSkinCompatibilities,
				boiler.MechModelSkinCompatibilityTableColumns.MechModelID,
				boiler.MechTableColumns.BlueprintID,
				boiler.MechModelSkinCompatibilityTableColumns.BlueprintMechSkinID,
				boiler.MechSkinTableColumns.BlueprintID,
				boiler.BlueprintMechSkinColumns.AvatarURL,
			),

			// owner info
			fmt.Sprintf("_ci.%s", boiler.CollectionItemColumns.OwnerID),
			boiler.PlayerTableColumns.Username,
			boiler.PlayerTableColumns.FactionID,
			boiler.PlayerTableColumns.Gid,
			boiler.PlayerTableColumns.Rank,

			// queue by player
			"TO_JSON(_queued_by_player.*) AS queued_by_player",
		),

		// from filtered collection item
		qm.From(fmt.Sprintf(
			`(
				SELECT %s, %s, %s, %s FROM %s 
				INNER JOIN %s ON %s = %s 
							AND %s %s
							AND %s ISNULL 
							AND %s ISNULL 
							AND %s ISNULL 
			) _ci`,
			// SELECT
			boiler.CollectionItemTableColumns.ItemID,
			boiler.CollectionItemTableColumns.OwnerID,
			boiler.BattleLobbiesMechTableColumns.BattleLobbyID,
			boiler.BattleLobbiesMechTableColumns.QueuedByID,

			// FROM
			boiler.TableNames.CollectionItems,

			// INNER JOIN
			boiler.TableNames.BattleLobbiesMechs,
			boiler.CollectionItemTableColumns.ItemID,
			boiler.BattleLobbiesMechTableColumns.MechID,
			boiler.BattleLobbiesMechTableColumns.BattleLobbyID,
			battleLobbyIDInClause,
			boiler.BattleLobbiesMechTableColumns.EndedAt,
			boiler.BattleLobbiesMechTableColumns.RefundTXID,
			boiler.BattleLobbiesMechTableColumns.DeletedAt,
		)),

		// inner join player info
		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = _ci.%s",
			boiler.TableNames.Players,
			boiler.PlayerTableColumns.ID,
			boiler.CollectionItemColumns.OwnerID,
		)),

		// inner join queue by player
		qm.InnerJoin(fmt.Sprintf(
			"%s _queued_by_player ON _queued_by_player.%s = _ci.%s",
			boiler.TableNames.Players,
			boiler.PlayerColumns.ID,
			boiler.BattleLobbiesMechColumns.QueuedByID,
		)),

		// inner join mech
		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = _ci.%s",
			boiler.TableNames.Mechs,
			boiler.MechTableColumns.ID,
			boiler.CollectionItemColumns.ItemID,
		)),
		// inner join blueprint mech
		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.BlueprintMechs,
			boiler.MechTableColumns.BlueprintID,
			boiler.BlueprintMechTableColumns.ID,
		)),
		// inner join mech skin
		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.MechSkin,
			boiler.MechTableColumns.ChassisSkinID,
			boiler.MechSkinTableColumns.ID,
		)),
		// inner join blueprint mech skin
		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.BlueprintMechSkin,
			boiler.MechSkinTableColumns.BlueprintID,
			boiler.BlueprintMechSkinTableColumns.ID,
		)),
	}

	rows, err := boiler.NewQuery(queries...).Query(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Interface("queries", queries).Msg("Failed to load battle lobbies")
		return nil, terror.Error(err, "Failed to load battle mechs")
	}

	impactedMechIDs := []string{}
	blms := []*BattleLobbiesMech{}
	for rows.Next() {
		blm := &BattleLobbiesMech{
			Owner: &boiler.Player{},
		}
		queuedByPlayer := &PublicPlayer{}

		bm := &BlueprintMech{}
		skinLevel := int64(0)
		err = rows.Scan(
			&blm.MechID,
			&blm.BattleLobbyID,
			&blm.Name,
			&bm,
			&skinLevel,
			&blm.Tier,
			&blm.AvatarURL,
			&blm.Owner.ID,
			&blm.Owner.Username,
			&blm.Owner.FactionID,
			&blm.Owner.Gid,
			&blm.Owner.Rank,
			&queuedByPlayer,
		)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to scan battle lobby.")
			return nil, terror.Error(err, "Failed to scan battle lobby mech")
		}

		blm.QueuedBy = queuedByPlayer
		bm.BoostedSpeed = int64(bm.Speed)
		bm.BoostedMaxHitpoints = int64(bm.MaxHitpoints)
		bm.BoostedShieldRechargeRate = int64(bm.ShieldRechargeRate)

		// mech boosted stat
		if bm.BoostStat.Valid {
			boostPercent := decimal.NewFromInt(skinLevel).Div(decimal.NewFromInt(100)).Add(decimal.NewFromInt(1))

			switch bm.BoostStat.String {
			case boiler.BoostStatMECH_SPEED:
				bm.BoostedSpeed = decimal.NewFromInt(bm.BoostedSpeed).Mul(boostPercent).IntPart()
			case boiler.BoostStatMECH_HEALTH:
				bm.BoostedMaxHitpoints = decimal.NewFromInt(bm.BoostedMaxHitpoints).Mul(boostPercent).IntPart()
			case boiler.BoostStatSHIELD_REGEN:
				bm.BoostedShieldRechargeRate = decimal.NewFromInt(bm.BoostedShieldRechargeRate).Mul(boostPercent).IntPart()
			}
		}

		blm.BlueprintMech = bm

		// set is destroyed flag
		blm.IsDestroyed = slices.Index(destroyedMechIDs, blm.MechID) != -1
		blm.FactionID = blm.Owner.FactionID

		blms = append(blms, blm)

		// record mech id for weapon query
		if slices.Index(impactedMechIDs, blm.MechID) == -1 {
			impactedMechIDs = append(impactedMechIDs, blm.MechID)
		}
	}

	// fill up mech weapon slots
	if len(impactedMechIDs) > 0 {
		impactedMechWhereInClause := fmt.Sprintf("WHERE %s IN (", boiler.MechWeaponTableColumns.ChassisID)
		for i, mechID := range impactedMechIDs {
			impactedMechWhereInClause += "'" + mechID + "'"

			if i < len(impactedMechIDs)-1 {
				impactedMechWhereInClause += ","
				continue
			}

			impactedMechWhereInClause += ")"
		}

		queries = []qm.QueryMod{
			qm.Select(
				boiler.MechWeaponTableColumns.ChassisID,
				boiler.MechWeaponTableColumns.WeaponID,
				boiler.MechWeaponTableColumns.SlotNumber,
				boiler.MechWeaponTableColumns.AllowMelee,
				boiler.MechWeaponTableColumns.IsSkinInherited,
				fmt.Sprintf("TO_JSON(%s)", boiler.TableNames.Weapons),
			),
			qm.From(fmt.Sprintf(
				"(SELECT %s, %s, %s, %s, %s FROM %s %s) %s",
				boiler.MechWeaponTableColumns.ChassisID,
				boiler.MechWeaponTableColumns.WeaponID,
				boiler.MechWeaponTableColumns.SlotNumber,
				boiler.MechWeaponTableColumns.AllowMelee,
				boiler.MechWeaponTableColumns.IsSkinInherited,
				boiler.TableNames.MechWeapons,
				impactedMechWhereInClause,
				boiler.TableNames.MechWeapons,
			)),

			qm.LeftOuterJoin(fmt.Sprintf(
				`(SELECT 
							%[1]s AS weapon_id, 
							%[2]s.*, 
							%[3]s.*,
							%[13]s.*
					FROM %[4]s
					INNER JOIN %[5]s ON %[6]s = %[7]s
					INNER JOIN %[8]s ON %[1]s = %[9]s
					INNER JOIN %[3]s ON %[10]s = %[11]s
					INNER JOIN %[13]s ON %[14]s = %[6]s AND %[15]s = %[11]s
				) %[4]s ON %[4]s.weapon_id = %[12]s`,
				boiler.WeaponTableColumns.ID,                                          // 1
				boiler.TableNames.BlueprintWeapons,                                    // 2
				boiler.TableNames.BlueprintWeaponSkin,                                 // 3
				boiler.TableNames.Weapons,                                             // 4
				boiler.TableNames.BlueprintWeapons,                                    // 5
				boiler.WeaponTableColumns.BlueprintID,                                 // 6
				boiler.BlueprintWeaponTableColumns.ID,                                 // 7
				boiler.TableNames.WeaponSkin,                                          // 8
				boiler.WeaponSkinTableColumns.EquippedOn,                              // 9
				boiler.BlueprintWeaponSkinTableColumns.ID,                             // 10
				boiler.WeaponSkinTableColumns.BlueprintID,                             // 11
				boiler.MechWeaponTableColumns.WeaponID,                                // 12
				boiler.TableNames.WeaponModelSkinCompatibilities,                      // 13
				boiler.WeaponModelSkinCompatibilityTableColumns.WeaponModelID,         // 14
				boiler.WeaponModelSkinCompatibilityTableColumns.BlueprintWeaponSkinID, // 15
			)),

			qm.OrderBy(boiler.MechWeaponTableColumns.SlotNumber),
		}

		rows, err = boiler.NewQuery(queries...).Query(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to load mech weapon slots")
			return nil, terror.Error(err, "Failed to load mech weapon slots.")
		}

		for rows.Next() {
			weaponSlot := &WeaponSlot{}
			var weapon *Weapon
			err = rows.Scan(&weaponSlot.MechID, &weaponSlot.WeaponID, &weaponSlot.SlotNumber, &weaponSlot.AllowMelee, &weaponSlot.IsSkinInherited, &weapon)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to scan mech weapon slots.")
				return nil, terror.Error(err, "Failed to scan mech weapon slots.")
			}

			weaponSlot.Weapon = weapon

			for _, blm := range blms {
				if weaponSlot.MechID != blm.MechID {
					continue
				}

				blm.WeaponSlots = append(blm.WeaponSlots, weaponSlot)
			}
		}
	}

	// fill mech in to lobby
	for _, bl := range resp {
		for _, blm := range blms {
			if bl.ID != blm.BattleLobbyID {
				continue
			}

			bl.BattleLobbiesMechs = append(bl.BattleLobbiesMechs, blm)

			// accumulate sups pool
			bl.SupsPool = bl.SupsPool.Add(bl.EntryFee)
		}
	}

	return resp, nil
}

// BattleLobbiesFactionFilter omit the mech owner and weapon slots of other faction mechs
func BattleLobbiesFactionFilter(bls []*BattleLobby, keepDataForFactionID string, toUserID string) []*BattleLobby {
	// generate a new struct
	battleLobbies := []*BattleLobby{}

	for _, bl := range bls {
		battleLobbies = append(battleLobbies, BattleLobbyInfoFilter(bl, keepDataForFactionID, bl.HostByID == toUserID))
	}

	return battleLobbies
}

// BattleLobbyInfoFilter filter single lobby at a time
func BattleLobbyInfoFilter(bl *BattleLobby, keepDataForFactionID string, keepAccessCode bool) *BattleLobby {
	// copy lobby data,
	// important: access code must be omitted
	battleLobby := &BattleLobby{
		BattleLobby:        bl.BattleLobby,
		HostBy:             bl.HostBy,
		GameMap:            bl.GameMap,
		BattleLobbiesMechs: []*BattleLobbiesMech{},
		IsPrivate:          bl.IsPrivate,
		StageOrder:         bl.StageOrder,
		SupsPool:           bl.SupsPool,
	}

	if !keepAccessCode {
		battleLobby.AccessCode = null.StringFromPtr(nil)
	}

	for _, blm := range bl.BattleLobbiesMechs {
		battleLobbyMech := &BattleLobbiesMech{
			MechID:        blm.MechID,
			BattleLobbyID: blm.BattleLobbyID,
			AvatarURL:     blm.AvatarURL,
			Name:          blm.Name,
			Tier:          blm.Tier,
			IsDestroyed:   blm.IsDestroyed,
			FactionID:     blm.Owner.FactionID,
		}

		if blm.Owner != nil && blm.Owner.FactionID.String == keepDataForFactionID {
			// copy owner detail
			battleLobbyMech.Owner = blm.Owner

			// copy weapon slots
			battleLobbyMech.WeaponSlots = blm.WeaponSlots

			battleLobbyMech.BlueprintMech = blm.BlueprintMech
		}

		if blm.QueuedBy != nil && blm.QueuedBy.FactionID.String == keepDataForFactionID {
			battleLobbyMech.QueuedBy = blm.QueuedBy
		}

		battleLobby.BattleLobbiesMechs = append(battleLobby.BattleLobbiesMechs, battleLobbyMech)
	}

	// supporters
	switch keepDataForFactionID {
	case RedMountainFactionID:
		battleLobby.OptedInRedMountSupporters = bl.OptedInRedMountSupporters
		battleLobby.SelectedRedMountSupporters = bl.SelectedRedMountSupporters
	case BostonCyberneticsFactionID:
		battleLobby.OptedInBostonSupporters = bl.OptedInBostonSupporters
		battleLobby.SelectedBostonSupporters = bl.SelectedBostonSupporters
	case ZaibatsuFactionID:
		battleLobby.OptedInZaiSupporters = bl.OptedInZaiSupporters
		battleLobby.SelectedZaiSupporters = bl.SelectedZaiSupporters
	}
	return battleLobby
}
