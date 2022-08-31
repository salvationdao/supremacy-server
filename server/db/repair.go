package db

import (
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
)

func RepairOfferDetail(offerID string) (*server.RepairOffer, error) {
	q := fmt.Sprintf(`
		SELECT
		    ro.%[2]s,
		    ro.%[3]s,
		    ro.%[4]s,
		    ro.%[5]s,
		    ro.%[6]s,
		    ro.%[7]s,
		    ro.%[8]s,
		    ro.%[9]s,
		    rc.%[13]s,
		    rc.%[14]s,
		    ro.%[8]s/ro.%[10]s AS sups_worth_per_block,
		    p.%[17]s,
		    p.%[18]s,
		    p.%[19]s,
		    COUNT(ra.%[21]s) AS working_agent_count
		FROM %[1]s ro
		INNER JOIN %[11]s rc ON rc.%[12]s = ro.%[3]s
		INNER JOIN %[15]s p ON p.%[16]s = ro.%[4]s
		LEFT JOIN %[20]s ra ON ra.%[22]s = ro.%[2]s AND ra.%[23]s ISNULL
		WHERE ro.%[2]s = $1
		GROUP BY ro.%[2]s, rc.%[13]s, rc.%[14]s, ro.%[8]s, ro.%[10]s, ro.%[6]s, ro.%[9]s, p.%[17]s, p.%[18]s, p.%[19]s
	`,
		boiler.TableNames.RepairOffers,              // 1
		boiler.RepairOfferColumns.ID,                // 2
		boiler.RepairOfferColumns.RepairCaseID,      // 3
		boiler.RepairOfferColumns.OfferedByID,       // 4
		boiler.RepairOfferColumns.ExpiresAt,         // 5
		boiler.RepairOfferColumns.ClosedAt,          // 6
		boiler.RepairOfferColumns.CreatedAt,         // 7
		boiler.RepairOfferColumns.OfferedSupsAmount, // 8
		boiler.RepairOfferColumns.FinishedReason,    // 9
		boiler.RepairOfferColumns.BlocksTotal,       // 10

		boiler.TableNames.RepairCases,                 // 11
		boiler.RepairCaseColumns.ID,                   // 12
		boiler.RepairCaseColumns.BlocksRequiredRepair, // 13
		boiler.RepairCaseColumns.BlocksRepaired,       // 14

		boiler.TableNames.Players,      // 15
		boiler.PlayerColumns.ID,        // 16
		boiler.PlayerColumns.Username,  // 17
		boiler.PlayerColumns.Gid,       // 18
		boiler.PlayerColumns.FactionID, // 19

		boiler.TableNames.RepairAgents,          // 20
		boiler.RepairAgentColumns.ID,            // 21
		boiler.RepairAgentColumns.RepairOfferID, // 22
		boiler.RepairAgentColumns.FinishedAt,    // 23

	)
	dro := &server.RepairOffer{
		RepairOffer: &boiler.RepairOffer{},
		JobOwner:    &server.PublicPlayer{},
	}
	err := gamedb.StdConn.QueryRow(q, offerID).Scan(
		&dro.ID,
		&dro.RepairCaseID,
		&dro.OfferedByID,
		&dro.ExpiresAt,
		&dro.ClosedAt,
		&dro.CreatedAt,
		&dro.OfferedSupsAmount,
		&dro.FinishedReason,
		&dro.BlocksRequiredRepair,
		&dro.BlocksRepaired,
		&dro.SupsWorthPerBlock,
		&dro.JobOwner.Username,
		&dro.JobOwner.Gid,
		&dro.JobOwner.FactionID,
		&dro.WorkingAgentCount,
	)
	if err != nil {
		return nil, err
	}

	return dro, nil
}

// IsRepairCaseOwner check the player is the owner of the repair case
func IsRepairCaseOwner(caseID string, playerID string) (bool, error) {
	isOwner := false
	q := fmt.Sprintf(`
		SELECT 
			COALESCE(
			    (SELECT TRUE FROM %[1]s ci WHERE ci.%[3]s = rc.%[4]s AND ci.%[5]s = $2), 
			    FALSE
			)    
		FROM %[2]s rc
		WHERE rc.%[6]s = $1
	`,
		boiler.TableNames.CollectionItems,    // 1
		boiler.TableNames.RepairCases,        // 2
		boiler.CollectionItemColumns.ItemID,  // 3
		boiler.RepairCaseColumns.MechID,      // 4
		boiler.CollectionItemColumns.OwnerID, // 5
		boiler.RepairCaseColumns.ID,          // 6
	)

	err := gamedb.StdConn.QueryRow(q, caseID, playerID).Scan(&isOwner)
	if err != nil {
		gamelog.L.Error().Err(err).Str("q", q).Str("$1", caseID).Str("$2", playerID).Msg("Failed to check repair case owner.")
		return false, terror.Error(err, "Failed to check repair case owner.")
	}

	return isOwner, nil
}

func TotalRepairBlocks(mechID string) int {
	totalRepairBlocks := GetIntWithDefault(KeyDefaultRepairBlocks, 5)
	bm, err := boiler.BlueprintMechs(
		qm.InnerJoin(
			fmt.Sprintf(
				"%s ON %s = %s AND %s = ?",
				boiler.TableNames.Mechs,
				qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.BlueprintID),
				qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ID),
				qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
			),
			mechID,
		),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load total repair blocks")
		return totalRepairBlocks
	}

	return bm.RepairBlocks
}

func DecrementRepairSlotNumber(conn Conn, playerID string) error {

	return nil
}
