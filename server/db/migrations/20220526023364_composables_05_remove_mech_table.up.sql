--  Here we are trying to remove the unneeded mechs table,
--  basically the mechs table turned out be redundant with the chassis table basically serving the same purpose.
--  1. We are going to swap all the FKs over to use chassis id
--  2. Rename chassis table mechs.
--  3. Yes I know I should have just updated the mechs table to begin with.

-- CREATE TABLE battle_queue (
-- mech_id UUID NOT NULL references mechs (id) PRIMARY KEY,

ALTER TABLE battle_queue
    ADD COLUMN chassis_id UUID REFERENCES chassis (id);

UPDATE battle_queue bq
SET chassis_id = (SELECT c.id
                  FROM mechs m
                           INNER JOIN chassis c ON m.chassis_id = c.id
                  WHERE m.id = bq.mech_id);

ALTER TABLE battle_queue
    DROP CONSTRAINT battle_queue_pkey,
    ADD PRIMARY KEY (chassis_id);

-- deal with table battle_queue_notifications that uses battle_queue.mech_id FK
-- CREATE TABLE battle_queue_notifications (
--     id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
--     battle_id UUID REFERENCES battles(id),
--     queue_mech_id UUID REFERENCES battle_queue(mech_id),
--     mech_id UUID NOT NULL REFERENCES mechs(id),

ALTER TABLE battle_queue_notifications
    ADD COLUMN chassis_id       UUID REFERENCES chassis (id),
    ADD COLUMN queue_chassis_id UUID REFERENCES battle_queue (chassis_id);

UPDATE battle_queue_notifications bqn
SET queue_chassis_id = (SELECT c.id
                        FROM mechs m
                                 INNER JOIN chassis c ON m.chassis_id = c.id
                        WHERE m.id = bqn.queue_mech_id),
    chassis_id       = (SELECT c.id
                        FROM mechs m
                                 INNER JOIN chassis c ON m.chassis_id = c.id
                        WHERE m.id = bqn.mech_id);

ALTER TABLE battle_queue_notifications
    DROP COLUMN queue_mech_id,
    DROP COLUMN mech_id;
ALTER TABLE battle_queue_notifications -- unsure why it wanted me to do a new alter table
    ALTER COLUMN chassis_id SET NOT NULL;
ALTER TABLE battle_queue_notifications -- unsure why it wanted me to do a new alter table
    RENAME COLUMN chassis_id TO mech_id;
ALTER TABLE battle_queue_notifications -- unsure why it wanted me to do a new alter table
    RENAME COLUMN queue_chassis_id TO queue_mech_id;


-- battle_queue_notifications_queue_mech_id_fkey
ALTER TABLE battle_queue
    DROP COLUMN mech_id;
ALTER TABLE battle_queue -- unsure why it wanted me to do a new alter table
    RENAME COLUMN chassis_id TO mech_id;



-- CREATE TABLE battle_mechs (
--     battle_id UUID NOT NULL references battles(id),
--     mech_id UUID NOT NULL references mechs(id),
--     owner_id UUID NOT NULL references players(id),
--     faction_id UUID NOT NULL references factions(id),
--     killed TIMESTAMPTZ NULL,
--     killed_by_id UUID NULL references mechs(id),

ALTER TABLE battle_mechs
    ADD COLUMN chassis_id           UUID REFERENCES chassis (id),
    ADD COLUMN killed_by_chassis_id UUID REFERENCES chassis (id);

UPDATE battle_mechs bm
SET chassis_id           = (SELECT c.id
                            FROM mechs m
                                     INNER JOIN chassis c ON m.chassis_id = c.id
                            WHERE m.id = bm.mech_id),
    killed_by_chassis_id = (SELECT c.id
                            FROM mechs m
                                     INNER JOIN chassis c ON m.chassis_id = c.id
                            WHERE m.id = bm.killed_by_id);

ALTER TABLE battle_mechs
    DROP CONSTRAINT battle_mechs_pkey,
    DROP COLUMN mech_id,
    DROP COLUMN killed_by_id;
ALTER TABLE battle_mechs -- unsure why it wanted me to do a new alter table
    RENAME COLUMN chassis_id TO mech_id;
ALTER TABLE battle_mechs -- unsure why it wanted me to do a new alter table
    RENAME COLUMN killed_by_chassis_id TO killed_by_id;
ALTER TABLE battle_mechs -- unsure why it wanted me to do a new alter table
    ADD PRIMARY KEY (battle_id, mech_id);



-- DROP TABLE mechs;

-- CREATE TABLE battle_wins (
--     battle_id UUID NOT NULL references battles(id),
--     mech_id UUID NOT NULL references mechs(id),


ALTER TABLE battle_wins
    ADD COLUMN chassis_id UUID REFERENCES chassis (id);

UPDATE battle_wins bw
SET chassis_id = (SELECT c.id
                  FROM mechs m
                           INNER JOIN chassis c ON m.chassis_id = c.id
                  WHERE m.id = bw.mech_id);

ALTER TABLE battle_wins
    DROP CONSTRAINT battle_wins_pkey,
    DROP COLUMN mech_id;
ALTER TABLE battle_wins -- unsure why it wanted me to do a new alter table
    RENAME COLUMN chassis_id TO mech_id;
ALTER TABLE battle_wins -- unsure why it wanted me to do a new alter table
    ADD PRIMARY KEY (battle_id, mech_id);


-- CREATE TABLE battle_kills (
--     battle_id UUID NOT NULL references battles(id),
--     mech_id UUID NOT NULL references mechs(id),
--     killed_id UUID NOT NULL references mechs(id),

ALTER TABLE battle_kills
    ADD COLUMN chassis_id        UUID REFERENCES chassis (id),
    ADD COLUMN killed_chassis_id UUID REFERENCES chassis (id);

UPDATE battle_kills bm
SET chassis_id        = (SELECT c.id
                         FROM mechs m
                                  INNER JOIN chassis c ON m.chassis_id = c.id
                         WHERE m.id = bm.mech_id),
    killed_chassis_id = (SELECT c.id
                         FROM mechs m
                                  INNER JOIN chassis c ON m.chassis_id = c.id
                         WHERE m.id = bm.killed_id);

ALTER TABLE battle_kills
    DROP CONSTRAINT battle_kills_pkey,
    DROP COLUMN mech_id,
    DROP COLUMN killed_id;
ALTER TABLE battle_kills -- unsure why it wanted me to do a new alter table
    RENAME COLUMN chassis_id TO mech_id;
ALTER TABLE battle_kills -- unsure why it wanted me to do a new alter table
    RENAME COLUMN killed_chassis_id TO killed_id;
ALTER TABLE battle_kills -- unsure why it wanted me to do a new alter table
    ADD PRIMARY KEY (battle_id, mech_id, killed_id);

-- CREATE TABLE battle_history (
--     id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
--     battle_id UUID NOT NULL references battles(id),
--     related_id UUID NULL references battle_history(id),
--     war_machine_one_id UUID NOT NULL references mechs(id),
--     war_machine_two_id UUID NULL references mechs(id),

ALTER TABLE battle_history
    ADD COLUMN war_machine_one_id_chassis UUID REFERENCES chassis (id),
    ADD COLUMN war_machine_two_id_chassis UUID REFERENCES chassis (id);

UPDATE battle_history bk
SET war_machine_one_id_chassis = (SELECT c.id
                                  FROM mechs m
                                           INNER JOIN chassis c ON m.chassis_id = c.id
                                  WHERE m.id = bk.war_machine_one_id),
    war_machine_two_id_chassis = (SELECT c.id
                                  FROM mechs m
                                           INNER JOIN chassis c ON m.chassis_id = c.id
                                  WHERE m.id = bk.war_machine_two_id);

ALTER TABLE battle_history
    DROP COLUMN war_machine_one_id,
    DROP COLUMN war_machine_two_id;
ALTER TABLE battle_history -- unsure why it wanted me to do a new alter table
    ALTER COLUMN war_machine_one_id_chassis SET NOT NULL;
ALTER TABLE battle_history -- unsure why it wanted me to do a new alter table
    RENAME COLUMN war_machine_one_id_chassis TO war_machine_one_id;
ALTER TABLE battle_history -- unsure why it wanted me to do a new alter table
    RENAME COLUMN war_machine_two_id_chassis TO war_machine_two_id;



-- CREATE TABLE mech_stats (
--     mech_id UUID PRIMARY KEY NOT NULL REFERENCES mechs (id),


ALTER TABLE mech_stats
    ADD COLUMN chassis_id UUID REFERENCES chassis (id);

UPDATE mech_stats ms
SET chassis_id = (SELECT c.id
                  FROM mechs m
                           INNER JOIN chassis c ON m.chassis_id = c.id
                  WHERE m.id = ms.mech_id);

ALTER TABLE mech_stats
    DROP CONSTRAINT mech_stats_pkey,
    DROP COLUMN mech_id;
ALTER TABLE mech_stats -- unsure why it wanted me to do a new alter table
    RENAME COLUMN chassis_id TO mech_id;
ALTER TABLE mech_stats -- unsure why it wanted me to do a new alter table
    ADD PRIMARY KEY (mech_id);

-- CREATE TABLE asset_repair(
--     id uuid primary key DEFAULT gen_random_uuid(),
--     mech_id UUID NOT NULL REFERENCES mechs (id),

ALTER TABLE asset_repair
    ADD COLUMN chassis_id UUID REFERENCES chassis (id);

UPDATE asset_repair ms
SET chassis_id = (SELECT c.id
                  FROM mechs m
                           INNER JOIN chassis c ON m.chassis_id = c.id
                  WHERE m.id = ms.mech_id);

ALTER TABLE asset_repair
    DROP COLUMN mech_id;
ALTER TABLE asset_repair -- unsure why it wanted me to do a new alter table
    ALTER COLUMN chassis_id SET NOT NULL;
ALTER TABLE asset_repair -- unsure why it wanted me to do a new alter table
    RENAME COLUMN chassis_id TO mech_id;

-- CREATE TABLE battle_contracts (
--     id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
--     mech_id UUID NOT NULL references mechs (id),

ALTER TABLE battle_contracts
    ADD COLUMN chassis_id UUID REFERENCES chassis (id);

UPDATE battle_contracts bc
SET chassis_id = (SELECT c.id
                  FROM mechs m
                           INNER JOIN chassis c ON m.chassis_id = c.id
                  WHERE m.id = bc.mech_id);

ALTER TABLE battle_contracts
    DROP CONSTRAINT bc_unique_mech_battle,
    DROP COLUMN mech_id;
ALTER TABLE battle_contracts -- unsure why it wanted me to do a new alter table
    ALTER COLUMN chassis_id SET NOT NULL;
ALTER TABLE battle_contracts -- unsure why it wanted me to do a new alter table
    RENAME COLUMN chassis_id TO mech_id;
ALTER TABLE battle_contracts
    ADD UNIQUE (mech_id, battle_id);


ALTER TABLE mechs
    RENAME TO mechs_old;

ALTER TABLE chassis
    RENAME TO mechs;

ALTER TABLE chassis_animation
    RENAME TO mech_animation;

ALTER TABLE chassis_skin
    RENAME TO mech_skin;

ALTER TABLE chassis_utility
    RENAME TO mech_utility;

-- ALTER TABLE blueprint_chassis
--     RENAME TO blueprint_mechs_old;

ALTER TABLE blueprint_chassis_animation
    RENAME TO blueprint_mech_animation;

ALTER TABLE chassis_weapons
    RENAME TO mech_weapons;


-- here we just update out seqs
SELECT SETVAL('collection_genesis', (SELECT COALESCE(MAX(genesis_token_id), 1) FROM mechs)::BIGINT, FALSE);
SELECT SETVAL('collection_limited_release', (SELECT COALESCE(MAX(limited_release_token_id), 1) FROM mechs)::BIGINT,
              FALSE);

-- clear up some columns
ALTER TABLE utility
    DROP COLUMN IF EXISTS owner_id,
    DROP COLUMN IF EXISTS tier;

ALTER TABLE weapons
    DROP COLUMN IF EXISTS owner_id,
    DROP COLUMN IF EXISTS tier;
