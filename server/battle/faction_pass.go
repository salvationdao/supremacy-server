package battle

import (
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"
)

func stackedAIMechsCheck() error {
	l := gamelog.L.With().Str("func", "stackedAIMechsCheck").Logger()
	// get AI mechs
	q := fmt.Sprintf(`
		INSERT INTO %s (%s, %s, %s)
		SELECT %s, %s, %s
		FROM %s
		INNER JOIN %s ON %s = %s AND %s = TRUE AND %s NOTNULL
		WHERE %s = $1
		ON CONFLICT DO NOTHING;
	`,
		boiler.TableNames.StakedMechs,
		boiler.StakedMechColumns.MechID,
		boiler.StakedMechColumns.OwnerID,
		boiler.StakedMechColumns.FactionID,
		boiler.CollectionItemTableColumns.ItemID,
		boiler.CollectionItemTableColumns.OwnerID,
		boiler.PlayerTableColumns.FactionID,
		boiler.TableNames.CollectionItems,
		boiler.TableNames.Players,
		boiler.PlayerTableColumns.ID,
		boiler.CollectionItemTableColumns.OwnerID,
		boiler.PlayerTableColumns.IsAi,
		boiler.PlayerTableColumns.FactionID,
		boiler.CollectionItemTableColumns.ItemType,
	)

	_, err := gamedb.StdConn.Exec(q, boiler.ItemTypeMech)
	if err != nil {
		l.Error().Err(err).Msg("Failed to upsert AI mechs into stake pool")
		return terror.Error(err, "Failed to upsert AI mechs into stake pool")
	}

	return nil
}

const (
	FactionStakedMechDashboardKeyStaked    = "STAKED"
	FactionStakedMechDashboardKeyQueue     = "QUEUE"
	FactionStakedMechDashboardKeyDamaged   = "DAMAGED"
	FactionStakedMechDashboardKeyRepairBay = "REPAIR_BAY"
	FactionStakedMechDashboardKeyMVP       = "MVP"
)

func (am *ArenaManager) FactionStakedMechDebounceBroadcaster() {
	shortDuration := 200 * time.Millisecond
	stakedMechTotalUpdateTimer := time.NewTimer(shortDuration)
	stakedMechQueueUpdateTimer := time.NewTimer(shortDuration)
	stakedMechDamagedUpdateTimer := time.NewTimer(shortDuration)
	stakedMechRepairBayUpdateTimer := time.NewTimer(shortDuration)

	longDuration := 30 * time.Second
	mvpStakedMechUpdateTimer := time.NewTimer(longDuration)

	for {
		select {
		// reset timer of specific status
		case keys := <-am.FactionStakedMechDashboardKeyChan:
			for _, key := range keys {
				switch key {
				case FactionStakedMechDashboardKeyStaked:
					// update all the data, when the base of staked mechs is changed
					stakedMechTotalUpdateTimer.Reset(shortDuration)
					stakedMechQueueUpdateTimer.Reset(shortDuration)
					stakedMechDamagedUpdateTimer.Reset(shortDuration)
					stakedMechRepairBayUpdateTimer.Reset(shortDuration)
					mvpStakedMechUpdateTimer.Reset(longDuration)

				case FactionStakedMechDashboardKeyQueue:
					stakedMechQueueUpdateTimer.Reset(shortDuration)
				case FactionStakedMechDashboardKeyDamaged:
					stakedMechDamagedUpdateTimer.Reset(shortDuration)
				case FactionStakedMechDashboardKeyRepairBay:
					stakedMechRepairBayUpdateTimer.Reset(shortDuration)
				case FactionStakedMechDashboardKeyMVP:
					mvpStakedMechUpdateTimer.Reset(longDuration)
				}
			}

		case <-stakedMechTotalUpdateTimer.C:
			go broadcastFactionStakedMechTotal()

		case <-stakedMechQueueUpdateTimer.C:
			go broadcastFactionStakedQueueMechTotal()

		case <-stakedMechDamagedUpdateTimer.C:
			go broadcastFactionStakedDamagedMechTotal()

		case <-stakedMechRepairBayUpdateTimer.C:
			go broadcastFactionStakedMechRepairBay()

		case <-mvpStakedMechUpdateTimer.C:
			go broadcastFactionStakedMechMVP()
		}
	}
}

type FactionStakedMechCount struct {
	factionID string
	count     int
}

