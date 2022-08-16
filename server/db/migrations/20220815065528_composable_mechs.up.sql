ALTER TABLE
    mech_weapons
ADD
    COLUMN IF NOT EXISTS is_skin_inherited bool NOT NULL DEFAULT false;