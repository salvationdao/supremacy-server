package db

import (
	"errors"
	"fmt"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
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

func Battle(id, mapID uuid.UUID, mechData []*BattleMechData) (*boiler.Battle, error) {
	btl := &boiler.Battle{ID: id.String(), GameMapID: mapID.String()}
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Str("db func", "Battle").Err(err).Msg("unable to begin tx")
		return nil, err
	}
	defer tx.Rollback()

	err = btl.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Interface("battle", btl).Str("db func", "Battle").Err(err).Msg("unable to insert Battle into database")
		return nil, err
	}
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
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Str("db Func", "Battle").Msg("unable to commit tx")
		return nil, err
	}

	return btl, nil
}

func UpdateBattleMech(battleID, mechID uuid.UUID, gotKill bool, gotKilled bool, killedByID ...uuid.UUID) (*boiler.BattleMech, error) {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Str("db func", "UpdateBattleMech").Err(err).Msg("unable to begin tx")
		return nil, err
	}
	defer tx.Rollback()

	bmd, err := boiler.FindBattleMech(tx, battleID.String(), mechID.String())
	if err != nil {
		gamelog.L.Error().
			Str("battleID", battleID.String()).
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

			killerBmd, err := boiler.FindBattleMech(tx, battleID.String(), kid.String())
			if err != nil {
				gamelog.L.Error().
					Str("battleID", battleID.String()).
					Str("killerBmdID", killedByID[0].String()).
					Str("db func", "UpdateBattleMech").
					Err(err).Msg("unable to retrieve Battle Mech from database")

				return nil, err
			}

			killerBmd.Kills++
			bk := &boiler.BattleKill{
				BattleID:  battleID.String(),
				MechID:    killedByID[0].String(),
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

func StoreBattleEvent(battleID uuid.UUID, relatedID uuid.UUID, wm1 uuid.UUID, wm2 uuid.UUID, eventType EventType, createdAt time.Time) (*boiler.BattleHistory, error) {
	if battleID.IsNil() || wm1.IsNil() {
		return nil, errors.New("no battle ID provided")
	}
	bh := &boiler.BattleHistory{BattleID: battleID.String(), WarMachineOneID: wm1.String(), EventType: eventType.String()}
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

func WinBattle(battleID uuid.UUID, winCondition string, mechs ...*MechWithOwner) error {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Str("db func", "WinBattle").Err(err).Msg("unable to begin tx")
		return err
	}
	defer tx.Rollback()
	for _, m := range mechs {
		mw := &boiler.BattleWin{
			BattleID:     battleID.String(),
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

	query := `select t.rn
		from (
			select
				mech_id,
				faction_id, 
				queued_at,
				count(*) as cnt,
				row_number() over ( order by max(queued_at) asc ) as rn
			from battle_queue
			group by mech_id
			order by queued_at asc
		) t
		where mech_id = $1 AND faction_id = $2`

	err := gamedb.Conn.QueryRow(context.Background(), query, mechID.String(), factionID.String()).Scan(&pos)
	if err != nil {
		gamelog.L.Error().
			Str("mech_id", mechID.String()).
			Str("faction_id", factionID.String()).
			Str("db func", "QueuePosition").Err(err).Msg("unable to get queue position of mech")
		return -1, err
	}

	return pos, nil
}

func JoinQueue(mech *BattleMechData) (int64, error) {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Str("db func", "JoinQueue").Err(err).Msg("unable to begin tx")
		return 0, err
	}
	defer tx.Rollback()
	bw := &boiler.BattleQueue{
		MechID:    mech.MechID.String(),
		QueuedAt:  time.Now(),
		FactionID: mech.FactionID.String(),
		OwnerID:   mech.OwnerID.String(),
	}
	err = bw.Insert(gamedb.StdConn, boil.Infer())
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

func LeaveQueue(mech *BattleMechData) error {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Str("db func", "LeaveQueue").Err(err).Msg("unable to begin tx")
		return err
	}
	defer tx.Rollback()
	bw := &boiler.BattleQueue{
		MechID: mech.MechID.String(),
	}
	_, err = bw.Delete(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().
			Interface("mech", mech).
			Str("db func", "LeaveQueue").Err(err).Msg("unable to remove mech from queue")
		return err
	}
	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().
			Interface("mech", mech).
			Str("db func", "LeaveQueue").Err(err).Msg("unable to commit mech deletion from queue")
		return err
	}

	return nil
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
