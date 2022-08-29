INSERT INTO kv (key, value)
VALUES ('battle_ability_location_select_duration', 20)
ON CONFLICT DO UPDATE SET value = excluded.value;