func broadcastFactionStakedMechTotal() {
	sms, err := boiler.StakedMechs().All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load faction staked mechs")
		return
	}

	fms := []*FactionStakedMechCount{}
	for _, sm := range sms {
		index := slices.IndexFunc(fms, func(fm *FactionStakedMechCount) bool { return fm.factionID == sm.FactionID })
		if index == -1 {
			fms = append(fms, &FactionStakedMechCount{
				factionID: sm.FactionID,
				count:     0,
			})

			index = len(fms) - 1
		}

		fms[index].count += 1
	}

	for _, fm := range fms {
		ws.PublishMessage(fmt.Sprintf("/faction/%s/staked_mech_count", fm.factionID), server.HubKeyFactionStakedMechCount, fm.count)
	}

	// free up memory
	fms = nil
}

func broadcastFactionStakedQueueMechTotal() {
	blms, err := boiler.BattleLobbiesMechs(
		qm.Where(fmt.Sprintf(
			"EXISTS ( SELECT 1 FROM %s WHERE %s = %s )",
			boiler.TableNames.StakedMechs,
			boiler.StakedMechTableColumns.MechID,
			boiler.BattleLobbiesMechTableColumns.MechID,
		)),
		boiler.BattleLobbiesMechWhere.EndedAt.IsNull(),
		boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load total staked mech in queue.")
		return
	}

	factionInQueueMechCount := []*FactionStakedMechCount{}
	factionBattleReadyMechCount := []*FactionStakedMechCount{}
	factionBattlingMechCount := []*FactionStakedMechCount{}

	for _, blm := range blms {

		// in queue mechs
		if !blm.LockedAt.Valid {
			index := slices.IndexFunc(factionInQueueMechCount, func(fq *FactionStakedMechCount) bool { return fq.factionID == blm.FactionID })
			if index == -1 {
				factionInQueueMechCount = append(factionInQueueMechCount, &FactionStakedMechCount{
					factionID: blm.FactionID,
					count:     0,
				})

				index = len(factionInQueueMechCount) - 1
			}

			factionInQueueMechCount[index].count += 1
			continue
		}

		// battle ready mechs
		if !blm.AssignedToBattleID.Valid {
			index := slices.IndexFunc(factionBattleReadyMechCount, func(fq *FactionStakedMechCount) bool { return fq.factionID == blm.FactionID })
			if index == -1 {
				factionBattleReadyMechCount = append(factionBattleReadyMechCount, &FactionStakedMechCount{
					factionID: blm.FactionID,
					count:     0,
				})

				index = len(factionBattleReadyMechCount) - 1
			}

			factionBattleReadyMechCount[index].count += 1
			continue
		}

		// battling mechs
		index := slices.IndexFunc(factionBattlingMechCount, func(fq *FactionStakedMechCount) bool { return fq.factionID == blm.FactionID })
		if index == -1 {
			factionBattlingMechCount = append(factionBattlingMechCount, &FactionStakedMechCount{
				factionID: blm.FactionID,
				count:     0,
			})

			index = len(factionBattlingMechCount) - 1
		}

		factionBattlingMechCount[index].count += 1
	}

	for _, fq := range factionInQueueMechCount {
		ws.PublishMessage(fmt.Sprintf("/faction/%s/in_queue_staked_mech_count", fq.factionID), server.HubKeyFactionStakedMechInQueueCount, fq.count)
	}

	for _, fq := range factionBattleReadyMechCount {
		ws.PublishMessage(fmt.Sprintf("/faction/%s/battle_ready_staked_mech_count", fq.factionID), server.HubKeyFactionStakedMechBattleReadyCount, fq.count)
	}

	for _, fq := range factionBattlingMechCount {
		ws.PublishMessage(fmt.Sprintf("/faction/%s/in_battle_staked_mech_count", fq.factionID), server.HubKeyFactionStakedMechInBattleCount, fq.count)
	}

	factionInQueueMechCount = nil
	factionBattleReadyMechCount = nil
	factionBattlingMechCount = nil
}

func broadcastFactionStakedDamagedMechTotal() {
	sms, err := boiler.StakedMechs(
		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s AND %s ISNULL AND %s ISNULL AND %s ISNULL",
			boiler.TableNames.RepairCases,
			boiler.RepairCaseTableColumns.MechID,
			boiler.StakedMechTableColumns.MechID,
			boiler.RepairCaseTableColumns.CompletedAt,
			boiler.RepairCaseTableColumns.PausedAt,
			boiler.RepairCaseTableColumns.DeletedAt,
		)),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load staked mechs.")
		return
	}

	factionDamagedMechs := []*FactionStakedMechCount{}

	for _, sm := range sms {
		index := slices.IndexFunc(factionDamagedMechs, func(fsm *FactionStakedMechCount) bool { return fsm.factionID == sm.FactionID })
		if index == -1 {
			factionDamagedMechs = append(factionDamagedMechs, &FactionStakedMechCount{
				factionID: sm.FactionID,
				count:     0,
			})

			index = len(factionDamagedMechs) - 1
		}

		factionDamagedMechs[index].count += 1
	}

	for _, fq := range factionDamagedMechs {
		ws.PublishMessage(fmt.Sprintf("/faction/%s/damaged_staked_mech_count", fq.factionID), server.HubKeyFactionStakedMechDamagedCount, fq.count)
	}

	factionDamagedMechs = nil
}

