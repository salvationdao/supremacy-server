DROP TYPE IF EXISTS MULTIPLIER_TYPE_ENUM;
DROP TYPE IF EXISTS ABILITY_TYPE_ENUM;
CREATE TYPE MULTIPLIER_TYPE_ENUM AS ENUM ('spend_average', 'most_sups_lost', 'gab_ability','combo_breaker','player_mech','hours_online','syndicate_win');
CREATE TYPE ABILITY_TYPE_ENUM AS ENUM ('AIRSTRIKE','NUKE','REPAIR', 'ROB','REINFORCEMENTS','ROBOT DOGS','OVERCHARGE');

CREATE TABLE battle_contributions (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    battle_id UUID NOT NULL references battles(id),
    ability_offering_id UUID NOT NULL,
    did_trigger BOOL NOT NULL default false,
    faction_id UUID NOT NULL references factions(id),
    ability_label ABILITY_TYPE_ENUM NOT NULL,
    is_all_syndicates BOOL NOT NULL default false,
    amount NUMERIC(28) NOT NULL,
    contributed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE battle_ability_triggers (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    player_id UUID NULL references players(id),
    battle_id UUID NOT NULL references battles(id),
    faction_id UUID NOT NULL references factions(id),
    is_all_syndicates BOOL NOT NULL default false,
    triggered_at TIMESTAMPTZ NOT NULL default NOW(),
    ability_label ABILITY_TYPE_ENUM NOT NULL,
    game_ability_id UUID NOT NULL references game_abilities(id)
);

CREATE TABLE multipliers (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    description TEXT NOT NULL,
    key TEXT NOT NULL UNIQUE,
    for_games INT NOT NULL default 1,
    multiplier_type MULTIPLIER_TYPE_ENUM NOT NULL,
    test_number INT NOT NULL,
    test_string TEXT NOT NULL,
    value INT NOT NULL default 1
);

INSERT INTO multipliers (key, value, description, for_games, multiplier_type, test_number, test_string) VALUES
    ('citizen', 10, 'When a player is within the top 80% of ability $SUPS average.', 1, 'spend_average', 80, ''),
    ('supporter', 25, 'When a player is within the top 50% of ability $SUPS average.', 1, 'spend_average', 50, ''),
    ('contributor', 50, 'When a player is within the top 25% of ability $SUPS average.', 1, 'spend_average', 25, ''),
    ('super contributor', 50, 'When a player is within the top 10% of ability $SUPS average.', 1, 'spend_average', 10, ''),

    ('a fool and his money', 50, 'A player who has put the most individual SUPS in but still didn''t trigger the ability.', 1, 'most_sups_lost', 0, ''),

    ('air support', 50, 'For a player who triggered an airstrike.', 1, 'gab_ability', 1, 'AIRSTRIKE'),
    ('air marshal', 50, 'For a player who triggered the last three airstrikes.', 1, 'gab_ability', 3, 'AIRSTRIKE'),
    ('now i am become death', 50, 'For a player who triggered a nuke.', 1,  'gab_ability', 1, 'NUKE'),
    ('destroyer of worlds', 100, 'For a player who has triggered the previous three nukes.', 1,  'gab_ability', 3, 'NUKE'),
    ('grease monkey', 25,'For a player who triggered a repair drop.', 1,  'gab_ability', 1, 'REPAIR'),
    ('field mechanic', 50, 'For a player who has triggered the previous three repair drops.', 1,  'gab_ability', 3, 'REPAIR'),

    ('combo breaker', 50, 'For a player who wins the vote for their syndicate after it has lost the last three rounds.', 1,  'combo_breaker', 3, ''),
    ('c-c-c-c-combo breaker', 50, 'For a player who wins the vote for their syndicate after it has lost the last ten rounds.', 3,  'combo_breaker', 10, ''),

    ('mech commander', 50, 'When a player''s mech wins the battles.', 1,  'player_mech', 1, ''),
    ('admiral', 100, 'When a player''s mech wins the last 3 battles.', 1, 'player_mech', 3, ''),

    ('fiend', 25, '3 hours of active playing.', 1, 'hours_online', 3, ''),
    ('juke-e', 50, '6 hours of active playing.', 1, 'hours_online', 6, ''),
    ('mech head', 100,'10 hours of active playing.', 1, 'hours_online', 10, ''),

    ('won battle', 50, 'When a player''s syndicate has won the last battle.', 1, 'syndicate_win', 1, ''),
    ('won last three battles', 100, 'When a player''s syndicate has won the last 3 battles.', 3, 'syndicate_win', 3, '');

CREATE TABLE user_multipliers (
    player_id UUID NOT NULL references players(id),
    from_battle_number INT NOT NULL references battles(battle_number),
    until_battle_number INT NOT NULL,
    multiplier UUID NOT NULL references multipliers(id),
    created_at TIMESTAMPTZ NOT NULL default NOW(),
    PRIMARY KEY(player_id, from_battle_number, multiplier)
);