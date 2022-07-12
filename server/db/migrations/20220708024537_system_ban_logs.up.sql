-- insert system battle user
INSERT INTO players (id, username, is_ai)
VALUES ('87c60803-b051-4abb-aa60-487104946bd7', 'Battle Arena System', true)
ON CONFLICT do nothing;

INSERT INTO players (id, username, is_ai)
VALUES ('7bba7172-932a-4293-9765-ebd0ae98f0ea', 'System Moderator', true);

INSERT INTO players (id, username, is_ai)
VALUES ('7bea1ab5-cc2e-46bb-95d4-e8082e141f1f', 'System Admin', true);

DROP TYPE IF EXISTS BAN_FROM_TYPE;
CREATE TYPE BAN_FROM_TYPE AS ENUM ('SYSTEM','ADMIN','PLAYER');

CREATE TABLE player_bans(
    id uuid primary key default gen_random_uuid(),
    ban_from BAN_FROM_TYPE NOT NULL,

    -- ban info
    battle_number integer references battles(battle_number),
    banned_player_id uuid not null references players (id),
    banned_by_id uuid not null references players (id),
    reason text not null,
    banned_at timestamptz not null default now(),
    end_at timestamptz not null,

    related_punish_vote_id uuid references punish_votes(id),

    -- unban mechanism
    manually_unban_by_id uuid,
    manually_unban_reason text,
    manually_unban_at timestamptz,

    -- ban option
    ban_sups_contribute bool not null default false,
    ban_location_select bool not null default false,
    ban_send_chat bool not null default false,
    ban_view_chat bool not null default false,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

create index idx_player_ban_search on player_bans(banned_player_id, end_at DESC);
create index idx_player_ban_sup_contribute on player_bans(banned_player_id,ban_sups_contribute, end_at DESC);
create index idx_player_ban_location_select on player_bans(ban_location_select, end_at DESC);
create index idx_player_ban_send_chat on player_bans(banned_player_id,ban_send_chat, end_at DESC);
create index idx_player_ban_view_chat on player_bans(banned_player_id,ban_view_chat, end_at DESC);

INSERT INTO player_bans (ban_from, banned_player_id, banned_by_id, reason, banned_at, end_at, related_punish_vote_id, ban_sups_contribute, ban_location_select, ban_send_chat, ban_view_chat)
SELECT  'PLAYER',
        pp.player_id,
        pv.issued_by_id,
        pv.reason,
        pp.created_at,
        pp.punish_until,
        pp.related_punish_vote_id,
        (SELECT true FROM punish_options po where po.id = pp.punish_option_id AND po.key = 'restrict_sups_contribution') = true,
        (SELECT true FROM punish_options po where po.id = pp.punish_option_id AND po.key = 'restrict_location_select') = true,
        (SELECT true FROM punish_options po where po.id = pp.punish_option_id AND po.key = 'restrict_chat') = true,
        false
from punished_players pp
INNER JOIN punish_votes pv on pv.id = pp.related_punish_vote_id;

DROP TABLE punished_players;

CREATE TABLE player_ips(
    player_id uuid not null references players (id),
    ip text not null,
    first_seen_at timestamptz not null,
    last_seen_at timestamptz not null,
    PRIMARY KEY (player_id, ip)
);