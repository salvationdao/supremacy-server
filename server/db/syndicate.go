package db

import (
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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
		queries = append(queries, boiler.SyndicateMotionWhere.ActualEndedAt.IsNotNull())
	} else {
		queries = append(queries, boiler.SyndicateMotionWhere.ActualEndedAt.IsNull())
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
