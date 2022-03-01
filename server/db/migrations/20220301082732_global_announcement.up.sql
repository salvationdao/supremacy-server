CREATE TABLE global_announcements
(
    id             UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    title          TEXT             NOT NULL,
    message        TEXT             NOT NULL,
    games_until                     INT NOT NULL, 
    show_until     TIMESTAMPTZ
);