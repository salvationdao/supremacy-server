ALTER TABLE mech_skin
    DROP COLUMN IF EXISTS image_url,
    DROP COLUMN IF EXISTS animation_url,
    DROP COLUMN IF EXISTS card_animation_url,
    DROP COLUMN IF EXISTS avatar_url,
    DROP COLUMN IF EXISTS large_image_url;

ALTER TABLE weapon_skin
    DROP COLUMN IF EXISTS weapon_type,
    DROP COLUMN IF EXISTS tier;

ALTER TABLE collection_items
    DROP COLUMN IF EXISTS image_url,
    DROP COLUMN IF EXISTS card_animation_url,
    DROP COLUMN IF EXISTS avatar_url,
    DROP COLUMN IF EXISTS large_image_url,
    DROP COLUMN IF EXISTS background_color,
    DROP COLUMN IF EXISTS animation_url,
    DROP COLUMN IF EXISTS youtube_url;


-- update currently owned mech skins to use the right mech skin blueprint id
----------------------------------------------- new id ----------------------------------------------------- old id ------------------
UPDATE mech_skin set blueprint_id = '72ffdb5c-c003-4734-b1d1-07898f5bcb39' where blueprint_id = 'f800d5e6-4654-46d5-957c-ecb3387cf588' OR  blueprint_id = '80df03b9-53ec-45cc-843d-2ed921973129'; -- this is gold skin
UPDATE mech_skin set blueprint_id = '11d6ec72-e75e-433a-9bdf-3d82f4ebec9a' where blueprint_id = '377060a2-ec4d-4e01-b5f9-a93657ad409f' OR  blueprint_id = 'ff1b7d82-40b1-4d1a-b93a-f5da92ec43ed'; -- this is ukarini skin
UPDATE mech_skin set blueprint_id = 'e9fe8e2d-031c-4d20-8928-4b32113108d2' where blueprint_id = '5a22a98f-d092-440e-a57d-3b13cefe3194';
UPDATE mech_skin set blueprint_id = '4aad6d66-9a2a-4f3c-9f14-0980842294c8' where blueprint_id = '20872580-713f-48af-8a08-b4f0b5e0887e';
UPDATE mech_skin set blueprint_id = 'b2448312-35ea-4592-8466-077e2138898e' where blueprint_id = 'c1c4ce93-04f2-4cde-b8f5-283544e9a32a';
UPDATE mech_skin set blueprint_id = 'c852467e-5ebd-4564-9955-4e542cd628f2' where blueprint_id = '1e434315-7a20-4ccb-b1aa-c180afde0778';
UPDATE mech_skin set blueprint_id = '243db24c-c946-4d78-bfe5-dfd6a1f0529a' where blueprint_id = '2456d608-60ad-4ee4-a1eb-50de615ddad4';
UPDATE mech_skin set blueprint_id = '2ac47dd7-4484-4c69-a7d2-aedbc539f504' where blueprint_id = 'd746665f-a863-457d-a38d-ae8b6008fb80';
UPDATE mech_skin set blueprint_id = '24e7f994-bd0c-4bfe-b836-0202d4fb0528' where blueprint_id = '965bb435-a9f7-46de-962b-4a8ead528085';
UPDATE mech_skin set blueprint_id = '08b93226-5a15-4f77-820d-fe9ef02ccfcc' where blueprint_id = 'e2e0972b-72b0-4ca0-842e-858b9cf4896d';
UPDATE mech_skin set blueprint_id = '26bee099-85a6-49cf-ba2d-f2d2f2ff42dc' where blueprint_id = 'bd35b891-be78-4c25-8be4-1d32aacfccf0';
UPDATE mech_skin set blueprint_id = '58b966f1-61c6-4222-9638-7f9f618e8f7b' where blueprint_id = 'f170f9e6-a97f-4527-a588-fbf332f7e687';
UPDATE mech_skin set blueprint_id = '4d39b18a-45fd-4490-a26b-1ac4ef97426b' where blueprint_id = 'fca191a3-41c2-45c8-b121-4ec78d1b8696';
UPDATE mech_skin set blueprint_id = 'aafcc80a-d716-4fd8-b1ef-cf7ef27b8fb8' where blueprint_id = '287372fe-41b6-47b2-acc5-83774345d4a5';
UPDATE mech_skin set blueprint_id = '00a8d4ea-1fd1-45c1-a0f4-1b117d28d0e1' where blueprint_id = 'c1e23aec-863f-4563-a31c-7d2a72806ac7';
UPDATE mech_skin set blueprint_id = 'bb12fcb6-1aba-4744-a5c2-247ed860e106' where blueprint_id = '6a357433-b071-4125-8110-20ace307f4ea';
UPDATE mech_skin set blueprint_id = 'd782fc51-a5e0-4b2e-81d9-5124269f5f0a' where blueprint_id = '372ab4d4-35d9-4f65-968a-7d506eef3d54';
UPDATE mech_skin set blueprint_id = 'd8a032bc-767a-427a-a204-1387579d3421' where blueprint_id = '7f9da0b4-560a-4839-8a40-216427e74d74';
UPDATE mech_skin set blueprint_id = 'ab0db88d-bd49-444d-90ad-a732c698583e' where blueprint_id = 'f6e54349-36ce-45d5-9712-8e3f005d5d27';
UPDATE mech_skin set blueprint_id = 'fefe62f3-519a-4cf3-acf6-06c89572c839' where blueprint_id = 'ad4f665f-2f62-4e93-9065-4c6d7af5163d';
UPDATE mech_skin set blueprint_id = 'd8f5c93d-4fb5-4ab3-a44a-2ffc298a53f7' where blueprint_id = 'be2a03ef-129d-49ba-ba71-9f7f10286905';
UPDATE mech_skin set blueprint_id = '3b527838-a2b7-4a93-876f-661862838b50' where blueprint_id = '48c90129-506f-4096-b82b-87954d354c95';
UPDATE mech_skin set blueprint_id = '982b4835-cb86-4f0e-b5a9-24f3de0ffcba' where blueprint_id = '729a33fd-6b87-4ea1-8a56-d415aa2ebcc6';
UPDATE mech_skin set blueprint_id = '0dd0559f-ac63-424d-b714-af32af65ef72' where blueprint_id = '31d3a161-532b-4edc-8aae-8a2d33a0ee43';


