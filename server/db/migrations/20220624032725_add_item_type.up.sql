ALTER TYPE item_type ADD VALUE 'weapon_skin';

ALTER TABLE weapon_skin
    DROP CONSTRAINT weapon_skin_equipped_on_fkey;

ALTER TABLE weapon_skin
    ADD CONSTRAINT weapon_skin_equipped_on_fkey FOREIGN KEY (equipped_on) REFERENCES weapons (id);