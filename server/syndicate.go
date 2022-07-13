package server

import "server/db/boiler"

type Syndicate struct {
	*boiler.Syndicate
	SymbolUrl string           `json:"symbol_url"`
	Members   []*boiler.Player `json:"members"`
	Directors []*boiler.Player `json:"directors"`
}

func SyndicateBoilerToServer(syndicate *boiler.Syndicate) *Syndicate {
	s := &Syndicate{
		Syndicate: syndicate,
	}

	if syndicate.R != nil {
		if syndicate.R.Symbol != nil {
			s.SymbolUrl = syndicate.R.Symbol.ImageURL
		}

		for _, p := range syndicate.R.Players {
			s.Members = append(s.Members, &boiler.Player{ID: p.ID, Username: p.Username, Gid: p.Gid, Rank: p.Rank})
		}

		for _, dp := range syndicate.R.DirectorOfSyndicatePlayers {
			s.Directors = append(s.Directors, &boiler.Player{ID: dp.ID, Username: dp.Username, Gid: dp.Gid, Rank: dp.Rank})
		}
	}

	return s
}

type SyndicateMotion struct {
	*boiler.SyndicateMotion
	IssuedByUser *boiler.Player `json:"issued_by_user"`
	Votes        []*boiler.SyndicateMotionVote
}

func SyndicateMotionBoilerToServer(motion *boiler.SyndicateMotion) *SyndicateMotion {
	sm := &SyndicateMotion{
		SyndicateMotion: motion,
	}

	if motion.R != nil {
		if motion.R.IssuedBy != nil {
			sm.IssuedByUser = &boiler.Player{ID: motion.R.IssuedBy.ID, Username: motion.R.IssuedBy.Username, Gid: motion.R.IssuedBy.Gid, Rank: motion.R.IssuedBy.Rank}
		}

		for _, vote := range sm.R.MotionSyndicateMotionVotes {
			sm.Votes = append(sm.Votes, vote)
		}
	}

	return sm
}
