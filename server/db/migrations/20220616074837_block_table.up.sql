CREATE TABLE block_marketplace
(
    id             uuid primary key not null default gen_random_uuid(),
    public_address text unique      not null,
    note           text,
    created_at     timestamptz      not null default now(),
    blocked_until  timestamptz      not null DEFAULT '22 February 2024 00:00:00 GMT+08:00'
);
