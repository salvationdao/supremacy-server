ALTER TYPE item_type ADD VALUE 'weapon_skin';
ALTER TABLE weapon_skin
    ADD CONSTRAINT fk_weapon_skin_weapon_model FOREIGN KEY (weapon_model_id) REFERENCES weapon_models (id);