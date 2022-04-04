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

	"github.com/georgysavva/scany/pgxscan"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/ninja-software/terror/v2"
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
			gamelog.L.Error().Interface("battle mech", bmd).Str("db func", "Battle").Err(err).Msg("unable to insert battle Mech into database")
			return err
		}
	}

	return tx.Commit()
}

func UpdateKilledBattleMech(battleID string, mechID uuid.UUID, ownerID string, factionID string, killedByID ...uuid.UUID) (*boiler.BattleMech, error) {
	bmd, err := boiler.FindBattleMech(gamedb.StdConn, battleID, mechID.String())
	if err != nil {
		gamelog.L.Error().
			Str("battleID", battleID).
			Str("mechID", mechID.String()).
			Str("db func", "UpdateBattleMech").
			Err(err).Msg("unable to retrieve battle Mech from database")

		bmd = &boiler.BattleMech{
			BattleID:  battleID,
			MechID:    mechID.String(),
			OwnerID:   ownerID,
			FactionID: factionID,
		}
		err = bmd.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().
				Str("battleID", battleID).
				Str("mechID", mechID.String()).
				Str("db func", "UpdateBattleMech").
				Err(err).Msg("unable to insert battle Mech into database after not being able to retrieve it")
			return nil, err
		}
	}

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

		killerBmd, err := boiler.FindBattleMech(gamedb.StdConn, battleID, kid.String())
		if err != nil {
			gamelog.L.Error().
				Str("battleID", battleID).
				Str("killerBmdID", killedByID[0].String()).
				Str("db func", "UpdateBattleMech").
				Err(err).Msg("unable to retrieve battle Mech from database")

			return nil, err
		}

		killerBmd.Kills = killerBmd.Kills + 1
		_, err = killerBmd.Update(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Warn().Err(err).
				Interface("boiler.BattleKill", killerBmd).
				Msg("unable to update killer battle mech")
		}

		bk := &boiler.BattleKill{
			MechID:    killedByID[0].String(),
			BattleID:  battleID,
			CreatedAt: bmd.Killed.Time,
			KilledID:  mechID.String(),
		}
		err = bk.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Warn().Err(err).
				Interface("boiler.BattleKill", bk).
				Msg("unable to insert battle kill")
		}
	}
	_, err = bmd.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).
			Interface("boiler.BattleMech", bmd).
			Msg("unable to update battle mech")
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
	query := `
		SELECT id, mech_id, queued_at, faction_id, owner_id, battle_id, notified
		FROM (
			SELECT ROW_NUMBER() OVER (PARTITION BY faction_id ORDER BY queued_at ASC) AS r, t.*
			FROM battle_queue t
		) x
		WHERE x.r <= $1
		`

	result, err := gamedb.Conn.Query(ctx, query, lengthPerFaction)
	if err != nil {
		gamelog.L.Error().Int("length", lengthPerFaction).Err(err).Msg("unable to retrieve mechs for load out")
		return nil, err
	}
	defer result.Close()

	queue := []*boiler.BattleQueue{}

	for result.Next() {
		mc := &boiler.BattleQueue{}
		err = result.Scan(&mc.ID, &mc.MechID, &mc.QueuedAt, &mc.FactionID, &mc.OwnerID, &mc.BattleID, &mc.Notified)
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
func AllMechsAfter(leavingMechPosition int, queuedAt time.Time, factionID uuid.UUID) ([]*MechAndPosition, error) {
	query := `
		WITH bqpos AS (
			SELECT t.*,
				   ROW_NUMBER() OVER(ORDER BY t.queued_at) AS position
			FROM battle_queue WHERE faction_id = $1 AND queued_at > $2)
			SELECT s.mech_id, s.position+$3-1
			FROM bqpos s
		`

	rows, err := gamedb.StdConn.Query(query, factionID.String(), queuedAt, leavingMechPosition)
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

// QueueOwnerList returns the mech's in queue from an owner.
func QueueOwnerList(userID uuid.UUID) ([]*MechAndPosition, error) {
	q := `
		SELECT q.mech_id, q.position
		FROM (
			SELECT _q.mech_id, ROW_NUMBER() OVER(ORDER BY _q.queued_at) AS position, _q.owner_id
			FROM battle_queue _q
			WHERE _q.faction_id = (
				SELECT _p.faction_id 
				FROM players _p
				WHERE _p.id = $1
			)
		) q
		WHERE q.owner_id = $1`
	rows, err := gamedb.StdConn.Query(q, userID.String())
	if err != nil {
		gamelog.L.Error().
			Str("user_id", userID.String()).
			Str("db func", "OueueOwnerList").Err(err).Msg("unable to grab queue status of mechs")
		return nil, err
	}
	defer rows.Close()

	output := []*MechAndPosition{}
	for rows.Next() {
		var (
			mechID   string
			position int64
		)
		err = rows.Scan(&mechID, &position)
		if err != nil {
			gamelog.L.Error().
				Str("user_id", userID.String()).
				Str("db func", "OueueOwnerList").Err(err).Msg("unable to scan queue status of mech")
			return nil, err
		}

		mechUUID, err := uuid.FromString(mechID)
		if err != nil {
			gamelog.L.Error().
				Str("user_id", userID.String()).
				Str("mech_id", mechID).
				Str("db func", "OueueOwnerList").Err(err).Msg("unable to parse queue mech id from queue status")
			return nil, err
		}

		obj := &MechAndPosition{
			MechID:        mechUUID,
			QueuePosition: position,
		}
		output = append(output, obj)
	}
	return output, nil
}

// QueuePosition returns the current queue position of the specified mech.
// QueuePosition returns -1 if the mech is in battle.
func QueuePosition(mechID uuid.UUID, factionID uuid.UUID) (int64, error) {
	var pos int64

	inBattle, err := boiler.BattleQueues(
		boiler.BattleQueueWhere.MechID.EQ(mechID.String()),
		boiler.BattleQueueWhere.BattleID.IsNotNull(),
	).Exists(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().
			Str("mech_id", mechID.String()).
			Str("faction_id", factionID.String()).
			Str("db func", "QueuePosition").Err(err).Msg("unable to check battle status of mech")
		return -1, err
	}
	if inBattle {
		return -1, nil
	}

	query := `WITH bqpos AS (
    SELECT t.*,
           ROW_NUMBER() OVER(ORDER BY t.queued_at) AS position
    FROM battle_queue t WHERE faction_id = $1)
	SELECT s.position
	FROM bqpos s
	WHERE s.mech_id = $2;`
	err = gamedb.StdConn.QueryRow(query, factionID.String(), mechID.String()).Scan(&pos)
	if err != nil {
		if !errors.Is(sql.ErrNoRows, err) {
			gamelog.L.Error().
				Str("mech_id", mechID.String()).
				Str("faction_id", factionID.String()).
				Bool("NoRows?", errors.Is(sql.ErrNoRows, err)).
				Str("db func", "QueuePosition").Err(err).Msg("unable to get queue position of mech")
		}
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

func QueueSetBattleID(battleID string, mechIDs ...uuid.UUID) error {
	if len(mechIDs) == 0 {
		gamelog.L.Warn().Str("battle_id", battleID).Msg("battle mech is empty")
		return nil
	}

	args := make([]interface{}, len(mechIDs))
	for i, id := range mechIDs {
		args[i] = id.String()
	}
	if len(args) == 0 {
		gamelog.L.Error().Interface("args", args).Str("db func", "QueueSetBattleID").Msg("zero mechs in queue")
		return nil
	}

	bq, err := boiler.BattleQueues(qm.WhereIn("mech_id IN ?", args...)).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Interface("args", args).Str("db func", "QueueSetBattleID").Err(err).Msg("unable to retrieve battle queue from WHERE IN query")
		return err
	}

	for _, b := range bq {
		b.BattleID = null.StringFrom(battleID)
		_, err = b.Update(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Str("battle_id", battleID).Interface("battle_queue", b).Str("db func", "QueueSetBattleID").Err(err).Msg("unable to set battle id for mechs from queue")
			continue
		}
		if b.BattleContractID.String == "" {
			gamelog.L.Error().Str("battle_id", battleID).Interface("battle_queue", b).Str("db func", "QueueSetBattleID").Msg("queue entry did not have a contract")
			continue
		}
		bc, err := boiler.FindBattleContract(gamedb.StdConn, b.BattleContractID.String)
		if err != nil {
			gamelog.L.Error().Str("battle_id", battleID).Interface("battle_queue", b).Str("db func", "QueueSetBattleID").Err(err).Msg("unable to set battle id for mechs for contract queue")
			continue
		}
		bc.BattleID = null.StringFrom(battleID)
		_, err = bc.Update(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Str("battle_id", battleID).Interface("battle_contract", bc).Str("db func", "QueueSetBattleID").Err(err).Msg("unable to set battle id for battle contract")
			continue
		}
	}

	return nil
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

	query := `UPDATE battle_queue SET deleted_at = NOW() WHERE mech_id IN (` + paramrefs + `)`

	_, err = gamedb.Conn.Exec(context.Background(), query, mechids...)
	if err != nil {
		gamelog.L.Error().Str("db func", "ClearQueue").Err(err).Msg("unable to delete mechs from queue")
		return err
	}

	return tx.Commit()
}

type BattleViewer struct {
	BattleID uuid.UUID `db:"battle_id"`
	PlayerID uuid.UUID `db:"player_id"`
}

func BattleViewerUpsert(ctx context.Context, conn Conn, battleID string, userID string) error {
	test := &BattleViewer{}
	q := `
		select bv.player_id from battles_viewers bv where battle_id = $1 and player_id = $2
	`
	err := pgxscan.Get(context.Background(), gamedb.Conn, test, q, battleID, userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		gamelog.L.Error().Str("battle_id", battleID).Str("player_id", userID).Err(err).Msg("failed to get battles viewer")
		return terror.Error(err)
	}

	// skip if user already insert
	if err == nil {
		return nil
	}

	// insert battle viewers
	q = `
	insert into battles_viewers (battle_id, player_id) VALUES ($1, $2) on conflict (battle_id, player_id) do nothing; 
	`
	_, err = conn.Exec(ctx, q, battleID, userID)
	if err != nil {
		gamelog.L.Error().Str("db func", "BattleViewerUpsert").Str("battle_id", battleID).Str("player_id", userID).Err(err).Msg("unable to upsert battle views")
		return err
	}

	// increase battle count
	_, err = UserStatAddViewBattleCount(userID)
	if err != nil {
		gamelog.L.Error().Str("battle_id", battleID).Str("player_id", userID).Err(err).Msg("failed to update user battle view")
		return terror.Error(err)
	}

	return nil
}
