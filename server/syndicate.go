package server

import "server/db/boiler"

type Syndicate struct {
	*boiler.Syndicate
	SymbolUrl string           `json:"symbol_url"`
	Members   []*boiler.Player `json:"members"`
}

func SyndicateBoilerToServer(syndicate *boiler.Syndicate) *Syndicate {
	s := &Syndicate{
		Syndicate: syndicate,
	}

	if syndicate.R != nil {
		if syndicate.R.Symbol != nil {
			s.SymbolUrl = syndicate.R.Symbol.ImageURL
		}

		if syndicate.R.Players != nil {
			for _, p := range syndicate.R.Players {
				s.Members = append(s.Members, p)
			}
		}
	}

	return s
}
