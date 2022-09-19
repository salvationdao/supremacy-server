-- repair block trigger
CREATE OR REPLACE FUNCTION check_repair_block() RETURNS TRIGGER AS
$check_repair_block$
DECLARE
    can_write_block BOOLEAN DEFAULT FALSE;
BEGIN

    SELECT (SELECT ro.expires_at > NOW() AND ro.closed_at IS NULL AND
                   ro.deleted_at IS NULL AND
                   rc.completed_at IS NULL AND
                   (SELECT COUNT(*) FROM repair_blocks rb WHERE rb.repair_case_id = rc.id) < rc.blocks_required_repair
            FROM repair_offers ro
                     INNER JOIN repair_cases rc ON ro.repair_case_id = rc.id
            WHERE ro.id = new.repair_offer_id)
    INTO can_write_block;
-- update blocks required in repair cases and continue the process
    IF can_write_block THEN
        UPDATE repair_cases SET blocks_repaired = blocks_repaired + 1 WHERE id = new.repair_case_id;
        UPDATE repair_agents SET finished_at = NOW(), finished_reason = 'SUCCEEDED' WHERE id = new.repair_agent_id;
        RETURN new;
    ELSE
        RAISE EXCEPTION 'unable to write block';
    END IF;
END
$check_repair_block$
    LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_check_repair_block ON repair_blocks;

CREATE TRIGGER trigger_check_repair_block
    BEFORE INSERT
    ON repair_blocks
    FOR EACH ROW
EXECUTE PROCEDURE check_repair_block();

-- repair agent check
CREATE OR REPLACE FUNCTION check_repair_agent() RETURNS TRIGGER AS
$check_repair_agent$
DECLARE
    can_register BOOLEAN DEFAULT FALSE;
BEGIN

    SELECT (SELECT ro.expires_at > NOW() AND ro.closed_at IS NULL AND
                   ro.deleted_at IS NULL AND
                   rc.completed_at IS NULL AND
                   (SELECT COUNT(*) FROM repair_blocks rb WHERE rb.repair_case_id = rc.id) < rc.blocks_required_repair
            FROM repair_offers ro
                     INNER JOIN repair_cases rc ON ro.repair_case_id = rc.id
            WHERE ro.id = new.repair_offer_id)
    INTO can_register;
-- update blocks required in repair cases and continue the process
    IF can_register THEN
        RETURN new;
    ELSE
        RAISE EXCEPTION 'unable to register repair agent';
    END IF;
END
$check_repair_agent$
    LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_check_repair_agent ON repair_agents;

CREATE TRIGGER trigger_check_repair_agent
    BEFORE INSERT
    ON repair_agents
    FOR EACH ROW
EXECUTE PROCEDURE check_repair_agent();

ALTER TABLE repair_cases
    DROP COLUMN IF EXISTS paused_at;