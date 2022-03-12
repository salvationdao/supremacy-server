package db

import (
	"database/sql"
	"errors"
	"fmt"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"golang.org/x/net/context"
)

type PlayerWithFaction struct {
	boiler.Player
	Faction boiler.Faction `json:"faction"`
}

type BattleMechData struct {
	MechID    uuid.UUID
	OwnerID   uuid.UUID
	FactionID uuid.UUID
}

func BattleMechs(btl *boiler.Battle, mechData []*BattleMechData) error {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Str("db func", "Battle").Err(err).Msg("unable to begin tx")
		return err
	}
	defer tx.Rollback()

	for _, md := range mechData {
		bmd := boiler.BattleMech{
			BattleID:  btl.ID,
			MechID:    md.MechID.String(),
			OwnerID:   md.OwnerID.String(),
			FactionID: md.FactionID.String(),
		}
		err = bmd.Insert(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().Interface("battle mech", bmd).Str("db func", "Battle").Err(err).Msg("unable to insert Battle Mech into database")
			return err
		}
	}

	return tx.Commit()
}

func UpdateBattleMech(battleID string, mechID uuid.UUID, gotKill bool, gotKilled bool, killedByID ...uuid.UUID) (*boiler.BattleMech, error) {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Str("db func", "UpdateBattleMech").Err(err).Msg("unable to begin tx")
		return nil, err
	}
	defer tx.Rollback()

	bmd, err := boiler.FindBattleMech(tx, battleID, mechID.String())
	if err != nil {
		gamelog.L.Error().
			Str("battleID", battleID).
			Str("mechID", mechID.String()).
			Str("db func", "UpdateBattleMech").
			Err(err).Msg("unable to retrieve Battle Mech from database")

		return nil, err
	}

	if gotKilled {
		bmd.Killed = null.TimeFrom(time.Now())
		if len(killedByID) > 0 && !killedByID[0].IsNil() {
			if len(killedByID) > 1 {
				warn := gamelog.L.Warn()
				for i, id := range killedByID {
					warn = warn.Str(fmt.Sprintf("killedByID[%d]", i), id.String())
				}
				warn.Str("db func", "UpdateBattleMech").Msg("more than 1 killer mech provided, only the zero indexed mech will be saved")
			}
			bmd.KilledByID = null.StringFrom(killedByID[0].String())
			kid, err := uuid.FromString(killedByID[0].String())

			killerBmd, err := boiler.FindBattleMech(tx, battleID, kid.String())
			if err != nil {
				gamelog.L.Error().
					Str("battleID", battleID).
					Str("killerBmdID", killedByID[0].String()).
					Str("db func", "UpdateBattleMech").
					Err(err).Msg("unable to retrieve Battle Mech from database")

				return nil, err
			}

			killerBmd.Kills++
			bk := &boiler.BattleKill{
				MechID:    killedByID[0].String(),
				BattleID:  battleID,
				CreatedAt: bmd.Killed.Time,
				KilledID:  mechID.String(),
			}
			err = bk.Insert(tx, boil.Infer())
		}
		_, err = bmd.Update(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).
				Interface("boiler.BattleMech", bmd).
				Msg("unable to update battle mech")
			return nil, err
		}

		return bmd, nil
	}

	if gotKill {
		bmd.Kills = bmd.Kills + 1
		_, err = bmd.Update(tx, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).
				Interface("boiler.BattleMech", bmd).
				Msg("unable to update battle mech")
			return nil, err
		}
		return bmd, nil
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Str("db Func", "UpdateBattleMech").Msg("unable to commit tx")
		return nil, err
	}

	return bmd, nil
}

type EventType byte

const (
	Btlevnt_Killed EventType = iota
	Btlevnt_Kill
	Btlevnt_Spawnedai
	Btlevnt_Ability_Triggered
)

func (ev EventType) String() string {
	return [...]string{"killed", "kill", "spawned_ai", "ability_triggered"}[ev]
}

