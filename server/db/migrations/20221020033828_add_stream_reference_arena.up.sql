INSERT INTO oven_streams (id, name, base_url, available_resolutions, default_resolution, active) VALUES ('f985de0c-903d-44fc-884b-92eaa6520cdc', 'Stream 2', 'wss://stream2.supremacy.game:3334/app/production-60008739-348e-4cf0-8cca-663685e30142', '{240,360,480,720,1080}', 1080, true);

ALTER TABLE battle_arena
    ADD COLUMN oven_stream_id UUID REFERENCES oven_streams(id);

UPDATE battle_arena SET oven_stream_id=(SELECT os.id from oven_streams os WHERE os.id != 'f985de0c-903d-44fc-884b-92eaa6520cdc') WHERE id='95774a8a-6b9c-411c-a298-20824d0f00ba';
UPDATE battle_arena SET oven_stream_id='f985de0c-903d-44fc-884b-92eaa6520cdc' WHERE id='60008739-348e-4cf0-8cca-663685e30142';

ALTER TABLE battle_arena
    ALTER COLUMN oven_stream_id SET NOT NULL;

