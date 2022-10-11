DROP TYPE IF EXISTS ROLE_NAME;
CREATE TYPE ROLE_NAME AS ENUM ('PLAYER', 'MODERATOR', 'ADMIN');

CREATE TABLE roles
(
    id        UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    role_type ROLE_NAME        NOT NULL DEFAULT 'PLAYER'
);

INSERT INTO roles (id, role_type)
VALUES ('8dd55355-fc22-4d1d-a825-b973bb075259', 'PLAYER'),
       ('72b62032-7a9b-4743-9bd7-4840440d2503', 'MODERATOR'),
       ('7e8f0c1d-f36c-437c-bee2-c14fedb4df93', 'ADMIN');

ALTER TABLE players
    ADD COLUMN role_id UUID NOT NULL REFERENCES roles (id) DEFAULT '8dd55355-fc22-4d1d-a825-b973bb075259'; -- defaults to 'PLAYER' role