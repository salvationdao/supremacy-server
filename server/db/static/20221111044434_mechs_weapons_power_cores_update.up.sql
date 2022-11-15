ALTER TABLE blueprint_mechs
    ADD COLUMN walk_speed_modifier numeric NOT NULL DEFAULT 1,
    ADD COLUMN sprint_spread_modifier numeric NOT NULL DEFAULT 1,
    ADD COLUMN idle_drain numeric NOT NULL DEFAULT 1,
    ADD COLUMN walk_drain numeric NOT NULL DEFAULT 2,
    ADD COLUMN run_drain numeric NOT NULL DEFAULT 3;

ALTER TABLE blueprint_weapons
    ADD COLUMN dot_tick_duration int NOT NULL DEFAULT 0,
    ADD COLUMN projectile_life_span int NOT NULL DEFAULT 50,
    ADD COLUMN recoil_force numeric NOT NULL DEFAULT 1,
    ADD COLUMN idle_power_cost int NOT NULL DEFAULT 0;

ALTER TABLE blueprint_power_cores
    ADD COLUMN weapon_share int NOT NULL DEFAULT 40,
    ADD COLUMN movement_share int NOT NULL DEFAULT 20,
    ADD COLUMN utility_share int NOT NULL DEFAULT 20,
    DROP CONSTRAINT blueprint_power_cores_size_check;

ALTER TABLE blueprint_power_cores
    ADD CONSTRAINT blueprint_power_cores_size_check CHECK ((size = ANY (ARRAY['SMALL'::text, 'MEDIUM'::text, 'LARGE'::text, 'TURBO'::text])));

ALTER TYPE POWERCORE_SIZE
    ADD VALUE 'TURBO';

