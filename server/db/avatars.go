package db

import (
	"database/sql"
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/db/boiler"
	"server/gamedb"
)

func GiveDefaultAvatars(playerID string, factionID string) error {
	fac, err := boiler.Factions(boiler.FactionWhere.ID.EQ(factionID)).One(gamedb.StdConn)
	if err != nil {
		return err
	}

	// get faction logo urls from profile avatars table
	ava, err := boiler.ProfileAvatars(boiler.ProfileAvatarWhere.AvatarURL.EQ(fac.LogoURL)).One(gamedb.StdConn)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		return err
	}

	// insert into player profile avatars
	ppa := &boiler.PlayersProfileAvatar{
		PlayerID:        playerID,
		ProfileAvatarID: ava.ID,
	}

	err = ppa.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return err
	}

	return nil
}

type AvatarListOpts struct {
	Search   string
	Filter   *ListFilterRequest
	Sort     *ListSortRequest
	PageSize int
	Page     int
	OwnerID  string
}

func AvatarList(opts *AvatarListOpts) (int64, []*boiler.ProfileAvatar, error) {
	var avatars []*boiler.ProfileAvatar

	queryMods := []qm.QueryMod{
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.PlayersProfileAvatars,
			qm.Rels(boiler.TableNames.PlayersProfileAvatars, boiler.PlayersProfileAvatarColumns.ProfileAvatarID),
			qm.Rels(boiler.TableNames.ProfileAvatars, boiler.ProfileAvatarColumns.ID),
		)),
		qm.Where("players_profile_avatars.player_id = ?", opts.OwnerID),
	}

	total, err := boiler.ProfileAvatars(
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
			qm.Rels(boiler.TableNames.ProfileAvatars, boiler.ProfileAvatarColumns.ID),
			qm.Rels(boiler.TableNames.ProfileAvatars, boiler.ProfileAvatarColumns.AvatarURL),
			qm.Rels(boiler.TableNames.ProfileAvatars, boiler.ProfileAvatarColumns.Tier),
		),
		qm.From(boiler.TableNames.ProfileAvatars),
	)

	rows, err := boiler.NewQuery(
		queryMods...,
	).Query(gamedb.StdConn)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		av := &boiler.ProfileAvatar{}

		scanArgs := []interface{}{
			&av.ID,
			&av.AvatarURL,
			&av.Tier,
		}

		err = rows.Scan(scanArgs...)
		if err != nil {
			return total, avatars, err
		}
		avatars = append(avatars, av)
	}

	return total, avatars, nil
}
