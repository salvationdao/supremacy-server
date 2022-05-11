-- Here we are refactoring how the template table works
-- Currently blueprints have joins to weapons, which they shouldn't. They shouldn't be aware of each other really, they are just the blueprint of a given item.
-- Templates need to hold pre packaged sets of blueprints, so for example we create a new template entry with its details and then a template_blueprint join table which holds various items

ALTER TABLE templates
    RENAME TO templates_old;

CREATE TABLE templates
(
    id                   UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    label                TEXT        NOT NULL UNIQUE,
    deleted_at           TIMESTAMPTZ,
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    blueprint_chassis_id UUID -- this is temp column we will remove, adding to make it easier to create tables
);

DROP TYPE IF EXISTS TEMPLATE_ITEM_TYPE;
CREATE TYPE TEMPLATE_ITEM_TYPE AS ENUM ('MECH', 'MECH_ANIMATION', 'MECH_SKIN', 'UTILITY','WEAPON', 'AMMO', 'POWER_CORE', 'WEAPON_SKIN', 'PLAYER_ABILITY');

CREATE TABLE template_blueprints
(
    id           UUID PRIMARY KEY            DEFAULT gen_random_uuid(),
    template_id  UUID               NOT NULL REFERENCES templates (id),
    type         TEMPLATE_ITEM_TYPE NOT NULL,
    blueprint_id UUID               NOT NULL,
    created_at   TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);

-- for each blueprint mech, create a template and add all its item to template_blueprints
WITH bm AS (SELECT bc.label AS label,
                   bc.id    AS blueprint_chassis_id
            FROM blueprint_chassis bc)
INSERT
INTO templates (label, blueprint_chassis_id)
SELECT label, blueprint_chassis_id
FROM bm;

-- for each blueprint mech utility join, create a template_blueprint
WITH bm_templates_utility AS (SELECT bc.label                  AS label,
                                     bc.id                     AS chassis_id,
                                     bcbu.blueprint_utility_id AS shield_id,
                                     t.id                      AS template_id
                              FROM blueprint_chassis bc
                                       INNER JOIN blueprint_chassis_blueprint_utility bcbu
                                                  ON bcbu.blueprint_chassis_id = bc.id
                                       INNER JOIN templates t ON bc.id = t.blueprint_chassis_id)
INSERT
INTO template_blueprints(template_id, type, blueprint_id)
SELECT bm_templates_utility.template_id, 'UTILITY', bm_templates_utility.shield_id
FROM bm_templates_utility;

-- for each blueprint mech weapon join, create a template_blueprint
WITH bm_templates_weapons AS (SELECT bc.label                 AS label,
                                     bc.id                    AS chassis_id,
                                     bcbw.blueprint_weapon_id AS weapon_id,
                                     t.id                     AS template_id
                              FROM blueprint_chassis bc
                                       INNER JOIN blueprint_chassis_blueprint_weapons bcbw
                                                  ON bcbw.blueprint_chassis_id = bc.id
                                       INNER JOIN templates t ON bc.id = t.blueprint_chassis_id)
INSERT
INTO template_blueprints(template_id, type, blueprint_id)
SELECT bm_templates_weapons.template_id, 'WEAPON', bm_templates_weapons.weapon_id
FROM bm_templates_weapons;


-- for each blueprint mech skin join, create a template_blueprint
WITH bm_templates_skins AS (SELECT bc.chassis_skin_id AS chassis_skin_id,
                                   t.id               AS template_id
                            FROM blueprint_chassis bc
                                     INNER JOIN templates t ON bc.id = t.blueprint_chassis_id)
INSERT
INTO template_blueprints(template_id, type, blueprint_id)
SELECT bm_templates_skins.template_id, 'MECH_SKIN', bm_templates_skins.chassis_skin_id
FROM bm_templates_skins;

-- for each blueprint mech, create a template_blueprint
WITH bm_templates_chassis AS (SELECT t.blueprint_chassis_id AS chassis_id,
                                     t.id                   AS template_id
                              FROM templates t)
INSERT
INTO template_blueprints(template_id, type, blueprint_id)
SELECT bm_templates_chassis.template_id, 'MECH', bm_templates_chassis.chassis_id
FROM bm_templates_chassis;

-- remove temp column
ALTER TABLE templates
    DROP COLUMN blueprint_chassis_id;

-- dont need this column anymore
ALTER TABLE blueprint_chassis
    DROP COLUMN chassis_skin_id;

DROP TABLE blueprint_chassis_blueprint_utility;
DROP TABLE blueprint_chassis_blueprint_weapons;
