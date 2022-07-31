package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/helpers"
	"strconv"

	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type BattleHistoryRecord struct {
	Number    int     `json:"number"`
	StartedAt int64   `json:"started_at"`
	EndedAt   *int64  `json:"ended_at"`
	Winner    *string `json:"winner"`
	RunnerUp  *string `json:"runner_up"`
	Loser     *string `json:"loser"`
}

// BattleHistoryController holds handlers for battle history requests
type BattleHistoryController struct {
}

func BattleHistoryRouter() chi.Router {
	c := &BattleHistoryController{}
	r := chi.NewRouter()
	r.Get("/", WithError(c.BattleHistoryCurrent))
	r.Get("/{battle_number}", WithError(c.BattleHistory))

	return r
}

// GET /api/battle_history
// {
//     "current_battle": {
//         "number": 12345,
//         "started_at": 1659280332,
//         "ended_at": null,
//         "winner": null,
//         "runner_up": null,
//         "loser": null
//     },
//     "previous_battle":{
//         "number": 12344,
//         "started_at": 1659279802,
//         "ended_at": 1659280302,
//         "winner": "ZHI",
//         "runner_up": "BC",
//         "loser": "RM"
//     }
// }
type BattleHistoryCurrent struct {
	CurrentBattle  *BattleHistoryRecord `json:"current_battle"`
	PreviousBattle *BattleHistoryRecord `json:"previous_battle"`
}

// BattleHistoryCurrent gets current battle and previous battle records
func (c *BattleHistoryController) BattleHistoryCurrent(w http.ResponseWriter, r *http.Request) (int, error) {
	battles, err := boiler.Battles(qm.OrderBy("started_at DESC"), qm.Limit(2)).All(gamedb.StdConn)
	if err != nil {
		return http.StatusBadRequest, errors.Wrap(err, "get battles")
	}
	if len(battles) != 2 {
		return http.StatusBadRequest, fmt.Errorf("expected 2 battles, got %d", len(battles))
	}

	curr := battles[0]
	prev := battles[1]

	currentBattleRecord := &BattleHistoryRecord{
		Number:    curr.BattleNumber,
		StartedAt: curr.StartedAt.Unix(),
		EndedAt:   nil,
		Winner:    nil,
		RunnerUp:  nil,
		Loser:     nil,
	}

	previousBattleRecord, err := BattleRecord(prev)
	if err != nil {
		return http.StatusBadRequest, errors.Wrap(err, "get battle record")
	}
	result := BattleHistoryCurrent{
		CurrentBattle:  currentBattleRecord,
		PreviousBattle: previousBattleRecord,
	}
	return helpers.EncodeJSON(w, result)
}

type BattleHistoryRequest struct {
	Battle *BattleHistoryRecord `json:"battle"`
}

// BattleHistory gets a single battle record
// GET /api/battle_history/{battle_number}
// {
//     "battle": {
//         "number": 12344,
//         "started_at": 1659279802,
//         "ended_at": 1659280302,
//         "winner": "ZHI",
//         "runner_up": "BC",
//         "loser": "RM"
//     }
// }
func (c *BattleHistoryController) BattleHistory(w http.ResponseWriter, r *http.Request) (int, error) {
	battleNumberStr := chi.URLParam(r, "battle_number")
	battleNumber, err := strconv.Atoi(battleNumberStr)
	if err != nil {
		return http.StatusBadRequest, err
	}
	battle, err := boiler.Battles(boiler.BattleWhere.BattleNumber.EQ(battleNumber)).One(gamedb.StdConn)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return http.StatusBadRequest, errors.Wrapf(err, "battle not found: %d", battleNumber)
	}
	if err != nil {
		return http.StatusBadRequest, errors.Wrapf(err, "get battle for battle: %d", battleNumber)
	}
	record, err := BattleRecord(battle)
	if err != nil {
		return http.StatusBadRequest, errors.Wrapf(err, "get battle record for battle: %d", battleNumber)
	}
	return helpers.EncodeJSON(w, &BattleHistoryRequest{record})
}

// BattleRecord processes the battle DB item and converts to to a battle history record
func BattleRecord(b *boiler.Battle) (*BattleHistoryRecord, error) {
	var endUnix *int64
	if b.EndedAt.Valid {
		endUnixNonPtr := b.EndedAt.Time.Unix()
		endUnix = &endUnixNonPtr
	}

	mechs, err := boiler.BattleMechs(
		boiler.BattleMechWhere.BattleID.EQ(b.ID),
		qm.OrderBy("killed DESC"), // Last mech to die with faction_won false is in the faction that got runner up
	).All(gamedb.StdConn)
	if err != nil {
		return nil, errors.Wrapf(err, "get battle mechs for battle: %d", b.BattleNumber)
	}

	ZaibatsuShortcode := "ZHI"
	RedMountainShortcode := "RMOMC"
	BostonShortcode := "BC"

	var winner *string
	var runnerUp *string
	var loser *string
	for _, mech := range mechs {
		// Mechs in here are connected to the winning faction only
		if mech.FactionWon.Bool {
			switch mech.FactionID {
			case server.ZaibatsuFactionID:
				*winner = ZaibatsuShortcode
			case server.RedMountainFactionID:
				*winner = RedMountainShortcode
			case server.BostonCyberneticsFactionID:
				*winner = BostonShortcode
			default:
				return nil, fmt.Errorf("faction not recognised: %s", mech.FactionID)
			}
			continue
		}

		// Only mechs processed here are not part of the winning faction
		// Because of the ORDER BY clause, the first mech (who is killed last) should be part of runner-up faction

		switch mech.FactionID {
		case server.ZaibatsuFactionID:
			*runnerUp = ZaibatsuShortcode
		case server.RedMountainFactionID:
			*runnerUp = RedMountainShortcode
		case server.BostonCyberneticsFactionID:
			*runnerUp = BostonShortcode
		default:
			return nil, fmt.Errorf("faction not recognised: %s", mech.FactionID)
		}

		// Remaining faction is the loser
		// TODO: Fix my sloppy conditionals
		if winner != &ZaibatsuShortcode && runnerUp != &ZaibatsuShortcode {
			loser = &ZaibatsuShortcode
		}

		if winner != &RedMountainShortcode && runnerUp != &RedMountainShortcode {
			loser = &RedMountainShortcode
		}

		if winner != &BostonShortcode && runnerUp != &BostonShortcode {
			loser = &BostonShortcode
		}

		// Got enough information, break
		break
	}

	result := &BattleHistoryRecord{
		Number:    b.BattleNumber,
		StartedAt: b.StartedAt.Unix(),
		EndedAt:   endUnix,
		Winner:    winner,
		RunnerUp:  runnerUp,
		Loser:     loser,
	}
	return result, nil
}
