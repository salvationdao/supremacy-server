CREATE TABLE multipliers (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    description TEXT NOT NULL,
    key TEXT NOT NULL UNIQUE,
    for_games INT NOT NULL default 1,
    value INT NOT NULL default 1
);

INSERT INTO multipliers (key, value, description, for_games) VALUES
    ('citizen', 10, 'When a player is within the top 80% of voting average.', 1),
    ('supporter', 25, 'When a player is within the top 50% of voting average.', 1),
    ('contributor', 50, 'When a player is within the top 75% of voting average.', 1),
    ('super contributor', 50, 'When a player is within the top 10% of voting average.', 1),
    ('a fool and his money', 50, 'A player who has put the most individual SUPS in but still didn''t trigger the ability.', 1),
    ('air support', 50, 'For a player who won an airstrike.', 1),
    ('now i am become death', 50, 'For a player who won a nuke.', 1),
    ('destroyer of worlds', 100, 'For a player who has won the previous three nukes.', 1),
    ('grease monkey', 25,'For a player who won a repair drop.', 1),
    ('field mechanic', 50, 'For a player who has won the previous three repair drops.', 1),
    ('combo breaker', 50, 'For a player who wins the vote for their syndicate after it has lost the last three rounds.', 1),
    ('mech commander', 50, 'When a player''s mech wins the battles.', 1),
    ('admiral', 100, 'When a player''s mech wins the last 3 battles.', 1),
    ('fiend', 25, '3 hours of active playing.', 1),
    ('juke-e', 50, '6 hours of active playing.', 1),
    ('mech head', 100,'10 hours of active playing.', 1),
    ('sniper', 100,'For a player who has won the vote by dropping in big.', 1),
    ('won battle', 50, 'When a player''s syndicate has won the last gamec.', 1),
    ('won last three battles', 100, 'When a player''s syndicate has won the last 3 battles.', 1);

CREATE TABLE user_multipliers (
    player_id UUID NOT NULL references players(id),
    from_battle_number INT NOT NULL references battles(battle_number),
    until_game_number INT NOT NULL,
    multiplier UUID NOT NULL references multipliers(id),
    created_at TIMESTAMPTZ NOT NULL default NOW(),
    PRIMARY KEY(player_id, from_battle_number, multiplier)
);