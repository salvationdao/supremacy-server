DROP TYPE IF EXISTS MOD_ACTION_TYPE;
CREATE TYPE MOD_ACTION_TYPE AS ENUM ('BAN', 'UNBAN', 'RESTART');

CREATE TABLE mod_action_audit
(
    id            UUID PRIMARY KEY                 NOT NULL DEFAULT gen_random_uuid(),
    action_type   MOD_ACTION_TYPE                  NOT NULL,
    mod_id        UUID references players (id)     NOT NULL,
    reason        TEXT                             NOT NULL,
    player_ban_id UUID references player_bans (id) NULL
);