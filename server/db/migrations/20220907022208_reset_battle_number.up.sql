ALTER TABLE spoils_of_war
    DROP CONSTRAINT spoils_of_war_battle_number_fkey,
    ADD CONSTRAINT spoils_of_war_battle_number_fkey FOREIGN KEY (battle_number) REFERENCES battles (battle_number) ON UPDATE CASCADE;

UPDATE battles
SET battle_number = battle_number * (-1);

ALTER SEQUENCE battles_battle_number_seq RESTART WITH 1;
