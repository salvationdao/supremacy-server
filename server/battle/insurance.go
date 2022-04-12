package battle

import (
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"
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

	// get contract data from each mech
	bcs, err := boiler.BattleContracts(
		boiler.BattleContractWhere.MechID.IN(requireRepairedWarMachinIDs),
		boiler.BattleContractWhere.BattleID.EQ(null.StringFrom(btl.ID)),
		qm.Load(
			boiler.BattleContractRels.Mech,
		),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Warn().Str("battle id", btl.ID).Interface("war machine ids", requireRepairedWarMachinIDs).Msg("Failed to get battle contract from war machine ids")
		return
	}

	now := time.Now()
	// insert mech repair table
	for _, bc := range bcs {
		// calc repair fee
		repairFee := bc.Fee.Div(decimal.NewFromInt(10))
		repairCompleteUntil := now.Add(30 * time.Minute)

		// if mech is not insured
		if !bc.R.Mech.IsInsured {
			repairFee = repairFee.Mul(decimal.NewFromInt(3)) // three time issurance fee
			repairCompleteUntil = now.Add(24 * time.Hour)    // change repair time to 24 hours
		}

		ar := boiler.AssetRepair{
			Hash:              bc.R.Mech.Hash,
			ExpectCompletedAt: repairCompleteUntil,
			RepairMode:        "FAST",
			FullRepairFee:     repairFee,
		}

		err = ar.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Str("mech hash", bc.R.Mech.Hash).Err(err).Msg("Failed to insert asset repair")
		}
	}
}

// InsurancePrice handle price calculation
func (arena *Arena) InsurancePrice() decimal.Decimal {
	return decimal.New(10, 18)
}
