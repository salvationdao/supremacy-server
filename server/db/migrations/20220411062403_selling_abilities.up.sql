create table blueprint_player_abilities (
    id uuid primary key not null default gen_random_uuid(),
    game_client_ability_id int4 NOT NULL,
    label text not null,
	colour text NOT NULL,
	image_url text NOT NULL,
	description text NOT NULL,
	text_colour text NOT NULL
);

create table player_abilities ( -- ephemeral, entries are removed on use
    blueprint_player_ability_id uuid not null references blueprint_player_abilities (id),
    owner_id uuid not null references players (id),
    constraint player_abilities_pkey primary key (blueprint_player_ability_id, owner_id)
);

create table sale_player_abilities (
    blueprint_player_ability_id uuid not null primary key references blueprint_player_abilities (id),
    current_price numeric(28) not null,
    available_until timestamptz
);

create table consumed_abilities (
    battle_id uuid not null primary key references battles (id),
    blueprint_player_ability_id uuid not null references blueprint_player_abilities (id),
    consumed_at timestamptz not null default now()
);
