INSERT INTO kv (key, value)
VALUES ('battle_ability_location_select_duration', 20)
ON CONFLICT (key) DO UPDATE SET value = excluded.value;

INSERT INTO kv (key, value)
VALUES ('battle_ability_bribe_duration', 20)
ON CONFLICT (key) DO UPDATE SET value = excluded.value;