alter table factions
    add column if not exists logo_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS background_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';

-- red mountain
UPDATE
    factions
SET
    logo_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/factions/rm-logo.svg',
    background_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/factions/rm-bg.webp',
    description = 'Red Mountain is the leader in autonomous mining operations in the Supremacy Era. It controls territory on Mars, as well as secure city locations on the continent formerly known as Australia on Earth. In addition to the production of War Machines, Red Mountain has an economy built on mining, space transportation and energy production. Its AI platforms are directed by REDNET and supported by AIs including CorpDroids, Juggernauts and XJs. The main tiers of humans include Executives, Engineers and Mechanics.'
WHERE
    id = '98bf7bb3-1a7c-4f21-8843-458d62884060';

-- boston
UPDATE
    factions
SET
    logo_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/factions/bc-logo.svg',
    background_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/factions/bc-bg.webp',
    description = 'Boston Cybernetics is the major commercial leader within the Supremacy Era. It has secure territories comprising 275 districts located on the east coast of the former United States. In addition to the production of War Machines, its economy is built on finance, memory production and exploration of the Asteroid Belt between Jupiter and Mars. Boston Cybernetics AI platforms are directed by BOSSDAN and supported by Synths and Rexeon Guards. The three main tiers of humans include Patrons, CyRiders and Dwellers.'
WHERE
    id = '7c6dde21-b067-46cf-9e56-155c88a520e2';

-- Zaibatsu
UPDATE
    factions
SET
    logo_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/factions/zai-logo.svg',
    background_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/factions/zai-bg.webp',
    description = 'Zaibatsu is the industrial leader within the Supremacy Era, with heavily populated territory on the islands formerly known as Japan. In addition to the production of War Machines, Zaibatsu''s economy is built on production optimized by human and AI interaction, as well as the development of cloud cities. Its AI platforms are directed by ZAIA and supported by AIs including HANCRs, XHANCERs and Boostmen. The three main tiers of humans include APEXRs, KODRs and DENZRs.'
WHERE
    id = '880db344-e405-428d-84e5-6ebebab1fe6d';
