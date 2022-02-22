-- insert 'Sydney' stream
INSERT INTO stream_list (host , "name" , url , stream_id , region , resolution, bit_rates_k_bits , user_max , users_now , active , status , latitude , longitude )
VALUES (
'https://staging-watch.supremacy.game:5443', 
'Sydney',
'wss://staging-watch.supremacy.game:5443/WebRTCAppEE/websocket',
'524280586954581049507513',
'au-east',
'1920x1080',
5000,
1000,
100,
true,
'online', 
'-33.9031982421875',
'151.15179443359375'
);

