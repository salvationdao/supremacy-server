CREATE TABLE discord_lobby_annoucements
(
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    message_id TEXT NOT NULL,
    battle_lobby_id UUID NOT NULL REFERENCES battle_lobbies(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE discord_lobby_followers
(
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    discord_member_id TEXT NOT NULL,
    discord_lobby_annoucements_id UUID NOT NULL REFERENCES discord_lobby_annoucements(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
)