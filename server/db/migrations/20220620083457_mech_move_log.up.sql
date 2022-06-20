CREATE TABLE mech_move_command_logs(
    id uuid primary key default gen_random_uuid(),
    mech_id uuid not null references mechs(id),
    triggered_by_id uuid not null references players(id),
    x int not null,
    y int not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

CREATE INDEX mech_move_command_logs_mech_id ON mech_move_command_logs(mech_id);
CREATE INDEX mech_move_command_logs_created_at_descending ON mech_move_command_logs(created_at DESC);
