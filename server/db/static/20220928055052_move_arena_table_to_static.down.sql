-- dont drop battle arena
ALTER TABLE battle_arena ALTER COLUMN gid DROP NOT NULL;
ALTER TABLE battle_arena ADD COLUMN IF NOT EXISTS type TEXT NOT NULL DEFAULT 'STORY';