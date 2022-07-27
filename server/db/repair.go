package db

import (
	"github.com/ninja-software/terror/v2"
	"server"
	"server/gamedb"
	"server/gamelog"
)

func RepairOfferDetail(offerID string) (*server.RepairOffer, error) {
	q := `
		SELECT 
			ro.id,
			ro.offered_by_id,
			ro.closed_at,
			ro.finished_reason,
			rc.blocks_required_repair,
			rc.blocks_repaired,
			ro.offered_sups_amount/ro.blocks_total as sups_worth_per_block,
			COUNT(ra.id) as working_agent_count
		FROM repair_offers ro
		INNER JOIN repair_cases rc on rc.id = ro.repair_case_id
		INNER JOIN repair_agents ra on ra.repair_offer_id = ro.id AND ra.finished_at ISNULL
		WHERE ro.id = $1
		GROUP BY ro.id, rc.blocks_required_repair, rc.blocks_repaired, ro.offered_sups_amount, ro.blocks_total, ro.closed_at, ro.finished_reason
	`

	dro := &server.RepairOffer{}
	err := gamedb.StdConn.QueryRow(q, offerID).Scan(
		&dro.ID,
		&dro.ClosedAt,
		&dro.FinishedReason,
		&dro.BlocksRequiredRepair,
		&dro.BlocksRequired,
		&dro.SupsWorthPerBlock,
		&dro.WorkingAgentCount,
	)
	if err != nil {
		return nil, terror.Error(err, "Failed to load repair offer detail.")
	}

	return dro, nil
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
