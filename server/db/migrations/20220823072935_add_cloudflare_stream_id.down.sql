DROP TABLE battle_replays;
DROP TYPE IF EXISTS RECORDING_STATUS;

UPDATE oven_streams SET base_url='wss://stream2.supremacy.game:3334/app/staging1' WHERE name='Stream 1';