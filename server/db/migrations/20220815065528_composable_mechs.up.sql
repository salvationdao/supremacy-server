ALTER TABLE mech_weapons
    ALTER COLUMN weapon_id DROP NOT NULL,
    DROP CONSTRAINT IF EXISTS chassis_weapons_chassis_id_slot_number_key,
    DROP CONSTRAINT IF EXISTS chassis_weapons_pkey,
    ADD CONSTRAINT chassis_weapons_pkey PRIMARY KEY (chassis_id, slot_number),
    DROP COLUMN IF EXISTS id,
    ADD COLUMN IF NOT EXISTS is_skin_inherited bool NOT NULL DEFAULT FALSE;

ALTER TABLE mech_utility
    ALTER COLUMN utility_id DROP NOT NULL,
    DROP CONSTRAINT IF EXISTS chassis_modules_chassis_id_slot_number_key,
    DROP CONSTRAINT IF EXISTS chassis_modules_pkey,
    ADD CONSTRAINT chassis_modules_pkey PRIMARY KEY (chassis_id, slot_number),
    DROP COLUMN IF EXISTS id;

-- New trigger, t_mech_insert for automatically creating mech_weapon and mech_utility entries
-- based on the newly created mech entry's weapon_hardpoints and utility_slots field
DROP FUNCTION IF EXISTS create_mech_slots ();

CREATE OR REPLACE FUNCTION create_mech_slots ()
    RETURNS TRIGGER
    LANGUAGE plpgsql
    AS $$
DECLARE
    weapon_hardpoints int4;
    utility_slots int4;
BEGIN
    SELECT
        bpm.weapon_hardpoints,
        bpm.utility_slots INTO weapon_hardpoints,
        utility_slots
    FROM
        blueprint_mechs bpm
    WHERE
        bpm.id = NEW.blueprint_id;
    FOR i IN 0..weapon_hardpoints - 1 LOOP
        INSERT INTO mech_weapons (chassis_id, slot_number)
            VALUES (NEW.id, i);
    END LOOP;
    FOR i IN 0..utility_slots - 1 LOOP
        INSERT INTO mech_utility (chassis_id, slot_number)
            VALUES (NEW.id, i);
    END LOOP;
    RETURN NEW;
END
$$;

DROP TRIGGER IF EXISTS t_mech_insert ON mechs;

CREATE TRIGGER "t_mech_insert"
    AFTER INSERT ON "mechs"
    FOR EACH ROW
    EXECUTE PROCEDURE create_mech_slots ();

