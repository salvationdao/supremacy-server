CREATE TABLE multipliers (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    description TEXT NOT NULL,
    key TEXT NOT NULL UNIQUE,
    value NUMERIC NOT NULL default 1
);

INSERT INTO multipliers (key, value) VALUES
    ('citizen', 5, 'When a player is within the top 80% of voting average.')
    ('supporter', 5, 'When a player is within the top 50% of voting average.')
    ('contributor', 5, 'When a player is within the top 75% of voting average.')
    ('super contributor', 5, 'When a player is within the top 10% of voting average.')
    ('a fool and his money', 5, 'For a player who has put the most individual SUPS in to vote but still lost.')
    ('air support', 5, 'For a player who won an airstrike.')
    ('now i am become death', 5, 'For a player who won a nuke.')
    ('destroyer of worlds', 5, 'For a player who has won the previous three nukes.')
    ('grease monkey', 5,'For a player who won a repair drop.' )
    ('field mechanic', 5, 'For a player who has won the previous three repair drops.')
    ('combo breaker', 5, 'For a player who wins the vote for their syndicate after it has lost the last three rounds.')
    ('mech commander', 5, 'When a player''s mech wins the battles.')
    ('admiral', 5, 'When a player''s mech wins the last 3 battles.')
    ('fiend', 5, '3 hours of active playing.')
    ('juke-e', 5, '6 hours of active playing.')
    ('mech head', 5,'10 hours of active playing.')
    ('sniper', 5,'For a player who has won the vote by dropping in big.' )
    ('won battle', 5, 'When a player''''s syndicate has won the last.')
    ('won last three battles' 5, 'When a player''s syndicate has won the last 3 battles.')
);






















CREATE TABLE user_multipliers (
    player_id UUID NOT NULL references players(id),
    from_battle_id UUID NOT NULL references battles(id),
    during_battle_id UUID NULL references battles(id),
    multiplier UUID NOT NULL references multipliers(id),
    created_at TIMESTAMPTZ NOT NULL default NOW(),
    PRIMARY KEY(player_id, from_battle_id, multiplier)
);