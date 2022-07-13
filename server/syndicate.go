package server

import "server/db/boiler"

type Syndicate struct {
	*boiler.Syndicate
	Founder *boiler.Player `json:"founder"`
	CEO     *boiler.Player `json:"ceo"`
	Admin   *boiler.Player `json:"admin"`
}

func SyndicateBoilerToServer(syndicate *boiler.Syndicate) *Syndicate {
	s := &Syndicate{
		Syndicate: syndicate,
	}

	if syndicate.R != nil {
		if s.R.FoundedBy != nil {
			s.Founder = &boiler.Player{ID: s.R.FoundedBy.ID, Username: s.R.FoundedBy.Username, FactionID: s.R.FoundedBy.FactionID, Gid: s.R.FoundedBy.Gid, Rank: s.R.FoundedBy.Rank}
		}
		if s.R.CeoPlayer != nil {
			s.Founder = &boiler.Player{ID: s.R.CeoPlayer.ID, Username: s.R.CeoPlayer.Username, FactionID: s.R.CeoPlayer.FactionID, Gid: s.R.CeoPlayer.Gid, Rank: s.R.CeoPlayer.Rank}
		}
		if s.R.Admin != nil {
			s.Founder = &boiler.Player{ID: s.R.Admin.ID, Username: s.R.Admin.Username, FactionID: s.R.Admin.FactionID, Gid: s.R.Admin.Gid, Rank: s.R.Admin.Rank}
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
