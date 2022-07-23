package db

import (
	"database/sql"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
)

type SyndicateMotionListFilter struct {
	Type    null.String `json:"type"`
	Result  null.String `json:"result"`
	IsEnded bool        `json:"is_ended"`
}

func SyndicateMotionList(syndicateID string, filter *SyndicateMotionListFilter, limit, offset int) ([]*boiler.SyndicateMotion, int64, error) {
	queries := []qm.QueryMod{
		boiler.SyndicateMotionWhere.SyndicateID.EQ(syndicateID),
		boiler.SyndicateMotionWhere.Result.EQ(filter.Result),
	}

	if filter.Type.Valid {
		queries = append(queries, boiler.SyndicateMotionWhere.Type.EQ(filter.Type.String))
	}

	if filter.IsEnded {
		queries = append(queries, boiler.SyndicateMotionWhere.FinalisedAt.IsNotNull())
	} else {
		queries = append(queries, boiler.SyndicateMotionWhere.FinalisedAt.IsNull())
	}

	count, err := boiler.SyndicateMotions(queries...).Count(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("queries", queries).Msg("Failed to get the total count of syndicate motions from db")
		return nil, 0, terror.Error(err, "Failed to get syndicate motion list")
	}

	queries = append(queries, qm.Limit(limit), qm.Offset(offset))

	sms, err := boiler.SyndicateMotions(queries...).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("queries", queries).Msg("Failed to get syndicate motion list from db")
		return nil, 0, terror.Error(err, "Failed to get syndicate motion list")
	}

	return sms, count, nil
}

func GetSyndicateDetail(syndicateID string) (*server.Syndicate, error) {
	syndicate, err := boiler.Syndicates(
		boiler.SyndicateWhere.ID.EQ(syndicateID),
		qm.Load(boiler.SyndicateRels.Players, qm.Select(boiler.PlayerColumns.ID, boiler.PlayerColumns.Username, boiler.PlayerColumns.Gid)),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msg("Failed to query syndicate from db")
		return nil, terror.Error(err, "Failed to get syndicate")
	}

	if syndicate == nil {
		return nil, terror.Error(fmt.Errorf("syndicate not found"), "Syndicate does not exist")
	}

	return server.SyndicateBoilerToServer(syndicate), nil
}

func GetSyndicateDirectors(syndicateID string) ([]*server.Player, error) {
	ps, err := boiler.Players(
		qm.Select(
			boiler.PlayerColumns.ID,
			boiler.PlayerColumns.FactionID,
			boiler.PlayerColumns.Username,
			boiler.PlayerColumns.Gid,
			boiler.PlayerColumns.Rank,
		),
		qm.InnerJoin(
			fmt.Sprintf(
				"%s on %s = %s",
				boiler.TableNames.SyndicateDirectors,
				qm.Rels(boiler.TableNames.SyndicateDirectors, boiler.SyndicateDirectorColumns.PlayerID),
				qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.ID),
			),
		),
		qm.Load(boiler.PlayerRels.IDPlayerStat),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate id", syndicateID).Msg("Failed to get syndicate directors from db")
		return nil, terror.Error(err, "Failed to get syndicate directors.")
	}

	players := []*server.Player{}
	for _, p := range ps {
		players = append(players, server.PlayerFromBoiler(p).Brief())
	}

	return players, nil
}

func GetSyndicateCommittees(syndicateID string) ([]*server.Player, error) {
	ps, err := boiler.Players(
		qm.Select(
			boiler.PlayerColumns.ID,
			boiler.PlayerColumns.FactionID,
			boiler.PlayerColumns.Username,
			boiler.PlayerColumns.Gid,
			boiler.PlayerColumns.Rank,
		),
		qm.InnerJoin(
			fmt.Sprintf(
				"%s on %s = %s",
				boiler.TableNames.SyndicateCommittees,
				qm.Rels(boiler.TableNames.SyndicateCommittees, boiler.SyndicateCommitteeColumns.PlayerID),
				qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.ID),
			),
		),
		qm.Load(boiler.PlayerRels.IDPlayerStat),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate id", syndicateID).Msg("Failed to get syndicate directors from db")
		return nil, terror.Error(err, "Failed to get syndicate directors.")
	}

	players := []*server.Player{}
	for _, p := range ps {
		players = append(players, server.PlayerFromBoiler(p).Brief())
	}

	return players, nil
}

func IsSyndicateDirector(syndicateID string, userID string) (bool, error) {
	// check availability
	exist, err := boiler.SyndicateDirectors(
		boiler.SyndicateDirectorWhere.SyndicateID.EQ(syndicateID),
		boiler.SyndicateDirectorWhere.PlayerID.EQ(userID),
	).Exists(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("syndicate id", syndicateID).Str("player id", userID).Msg("Failed to check syndicate director list from db.")
		return false, terror.Error(err, "Failed to check whether user is a director of the syndicate")
	}

	return exist, nil
}

// GetSyndicateTotalAvailableMotionVoters return total of available motion voter base on syndicate type
func GetSyndicateTotalAvailableMotionVoters(syndicateID string) (int64, error) {
	var total int64
	var err error

	s, err := boiler.FindSyndicate(gamedb.StdConn, syndicateID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate id", s.ID).Msg("Failed to get syndicate detail from db")
		return 0, terror.Error(err, "Failed to load syndicate detail.")
	}

	switch s.Type {
	case boiler.SyndicateTypeCORPORATION:
		total, err = s.SyndicateDirectors().Count(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Str("syndicate id", s.ID).Msg("Failed to get syndicate director number")
			return 0, terror.Error(err, "Failed to get syndicate directors number")
		}

		return total, nil
	case boiler.SyndicateTypeDECENTRALISED:
		total, err = s.Players().Count(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Str("syndicate id", s.ID).Msg("Failed to get syndicate members number")
			return 0, terror.Error(err, "Failed to get syndicate members number")
		}

		return total, nil

	default:
		gamelog.L.Error().Err(err).Str("syndicate id", s.ID).Str("syndicate type", s.Type).Msg("Failed to get total available motion voters")
		return 0, terror.Error(fmt.Errorf("invalid syndicate type"), "Invalid syndicate type")
	}
}
