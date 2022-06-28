package profanities

import (
	"server/db/boiler"
	"server/gamedb"

	goaway "github.com/TwiN/go-away"
	"github.com/davecgh/go-spew/spew"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type ProfanityManager struct {
	Detector *goaway.ProfanityDetector
}

func NewProfanityManager() (*ProfanityManager, error) {
	profanities, err := boiler.Profanities().All(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err, "failed to fetch profanity phrases from db")
	}

	dictionary := []string{}
	for _, p := range profanities {
		dictionary = append(dictionary, p.Phrase)
	}

	spew.Dump(dictionary)

	return &ProfanityManager{
		Detector: goaway.NewProfanityDetector().WithCustomDictionary(dictionary, []string{}, []string{}),
	}, nil
}

func (pm ProfanityManager) ReloadDictionary() error {
	profanities, err := boiler.Profanities().All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "failed to fetch profanity phrases from db")
	}

	dictionary := []string{}
	for _, p := range profanities {
		dictionary = append(dictionary, p.Phrase)
	}

	pm.Detector = goaway.NewProfanityDetector().WithCustomDictionary(dictionary, []string{}, []string{})
	return nil
}

func (pm ProfanityManager) AddToDictionary(phrase string) error {
	p := &boiler.Profanity{
		Phrase: phrase,
	}

	err := p.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "failed to update profanity dictionary")
	}

	err = pm.ReloadDictionary()
	if err != nil {
		return terror.Error(err, "failed to reload profanity dictionary")
	}
	return nil
}
