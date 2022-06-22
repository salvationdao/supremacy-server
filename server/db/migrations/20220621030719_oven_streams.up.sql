ALTER TABLE stream_list ADD COLUMN service TEXT;

-- insert experimental stream
insert into stream_list
(host, "name", url, stream_id, region, resolution, bit_rates_k_bits, user_max, users_now, active, status, latitude, longitude, service )
values 
('https://stream2.supremacy.game:3334/app/stream2', 'Experimental 🌟', 'wss://stream2.supremacy.game:3334/app/stream2','Experimental', 'au', '1920x1080', 4000, 1000, 100, true, 'online', 1.3521000146865845, 103.8198013305664, 'OvenMediaEngine' );

-- update services on other streams to 'AntMedia'
UPDATE stream_list 
SET service = 'AntMedia'
WHERE stream_id != 'Experimental';
