package battle

import (
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"
)

const (
	RepairModeStandard = "STANDARD"
	RepairModeFast     = "FAST"
)

func (btl *Battle) processWarMachineRepair(payload *BattleEndPayload) {
	// get war machines that required repair
	requireRepairedWarMachinIDs := []string{}
	for _, wm := range btl.WarMachines {
		isWin := false
		for _, wwm := range payload.WinningWarMachines {
			if wm.Hash == wwm.Hash {
				isWin = true
				break
			}
		}
		if !isWin {
			requireRepairedWarMachinIDs = append(requireRepairedWarMachinIDs, wm.ID)
		}
	}

	if len(requireRepairedWarMachinIDs) == 0 {
		gamelog.L.Warn().Str("battle id", btl.ID).Msg("There is no war machine needs repair, which shouldn't happen!!!")
		return
	}

	mechs, err := boiler.Mechs(
		qm.Select(boiler.MechColumns.ID, boiler.MechColumns.IsInsured),
		boiler.MechWhere.ID.IN(requireRepairedWarMachinIDs),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("battle id", btl.ID).Interface("mech id list", requireRepairedWarMachinIDs).Msg("Failed to get mechs from db")
		return
	}

	now := time.Now()
	for _, mech := range mechs {
		repairFee := btl.arena.InsurancePrice(mech.ID)

		ar := boiler.AssetRepair{
			MechID:        mech.ID,
			RepairMode:    RepairModeFast,
			CompleteUntil: now.Add(30 * time.Minute),
			FullRepairFee: repairFee,
		}

		// if mech is not insured
		if !mech.IsInsured {
			ar.RepairMode = RepairModeStandard
			ar.CompleteUntil = now.Add(24 * time.Hour)              // change repair time to 24 hours
			ar.FullRepairFee = repairFee.Mul(decimal.NewFromInt(3)) // three time insurance fee
		}

		err := ar.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Str("mech id", mech.ID).Err(err).Msg("Failed to insert asset repair")
		}
	}
}

// InsurancePrice handle price calculation
func (arena *Arena) InsurancePrice(mechID string) decimal.Decimal {
	// get insurance price from mech

	// else get current global insurance price

	return decimal.New(10, 18)
}
