CREATE TABLE blobs (
    id              UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    file_name       TEXT             NOT NULL,
    mime_type       TEXT             NOT NULL,
    file_size_bytes BIGINT           NOT NULL,
    extension       TEXT             NOT NULL,
    file            BYTEA            NOT NULL,
    views           INTEGER          NOT NULL DEFAULT 0,
    hash            TEXT,
    deleted_at      TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

UPDATE game_abilities
SET image_url = '/api/blobs/dc713e47-4119-494a-a81b-8ac92cf3222b',
	description = 'Rain fury on the arena with a targeted airstrike.'
WHERE label = 'AIRSTRIKE';

UPDATE game_abilities
SET image_url = '/api/blobs/8e0e1918-556c-4370-85f9-b8960fd19554',
	description = 'The show-stopper. A tactical nuke at your fingertips.'
WHERE label = 'NUKE';

UPDATE game_abilities
SET image_url = '/api/blobs/5d0a0028-c074-4ab5-b46e-14d0ff07795d',
	description = 'Red Mountain unique ability. Call an additional Mech to the arena.'
WHERE label = 'REINFORCEMENTS';

UPDATE game_abilities
SET image_url = '/api/blobs/f40e90b7-1ea2-4a91-bf0f-feb052a019be',
	description = 'Support your Syndicate with a well-timed repair.'
WHERE label = 'REPAIR';

UPDATE game_abilities
SET image_url = '/api/blobs/3b4ae24a-7ccb-4d3b-8d88-905b406da0e1',
	description = 'Boston Cybernetic unique ability. Release the hounds!'
WHERE label = 'ROBOT DOGS';

UPDATE battle_abilities SET description = 'Rain fury on the arena with a targeted airstrike.' WHERE label = 'AIRSTRIKE';
UPDATE battle_abilities SET description = 'The show-stopper. A tactical nuke at your fingertips.' WHERE label = 'NUKE';
UPDATE battle_abilities SET description = 'Support your Syndcate with a well-timed repair.' WHERE label = 'REPAIR';

ALTER TABLE battle_abilities
	ALTER COLUMN description SET NOT NULL;