type BattleEvent struct {
	ID        uuid.UUID
	BattleID  uuid.UUID
	RelatedID string
	WM1       uuid.UUID
	WM2       uuid.UUID
	EventType EventType
	CreatedAt time.Time
}

func StoreBattleEvent(battleID string, relatedID uuid.UUID, wm1 uuid.UUID, wm2 uuid.UUID, eventType EventType, createdAt time.Time) (*boiler.BattleHistory, error) {
	if battleID == "" || wm1.IsNil() {
		return nil, errors.New("no battle ID provided")
	}
	bh := &boiler.BattleHistory{BattleID: battleID, WarMachineOneID: wm1.String(), EventType: eventType.String()}
	if !relatedID.IsNil() {
		bh.RelatedID = null.StringFrom(relatedID.String())
	}
	if !wm2.IsNil() {
		bh.WarMachineTwoID = null.StringFrom(wm2.String())
	}
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Str("db func", "StoreBattleEvent").Err(err).Msg("unable to begin tx")
		return nil, err
	}
	defer tx.Rollback()
	err = bh.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("boiler.BattleHistory", bh).Msg("unable to insert battle history event")
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Interface("boiler.BattleHistory", bh).Msg("unable to commit tx")
		return nil, err
	}

	return bh, err
}

type MechWithOwner struct {
	OwnerID   uuid.UUID
	MechID    uuid.UUID
	FactionID uuid.UUID
}

func WinBattle(battleID string, winCondition string, mechs ...*MechWithOwner) error {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Str("db func", "WinBattle").Err(err).Msg("unable to begin tx")
		return err
	}
	defer tx.Rollback()
	for _, m := range mechs {
		mw := &boiler.BattleWin{
			BattleID:     battleID,
			WinCondition: winCondition,
			MechID:       m.MechID.String(),
			OwnerID:      m.OwnerID.String(),
			FactionID:    m.FactionID.String(),
		}
		err = mw.Insert(tx, boil.Infer())
	}
	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Str("db func", "WinBattle").Err(err).Msg("unable to commit tx")
		return err
	}
	return err
}

//DefaultFactionPlayers return default mech players
func DefaultFactionPlayers() (map[string]PlayerWithFaction, error) {
	players, err := boiler.Players(qm.Where("is_ai = true")).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	factionids := make([]interface{}, len(players))

	for i, player := range players {
		factionids[i] = player.FactionID.String
	}

	factions, err := boiler.Factions(qm.WhereIn("id IN ?", factionids...)).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	result := make(map[string]PlayerWithFaction, len(players))
	for _, player := range players {
		var faction *boiler.Faction
		for _, f := range factions {
			if f.ID == player.FactionID.String {
				faction = f
			}
		}
		if faction == nil {
			gamelog.L.Warn().Str("player id", player.ID).Str("faction id", player.FactionID.String).Msg("AI player has no faction")
			continue
		}
		result[player.ID] = PlayerWithFaction{*player, *faction}
	}

	return result, err
}

func LoadBattleQueue(ctx context.Context, lengthPerFaction int) ([]*boiler.BattleQueue, error) {
	query := `SELECT
mech_id, queued_at, faction_id, owner_id, battle_id
FROM (
SELECT
ROW_NUMBER() OVER (PARTITION BY faction_id ORDER BY queued_at ASC) AS r,
t.*
FROM
battle_queue t) x
WHERE
x.r <= $1`

	result, err := gamedb.Conn.Query(ctx, query, lengthPerFaction)
	if err != nil {
		gamelog.L.Error().Int("length", lengthPerFaction).Err(err).Msg("unable to retrieve mechs for load out")
		return nil, err
	}
	defer result.Close()

	queue := []*boiler.BattleQueue{}

	for result.Next() {
		mc := &boiler.BattleQueue{}
		err = result.Scan(&mc.MechID, &mc.QueuedAt, &mc.FactionID, &mc.OwnerID, &mc.BattleID)
		if err != nil {
			return nil, err
		}
		queue = append(queue, mc)
	}

	return queue, nil
}

