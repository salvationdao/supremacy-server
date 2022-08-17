
DROP INDEX IF EXISTS chat_history_search;
DROP INDEX IF EXISTS chat_history_created_at_descending;

ALTER TABLE chat_history
    DROP COLUMN IF EXISTS arena_id;

DROP INDEX IF EXISTS
    mech_move_command_logs_mech_id,
    mech_move_command_logs_available,
    mech_move_command_logs_record_search;


create index mech_move_command_logs_mech_id ON gameserver.public.mech_move_command_logs (mech_id);
create index mech_move_command_logs_available ON gameserver.public.mech_move_command_logs (cancelled_at, reached_at, deleted_at);
create index mech_move_command_logs_record_search ON gameserver.public.mech_move_command_logs (mech_id, battle_id, cancelled_at, reached_at, deleted_at);

-- add arena id to battle table
ALTER TABLE battles
    DROP COLUMN IF EXISTS arena_id;

-- add arena id to mech move command table
ALTER TABLE mech_move_command_logs
    DROP COLUMN IF EXISTS arena_id;

DROP TABLE IF EXISTS battle_arena;

DROP TYPE IF EXISTS ARENA_TYPE_ENUM;

DROP SEQUENCE IF EXISTS battle_arena_gid_seq;
