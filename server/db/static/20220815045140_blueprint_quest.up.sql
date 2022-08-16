DROP TYPE IF EXISTS QUEST_EVENT_TYPE;
CREATE TYPE QUEST_EVENT_TYPE AS ENUM ( 'daily_quest', 'weekly_quest', 'monthly_quest', 'proving_grounds' );

DROP TYPE IF EXISTS QUEST_KEY;
CREATE TYPE QUEST_KEY AS ENUM ( 'ability_kill', 'mech_kill', 'total_battle_used_mech_commander', 'repair_for_other', 'chat_sent', 'mech_join_battle' );

CREATE TABLE IF NOT EXISTS blueprint_quests
(
    id               UUID PRIMARY KEY          DEFAULT gen_random_uuid(),
    quest_event_type QUEST_EVENT_TYPE NOT NULL,
    key              QUEST_KEY        NOT NULL,
    name             TEXT             NOT NULL,
    description      TEXT             NOT NULL,
    request_amount   INT              NOT NULL,

    created_at       TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX idx_blueprint_quest_round_type ON blueprint_quests (quest_event_type);