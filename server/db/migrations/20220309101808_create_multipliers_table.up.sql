DROP TYPE IF EXISTS MULTIPLIER_TYPE_ENUM;

DROP TYPE IF EXISTS ABILITY_TYPE_ENUM;

CREATE TYPE MULTIPLIER_TYPE_ENUM AS ENUM (
    'spend_average',
    'most_sups_lost',
    'gab_ability',
    'combo_breaker',
    'player_mech',
    'hours_online',
    'syndicate_win'
);

CREATE TYPE ABILITY_TYPE_ENUM AS ENUM (
    'AIRSTRIKE',
    'NUKE',
    'REPAIR',
    'ROB',
    'REINFORCEMENTS',
    'ROBOT DOGS',
    'OVERCHARGE'
);

CREATE TABLE battle_contributions (
    id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
    battle_id uuid NOT NULL REFERENCES battles (id),
    player_id uuid NOT NULL REFERENCES players (id),
    ability_offering_id uuid NOT NULL,
    did_trigger bool NOT NULL DEFAULT FALSE,
    faction_id uuid NOT NULL REFERENCES factions (id),
    ability_label ABILITY_TYPE_ENUM NOT NULL,
    is_all_syndicates bool NOT NULL DEFAULT FALSE,
    amount numeric(28) NOT NULL,
    contributed_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE battle_ability_triggers (
    id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
    player_id uuid NULL REFERENCES players (id),
    battle_id uuid NOT NULL REFERENCES battles (id),
    faction_id uuid NOT NULL REFERENCES factions (id),
    is_all_syndicates bool NOT NULL DEFAULT FALSE,
    triggered_at timestamptz NOT NULL DEFAULT NOW(),
    ability_label ABILITY_TYPE_ENUM NOT NULL,
    ability_offering_id uuid NOT NULL UNIQUE,
    game_ability_id uuid NOT NULL REFERENCES game_abilities (id)
);

CREATE TABLE multipliers (
    id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
    description text NOT NULL,
    key TEXT NOT NULL UNIQUE,
    for_games int NOT NULL DEFAULT 1,
    multiplier_type MULTIPLIER_TYPE_ENUM NOT NULL,
    must_be_online bool NOT NULL DEFAULT TRUE,
    test_number int NOT NULL,
    test_string text NOT NULL,
    value numeric(28) NOT NULL DEFAULT 1
);

INSERT INTO multipliers (key, value, description, for_games, multiplier_type, test_number, test_string, must_be_online)
    VALUES ('citizen', 10, 'When a player is within the top 80% of ability $SUPS average.', 1, 'spend_average', 80, '', TRUE), ('supporter', 25, 'When a player is within the top 50% of ability $SUPS average.', 1, 'spend_average', 50, '', TRUE), ('contributor', 50, 'When a player is within the top 25% of ability $SUPS average.', 1, 'spend_average', 25, '', TRUE), ('super contributor', 50, 'When a player is within the top 10% of ability $SUPS average.', 1, 'spend_average', 10, '', TRUE), ('a fool and his money', 50, 'A player who has put the most individual SUPS in but still didn''t trigger the ability.', 1, 'most_sups_lost', 0, '', TRUE), ('air support', 50, 'For a player who triggered the last airstrike of the battle', 1, 'gab_ability', 1, 'AIRSTRIKE', TRUE), ('air marshal', 50, 'For a player who triggered the last three airstrikes', 1, 'gab_ability', 3, 'AIRSTRIKE', TRUE), ('now i am become death', 50, 'For a player who triggered a nuke.', 1, 'gab_ability', 1, 'NUKE', TRUE), ('destroyer of worlds', 100, 'For a player who has triggered the previous three nukes.', 1, 'gab_ability', 3, 'NUKE', TRUE), ('grease monkey', 25, 'For a player who triggered a repair drop.', 1, 'gab_ability', 1, 'REPAIR', TRUE), ('field mechanic', 50, 'For a player who has triggered the previous three repair drops.', 1, 'gab_ability', 3, 'REPAIR', TRUE), ('combo breaker', 50, 'For a player who triggers an ability for their syndicate after it has lost the last three rounds.', 1, 'combo_breaker', 3, '', TRUE), ('c-c-c-c-combo breaker', 50, 'For a player who triggers an ability for their syndicate after it has lost the last ten rounds.', 3, 'combo_breaker', 10, '', TRUE), ('mech commander', 50, 'When a player''s mech wins the battles.', 1, 'player_mech', 1, '', FALSE), ('admiral', 100, 'When a player''s mech wins the last 3 battles.', 1, 'player_mech', 3, '', FALSE), ('fiend', 25, '3 hours of active playing.', 1, 'hours_online', 3, '', TRUE), ('junk-e', 50, '6 hours of active playing.', 1, 'hours_online', 6, '', TRUE), ('mech head', 100, '10 hours of active playing.', 1, 'hours_online', 10, '', TRUE), ('won battle', 50, 'When a player''s syndicate has won the last battle.', 1, 'syndicate_win', 1, '', TRUE), ('won last three battles', 100, 'When a player''s syndicate has won the last 3 battles.', 3, 'syndicate_win', 3, '', TRUE);

-- user multiplier must have the value as of when it is set
CREATE TABLE user_multipliers (
    player_id uuid NOT NULL REFERENCES players (id),
    from_battle_number int NOT NULL REFERENCES battles (battle_number),
    until_battle_number int NOT NULL,
    multiplier_id uuid NOT NULL REFERENCES multipliers (id),
    value numeric(28) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    PRIMARY KEY (player_id, from_battle_number, multiplier_id)
);

