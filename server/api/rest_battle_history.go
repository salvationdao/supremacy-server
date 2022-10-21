package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"strconv"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-syndicate/supremacy-bridge/bridge"
	cache "github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type FactionShortcode string

var FactionMap = map[FactionShortcode]int{
	NoneShortcode:        0,
	ZaibatsuShortcode:    1,
	RedMountainShortcode: 2,
	BostonShortcode:      3,
}

const NoneShortcode FactionShortcode = "NONE"
const ZaibatsuShortcode FactionShortcode = "ZHI"
const RedMountainShortcode FactionShortcode = "RMOMC"
const BostonShortcode FactionShortcode = "BC"

type CurrentBattle struct {
	Number    int    `json:"number"`
	StartedAt int64  `json:"started_at"`
	ExpiresAt int64  `json:"expires_at"`
	Signature string `json:"signature"`
}

type MechDetails struct {
	Faction    int64                 `json:"faction"`
	MechOwners []*PlayerShortDetails `json:"mech_owners"`
}

type OnlineCitizens struct {
	Faction       int64                 `json:"faction"`
	PlayerDetails []*PlayerShortDetails `json:"player_details"`
}

type PlayerShortDetails struct {
	Username string `json:"username"`
	GID      int    `json:"gid"`
}

type BattleHistoryRecord struct {
	Number         int               `json:"number"`
	StartedAt      int64             `json:"started_at"`
	EndedAt        *int64            `json:"ended_at"`
	Winner         int64             `json:"winner"`
	RunnerUp       int64             `json:"runner_up"`
	Loser          int64             `json:"loser"`
	Signature      string            `json:"signature"`
	MechDetails    []*MechDetails    `json:"mech_details,omitempty"`
	OnlineCitizens []*OnlineCitizens `json:"online_citizens,omitempty"`
}

// BattleHistoryController holds handlers for battle history requests
type BattleHistoryController struct {
	signerPrivateKeyHex string
}

func BattleHistoryRouter(signerPrivateKeyHex string) chi.Router {
	quick, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(10000000),
	)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("quick adaptor: failed to create")
		os.Exit(1)
	}
	long, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(10000000),
	)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("long adaptor: failed to create")
		os.Exit(1)
	}
	quickCache, err := cache.NewClient(
		cache.ClientWithAdapter(quick),
		cache.ClientWithTTL(5*time.Second),
		cache.ClientWithRefreshKey("opn"),
	)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("quick cache: failed to initialise")
		os.Exit(1)
	}
	longCache, err := cache.NewClient(
		cache.ClientWithAdapter(long),
		cache.ClientWithTTL(10*time.Minute),
		cache.ClientWithRefreshKey("opn"),
	)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("long cache: failed to initialise")
		os.Exit(1)
	}
	c := &BattleHistoryController{signerPrivateKeyHex}
	r := chi.NewRouter()
	r.Get("/", quickCache.Middleware(http.HandlerFunc(WithError(c.BattleHistoryCurrent))).ServeHTTP)
	r.Get("/{battle_number}", longCache.Middleware(http.HandlerFunc(WithError(c.BattleHistory))).ServeHTTP)

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
//     "previous_battles":[
//			{
//				"number": 12344,
//				"started_at": 1659279802,
//				"ended_at": 1659280302,
//				"winner": "ZHI",
//				"runner_up": "BC",
//				"loser": "RM"
//    		}
// 			...
//		]
// }
type BattleHistoryCurrent struct {
	CurrentBattle   *CurrentBattle         `json:"current_battle"`
	PreviousBattles []*BattleHistoryRecord `json:"previous_battles"`
}

