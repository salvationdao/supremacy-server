BEGIN;
ALTER TYPE FEATURE_NAME ADD VALUE 'PROFILE_AVATAR';
COMMIT;

INSERT INTO features (name) VALUES ('PROFILE_AVATAR');