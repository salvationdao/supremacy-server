package db

import (
	"database/sql"
	"errors"
	"fmt"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
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
			Str("db func", "UpdateKilledBattleMech").
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
				Str("db func", "UpdateKilledBattleMech").
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
			warn.Str("db func", "UpdateKilledBattleMech").Msg("more than 1 killer mech provided, only the zero indexed mech will be saved")
		}
		bmd.KilledByID = null.StringFrom(killedByID[0].String())
		kid, err := uuid.FromString(killedByID[0].String())

		killerBmd, err := boiler.FindBattleMech(gamedb.StdConn, battleID, kid.String())
		if err != nil {
			gamelog.L.Error().
				Str("battleID", battleID).
				Str("killerBmdID", killedByID[0].String()).
				Str("db func", "UpdateKilledBattleMech").
				Err(err).Msg("unable to retrieve killer battle Mech from database")

			return nil, err
		}

		killerBmd.Kills = killerBmd.Kills + 1
		_, err = killerBmd.Update(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Warn().Err(err).
				Interface("boiler.BattleMech", killerBmd).
				Msg("unable to update killer battle mech")
		}

		// Update mech_stats for killer mech
		killerMS, err := boiler.MechStats(boiler.MechStatWhere.MechID.EQ(killedByID[0].String())).One(gamedb.StdConn)
		if errors.Is(err, sql.ErrNoRows) {
			// If mech stats not exist then create it
			newMs := boiler.MechStat{
				MechID:     killedByID[0].String(),
				TotalKills: 1,
			}
			err := newMs.Insert(gamedb.StdConn, boil.Infer())
			gamelog.L.Warn().Err(err).
				Interface("boiler.MechStat", newMs).
				Msg("unable to create killer mech stat")
		} else if err != nil {
			gamelog.L.Warn().Err(err).
				Str("mechID", killedByID[0].String()).
				Msg("unable to get killer mech stat")
		} else {
			killerMS.TotalKills = killerMS.TotalKills + 1
			_, err = killerMS.Update(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Warn().Err(err).
					Interface("boiler.MechStat", killerMS).
					Msg("unable to update killer mech stat")
			}
		}

		// Create new entry in battle_kills
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

	// Update mech_stats for killed mech
	ms, err := boiler.MechStats(boiler.MechStatWhere.MechID.EQ(mechID.String())).One(gamedb.StdConn)
	if errors.Is(err, sql.ErrNoRows) {
		// If mech stats not exist then create it
		newMs := boiler.MechStat{
			MechID:      mechID.String(),
			TotalDeaths: 1,
		}
		err := newMs.Insert(gamedb.StdConn, boil.Infer())
		gamelog.L.Warn().Err(err).
			Interface("boiler.MechStat", newMs).
			Msg("unable to create mech stat")
	} else if err != nil {
		gamelog.L.Warn().Err(err).
			Str("mechID", mechID.String()).
			Msg("unable to get mech stat")
	} else {
		ms.TotalDeaths = ms.TotalDeaths + 1
		_, err = ms.Update(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Warn().Err(err).
				Interface("boiler.MechStat", ms).
				Msg("unable to update mech stat")
		}
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
	players, err := boiler.Players(qm.Where("is_ai = true"), boiler.PlayerWhere.FactionID.IsNotNull()).All(gamedb.StdConn)
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
	query := fmt.Sprintf(`
		SELECT %s, %s, %s, %s, %s, %s, %s, %s
		FROM (
			SELECT ROW_NUMBER() OVER (PARTITION BY faction_id ORDER BY queued_at ASC) AS r, t.*
			FROM battle_queue t
		) x
		WHERE x.r <= $1
	`,
		boiler.BattleQueueColumns.ID,
		boiler.BattleQueueColumns.MechID,
		boiler.BattleQueueColumns.QueuedAt,
		boiler.BattleQueueColumns.FactionID,
		boiler.BattleQueueColumns.OwnerID,
		boiler.BattleQueueColumns.BattleID,
		boiler.BattleQueueColumns.Notified,
		boiler.BattleQueueColumns.SystemMessageNotified,
	)

	result, err := gamedb.StdConn.Query(query, lengthPerFaction)
	if err != nil {
		gamelog.L.Error().Int("length", lengthPerFaction).Err(err).Msg("unable to retrieve mechs for load out")
		return nil, err
	}
	defer result.Close()

	queue := []*boiler.BattleQueue{}

	for result.Next() {
		mc := &boiler.BattleQueue{}
		err = result.Scan(&mc.ID, &mc.MechID, &mc.QueuedAt, &mc.FactionID, &mc.OwnerID, &mc.BattleID, &mc.Notified, &mc.SystemMessageNotified)
		if err != nil {
			return nil, err
		}
		queue = append(queue, mc)
	}

	return queue, nil
}

type MechAndPosition struct {
	MechID        uuid.UUID `db:"mech_id"`
	QueuePosition int64     `db:"queue_position"`
}

func QueueLength(factionID uuid.UUID) (int64, error) {
	bqs, err := boiler.BattleQueues(
		qm.Select(
			boiler.BattleQueueColumns.ID,
		),
		boiler.BattleQueueWhere.FactionID.EQ(factionID.String()),
		boiler.BattleQueueWhere.BattleID.IsNull(), // only count the mech that is not in battle
	).All(gamedb.StdConn)
	if err != nil {
		return -1, err
	}

	return int64(len(bqs)), nil
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

func QueueFee(mechID uuid.UUID, factionID uuid.UUID) (*decimal.Decimal, error) {
	var queueCost decimal.Decimal

	// Get latest queue contract
	query := `select fee
		from battle_contracts
		where mech_id = $1 AND faction_id = $2
		order by queued_at desc
		limit 1
	`

	err := gamedb.StdConn.QueryRow(query, mechID.String(), factionID.String()).Scan(&queueCost)
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
		_, err = b.Update(gamedb.StdConn, boil.Whitelist(boiler.BattleQueueColumns.BattleID))
		if err != nil {
			gamelog.L.Error().Str("battle_id", battleID).Interface("battle_queue", b).Str("db func", "QueueSetBattleID").Err(err).Msg("unable to set battle id for mechs from queue")
			continue
		}
	}

	return nil
}

type BattleViewer struct {
	BattleID uuid.UUID `db:"battle_id"`
	PlayerID uuid.UUID `db:"player_id"`
}

func BattleViewerUpsert(battleID string, userID string) error {
	test := &BattleViewer{}
	q := `
		select bv.player_id from battle_viewers bv where battle_id = $1 and player_id = $2
	`
	err := gamedb.StdConn.QueryRow(q, battleID, userID).Scan(&test.PlayerID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("battle_id", battleID).Str("player_id", userID).Err(err).Msg("failed to get battles viewer")
		return err
	}

	// skip if user already insert
	if err == nil {
		return nil
	}

	// insert battle viewers
	q = `
		insert into battle_viewers (battle_id, player_id) VALUES ($1, $2) on conflict (battle_id, player_id) do nothing; 
	`
	_, err = gamedb.StdConn.Exec(q, battleID, userID)
	if err != nil {
		gamelog.L.Error().Str("db func", "BattleViewerUpsert").Str("battle_id", battleID).Str("player_id", userID).Err(err).Msg("unable to upsert battle views")
		return err
	}

	// increase battle count
	_, err = UserStatAddViewBattleCount(userID)
	if err != nil {
		gamelog.L.Error().Str("battle_id", battleID).Str("player_id", userID).Err(err).Msg("failed to update user battle view")
		return err
	}

	return nil
}
