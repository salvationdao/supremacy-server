CREATE TABLE oven_streams
(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    base_url TEXT NOT NULL,
    available_resolutions TEXT[] NOT NULL,
    active BOOLEAN NOT NULL
); 

-- seed 
insert into oven_streams 
(name, base_url, available_resolutions, active)
values
('thing', ' wss://stream2.supremacy.game:3334/app/staging', '{1080, potato}', true);