DROP TABLE IF EXISTS global_announcements;

CREATE TABLE global_announcements
(
    id             UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    title          TEXT             NOT NULL,
    message        TEXT             NOT NULL,
    show_from_battle_number         INT, 
    show_until_battle_number        INT
);