-- update mystery crates
---------------------------------------------------------------- new id ----------------------------------------------------- old id ------------------
UPDATE mystery_crate_blueprints set blueprint_id = 'e9fe8e2d-031c-4d20-8928-4b32113108d2' where blueprint_id = '5a22a98f-d092-440e-a57d-3b13cefe3194';
UPDATE mystery_crate_blueprints set blueprint_id = '4aad6d66-9a2a-4f3c-9f14-0980842294c8' where blueprint_id = '20872580-713f-48af-8a08-b4f0b5e0887e';
UPDATE mystery_crate_blueprints set blueprint_id = 'b2448312-35ea-4592-8466-077e2138898e' where blueprint_id = 'c1c4ce93-04f2-4cde-b8f5-283544e9a32a';
UPDATE mystery_crate_blueprints set blueprint_id = 'c852467e-5ebd-4564-9955-4e542cd628f2' where blueprint_id = '1e434315-7a20-4ccb-b1aa-c180afde0778';
UPDATE mystery_crate_blueprints set blueprint_id = '243db24c-c946-4d78-bfe5-dfd6a1f0529a' where blueprint_id = '2456d608-60ad-4ee4-a1eb-50de615ddad4';
UPDATE mystery_crate_blueprints set blueprint_id = '2ac47dd7-4484-4c69-a7d2-aedbc539f504' where blueprint_id = 'd746665f-a863-457d-a38d-ae8b6008fb80';
UPDATE mystery_crate_blueprints set blueprint_id = '24e7f994-bd0c-4bfe-b836-0202d4fb0528' where blueprint_id = '965bb435-a9f7-46de-962b-4a8ead528085';
UPDATE mystery_crate_blueprints set blueprint_id = '08b93226-5a15-4f77-820d-fe9ef02ccfcc' where blueprint_id = 'e2e0972b-72b0-4ca0-842e-858b9cf4896d';
UPDATE mystery_crate_blueprints set blueprint_id = '26bee099-85a6-49cf-ba2d-f2d2f2ff42dc' where blueprint_id = 'bd35b891-be78-4c25-8be4-1d32aacfccf0';
UPDATE mystery_crate_blueprints set blueprint_id = '58b966f1-61c6-4222-9638-7f9f618e8f7b' where blueprint_id = 'f170f9e6-a97f-4527-a588-fbf332f7e687';
UPDATE mystery_crate_blueprints set blueprint_id = '4d39b18a-45fd-4490-a26b-1ac4ef97426b' where blueprint_id = 'fca191a3-41c2-45c8-b121-4ec78d1b8696';
UPDATE mystery_crate_blueprints set blueprint_id = 'aafcc80a-d716-4fd8-b1ef-cf7ef27b8fb8' where blueprint_id = '287372fe-41b6-47b2-acc5-83774345d4a5';
UPDATE mystery_crate_blueprints set blueprint_id = '00a8d4ea-1fd1-45c1-a0f4-1b117d28d0e1' where blueprint_id = 'c1e23aec-863f-4563-a31c-7d2a72806ac7';
UPDATE mystery_crate_blueprints set blueprint_id = 'bb12fcb6-1aba-4744-a5c2-247ed860e106' where blueprint_id = '6a357433-b071-4125-8110-20ace307f4ea';
UPDATE mystery_crate_blueprints set blueprint_id = 'd782fc51-a5e0-4b2e-81d9-5124269f5f0a' where blueprint_id = '372ab4d4-35d9-4f65-968a-7d506eef3d54';
UPDATE mystery_crate_blueprints set blueprint_id = 'd8a032bc-767a-427a-a204-1387579d3421' where blueprint_id = '7f9da0b4-560a-4839-8a40-216427e74d74';
UPDATE mystery_crate_blueprints set blueprint_id = 'ab0db88d-bd49-444d-90ad-a732c698583e' where blueprint_id = 'f6e54349-36ce-45d5-9712-8e3f005d5d27';
UPDATE mystery_crate_blueprints set blueprint_id = 'fefe62f3-519a-4cf3-acf6-06c89572c839' where blueprint_id = 'ad4f665f-2f62-4e93-9065-4c6d7af5163d';
UPDATE mystery_crate_blueprints set blueprint_id = 'd8f5c93d-4fb5-4ab3-a44a-2ffc298a53f7' where blueprint_id = 'be2a03ef-129d-49ba-ba71-9f7f10286905';
UPDATE mystery_crate_blueprints set blueprint_id = '3b527838-a2b7-4a93-876f-661862838b50' where blueprint_id = '48c90129-506f-4096-b82b-87954d354c95';
UPDATE mystery_crate_blueprints set blueprint_id = '982b4835-cb86-4f0e-b5a9-24f3de0ffcba' where blueprint_id = '729a33fd-6b87-4ea1-8a56-d415aa2ebcc6';
UPDATE mystery_crate_blueprints set blueprint_id = '0dd0559f-ac63-424d-b714-af32af65ef72' where blueprint_id = '31d3a161-532b-4edc-8aae-8a2d33a0ee43';
