ALTER TABLE blueprint_mech_skin
    ADD COLUMN IF NOT EXISTS blueprint_weapon_skin_id UUID REFERENCES blueprint_weapon_skin (id);

