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
	"strconv"
)

type BattleLobby struct {
	*boiler.BattleLobby
	HostBy             *boiler.Player       `json:"host_by"`
	GameMap            *boiler.GameMap      `json:"game_map"`
	BattleLobbiesMechs []*BattleLobbiesMech `json:"battle_lobbies_mechs"`
	IsPrivate          bool                 `json:"is_private"`
}

type BattleLobbiesMech struct {
	MechID        string `json:"mech_id" db:"mech_id"`
	BattleLobbyID string `json:"battle_lobby_id" db:"battle_lobby_id"`
	Name          string `json:"name" db:"name"`
	Label         string `json:"label" db:"label"`
	Tier          string `json:"tier" db:"tier"`

	IsDestroyed bool           `json:"is_destroyed"`
	Owner       *boiler.Player `json:"owner"`
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
			BattleLobby: &copiedBattleLobby,
			IsPrivate:   copiedBattleLobby.Password.Valid,
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
	var args []interface{}
	battleLobbyIDInClause := " IN ("
	for i, bl := range bls {
		args = append(args, bl.ID)
		battleLobbyIDInClause += "$" + strconv.Itoa(len(args))

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
			boiler.BlueprintMechSkinColumns.ID,
		)),
	}

	rows, err := boiler.NewQuery(queries...).Query(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
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
			&blm.Owner.ID,
			&blm.Owner.Username,
			&blm.Owner.FactionID,
			&blm.Owner.Gid,
			&blm.Owner.Rank,
		)
		if err != nil {
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

type BattleBounty struct {
	*boiler.BattleBounty
	IsClosed        bool           `json:"is_closed"`
	OfferedByPlayer *boiler.Player `json:"offered_by_player"`
}

func BattleBountiesFromBoiler(bbs []*boiler.BattleBounty) []*BattleBounty {
	resp := []*BattleBounty{}
	if bbs == nil || len(bbs) == 0 {
		return resp
	}

	for _, bb := range bbs {
		battleBounty := &BattleBounty{
			BattleBounty: bb,
			IsClosed:     bb.PayoutTXID.Valid && bb.RefundTXID.Valid,
		}

		// clean up transaction id
		battleBounty.PayoutTXID = null.StringFromPtr(nil)
		battleBounty.RefundTXID = null.StringFromPtr(nil)

		if bb.R != nil && bb.R.OfferedBy != nil {
			player := bb.R.OfferedBy
			battleBounty.OfferedByPlayer = &boiler.Player{
				ID:        player.ID,
				Username:  player.Username,
				FactionID: player.FactionID,
				Gid:       player.Gid,
				Rank:      player.Rank,
			}
		}

		resp = append(resp, battleBounty)
	}

	return resp
}
