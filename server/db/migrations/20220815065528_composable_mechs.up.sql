ALTER TABLE mech_weapons
    ALTER COLUMN weapon_id DROP NOT NULL,
    DROP CONSTRAINT IF EXISTS chassis_weapons_chassis_id_slot_number_key,
    DROP CONSTRAINT IF EXISTS chassis_weapons_pkey,
    ADD CONSTRAINT chassis_weapons_pkey PRIMARY KEY (chassis_id, slot_number),
    DROP COLUMN IF EXISTS id,
    ADD COLUMN IF NOT EXISTS is_skin_inherited bool NOT NULL DEFAULT FALSE;

-- New trigger, t_mech_insert for automatically creating mech_weapon entries
-- based on the newly created mech entry's weapon_hardpoints field
DROP FUNCTION IF EXISTS create_mech_weapons ();

CREATE OR REPLACE FUNCTION create_mech_weapons ()
    RETURNS TRIGGER
    LANGUAGE plpgsql
    AS $$
BEGIN
    FOR i IN 0..NEW.weapon_hardpoints - 1 LOOP
        INSERT INTO mech_weapons (chassis_id, slot_number)
            VALUES (NEW.id, i);
    END LOOP;
    RETURN NEW;
END
$$;

DROP TRIGGER IF EXISTS t_mech_insert ON mechs;

CREATE TRIGGER "t_mech_insert"
    AFTER INSERT ON "mechs"
    FOR EACH ROW
    EXECUTE PROCEDURE create_mech_weapons ();

