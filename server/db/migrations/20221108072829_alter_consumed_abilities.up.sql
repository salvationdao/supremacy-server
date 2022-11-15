ALTER TABLE consumed_abilities
    ADD COLUMN IF NOT EXISTS offering_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';

ALTER TABLE consumed_abilities
    ALTER COLUMN offering_id DROP DEFAULT;

CREATE INDEX IF NOT EXISTS idx_consumed_ability_offering_id ON consumed_abilities(offering_id);

ALTER TYPE battle_event ADD VALUE 'stunned';
ALTER TYPE battle_event ADD VALUE 'hacked';

ALTER TABLE battle_ability_triggers
    ALTER COLUMN ability_label TYPE TEXT USING ability_label :: TEXT;

ALTER TABLE battle_history
    RENAME COLUMN related_id TO battle_ability_offering_id;

ALTER TABLE battle_history
    ADD COLUMN IF NOT EXISTS player_ability_offering_id UUID;