// MechBattleStatus returns true if the mech is currently in battle, and false if not
func MechBattleStatus(mechID uuid.UUID) (bool, error) {
	var count int64

	query := `
	select count(*) from battles b
	inner join battle_mechs bm on bm.battle_id = b.id
	where b.id = (select id from battles order by battle_number desc limit 1) and bm.mech_id = $1
	`

	err := gamedb.Conn.QueryRow(context.Background(), query, mechID.String()).Scan(&count)
	if err != nil {
		gamelog.L.Error().
			Str("mech_id", mechID.String()).
			Str("db func", "MechBattleStatus").Err(err).Msg("unable to get queue position of mech")
		return false, err
	}

	return count > 0, nil
}

type MechAndPosition struct {
	MechID        uuid.UUID
	QueuePosition int64
}

// AllMechsAfter gets all mechs that come after the specified position in the queue
// It returns a list of mech IDs
func AllMechsAfter(queuedAt time.Time, factionID uuid.UUID) ([]*MechAndPosition, error) {
	query := `
		WITH bqpos AS (
			SELECT t.*,
				   ROW_NUMBER() OVER(ORDER BY t.queued_at) AS position
			FROM battle_queue t WHERE faction_id = $1 AND queued_at > $2)
			SELECT s.mech_id, s.position
			FROM bqpos s
		`

	rows, err := gamedb.StdConn.Query(query, factionID.String(), queuedAt)
	if err != nil {
		gamelog.L.Error().
			Time("queued_at", queuedAt).
			Str("faction_id", factionID.String()).
			Str("db func", "AllMechsAfter").Err(err).Msg("unable to get mechs after")
		return nil, err
	}
	defer rows.Close()

	mechsAfter := make([]*MechAndPosition, 0)
	for rows.Next() {
		item := &MechAndPosition{}
		err := rows.Scan(&item.MechID, &item.QueuePosition)
		if err != nil {
			gamelog.L.Error().
				Time("queued_at", queuedAt).
				Str("faction_id", factionID.String()).
				Str("db func", "AllMechsAfter").Err(err).Msg("unable to get mechs after")
			return nil, err
		}
		mechsAfter = append(mechsAfter, item)
	}
	rows.Close()

	return mechsAfter, nil
}

func QueueLength(factionID uuid.UUID) (int64, error) {
	var count int64

	err := gamedb.Conn.QueryRow(context.Background(), `SELECT COUNT(mech_id) FROM battle_queue WHERE faction_id = $1`, factionID.String()).Scan(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}

func QueuePosition(mechID uuid.UUID, factionID uuid.UUID) (int64, error) {
	var pos int64

	exists, _ := boiler.BattleQueueExists(gamedb.StdConn, mechID.String())
	if !exists {
		return -1, nil
	}

	query := `WITH bqpos AS (
    SELECT t.*,
           ROW_NUMBER() OVER(ORDER BY t.queued_at) AS position
    FROM battle_queue t WHERE faction_id = $1)
	SELECT s.position
	FROM bqpos s
	WHERE s.mech_id = $2;`

	err := gamedb.StdConn.QueryRow(query, factionID.String(), mechID.String()).Scan(&pos)

	if errors.Is(sql.ErrNoRows, err) {
		return -1, nil
	}

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().
			Str("mech_id", mechID.String()).
			Str("faction_id", factionID.String()).
			Bool("NoRows?", errors.Is(sql.ErrNoRows, err)).
			Str("db func", "QueuePosition").Err(err).Msg("unable to get queue position of mech")
		return -1, err
	}

	return pos, nil
}

