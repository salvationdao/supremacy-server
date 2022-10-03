package server

import (
	"database/sql"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
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
}

type BattleLobbiesMech struct {
	MechID        string `json:"mech_id" db:"mech_id"`
	BattleLobbyID string `json:"battle_lobby_id" db:"battle_lobby_id"`
	AvatarURL     string `json:"avatar_url" db:"avatar_url"`
	Name          string `json:"name" db:"name"`
	Label         string `json:"label" db:"label"`
	Tier          string `json:"tier" db:"tier"`

	IsDestroyed bool           `json:"is_destroyed"`
	Owner       *boiler.Player `json:"owner"`
}

type BattleLobbySupporter struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	FactionID      string `json:"faction_id"`
	AvatarID       string `json:"avatar_id,omitempty"`
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
			BattleLobby:               &copiedBattleLobby,
			IsPrivate:                 copiedBattleLobby.Password.Valid,
			BattleLobbiesMechs:        []*BattleLobbiesMech{},
			OptedInRedMountSupporters: []*BattleLobbySupporter{},
			OptedInZaiSupporters:      []*BattleLobbySupporter{},
			OptedInBostonSupporters:   []*BattleLobbySupporter{},
		}

		// omit password
		sbl.Password = null.StringFromPtr(nil)

		if bl.R != nil {
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
		}

		if bl.R != nil && bl.R.BattleLobbySupporters != nil && len(bl.R.BattleLobbySupporters) > 0 {
			for _, sup := range bl.R.BattleLobbySupporters {
				switch sup.FactionID {
				case RedMountainFactionID:
					sbl.OptedInRedMountSupporters = append(sbl.OptedInRedMountSupporters, &BattleLobbySupporter{
						ID:             sup.ID,
						Username:       sup.R.Supporter.Username.String,
						FactionID:      sup.R.Supporter.FactionID.String,
						AvatarID:       sup.R.Supporter.ProfileAvatarID.String,
						CustomAvatarID: sup.R.Supporter.CustomAvatarID.String,
					})
				case BostonCyberneticsFactionID:
					sbl.OptedInBostonSupporters = append(sbl.OptedInBostonSupporters, &BattleLobbySupporter{
						ID:             sup.ID,
						Username:       sup.R.Supporter.Username.String,
						FactionID:      sup.R.Supporter.FactionID.String,
						AvatarID:       sup.R.Supporter.ProfileAvatarID.String,
						CustomAvatarID: sup.R.Supporter.CustomAvatarID.String,
					})
				case ZaibatsuFactionID:
					sbl.OptedInZaiSupporters = append(sbl.OptedInZaiSupporters, &BattleLobbySupporter{
						ID:             sup.ID,
						Username:       sup.R.Supporter.Username.String,
						FactionID:      sup.R.Supporter.FactionID.String,
						AvatarID:       sup.R.Supporter.ProfileAvatarID.String,
						CustomAvatarID: sup.R.Supporter.CustomAvatarID.String,
					})
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
			boiler.BlueprintMechTableColumns.Label,
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
		),

		// from filtered collection item
		qm.From(fmt.Sprintf(
			`(
				SELECT %s, %s, %s FROM %s 
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

	for rows.Next() {
		blm := &BattleLobbiesMech{
			Owner: &boiler.Player{},
		}
		err = rows.Scan(
			&blm.MechID,
			&blm.BattleLobbyID,
			&blm.Name,
			&blm.Label,
			&blm.Tier,
			&blm.AvatarURL,
			&blm.Owner.ID,
			&blm.Owner.Username,
			&blm.Owner.FactionID,
			&blm.Owner.Gid,
			&blm.Owner.Rank,
		)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to scan battle lobby.")
			return nil, terror.Error(err, "Failed to scan battle lobby mech")
		}

		// append mech to battle lobby
		for _, bl := range resp {
			if bl.ID != blm.BattleLobbyID {
				continue
			}

			// set is destroyed flag
			blm.IsDestroyed = slices.Index(destroyedMechIDs, blm.MechID) != -1

			bl.BattleLobbiesMechs = append(bl.BattleLobbiesMechs, blm)
			break
		}
	}

	return resp, nil
}
