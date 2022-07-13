package syndicate

import (
	"fmt"
	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"strings"
)

func (ss *System) SyndicateNameVerification(inputName string) (string, error) {
	syndicateName := strings.TrimSpace(inputName)

	if len(syndicateName) == 0 {
		return "", terror.Error(fmt.Errorf("empty syndicate name"), "The name of syndicate is empty")
	}

	// check profanity
	if ss.profanityManager.Detector.IsProfane(syndicateName) {
		return "", terror.Error(fmt.Errorf("profanity detected"), "The syndicate name contains profanity")
	}

	if len(syndicateName) > 64 {
		return "", terror.Error(fmt.Errorf("too many characters"), "The syndicate name should not be longer than 64 characters")
	}

	// name similarity check
	syndicates, err := boiler.Syndicates(
		qm.Select(boiler.SyndicateColumns.ID, boiler.SyndicateColumns.Name),
		qm.Where(
			fmt.Sprintf("SIMILARITY(%s,$1) > 0.4", qm.Rels(boiler.TableNames.Syndicates, boiler.SyndicateColumns.Name)),
			syndicateName,
		),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate name", syndicateName).Msg("Failed to get syndicate by name from db")
		return "", terror.Error(err, "Failed to verify syndicate name")
	}

	// https://github.com/adrg/strutil
	for _, s := range syndicates {
		if strutil.Similarity(syndicateName, s.Name, metrics.NewHamming()) >= 0.9 {
			return "", terror.Error(fmt.Errorf("similar syndicate name"), fmt.Sprintf("'%s' has already been taken by other syndicate", syndicateName))
		}
	}

	return syndicateName, nil
}

func (ss *System) SyndicateSymbolVerification(inputSymbol string) (string, error) {
	// get rid of all the spaces
	symbol := strings.ToUpper(strings.ReplaceAll(inputSymbol, " ", ""))

	if len(symbol) < 4 || len(symbol) > 5 {
		return "", terror.Error(fmt.Errorf("must be 4 or 5 character"), "The length of symbol must be at least four and no more than five excluding spaces.")
	}

	// check profanity
	if ss.profanityManager.Detector.IsProfane(symbol) {
		return "", terror.Error(fmt.Errorf("profanity detected"), "The syndicate symbol contains profanity")
	}

	// name similarity check
	syndicates, err := boiler.Syndicates(
		qm.Select(boiler.SyndicateColumns.ID, boiler.SyndicateColumns.Symbol),
		qm.Where(
			fmt.Sprintf("SIMILARITY(%s,$1) > 0.4", qm.Rels(boiler.TableNames.Syndicates, boiler.SyndicateColumns.Symbol)),
			symbol,
		),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("syndicate name", symbol).Msg("Failed to get syndicate by name from db")
		return "", terror.Error(err, "Failed to verify syndicate name")
	}

	// https://github.com/adrg/strutil
	for _, s := range syndicates {
		if strutil.Similarity(symbol, s.Symbol, metrics.NewHamming()) >= 0.9 {
			return "", terror.Error(fmt.Errorf("similar syndicate name"), fmt.Sprintf("'%s' has already been taken by other syndicate", symbol))
		}
	}

	return symbol, nil
}

func (s *Syndicate) getTotalAvailableMotionVoter() (int64, error) {
	var total int64
	var err error

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