// BattleHistoryCurrent gets current battle and previous battle records (100 records)
func (c *BattleHistoryController) BattleHistoryCurrent(w http.ResponseWriter, r *http.Request) (int, error) {
	battles, err := boiler.Battles(qm.OrderBy("started_at DESC"), qm.Limit(11)).All(gamedb.StdConn)
	if err != nil {
		return http.StatusBadRequest, errors.Wrap(err, "get battles")
	}

	// Head of battle array
	curr := battles[0]
	expiry := time.Now().Add(30 * time.Second).Unix()
	signer := bridge.NewSigner(c.signerPrivateKeyHex)
	_, sig, err := signer.GenerateCurrentBattleSignature(
		int64(curr.BattleNumber),
		curr.StartedAt.Unix(),
		expiry,
	)
	if err != nil {
		return 0, fmt.Errorf("generate signature: %w", err)
	}

	currentBattleRecord := &CurrentBattle{
		Number:    curr.BattleNumber,
		StartedAt: curr.StartedAt.Unix(),
		ExpiresAt: expiry,
		Signature: hexutil.Encode(sig),
	}

	previousBattleRecords := []*BattleHistoryRecord{}

	// Tail of battle array
	for _, battle := range battles[1:] {
		previousBattleRecord, err := BattleRecord(battle, c.signerPrivateKeyHex)
		if err != nil {
			return http.StatusBadRequest, errors.Wrap(err, "get battle record")
		}
		previousBattleRecords = append(previousBattleRecords, previousBattleRecord)
	}

	result := BattleHistoryCurrent{
		CurrentBattle:   currentBattleRecord,
		PreviousBattles: previousBattleRecords,
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
	record, err := BattleRecord(battle, c.signerPrivateKeyHex)
	if err != nil {
		return http.StatusBadRequest, errors.Wrapf(err, "get battle record for battle: %d", battleNumber)
	}
	return helpers.EncodeJSON(w, &BattleHistoryRequest{record})
}

// BattleRecord processes the battle DB item and converts to to a battle history record
func BattleRecord(b *boiler.Battle, signerPrivateKeyHex string) (*BattleHistoryRecord, error) {
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

	var winner FactionShortcode = NoneShortcode
	var runnerUp FactionShortcode = NoneShortcode
	var loser FactionShortcode = NoneShortcode

	for _, mech := range mechs {
		// Mechs in here are connected to the winning faction only
		if mech.FactionWon.Bool {
			switch mech.FactionID {
			case server.ZaibatsuFactionID:
				winner = ZaibatsuShortcode
			case server.RedMountainFactionID:
				winner = RedMountainShortcode
			case server.BostonCyberneticsFactionID:
				winner = BostonShortcode
			default:
				return nil, fmt.Errorf("faction not recognised: %s", mech.FactionID)
			}
			continue
		}

		// Only mechs processed here are not part of the winning faction
		// Because of the ORDER BY clause, the first mech (who is killed last) should be part of runner-up faction

		switch mech.FactionID {
		case server.ZaibatsuFactionID:
			runnerUp = ZaibatsuShortcode
		case server.RedMountainFactionID:
			runnerUp = RedMountainShortcode
		case server.BostonCyberneticsFactionID:
			runnerUp = BostonShortcode
		default:
			return nil, fmt.Errorf("faction not recognised: %s", mech.FactionID)
		}

		// Remaining faction is the loser
		// TODO: Fix my sloppy conditionals
		if winner != ZaibatsuShortcode && runnerUp != ZaibatsuShortcode {
			loser = ZaibatsuShortcode
		}

		if winner != RedMountainShortcode && runnerUp != RedMountainShortcode {
			loser = RedMountainShortcode
		}

		if winner != BostonShortcode && runnerUp != BostonShortcode {
			loser = BostonShortcode
		}

		// Got enough information, break
		break
	}

	result := &BattleHistoryRecord{
		Number:    b.BattleNumber,
		StartedAt: b.StartedAt.Unix(),
		EndedAt:   endUnix,
		Winner:    int64(FactionMap[winner]),
		RunnerUp:  int64(FactionMap[runnerUp]),
		Loser:     int64(FactionMap[loser]),
	}

	func() {
		zaiMechDetails, err := getMechMechOwnerDetails(ZaibatsuShortcode, b.ID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("Faction Shortcode", string(ZaibatsuShortcode)).Str("Battle ID", b.ID).Msg("Failed to get online mech owner details")
			return
		}
		rmMechDetails, err := getMechMechOwnerDetails(RedMountainShortcode, b.ID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("Faction Shortcode", string(RedMountainShortcode)).Str("Battle ID", b.ID).Msg("Failed to get online mech owner details")
			return
		}
		bcMechDetails, err := getMechMechOwnerDetails(BostonShortcode, b.ID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("Faction Shortcode", string(BostonShortcode)).Str("Battle ID", b.ID).Msg("Failed to get online mech owner details")
			return
		}

		mechDetails := []*MechDetails{}

		mechDetails = append(mechDetails, zaiMechDetails, rmMechDetails, bcMechDetails)

		result.MechDetails = mechDetails
	}()

	func() {
		zaiOnlineCitizens, err := getOnlineCitizensDetails(ZaibatsuShortcode, b.ID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("Faction Shortcode", string(ZaibatsuShortcode)).Str("Battle ID", b.ID).Msg("Failed to get online citizen details")
			return
		}
		rmOnlineCitizens, err := getOnlineCitizensDetails(RedMountainShortcode, b.ID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("Faction Shortcode", string(RedMountainShortcode)).Str("Battle ID", b.ID).Msg("Failed to get online citizen details")
			return
		}
		bcOnlineCitizens, err := getOnlineCitizensDetails(BostonShortcode, b.ID)
		if err != nil {
			gamelog.L.Error().Err(err).Str("Faction Shortcode", string(BostonShortcode)).Str("Battle ID", b.ID).Msg("Failed to get online citizen details")
			return
		}

		onlineCitizens := []*OnlineCitizens{}

		onlineCitizens = append(onlineCitizens, zaiOnlineCitizens, rmOnlineCitizens, bcOnlineCitizens)

		result.OnlineCitizens = onlineCitizens
	}()

	if winner != NoneShortcode && runnerUp != NoneShortcode && loser != NoneShortcode {
		signer := bridge.NewSigner(signerPrivateKeyHex)
		_, sig, err := signer.GenerateBattleRecordSignature(
			int64(b.BattleNumber),
			b.StartedAt.Unix(),
			b.EndedAt.Time.Unix(),
			int64(FactionMap[winner]),
			int64(FactionMap[runnerUp]),
			int64(FactionMap[loser]),
		)
		if err != nil {
			return nil, fmt.Errorf("generate signature: %w", err)
		}
		result.Signature = hexutil.Encode(sig)
	}

	return result, nil
}

func getMechMechOwnerDetails(factionShortcode FactionShortcode, battleID string) (*MechDetails, error) {
	factionID := ""
	switch FactionMap[factionShortcode] {
	case 1:
		factionID = server.ZaibatsuFactionID
	case 2:
		factionID = server.RedMountainFactionID
	case 3:
		factionID = server.BostonCyberneticsFactionID
	}

	if factionID == "" {
		return nil, terror.Error(fmt.Errorf("failed to get faction id from faction shortcode"), "Failed to get faction id from faction shortcode")
	}

	mechDetails := &MechDetails{
		Faction: int64(FactionMap[factionShortcode]),
	}

	battleMechs, err := boiler.BattleMechs(
		boiler.BattleMechWhere.BattleID.EQ(battleID),
		boiler.BattleMechWhere.FactionID.EQ(factionID),
		qm.Load(boiler.BattleMechRels.PilotedBy),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err, "Failed to find battle mech details")
	}

	mechOwnersShortDetails := []*PlayerShortDetails{}

	for _, mech := range battleMechs {
		if mech.R != nil && mech.R.PilotedBy != nil {
			mechOwnersShortDetail := &PlayerShortDetails{
				GID: mech.R.PilotedBy.Gid,
			}

			if mech.R.PilotedBy.Username.Valid {
				mechOwnersShortDetail.Username = mech.R.PilotedBy.Username.String
			}

			mechOwnersShortDetails = append(mechOwnersShortDetails, mechOwnersShortDetail)
		}
	}

	mechDetails.MechOwners = mechOwnersShortDetails
	return mechDetails, nil
}

