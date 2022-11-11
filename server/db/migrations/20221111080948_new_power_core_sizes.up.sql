-- Update genesis mechs power cores to be of size TURBO
UPDATE
    power_cores
SET
    blueprint_id = ('6921c16c-44f0-46af-b61b-f8106e40530f')
WHERE
    id IN (
        SELECT
            power_core_id
        FROM
            mechs m
        WHERE
            m.blueprint_id IN ('5d3a973b-c62b-4438-b746-d3de2699d42a', 'ac27f3b9-753d-4ace-84a9-21c041195344', '02ba91b7-55dc-450a-9fbd-e7337ae97a2b'));

-- Update nexus platform mechs power cores to be of size TURBO
UPDATE
    power_cores
SET
    blueprint_id = ('5bd363a0-0a15-41a0-af3e-33770b67fd3a')
WHERE
    id IN (
        SELECT
            power_core_id
        FROM
            mechs m
        WHERE
            m.blueprint_id IN ('7068ab3e-89dc-4ac1-bcbb-1089096a5eda', 'df1ac803-0a90-4631-b9e0-b62a44bdadff', '0639ebde-fbba-498b-88ac-f7122ead9c90'));

-- Update all power cores to TURBO in mystery crates containing platform mechs
UPDATE
    mystery_crate_blueprints
SET
    blueprint_id = '5bd363a0-0a15-41a0-af3e-33770b67fd3a'
WHERE
    blueprint_type = 'POWER_CORE'
    AND mystery_crate_id IN (
        -- Get all mystery crates that contain platform mechs
        SELECT
            id FROM mystery_crate mc
            WHERE
                mc.id IN (
                    SELECT
                        mcb.mystery_crate_id
                    FROM mystery_crate_blueprints mcb
                    INNER JOIN blueprint_mechs bm ON bm.id = mcb.blueprint_id
                    WHERE
                        blueprint_type = 'MECH'
                        AND bm.mech_type = 'PLATFORM'));

