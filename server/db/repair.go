package db

import (
	"github.com/ninja-software/terror/v2"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
)

func RepairOfferDetail(offerID string) (*server.RepairOffer, error) {
	q := `
		SELECT
		    ro.id,
		    ro.repair_case_id,
		    ro.offered_by_id,
		    ro.expires_at,
		    ro.closed_at,
		    ro.created_at,
		    ro.offered_sups_amount,
		    ro.finished_reason,
		    rc.blocks_required_repair,
		    rc.blocks_repaired,
		    ro.offered_sups_amount/ro.blocks_total as sups_worth_per_block,
		    p.username,
		    p.gid,
		    p.faction_id,
		    COUNT(ra.id) as working_agent_count
		FROM repair_offers ro
		INNER JOIN repair_cases rc on rc.id = ro.repair_case_id
		INNER JOIN players p on p.id = ro.offered_by_id
		LEFT JOIN repair_agents ra on ra.repair_offer_id = ro.id AND ra.finished_at ISNULL
		WHERE ro.id = $1
		GROUP BY ro.id, rc.blocks_required_repair, rc.blocks_repaired, ro.offered_sups_amount, ro.blocks_total, ro.closed_at, ro.finished_reason,p.username,p.gid,p.faction_id
	`
	dro := &server.RepairOffer{
		RepairOffer: &boiler.RepairOffer{},
		JobOwner:    &boiler.Player{},
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

// AbandonRepairAgent abandon repair agent and return the repair offer id
func AbandonRepairAgent(repairAgentID string) (string, error) {
	offerID := ""

	q := `
		UPDATE
			repair_agents
		SET
		    finished_at = now(),
		    finished_reason = $2
		WHERE
		    finished_at ISNULL AND id = $1
	`

	err := gamedb.StdConn.QueryRow(q, repairAgentID, boiler.RepairAgentFinishReasonABANDONED).Scan(&offerID)
	if err != nil {
		return "", err
	}

	return offerID, nil
}

// IsRepairCaseOwner check the player is the owner of the repair case
func IsRepairCaseOwner(caseID string, playerID string) (bool, error) {
	isOwner := false
	q := `
		SELECT 
			COALESCE(
			    (SELECT true FROM collection_items ci WHERE ci.item_id = rc.mech_id AND ci.owner_id = $2), 
			    false
			)    
		FROM repair_cases rc
		WHERE rc.id = $1
	`

	err := gamedb.StdConn.QueryRow(q, caseID, playerID).Scan(&isOwner)
	if err != nil {
		gamelog.L.Error().Err(err).Str("q", q).Str("$1", caseID).Str("$2", playerID).Msg("Failed to check repair case owner.")
		return false, terror.Error(err, "Failed to check repair case owner.")
	}

	return isOwner, nil
}
