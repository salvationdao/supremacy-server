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
	"github.com/volatiletech/null/v8"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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

func UpdateKilledBattleMech(battleID string, mechID uuid.UUID, ownerID string, factionID string, killedByID ...uuid.UUID) (*boiler.BattleMech, error) {
	bmd, err := boiler.FindBattleMech(gamedb.StdConn, battleID, mechID.String())
	if err != nil {
		gamelog.L.Error().
			Str("battleID", battleID).
			Str("mechID", mechID.String()).
			Str("db func", "UpdateKilledBattleMech").
			Err(err).Msg("unable to retrieve battle Mech from database")

		bmd = &boiler.BattleMech{
			BattleID:    battleID,
			MechID:      mechID.String(),
			PilotedByID: ownerID,
			FactionID:   factionID,
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
			if err != nil {
				gamelog.L.Warn().Err(err).
					Interface("boiler.MechStat", newMs).
					Msg("unable to create killer mech stat")
			}
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
		if err != nil {
			gamelog.L.Warn().Err(err).
				Interface("boiler.MechStat", newMs).
				Msg("unable to create mech stat")
		}
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

// DefaultFactionPlayers return default mech players
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

type BattleViewer struct {
	BattleID uuid.UUID `db:"battle_id"`
	PlayerID uuid.UUID `db:"player_id"`
}

func BattleViewerUpsert(battleID string, userID string) error {
	test := &BattleViewer{}
	q := `
		SELECT bv.player_id FROM battle_viewers bv WHERE battle_id = $1 AND player_id = $2
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
		INSERT INTO battle_viewers (battle_id, player_id) VALUES ($1, $2) ON CONFLICT (battle_id, player_id) DO NOTHING; 
	`
	_, err = gamedb.StdConn.Exec(q, battleID, userID)
	if err != nil {
		gamelog.L.Error().Str("db func", "BattleViewerUpsert").Str("battle_id", battleID).Str("player_id", userID).Err(err).Msg("unable to upsert battle views")
	}

	// increase battle count
	_, err = UserStatAddViewBattleCount(userID)
	if err != nil {
		gamelog.L.Error().Str("battle_id", battleID).Str("player_id", userID).Err(err).Msg("failed to update user battle view")
		return err
	}

	return nil
}