func QueueContract(mechID uuid.UUID, factionID uuid.UUID) (*decimal.Decimal, error) {
	var contractReward decimal.Decimal

	// Get latest queue contract
	query := `select contract_reward
		from battle_contracts
		where mech_id = $1 AND faction_id = $2
		order by queued_at desc
		limit 1
	`

	err := gamedb.Conn.QueryRow(context.Background(), query, mechID.String(), factionID.String()).Scan(&contractReward)
	if err != nil {
		gamelog.L.Error().
			Str("mech_id", mechID.String()).
			Str("faction_id", factionID.String()).
			Str("db func", "QueueContract").Err(err).Msg("unable to get battle contract of mech")
		return nil, err
	}

	return &contractReward, nil
}

func QueueFee(mechID uuid.UUID, factionID uuid.UUID) (*decimal.Decimal, error) {
	var queueCost decimal.Decimal

	// Get latest queue contract
	query := `select fee
		from battle_contracts
		where mech_id = $1 AND faction_id = $2
		order by queued_at desc
		limit 1
	`

	err := gamedb.Conn.QueryRow(context.Background(), query, mechID.String(), factionID.String()).Scan(&queueCost)
	if err != nil {
		gamelog.L.Error().
			Str("mech_id", mechID.String()).
			Str("faction_id", factionID.String()).
			Str("db func", "QueueFee").Err(err).Msg("unable to get battle contract of mech")
		return nil, err
	}

	return &queueCost, nil
}

func JoinQueue(mech *BattleMechData, contractReward decimal.Decimal, queueFee decimal.Decimal) (int64, error) {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Str("db func", "JoinQueue").Err(err).Msg("unable to begin tx")
		return 0, err
	}
	defer tx.Rollback()

	exists, err := boiler.BattleQueueExists(tx, mech.MechID.String())
	if err != nil {
		gamelog.L.Error().Str("db func", "JoinQueue").Str("mech_id", mech.MechID.String()).Err(err).Msg("check mech exists in queue")
	}
	if exists {
		gamelog.L.Debug().Str("db func", "JoinQueue").Str("mech_id", mech.MechID.String()).Err(err).Msg("mech already in queue")
		return QueuePosition(mech.MechID, mech.FactionID)
	}

	bc := &boiler.BattleContract{
		MechID:         mech.MechID.String(),
		FactionID:      mech.FactionID.String(),
		PlayerID:       mech.OwnerID.String(),
		ContractReward: contractReward,
		Fee:            queueFee,
	}
	err = bc.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().
			Interface("mech", mech).
			Str("contractReward", contractReward.String()).
			Str("queueFee", queueFee.String()).
			Str("db func", "JoinQueue").Err(err).Msg("unable to create battle contract")
	}

	bq := &boiler.BattleQueue{
		MechID:           mech.MechID.String(),
		QueuedAt:         time.Now(),
		FactionID:        mech.FactionID.String(),
		OwnerID:          mech.OwnerID.String(),
		BattleContractID: null.StringFrom(bc.ID),
	}
	err = bq.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().
			Interface("mech", mech).
			Str("db func", "JoinQueue").Err(err).Msg("unable to insert mech into queue")
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().
			Interface("mech", mech).
			Str("db func", "JoinQueue").Err(err).Msg("unable to commit mech insertion into queue")
	}

	return QueuePosition(mech.MechID, mech.FactionID)
}

func LeaveQueue(mech *BattleMechData) (int64, error) {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Str("db func", "LeaveQueue").Err(err).Msg("unable to begin tx")
		return -1, err
	}
	defer tx.Rollback()
	// Get queue position before deleting
	position, err := QueuePosition(mech.MechID, mech.FactionID)
	if err != nil {
		gamelog.L.Error().
			Interface("mech", mech).
			Str("db func", "LeaveQueue").Err(err).Msg("unable to get mech position")
		return -1, err
	}

	canxq := `UPDATE battle_contracts SET cancelled = true WHERE id = (SELECT battle_contract_id FROM battle_queue WHERE mech_id = $1)`
	_, err = gamedb.StdConn.Exec(canxq, mech.MechID.String())
	if err != nil {
		gamelog.L.Warn().Err(err).Msg("unable to cancel battle contract. mech has left queue though.")
	}

	bw := &boiler.BattleQueue{
		MechID: mech.MechID.String(),
	}
	_, err = bw.Delete(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().
			Interface("mech", mech).
			Str("db func", "LeaveQueue").Err(err).Msg("unable to remove mech from queue")
		return -1, err
	}
	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().
			Interface("mech", mech).
			Str("db func", "LeaveQueue").Err(err).Msg("unable to commit mech deletion from queue")
		return -1, err
	}

	//err = boiler.FindBattleContract()

	return position, nil
}

