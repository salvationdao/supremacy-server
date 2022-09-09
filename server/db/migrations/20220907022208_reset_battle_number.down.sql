UPDATE battles
SET battle_number = battle_number * (-1);

SELECT SETVAL('battles_battle_number_seq', (SELECT MAX(battle_number) + 1 FROM battles), FALSE)
