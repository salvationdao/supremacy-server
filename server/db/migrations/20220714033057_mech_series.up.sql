DROP TYPE IF EXISTS ITEM_SERIES;
CREATE TYPE ITEM_SERIES AS ENUM ('Genesis', 'Limited', 'Nexus');

ALTER TABLE blueprint_mechs 
	ADD COLUMN series ITEM_SERIES;

-- Label Genesis Mechs
-- SEE: IsCompleteGenesis() on server/mechs.go
UPDATE blueprint_mechs bm
SET series = 'Genesis'
FROM mechs m
WHERE bm.series IS NULL
	AND bm.id = m.blueprint_id
	AND m.genesis_token_id IS NOT NULL
	AND EXISTS (
		SELECT 1
		FROM mech_weapons _mw
			INNER JOIN weapons _w ON _w.id = _mw.weapon_id
				AND _w.genesis_token_id = m.genesis_token_id
		WHERE _mw.chassis_id = m.id
			AND _mw.slot_number = 0
		LIMIT 1
	)
	AND EXISTS (
		SELECT 1
		FROM mech_weapons _mw
			INNER JOIN weapons _w ON _w.id = _mw.weapon_id
				AND _w.genesis_token_id = m.genesis_token_id
		WHERE _mw.chassis_id = m.id
			AND _mw.slot_number = 1
		LIMIT 1
	)
	AND EXISTS (
		SELECT 1
		FROM mech_weapons _mw
			INNER JOIN weapons _w ON _w.id = _mw.weapon_id
				AND _w.genesis_token_id = m.genesis_token_id
		WHERE _mw.chassis_id = m.id
			AND _mw.slot_number = 2
		LIMIT 1
	)
;

-- Label Limited Mechs
-- SEE: IsCompleteLimited() on server/mechs.go
UPDATE blueprint_mechs bm
SET series = 'Limited'
FROM mechs m
WHERE bm.series IS NULL
	AND bm.id = m.blueprint_id
	AND m.limited_release_token_id IS NOT NULL
	AND EXISTS (
		SELECT 1
		FROM mech_weapons _mw
			INNER JOIN weapons _w ON _w.id = _mw.weapon_id
				AND _w.limited_release_token_id = m.limited_release_token_id
		WHERE _mw.chassis_id = m.id
			AND _mw.slot_number = 0
		LIMIT 1
	)
	AND EXISTS (
		SELECT 1
		FROM mech_weapons _mw
			INNER JOIN weapons _w ON _w.id = _mw.weapon_id
				AND _w.limited_release_token_id = m.limited_release_token_id
		WHERE _mw.chassis_id = m.id
			AND _mw.slot_number = 1
		LIMIT 1
	)
	AND EXISTS (
		SELECT 1
		FROM mech_weapons _mw
			INNER JOIN weapons _w ON _w.id = _mw.weapon_id
				AND _w.limited_release_token_id = m.limited_release_token_id
		WHERE _mw.chassis_id = m.id
			AND _mw.slot_number = 2
		LIMIT 1
	)
;
