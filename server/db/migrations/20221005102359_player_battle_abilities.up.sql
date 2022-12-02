CREATE TABLE player_battle_abilities (
     id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
     player_id UUID NOT NULL REFERENCES players(id),
     game_ability_id UUID NOT NULL REFERENCES game_abilities(id),
     battle_id UUID NOT NULL REFERENCES battles(id),
     used_at TIMESTAMPTZ,
     deleted_at TIMESTAMPTZ,
     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
