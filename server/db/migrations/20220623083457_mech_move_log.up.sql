CREATE TABLE mech_move_command_logs(
    id uuid primary key default gen_random_uuid(),
    battle_id uuid not null references battles (id),
    mech_id uuid not null references mechs(id),
    triggered_by_id uuid not null references players(id),
    cell_x int not null,
    cell_y int not null,
    tx_id text not null,
    cancelled_at timestamptz,
    reached_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

CREATE INDEX mech_move_command_logs_mech_id ON mech_move_command_logs(mech_id);
CREATE INDEX mech_move_command_logs_record_search ON mech_move_command_logs(mech_id, battle_id, cancelled_at, reached_at);
CREATE INDEX mech_move_command_logs_created_at_descending ON mech_move_command_logs(created_at DESC);