func broadcastFactionStakedMechRepairBay() {
	queries := []qm.QueryMod{
		qm.Select(
			boiler.StakedMechTableColumns.FactionID,
			boiler.RepairCaseTableColumns.BlocksRequiredRepair,
			boiler.RepairCaseTableColumns.BlocksRepaired,
		),

		qm.From(boiler.TableNames.StakedMechs),

		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s AND %s != '%s'",
			boiler.TableNames.PlayerMechRepairSlots,
			boiler.PlayerMechRepairSlotTableColumns.MechID,
			boiler.StakedMechTableColumns.MechID,
			boiler.PlayerMechRepairSlotTableColumns.Status,
			boiler.RepairSlotStatusDONE,
		)),

		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s AND %s ISNULL AND %s ISNULL",
			boiler.TableNames.RepairCases,
			boiler.RepairCaseTableColumns.MechID,
			boiler.StakedMechTableColumns.MechID,
			boiler.RepairCaseTableColumns.CompletedAt,
			boiler.RepairCaseTableColumns.DeletedAt,
		)),
	}

	rows, err := boiler.NewQuery(queries...).Query(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to repair bay mechs from db.")
		return
	}

	frb := []*server.FactionStakedMechRepairBayResponse{}
	for rows.Next() {
		factionID := ""
		requiredRepairedBlocks := 0
		repairedBlocks := 0

		err = rows.Scan(&factionID, &requiredRepairedBlocks, &repairedBlocks)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to scan mech repair detail.")
			return
		}

		index := slices.IndexFunc(frb, func(rb *server.FactionStakedMechRepairBayResponse) bool { return rb.FactionID == factionID })
		if index == -1 {
			frb = append(frb, &server.FactionStakedMechRepairBayResponse{
				FactionID: factionID,
			})

			index = len(frb) - 1
		}

		frb[index].MechCount += 1
		frb[index].TotalRequiredRepairedBlocks += requiredRepairedBlocks
		frb[index].TotalRepairedBlocks += repairedBlocks
	}

	for _, rb := range frb {
		ws.PublishMessage(fmt.Sprintf("/faction/%s/in_repair_bay_staked_mech", rb.FactionID), server.HubKeyFactionStakedMechInRepairBay, rb)
	}

	frb = nil
}

func broadcastFactionStakedMechMVP() {
	queries := []qm.QueryMod{
		qm.Select(
			boiler.StakedMechTableColumns.FactionID,
			boiler.StakedMechTableColumns.MechID,
		),

		qm.From(boiler.TableNames.StakedMechs),

		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s AND %s = %s AND %s > ?",
			boiler.TableNames.StakedMechBattleLogs,
			boiler.StakedMechBattleLogTableColumns.StakedMechID,
			boiler.StakedMechTableColumns.MechID,
			boiler.StakedMechBattleLogTableColumns.FactionID,
			boiler.StakedMechTableColumns.FactionID,
			boiler.StakedMechBattleLogTableColumns.CreatedAt,
		),
			time.Now().AddDate(0, -3, 0),
		),

		qm.GroupBy(fmt.Sprintf("%s, %s", boiler.StakedMechTableColumns.MechID, boiler.StakedMechTableColumns.FactionID)),
		qm.OrderBy(fmt.Sprintf("COUNT(%s) DESC", boiler.StakedMechBattleLogTableColumns.ID)),
	}

	rows, err := boiler.NewQuery(queries...).Query(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load staked mech battle log")
		return
	}

	limit := 3 // three faction
	var recordedFactionID []string
	var recordedMechID []string
	for rows.Next() {
		factionID := ""
		mechID := ""
		err = rows.Scan(&factionID, &mechID)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to scan faction mech detail.")
			return
		}

		if slices.Index(recordedFactionID, factionID) != -1 {
			continue
		}

		recordedFactionID = append(recordedFactionID, factionID)
		recordedMechID = append(recordedMechID, mechID)

		// skip, if already got mvp mech from each faction
		if len(recordedFactionID) >= limit {
			break
		}
	}

	if len(recordedMechID) == 0 {
		return
	}

	// load mechs
	lms, err := db.LobbyMechsBrief("", recordedMechID...)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load mvp mech")
		return
	}

	for _, lm := range lms {
		ws.PublishMessage(fmt.Sprintf("/faction/%s/mvp_staked_mech", lm.FactionID.String), server.HubKeyFactionMostPopularStakedMech, lm)
	}
}
