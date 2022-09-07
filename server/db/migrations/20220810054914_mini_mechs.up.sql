ALTER TABLE
    mech_move_command_logs
ADD
    COLUMN is_moving bool NOT NULL DEFAULT false;
