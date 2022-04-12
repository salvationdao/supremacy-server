create table blueprint_player_abilities (
    id uuid primary key not null default gen_random_uuid(),
    game_client_ability_id int4 NOT NULL,
    label text not null,
	colour text NOT NULL,
	image_url text NOT NULL,
	description text NOT NULL,
	text_colour text NOT NULL,
    "type" TEXT CHECK ("type" IN ('MECH_SELECT', 'LOCATION_SELECT', 'GLOBAL'))
);

create table player_abilities ( -- ephemeral, entries are removed on use
    id uuid primary key not null default gen_random_uuid(),
    owner_id uuid not null references players (id),
    blueprint_id uuid not null references blueprint_player_abilities (id),
    game_client_ability_id int4 NOT NULL,
    label text not null,
	colour text NOT NULL,
	image_url text NOT NULL,
	description text NOT NULL,
	text_colour text NOT NULL,
    "type" TEXT CHECK ("type" IN ('MECH_SELECT', 'LOCATION_SELECT', 'GLOBAL')),
    purchased_at timestamptz not null default now()
);

create table sale_player_abilities (
    blueprint_id uuid not null primary key references blueprint_player_abilities (id),
    current_price numeric(28) not null,
    available_until timestamptz
);

create table consumed_abilities (
    battle_id uuid not null primary key references battles (id),
    player_ability_id uuid not null references player_abilities (id),
    consumed_at timestamptz not null default now()
);
