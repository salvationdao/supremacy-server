BEGIN;

ALTER TYPE FEATURE_NAME
ADD
    VALUE 'CHAT_BAN';

COMMIT;

INSERT INTO
    features (name)
values
    ('CHAT_BAN');