func QueueSetBattleID(battleID string, mechIDs ...uuid.UUID) error {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Str("db func", "ClearQueue").Err(err).Msg("unable to begin tx")
		return err
	}
	defer tx.Rollback()

	args := make([]interface{}, len(mechIDs)+1)
	args[0] = battleID
	var paramrefs string
	for i, id := range mechIDs {
		paramrefs += `$` + strconv.Itoa(i+2) + `,`
		args[i+1] = id.String()
	}
	if len(args) == 1 {
		fmt.Println("no mechs", len(mechIDs))
	}

	paramrefs = paramrefs[:len(paramrefs)-1]

	query := `UPDATE battle_queue SET battle_id=$1 WHERE mech_id IN (` + paramrefs + `)`
	_, err = gamedb.Conn.Exec(context.Background(), query, args...)
	if err != nil {
		gamelog.L.Error().Interface("paramrefs", paramrefs).Interface("args", args).Str("db func", "ClearQueue").Err(err).Msg("unable to set battle id for mechs from queue")
		return err
	}

	query = `UPDATE battle_contracts SET battle_id=$1 WHERE mech_id IN (` + paramrefs + `)`
	_, err = gamedb.Conn.Exec(context.Background(), query, args...)
	if err != nil {
		gamelog.L.Error().Interface("paramrefs", paramrefs).Interface("args", args).Str("db func", "ClearQueue").Err(err).Msg("unable to set battle id for mechs from battle_contracts")
		return err
	}

	return tx.Commit()
}

func ClearQueueByBattle(battleID string) error {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Str("db func", "ClearQueue").Err(err).Msg("unable to begin tx")
		return err
	}
	defer tx.Rollback()

	query := `DELETE FROM battle_queue WHERE battle_id = $1`
	_, err = gamedb.StdConn.Exec(query, battleID)
	if err != nil {
		gamelog.L.Error().Str("db func", "ClearQueue").Err(err).Msg("unable to delete mechs from queue")
		return err
	}

	return tx.Commit()
}

func ClearQueue(mechIDs ...uuid.UUID) error {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Str("db func", "ClearQueue").Err(err).Msg("unable to begin tx")
		return err
	}
	defer tx.Rollback()

	mechids := make([]interface{}, len(mechIDs))
	var paramrefs string
	for i, id := range mechIDs {
		paramrefs += `$` + strconv.Itoa(i+1) + `,`
		mechids[i] = id.String()
	}
	if len(mechids) == 0 {
		fmt.Println("no mechs", len(mechIDs))
	}

	paramrefs = paramrefs[:len(paramrefs)-1]

	query := `DELETE FROM battle_queue WHERE mech_id IN (` + paramrefs + `)`

	_, err = gamedb.Conn.Exec(context.Background(), query, mechids...)
	if err != nil {
		gamelog.L.Error().Str("db func", "ClearQueue").Err(err).Msg("unable to delete mechs from queue")
		return err
	}

	return tx.Commit()
}

func BattleViewerUpsert(ctx context.Context, conn Conn, battleID string, userID string) error {

	q := `
	insert into battles_viewers (battle_id, player_id) VALUES ($1, $2) on conflict (battle_id, player_id) do nothing; 
	`
	_, err := conn.Exec(ctx, q, battleID, userID)
	if err != nil {
		gamelog.L.Error().Str("db func", "BattleViewerUpsert").Err(err).Msg("unable to upsert battle views")
		return err
	}

	return nil
}