func getOnlineCitizensDetails(factionShortcode FactionShortcode, battleID string) (*OnlineCitizens, error) {
	factionID := ""
	switch FactionMap[factionShortcode] {
	case 1:
		factionID = server.ZaibatsuFactionID
	case 2:
		factionID = server.RedMountainFactionID
	case 3:
		factionID = server.BostonCyberneticsFactionID
	}

	if factionID == "" {
		return nil, terror.Error(fmt.Errorf("failed to get faction id from faction shortcode"), "Failed to get faction id from faction shortcode")
	}

	onlineCitizen := &OnlineCitizens{
		Faction: int64(FactionMap[factionShortcode]),
	}

	battleViewers, err := boiler.Players(
		boiler.PlayerWhere.FactionID.EQ(null.StringFrom(factionID)),
		qm.Where(fmt.Sprintf(
			"EXISTS (SELECT 1 FROM %s WHERE player_id = %s AND battle_id = ?)",
			boiler.TableNames.BattleViewers,
			boiler.PlayerTableColumns.ID,
		),
			battleID,
		),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err, "failed to get battle viewers for citizen details")
	}

	playerShortDetails := []*PlayerShortDetails{}
	for _, viewer := range battleViewers {
		playerShortDetail := &PlayerShortDetails{
			GID: viewer.Gid,
		}

		if viewer.Username.Valid {
			playerShortDetail.Username = viewer.Username.String
		}

		playerShortDetails = append(playerShortDetails, playerShortDetail)
	}

	onlineCitizen.PlayerDetails = playerShortDetails
	return onlineCitizen, nil
}
