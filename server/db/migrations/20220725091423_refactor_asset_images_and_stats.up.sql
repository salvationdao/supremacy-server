ALTER TABLE mechs
    DROP COLUMN IF EXISTS label,
    DROP COLUMN IF EXISTS brand_id,
    DROP COLUMN IF EXISTS model_id;

ALTER TABLE weapons
    DROP COLUMN IF EXISTS owner_id,
    DROP COLUMN IF EXISTS weapon_type,
    DROP COLUMN IF EXISTS label;

ALTER TABLE mech_skin
    DROP COLUMN IF EXISTS label,
    DROP COLUMN IF EXISTS mech_model,
    DROP COLUMN IF EXISTS image_url,
    DROP COLUMN IF EXISTS animation_url,
    DROP COLUMN IF EXISTS card_animation_url,
    DROP COLUMN IF EXISTS avatar_url,
    DROP COLUMN IF EXISTS large_image_url;

ALTER TABLE weapon_skin
    DROP COLUMN IF EXISTS label,
    DROP COLUMN IF EXISTS owner_id,
    DROP COLUMN IF EXISTS weapon_type,
    DROP COLUMN IF EXISTS weapon_model_id,
    DROP COLUMN IF EXISTS tier;

ALTER TABLE collection_items
    DROP COLUMN IF EXISTS image_url,
    DROP COLUMN IF EXISTS card_animation_url,
    DROP COLUMN IF EXISTS avatar_url,
    DROP COLUMN IF EXISTS large_image_url,
    DROP COLUMN IF EXISTS background_color,
    DROP COLUMN IF EXISTS animation_url,
    DROP COLUMN IF EXISTS youtube_url;


-- region -- update currently owned mech skins to use the right mech skin blueprint id
----------------------------------------------- new id ----------------------------------------------------- old id ------------------
UPDATE mech_skin SET blueprint_id = '72ffdb5c-c003-4734-b1d1-07898f5bcb39' WHERE blueprint_id = 'f800d5e6-4654-46d5-957c-ecb3387cf588' OR  blueprint_id = '80df03b9-53ec-45cc-843d-2ed921973129'; -- this is gold skin
UPDATE mech_skin SET blueprint_id = '11d6ec72-e75e-433a-9bdf-3d82f4ebec9a' WHERE blueprint_id = '377060a2-ec4d-4e01-b5f9-a93657ad409f' OR  blueprint_id = 'ff1b7d82-40b1-4d1a-b93a-f5da92ec43ed'; -- this is ukarini skin
UPDATE mech_skin SET blueprint_id = 'e9fe8e2d-031c-4d20-8928-4b32113108d2' WHERE blueprint_id = '5a22a98f-d092-440e-a57d-3b13cefe3194';
UPDATE mech_skin SET blueprint_id = '4aad6d66-9a2a-4f3c-9f14-0980842294c8' WHERE blueprint_id = '20872580-713f-48af-8a08-b4f0b5e0887e';
UPDATE mech_skin SET blueprint_id = 'b2448312-35ea-4592-8466-077e2138898e' WHERE blueprint_id = 'c1c4ce93-04f2-4cde-b8f5-283544e9a32a';
UPDATE mech_skin SET blueprint_id = 'c852467e-5ebd-4564-9955-4e542cd628f2' WHERE blueprint_id = '1e434315-7a20-4ccb-b1aa-c180afde0778';
UPDATE mech_skin SET blueprint_id = '243db24c-c946-4d78-bfe5-dfd6a1f0529a' WHERE blueprint_id = '2456d608-60ad-4ee4-a1eb-50de615ddad4';
UPDATE mech_skin SET blueprint_id = '2ac47dd7-4484-4c69-a7d2-aedbc539f504' WHERE blueprint_id = 'd746665f-a863-457d-a38d-ae8b6008fb80';
UPDATE mech_skin SET blueprint_id = '24e7f994-bd0c-4bfe-b836-0202d4fb0528' WHERE blueprint_id = '965bb435-a9f7-46de-962b-4a8ead528085';
UPDATE mech_skin SET blueprint_id = '08b93226-5a15-4f77-820d-fe9ef02ccfcc' WHERE blueprint_id = 'e2e0972b-72b0-4ca0-842e-858b9cf4896d';
UPDATE mech_skin SET blueprint_id = '26bee099-85a6-49cf-ba2d-f2d2f2ff42dc' WHERE blueprint_id = 'bd35b891-be78-4c25-8be4-1d32aacfccf0';
UPDATE mech_skin SET blueprint_id = '58b966f1-61c6-4222-9638-7f9f618e8f7b' WHERE blueprint_id = 'f170f9e6-a97f-4527-a588-fbf332f7e687';
UPDATE mech_skin SET blueprint_id = '4d39b18a-45fd-4490-a26b-1ac4ef97426b' WHERE blueprint_id = 'fca191a3-41c2-45c8-b121-4ec78d1b8696';
UPDATE mech_skin SET blueprint_id = 'aafcc80a-d716-4fd8-b1ef-cf7ef27b8fb8' WHERE blueprint_id = '287372fe-41b6-47b2-acc5-83774345d4a5';
UPDATE mech_skin SET blueprint_id = '00a8d4ea-1fd1-45c1-a0f4-1b117d28d0e1' WHERE blueprint_id = 'c1e23aec-863f-4563-a31c-7d2a72806ac7';
UPDATE mech_skin SET blueprint_id = 'bb12fcb6-1aba-4744-a5c2-247ed860e106' WHERE blueprint_id = '6a357433-b071-4125-8110-20ace307f4ea';
UPDATE mech_skin SET blueprint_id = 'd782fc51-a5e0-4b2e-81d9-5124269f5f0a' WHERE blueprint_id = '372ab4d4-35d9-4f65-968a-7d506eef3d54';
UPDATE mech_skin SET blueprint_id = 'd8a032bc-767a-427a-a204-1387579d3421' WHERE blueprint_id = '7f9da0b4-560a-4839-8a40-216427e74d74';
UPDATE mech_skin SET blueprint_id = 'ab0db88d-bd49-444d-90ad-a732c698583e' WHERE blueprint_id = 'f6e54349-36ce-45d5-9712-8e3f005d5d27';
UPDATE mech_skin SET blueprint_id = 'fefe62f3-519a-4cf3-acf6-06c89572c839' WHERE blueprint_id = 'ad4f665f-2f62-4e93-9065-4c6d7af5163d';
UPDATE mech_skin SET blueprint_id = 'd8f5c93d-4fb5-4ab3-a44a-2ffc298a53f7' WHERE blueprint_id = 'be2a03ef-129d-49ba-ba71-9f7f10286905';
UPDATE mech_skin SET blueprint_id = '3b527838-a2b7-4a93-876f-661862838b50' WHERE blueprint_id = '48c90129-506f-4096-b82b-87954d354c95';
UPDATE mech_skin SET blueprint_id = '982b4835-cb86-4f0e-b5a9-24f3de0ffcba' WHERE blueprint_id = '729a33fd-6b87-4ea1-8a56-d415aa2ebcc6';
UPDATE mech_skin SET blueprint_id = '0dd0559f-ac63-424d-b714-af32af65ef72' WHERE blueprint_id = '31d3a161-532b-4edc-8aae-8a2d33a0ee43';
-- endregion

-- region -- update mystery crates
---------------------------------------------------------------- new id ----------------------------------------------------- old id ------------------
UPDATE mystery_crate_blueprints SET blueprint_id = 'e9fe8e2d-031c-4d20-8928-4b32113108d2' WHERE blueprint_id = '5a22a98f-d092-440e-a57d-3b13cefe3194';
UPDATE mystery_crate_blueprints SET blueprint_id = '4aad6d66-9a2a-4f3c-9f14-0980842294c8' WHERE blueprint_id = '20872580-713f-48af-8a08-b4f0b5e0887e';
UPDATE mystery_crate_blueprints SET blueprint_id = 'b2448312-35ea-4592-8466-077e2138898e' WHERE blueprint_id = 'c1c4ce93-04f2-4cde-b8f5-283544e9a32a';
UPDATE mystery_crate_blueprints SET blueprint_id = 'c852467e-5ebd-4564-9955-4e542cd628f2' WHERE blueprint_id = '1e434315-7a20-4ccb-b1aa-c180afde0778';
UPDATE mystery_crate_blueprints SET blueprint_id = '243db24c-c946-4d78-bfe5-dfd6a1f0529a' WHERE blueprint_id = '2456d608-60ad-4ee4-a1eb-50de615ddad4';
UPDATE mystery_crate_blueprints SET blueprint_id = '2ac47dd7-4484-4c69-a7d2-aedbc539f504' WHERE blueprint_id = 'd746665f-a863-457d-a38d-ae8b6008fb80';
UPDATE mystery_crate_blueprints SET blueprint_id = '24e7f994-bd0c-4bfe-b836-0202d4fb0528' WHERE blueprint_id = '965bb435-a9f7-46de-962b-4a8ead528085';
UPDATE mystery_crate_blueprints SET blueprint_id = '08b93226-5a15-4f77-820d-fe9ef02ccfcc' WHERE blueprint_id = 'e2e0972b-72b0-4ca0-842e-858b9cf4896d';
UPDATE mystery_crate_blueprints SET blueprint_id = '26bee099-85a6-49cf-ba2d-f2d2f2ff42dc' WHERE blueprint_id = 'bd35b891-be78-4c25-8be4-1d32aacfccf0';
UPDATE mystery_crate_blueprints SET blueprint_id = '58b966f1-61c6-4222-9638-7f9f618e8f7b' WHERE blueprint_id = 'f170f9e6-a97f-4527-a588-fbf332f7e687';
UPDATE mystery_crate_blueprints SET blueprint_id = '4d39b18a-45fd-4490-a26b-1ac4ef97426b' WHERE blueprint_id = 'fca191a3-41c2-45c8-b121-4ec78d1b8696';
UPDATE mystery_crate_blueprints SET blueprint_id = 'aafcc80a-d716-4fd8-b1ef-cf7ef27b8fb8' WHERE blueprint_id = '287372fe-41b6-47b2-acc5-83774345d4a5';
UPDATE mystery_crate_blueprints SET blueprint_id = '00a8d4ea-1fd1-45c1-a0f4-1b117d28d0e1' WHERE blueprint_id = 'c1e23aec-863f-4563-a31c-7d2a72806ac7';
UPDATE mystery_crate_blueprints SET blueprint_id = 'bb12fcb6-1aba-4744-a5c2-247ed860e106' WHERE blueprint_id = '6a357433-b071-4125-8110-20ace307f4ea';
UPDATE mystery_crate_blueprints SET blueprint_id = 'd782fc51-a5e0-4b2e-81d9-5124269f5f0a' WHERE blueprint_id = '372ab4d4-35d9-4f65-968a-7d506eef3d54';
UPDATE mystery_crate_blueprints SET blueprint_id = 'd8a032bc-767a-427a-a204-1387579d3421' WHERE blueprint_id = '7f9da0b4-560a-4839-8a40-216427e74d74';
UPDATE mystery_crate_blueprints SET blueprint_id = 'ab0db88d-bd49-444d-90ad-a732c698583e' WHERE blueprint_id = 'f6e54349-36ce-45d5-9712-8e3f005d5d27';
UPDATE mystery_crate_blueprints SET blueprint_id = 'fefe62f3-519a-4cf3-acf6-06c89572c839' WHERE blueprint_id = 'ad4f665f-2f62-4e93-9065-4c6d7af5163d';
UPDATE mystery_crate_blueprints SET blueprint_id = 'd8f5c93d-4fb5-4ab3-a44a-2ffc298a53f7' WHERE blueprint_id = 'be2a03ef-129d-49ba-ba71-9f7f10286905';
UPDATE mystery_crate_blueprints SET blueprint_id = '3b527838-a2b7-4a93-876f-661862838b50' WHERE blueprint_id = '48c90129-506f-4096-b82b-87954d354c95';
UPDATE mystery_crate_blueprints SET blueprint_id = '982b4835-cb86-4f0e-b5a9-24f3de0ffcba' WHERE blueprint_id = '729a33fd-6b87-4ea1-8a56-d415aa2ebcc6';
UPDATE mystery_crate_blueprints SET blueprint_id = '0dd0559f-ac63-424d-b714-af32af65ef72' WHERE blueprint_id = '31d3a161-532b-4edc-8aae-8a2d33a0ee43';
-- endregion

-- region -- update templates
---------------------------------------------------------------- new id ----------------------------------------------------- old id ------------------
UPDATE template_blueprints SET blueprint_id = 'e9fe8e2d-031c-4d20-8928-4b32113108d2' WHERE blueprint_id = '5a22a98f-d092-440e-a57d-3b13cefe3194';
UPDATE template_blueprints SET blueprint_id = '4aad6d66-9a2a-4f3c-9f14-0980842294c8' WHERE blueprint_id = '20872580-713f-48af-8a08-b4f0b5e0887e';
UPDATE template_blueprints SET blueprint_id = 'b2448312-35ea-4592-8466-077e2138898e' WHERE blueprint_id = 'c1c4ce93-04f2-4cde-b8f5-283544e9a32a';
UPDATE template_blueprints SET blueprint_id = 'c852467e-5ebd-4564-9955-4e542cd628f2' WHERE blueprint_id = '1e434315-7a20-4ccb-b1aa-c180afde0778';
UPDATE template_blueprints SET blueprint_id = '243db24c-c946-4d78-bfe5-dfd6a1f0529a' WHERE blueprint_id = '2456d608-60ad-4ee4-a1eb-50de615ddad4';
UPDATE template_blueprints SET blueprint_id = '2ac47dd7-4484-4c69-a7d2-aedbc539f504' WHERE blueprint_id = 'd746665f-a863-457d-a38d-ae8b6008fb80';
UPDATE template_blueprints SET blueprint_id = '24e7f994-bd0c-4bfe-b836-0202d4fb0528' WHERE blueprint_id = '965bb435-a9f7-46de-962b-4a8ead528085';
UPDATE template_blueprints SET blueprint_id = '08b93226-5a15-4f77-820d-fe9ef02ccfcc' WHERE blueprint_id = 'e2e0972b-72b0-4ca0-842e-858b9cf4896d';
UPDATE template_blueprints SET blueprint_id = '26bee099-85a6-49cf-ba2d-f2d2f2ff42dc' WHERE blueprint_id = 'bd35b891-be78-4c25-8be4-1d32aacfccf0';
UPDATE template_blueprints SET blueprint_id = '58b966f1-61c6-4222-9638-7f9f618e8f7b' WHERE blueprint_id = 'f170f9e6-a97f-4527-a588-fbf332f7e687';
UPDATE template_blueprints SET blueprint_id = '4d39b18a-45fd-4490-a26b-1ac4ef97426b' WHERE blueprint_id = 'fca191a3-41c2-45c8-b121-4ec78d1b8696';
UPDATE template_blueprints SET blueprint_id = 'aafcc80a-d716-4fd8-b1ef-cf7ef27b8fb8' WHERE blueprint_id = '287372fe-41b6-47b2-acc5-83774345d4a5';
UPDATE template_blueprints SET blueprint_id = '00a8d4ea-1fd1-45c1-a0f4-1b117d28d0e1' WHERE blueprint_id = 'c1e23aec-863f-4563-a31c-7d2a72806ac7';
UPDATE template_blueprints SET blueprint_id = 'bb12fcb6-1aba-4744-a5c2-247ed860e106' WHERE blueprint_id = '6a357433-b071-4125-8110-20ace307f4ea';
UPDATE template_blueprints SET blueprint_id = 'd782fc51-a5e0-4b2e-81d9-5124269f5f0a' WHERE blueprint_id = '372ab4d4-35d9-4f65-968a-7d506eef3d54';
UPDATE template_blueprints SET blueprint_id = 'd8a032bc-767a-427a-a204-1387579d3421' WHERE blueprint_id = '7f9da0b4-560a-4839-8a40-216427e74d74';
UPDATE template_blueprints SET blueprint_id = 'ab0db88d-bd49-444d-90ad-a732c698583e' WHERE blueprint_id = 'f6e54349-36ce-45d5-9712-8e3f005d5d27';
UPDATE template_blueprints SET blueprint_id = 'fefe62f3-519a-4cf3-acf6-06c89572c839' WHERE blueprint_id = 'ad4f665f-2f62-4e93-9065-4c6d7af5163d';
UPDATE template_blueprints SET blueprint_id = 'd8f5c93d-4fb5-4ab3-a44a-2ffc298a53f7' WHERE blueprint_id = 'be2a03ef-129d-49ba-ba71-9f7f10286905';
UPDATE template_blueprints SET blueprint_id = '3b527838-a2b7-4a93-876f-661862838b50' WHERE blueprint_id = '48c90129-506f-4096-b82b-87954d354c95';
UPDATE template_blueprints SET blueprint_id = '982b4835-cb86-4f0e-b5a9-24f3de0ffcba' WHERE blueprint_id = '729a33fd-6b87-4ea1-8a56-d415aa2ebcc6';
UPDATE template_blueprints SET blueprint_id = '0dd0559f-ac63-424d-b714-af32af65ef72' WHERE blueprint_id = '31d3a161-532b-4edc-8aae-8a2d33a0ee43';
-- endregion

-- region -- update currently owned weapon skins to use the right weapon skin blueprint id
UPDATE weapon_skin SET blueprint_id = 'f6b4a628-78e5-47d1-9f75-c95d70b3472e'
WHERE blueprint_id IN (
                        '6bee4231-85d3-4b0d-91d3-dba759d8fb4f',
                        'a40d9c2c-812f-45f9-a059-f52dc7980982',
                        '22c26dc4-2107-4a0b-ab9b-c8da99e9abfb',
                        '561960ca-bb1f-4349-ad1e-33484be8dec0',
                        '2f2cbac4-8c40-43ef-ba9c-d91f91cd355a',
                        '9845bf88-f9b2-4c07-ac62-54accc1d4bc4',
                        '77cc363f-ac45-4545-af4a-3c9dfb704789',
                        '2e20cb8a-b40a-4b64-b21c-16fa94a9392e'
    );
UPDATE weapon_skin SET blueprint_id = 'e9886f77-1ed9-4aed-8f5a-abd3e6275464'
WHERE blueprint_id IN (
                       'a078b4d4-2e3d-49e4-86d2-d2b6a839534a',
                       '9bbcda1f-7c1e-4ea7-8de7-fa507bb0b39f',
                       'cdcde923-430e-4b24-9916-56b48150bbed',
                       '40740724-d891-469f-b74e-54e8b5685a19',
                       'b1b78eed-7371-4ad9-a3ea-6ef86d61e278',
                       '1cfe0387-33f3-4d80-843e-1998eb498875',
                       'aa67c963-9a4d-4d0d-bd01-d8075eaac56a',
                       '05c8c140-a073-4f7c-96a6-297ae99c86a9'
    );
UPDATE weapon_skin SET blueprint_id = 'f1ebf586-7ce1-4e5e-b6ba-113e96c4accd'
WHERE blueprint_id IN (
                       '4ad520a2-e033-4ca1-ba65-0ebbc3513697',
                       '97246188-09cd-4858-bbbb-49967c55cd38',
                       '7d5d8f82-27b0-4ef2-b44e-7476027a0371',
                       'd8e49315-62f8-47f4-9a8c-92e9ea4ebfa9',
                       'fe2b900f-5e56-4ce5-bf6c-31b498eb26af',
                       'e0427ba7-a4a4-46fc-820c-22d2f693d8eb',
                       '158189c0-9ac6-4323-8008-71ff210e2287',
                       '8f2bf410-3cda-459e-ad14-5e1271faf926'
    );
UPDATE weapon_skin SET blueprint_id = '8c944ba8-bf7a-46b3-831f-641017430608'
WHERE blueprint_id IN (
                       '187c9bfb-d2e8-4af9-ac5a-9f8e86ee24ed',
                       '3bccd240-0378-4d13-a23e-b263be784178',
                       '8a9b48ed-48ef-44da-88a1-5f8e7988505d',
                       '61cf7d24-12dc-4993-88c3-037fc199da58',
                       'f83db0be-47a5-49d6-bd27-0315e3b0fd48',
                       'b38a5ab4-83e1-4183-ba81-c2834d4a6815',
                       '07b84501-67e3-42dc-92e1-e333f4504ae9',
                       'f90d8ee3-ac8c-46f2-85e8-9f522ce9b552'
    );
UPDATE weapon_skin SET blueprint_id = '8c990899-35a3-4eb3-b710-79186b735fa4'
WHERE blueprint_id IN (
                       'e8b2be8f-04e8-48ee-a4cd-abe78fb717ca',
                       'eccff0f5-784d-4e87-a29d-b5a110b3af04',
                       '5c40580e-4ce6-4245-bc1d-a82e3c6f6d00',
                       '9303203d-1446-4a7a-83ca-f33f6b556534',
                       '7aaf7074-8acc-4594-a1ef-0d3b3252ebfa',
                       '3ffee0cc-b2bd-4044-9f0f-da79cc720acc',
                       '7ec4c431-1bce-42b0-b1a2-bfe4447f3ef8',
                       'c7e2c330-7241-46ae-bb8b-8fbe45295905'
    );
UPDATE weapon_skin SET blueprint_id = 'f5036def-a979-4d03-89de-0281be618d57'
WHERE blueprint_id IN (
                       '3a5f326d-e6cb-4e81-84aa-10633e07cff7',
                       'c0a4ae40-6dfa-486f-92b3-c377d07a7be6',
                       '589156dd-47fc-4d1b-a745-64d719993368',
                       '1770e130-d005-4ca8-846d-c3a4286ea55b',
                       '6c6b46dc-021a-4ba0-9d06-1b41e2524da4',
                       'afc1793f-da3f-42d2-a45f-b8524c92d420',
                       '38c91bc6-4bf6-4ac8-b631-63a8c059da76',
                       'd3c39a8e-85fa-499f-b346-d3aa459bcd13'
    );
UPDATE weapon_skin SET blueprint_id = '3f7076cc-0cc0-4d2a-a79f-2462d2e02c20'
WHERE blueprint_id IN (
                       '53d0c32b-66f7-4635-8a23-dab98d60348b',
                       '4bffd3ef-6337-481b-8daa-4f519b82374e',
                       'd003120e-fd38-4d9c-af4a-eb54637e4bf5',
                       'e9c5b973-b8d8-4ee9-a47d-92a601c69126',
                       'a9b8516c-9be4-4ec8-9e8b-0ce155362ba6',
                       '54b5ec13-6468-4765-a97c-7063d0fda2f3',
                       '272635b1-8899-4904-9912-ebb14e0237a0',
                       '762a84f2-fc77-4522-b441-0d204274aa1c'
    );
UPDATE weapon_skin SET blueprint_id = 'ab8a5813-53cd-40f9-96db-65f3ee3f977d'
WHERE blueprint_id IN (
                       '971db084-ef1a-472d-96cb-c28a15264d0d',
                       '7875e080-6f37-429d-9c6c-ef0c97c6408a',
                       '5ddc9895-439f-48ad-a0cf-a5b841322c8e',
                       '1b42c687-57b8-4652-a4b0-d3abda6f473a',
                       '9b9b92cf-d07d-464f-941e-e934d9dc6997',
                       '26d276d2-0c91-4774-a10e-c4e358644ca3',
                       'f976202d-347a-49bb-b669-091705d1cb99',
                       '740cb444-68e1-4bf0-bdfd-f5b0883bcdf4'
    );
UPDATE weapon_skin SET blueprint_id = '3dab2057-359d-422b-a0b3-b9714096fc00'
WHERE blueprint_id IN (
                       'a4d1500a-e879-4057-8b19-bbd1594a7c5f',
                       '4538e24e-65ef-4da2-a47a-30b675ffc714',
                       'e52ab94c-f710-4957-ab4b-cdfd61b0a8e5',
                       '463edf3b-72ed-4b89-9f8b-3068b2769d3c',
                       'dcce1019-867e-48a5-ad48-fcaa1980efc9',
                       'c12b704a-02da-4791-9d67-5c75a2213c5a',
                       '53a24b70-22db-4793-bc03-8a988782df19',
                       '27047b50-8c4e-453c-83f4-a42f7e7bc601'
    );
UPDATE weapon_skin SET blueprint_id = '11faa37d-5c33-43ee-84ac-f01ba8ef8875'
WHERE blueprint_id IN (
                       'e62003dd-69aa-4c41-8a97-01d7f806015a',
                       '3fe2c440-aaf6-421f-9194-0991dd5120cd',
                       'fbba705b-775a-454d-9307-f558e677acc9',
                       'fa6eb8cb-0822-42a2-8b4c-edc0147cf18d',
                       '5599cfa8-9a16-47f7-aa5e-1581c362df8b',
                       '4d192e46-ecf6-4514-9ab1-e67d71003513',
                       'fdfe3a17-ec10-4cb6-a094-e7bd828f5e76',
                       '3b80c589-9eed-4082-baad-e7c8f237fa27'
    );
UPDATE weapon_skin SET blueprint_id = '96c51cff-c2bb-49ae-a186-8477fd850d84'
WHERE blueprint_id IN (
                       'cd6151b8-bb2a-46a8-a965-a74c22578f9c',
                       '7df80cf3-7524-406a-ad18-4e32688c6f6c',
                       '92e7fdb6-1eee-4ba0-8d65-c58e81972578',
                       '7a4479cf-196f-4192-af6b-bf2792e23fc5',
                       'f63f5bd8-7520-401a-b362-dea2c44b830f',
                       '6ceb0b4a-73fb-4c07-949f-9ee3c67cec70',
                       '8a69404c-8436-4c2c-b70b-5bead79a76cb',
                       '1a50e776-32e9-4861-b2f4-d4417a36fa9d'
    );
UPDATE weapon_skin SET blueprint_id = 'f053db4f-cac9-4445-aa60-f75e691f4fe6'
WHERE blueprint_id IN (
                       '3f6b0357-f644-430c-a7df-5489537c046d',
                       'b992d386-b6aa-4f14-9bd1-37cff98bd98b',
                       '2b7f9c5d-1014-4f08-afdd-fdd8433e0dc3',
                       '74cb10fd-f958-4475-abbf-b026fc820785',
                       'e746a916-bcda-480f-9f89-2071b818c5df',
                       '590f895a-6bc9-46a4-933b-08c8f4ec6ffc',
                       '3de8313f-87df-4b18-b286-fcc5dff81fd5',
                       '587cac53-82af-4404-b8b7-7541797a8b9b'
    );
UPDATE weapon_skin SET blueprint_id = 'fb1a375e-4ce6-4e1f-b906-629198584c00'
WHERE blueprint_id IN (
                       '98c45848-677e-4a08-b890-b23ba09040fb',
                       'fdd79139-83aa-4657-83ba-db0e296c9ee0',
                       '0df33b12-ae2c-4529-9a8e-00d5da04eb53',
                       'd11f6c43-5707-4161-aea1-77d5dab23bff',
                       'a420f0b5-3286-4788-a71a-9782b54f6008',
                       '3e16f30e-a9ef-409c-807b-76f4dc556732',
                       'dd1e5ca9-3f18-4bdf-827b-ce3f90842374',
                       'c194a857-b558-4344-84aa-3f37d87a537d'
    );
UPDATE weapon_skin SET blueprint_id = '7dc7a664-9bb1-49ba-ab3d-ed7a300bacfe'
WHERE blueprint_id IN (
                       'a5d0e25d-edfa-4f23-8d3a-e9d16e227048',
                       '53e3ce09-6a84-47bb-bb31-2e8cfd966cb7',
                       '8d3e1223-3527-406b-b310-aaa9bbd0c939',
                       '6592b54d-9ed4-4e2b-a368-4a84d5bf3987',
                       '1bf33888-cefd-42f4-b89b-3253a2b53e56',
                       '9d9cf7dc-e210-404c-b1e9-922b4e849514',
                       '09090725-a7fd-49ea-9a5c-6203344de741',
                       '13700be7-26fd-4e73-9c64-d3649a5a1f3a'
    );
UPDATE weapon_skin SET blueprint_id = 'f643bd16-e6df-4a1c-a2d6-283717843973'
WHERE blueprint_id IN (
                       '1109a7c8-03b4-4658-92ca-bd1c540b805f',
                       'ce8be4c0-c7b3-4d17-b3db-641c2a7faffa',
                       'eddc42ea-294e-40ba-8927-e28d5fd896bb',
                       'c5cc3968-25e1-443e-af7f-cb78f7754222',
                       'c5a1c6c0-ba9b-4509-9c26-37faad6a6986',
                       '2bfb4c29-5e1c-44f8-9867-0deedea59142',
                       'b74a7f4b-469b-4233-951e-cadea4376cd2',
                       '6689d421-a3b1-4b68-a438-562879aa9c01'
    );
UPDATE weapon_skin SET blueprint_id = '889b7106-7a66-41b2-8d15-1e2960f2620b'
WHERE blueprint_id IN (
                       '9844b794-b802-4658-b778-3c1300998ef1',
                       '74727c69-55a4-4ecc-a90a-03eacfab7581',
                       '1ea1551c-b66a-46e9-b3ce-da0fb22c5bc5',
                       'b27f5dd7-393a-4e0a-a9d1-07520227ffa6',
                       '7f088002-5250-4447-a8f1-d39c3ca2ce0d',
                       '5193defd-40c5-4b13-83c2-03199677a0fb',
                       'f52f1ccf-59ba-46d4-9ed0-7ff39558d00b',
                       'c51a094b-3e5e-4b43-bb57-af0f82b38666'
    );
UPDATE weapon_skin SET blueprint_id = '19895ef4-8951-4ed2-85a9-ef02d1c1efa0'
WHERE blueprint_id IN (
                       '6b61b82c-b5d9-4589-8303-57f418b0bb9d',
                       '0db51c94-95e7-4fa3-8f09-9ba9f60c2964',
                       '5f531492-82bc-434e-ab88-a1a88e002771',
                       '51528eb1-dd7a-41a4-a72a-e9490266407b',
                       'fa58696e-e410-40b7-a48c-16be320f9211',
                       '8afc0964-1573-4adb-aeff-2185d71c939d',
                       'a93f8b44-e149-4933-8f5a-4053c386580e'
    );
UPDATE weapon_skin SET blueprint_id = '78a3b8a6-f4d6-497d-b621-87893950fd2b'
WHERE blueprint_id IN (
                       'e36ecf20-40e5-4ac0-ba52-757bf3372619',
                       'ae3722db-f5d8-410b-81b1-afcf601bd171',
                       '78349dbc-8197-476c-b84d-5d261034928e',
                       '772d0087-5409-4383-8c24-98d8f55e35bc',
                       'cb467e60-7305-4ba1-94a6-83229936bfaa',
                       'd79619ba-f897-49c5-ab40-0e879e9accaf',
                       'd1383c38-2047-46cc-b530-94e47d197cf1',
                       'b59e3e9c-77d3-4a63-be63-c4598e43eec1'
    );
UPDATE weapon_skin SET blueprint_id = '23aabcd6-9e40-44b0-a0f1-b2cece19a54c'
WHERE blueprint_id IN (
                       '72852593-84b5-4883-b3ab-1496de44e7be',
                       '9d125cf1-5123-4bc8-a598-b3d0bc3890be',
                       'c1012df4-f8f2-4e04-8fe0-021f0d0cdc4f',
                       'f045ab10-f03b-4b1d-890b-9a1a3e973371',
                       '9c3b93ac-e1c5-45c9-bc72-ad49ec57c3a9',
                       '149fd8d0-459b-4516-9056-f8ffd9779d60',
                       'c0385acf-73f2-441f-b6ad-e1f622f4dd53',
                       '01259820-acd4-481a-9413-06aede2c4926'
);
UPDATE weapon_skin SET blueprint_id = '4fa842e9-9261-4153-a695-142c57155248'
WHERE blueprint_id IN (
                       '8e5dd2c1-b52e-48cd-8539-aed12feae5d8',
                       '6e369636-5ac3-4217-9074-0b1350d1ec6d',
                       'fffdd01f-5b14-432a-8f7e-6f86d557be8f',
                       '05b337b1-ce3b-4fcf-a3d2-b94aa76628a9',
                       'c6a8df72-896c-458f-b679-8162396ab2e0',
                       '6e897de9-f333-4a15-aeda-86c6eecc8a4e',
                       '244c979d-87de-4af7-a3a7-196a3d4c815a',
                       'f654b52b-ebea-4edf-abc9-b9431d7c299b'
);
UPDATE weapon_skin SET blueprint_id = 'd2a5d45a-eca3-49c0-8c66-bda97626cd2d'
WHERE blueprint_id IN (
                       '98bc7460-2f0b-479b-befe-de663cc3a8f0',
                       '4e0e7f14-b237-484f-98bc-993ccbc3e887',
                       '4d939a57-ed63-47bf-a941-d654faf3b343',
                       '86acd9ff-f51f-433b-b58b-ce947d0e66e9',
                       'c0ced55a-1cc7-42f7-8749-9f616dd59f56',
                       'c4935402-4b1d-401f-996e-77876a97761c',
                       'e6769598-b69e-448b-abeb-20222208bb34',
                       'a91491fa-4bd0-49b1-8aff-fe35d1a3c20f'
);
UPDATE weapon_skin SET blueprint_id = '3029890f-5336-464b-ba0e-d701eb20a360'
WHERE blueprint_id IN (
                       'c5cf853c-758a-4169-ba05-866548b86a59',
                       'd4912e43-62c1-4891-a2f5-4cd1d0d0f6b3',
                       '5c766a22-7058-4456-abe6-97dcc189fd34',
                       '9f4255de-9d4b-4ffd-9ad3-bcfc16175497',
                       '1761bc66-5199-46be-a4e8-06c6d3e8566f',
                       '98899681-ba45-4fa6-a8c1-00f279531f1d',
                       '693adb11-2b9d-4921-ba04-93eb7aa744e3',
                       '7ff51d7a-8753-458c-953a-8b382fc98e51'
);
UPDATE weapon_skin SET blueprint_id = '9c20c27c-e0ed-406a-ad02-0ab05ba10364'
WHERE blueprint_id IN (
                       'c4ec30b1-3301-4bf6-ad94-a91c306df5ca',
                       '867daad5-1ce3-4ae0-93e2-6f9d0f465435',
                       '583fe88f-66e0-442a-96be-e2877fd251ca',
                       '1e38c064-08ff-4bac-a280-8325bab4762f',
                       'beb80df7-1344-41f7-934e-56272a1d2217',
                       '71b85a94-b35d-464a-8e37-0eed8f2e4c22',
                       '5b038029-6684-410b-8438-d0da29f929cb',
                       '66285093-fb4d-4bcc-a7b9-15b5c1fdb821'
);
UPDATE weapon_skin SET blueprint_id = '1eb3a17c-d157-4bb1-a6dc-7cd3fe7ab7b9'
WHERE blueprint_id IN (
                       '3dbe2090-064f-4430-bcdf-12d017689181',
                       '61d2184b-b4fd-4623-80e2-78ddf8fa8ebf',
                       '8303fa1c-ddb8-446a-8e0e-4d31709565fc',
                       '86e1200e-04af-4a4b-8a99-0844a69d2fa4',
                       'f6627fe6-10ef-4b37-8304-defffef98e6e',
                       'd4f65a17-991c-41e9-ac6e-9c28375660bc',
                       '53f57c4f-e8b5-45c7-b1c6-0ccce0e52ba1',
                       'b252c620-20c6-4465-a683-82956d41fec1'
);
UPDATE weapon_skin SET blueprint_id = '06e1894d-3872-49cd-a4cf-0c42ee4cedf0'
WHERE blueprint_id IN (
                       '555503c1-3086-42cf-a967-01e40b521897',
                       '4f7dff5f-a6e6-4d21-b638-f025b203ea9d',
                       '720e3668-2bce-4670-a0e9-7838e2929c89',
                       '9491926d-c234-4ebb-b96e-e7a8e50791fb',
                       'e44570ed-aa12-46a2-af26-a16d8692059b',
                       'ef7f0181-f36f-4622-9993-746ac90d50b7',
                       '6b4a8dfb-77dd-45ab-b2c3-bfcf01f54ce0',
                       '0d367f31-b3e2-4b53-960d-5200fe53bb22'
);
UPDATE weapon_skin SET blueprint_id = 'c609283d-82bd-4a04-8e80-450ec5c64a56'
WHERE blueprint_id IN (
                       '8142a60a-ecb4-4f99-9a34-9936478370e4',
                       'bd9db924-a289-4039-84bf-2b9a9e236327',
                       '53192802-57e4-4fba-80d7-d5f0785b30b0',
                       '34083108-555d-490d-b4df-8dc3ee63c749',
                       'c6d3de5c-eeb6-4686-8a8a-f4ee69a6d86e',
                       '0cc56beb-b5bb-4a47-92bc-3a386ba0b9f1',
                       '6e72ff2d-4a23-4145-b8d9-8decbc9967b2',
                       'fbf3074c-e516-459d-af56-08149ff20615'
);
UPDATE weapon_skin SET blueprint_id = 'fb11780c-c6f7-43a8-a5fe-62e77add9e9d'
WHERE blueprint_id IN (
                       '2b09d985-085b-440e-a0d1-752b1f5e0a2b',
                       'bf235fa4-df21-45e6-ac6d-c6981ea1c699',
                       '8de1f3d6-c41f-42cb-9c7c-f1eb895030e3',
                       '6e6eb9f9-a547-48cf-8681-5166e22785fd',
                       'dc3985e2-7689-4332-936c-7cd0e5a9f10c',
                       '1a388404-8811-47b8-8c67-f5c21fb2cbed',
                       '09ecb2e8-3783-4639-ae5d-8e584d346f48',
                       '393c003a-d23d-4f82-a93b-fbaa8a9d9dfe'
);
UPDATE weapon_skin SET blueprint_id = '3df394d1-3f01-45fb-a2c7-1bdecbbc90a6'
WHERE blueprint_id IN (
                       '492e22c6-ba3e-47d6-955e-53e362ea0f20',
                       '447af743-9b46-447a-aa5f-a683b5f90b28',
                       '33b32552-9639-4eef-97f5-0b570739976f',
                       '770bd969-6d3f-4236-94a7-325eaa4e998b',
                       'a481c00c-4c3e-47b3-ab8e-948da3150dcc',
                       'b0261840-d31a-491d-9591-9167f0f2720c',
                       '4e458fb4-01e8-4770-9574-45ea996acb95',
                       '2fafd818-95ad-4eaf-a744-c9a09f82535b'
);
UPDATE weapon_skin SET blueprint_id = '5c0b5d6b-1300-4727-af15-9c68c377554f'
WHERE blueprint_id IN (
                       '205b3a4e-e192-430c-b83f-7182c5e5aea8',
                       'f85c95e9-35f1-4e9d-8a6a-56b86af8344d',
                       '2fc63f40-1c70-4177-b0c3-fc59ac1a25a4',
                       '1cc07d02-c23c-4af4-a570-34478bea8185',
                       '043c84e9-de04-463f-9214-9c750d64bcf3',
                       '65b44826-d66a-456a-850b-348686bba714',
                       '805ab241-7f31-4673-b4cd-5cb26d4f440c',
                       'f7f9a1a9-1b5d-4207-a90b-89811ed58351'
);
UPDATE weapon_skin SET blueprint_id = '21eccaa7-fa81-4aab-837e-f17ce627b399'
WHERE blueprint_id IN (
                       'df77a278-0daf-46ad-8433-79e3efd63fff',
                       '22e9bf6b-e7da-4551-83fd-8c1583eebdf7',
                       'acc19db8-ea3e-4159-bdd5-397a95fb0662',
                       '79f76bea-efee-4046-8f79-e0b3cd29bd4e',
                       '5039d66d-a516-4ed7-8f92-cb4715a23285',
                       '858f7dac-8b53-47ca-96e5-d5ed2ce4e4d7',
                       '54d6ea9f-c896-4463-8def-ccc958f925c4',
                       'd61026b5-941c-4fa4-9b08-3ddfbdf1d131'
);
UPDATE weapon_skin SET blueprint_id = 'c49ddbb9-ee57-4209-9956-7d605717f19f'
WHERE blueprint_id IN (
                       '2a5a359f-4159-4862-8ca6-16aa51f30e3d',
                       '8dec706b-e2a3-47bb-919b-bcd06c4c67bf',
                       'e58526c4-a7b6-4908-9550-3e1034cd93a6',
                       '4b3b9774-6d5c-4885-b494-637473a7ba89',
                       '0df47dba-bd6a-49cd-951f-9bb773e085b9',
                       '8370fe8c-6fa8-48df-a56e-0b716955b80f',
                       'caf52fae-438e-42d2-b616-36d74f375134',
                       '636437a4-8bba-4f92-9901-4cf9837e1cb3'
);
UPDATE weapon_skin SET blueprint_id = '03693e90-ece4-40f2-bb22-35dc32ed02ea'
WHERE blueprint_id IN (
                       'ea105045-cc90-4191-94f1-303d548fe7fb',
                       'eaccf9e5-cb47-435b-bbf7-9660deec5db0',
                       '9845213f-7d1d-4c8d-8729-df7acebf506d',
                       '0d584cc3-93c1-4d84-9549-e302e27249bb',
                       '991d2d86-1659-47a2-861c-ac6250fb7752',
                       '897aaf4d-2a6c-4ac8-882f-65b531c42d63',
                       '72ed0bb9-8aca-4746-9637-fe1a661483e6',
                       'a6944388-1cf1-4b42-b532-58e20307e75b'
);
UPDATE weapon_skin SET blueprint_id = '00fb66a4-c081-46cc-aa62-0bcc05f416e4'
WHERE blueprint_id IN (
                       '9ae03bae-d22c-419d-8968-cbb138d47754',
                       '6e246144-44a7-4ca0-95a7-0e5a43f2566b',
                       'b302967e-aa6c-4aec-90ef-7c09a9db0e6b',
                       '874db321-38e4-4aa9-aedc-b36feb1e31b5',
                       '706e2592-a959-41b7-bdd7-21e60c3651fb',
                       '2d37aa18-4405-491d-9dec-e5a2defa6a89',
                       '8fde9903-578e-401d-a9e2-7cca9f11a9bf',
                       '37bd7ab2-e045-4256-8928-e7759d61e62c'
);
UPDATE weapon_skin SET blueprint_id = 'abac4eb4-95e5-4399-947f-54612024aefb'
WHERE blueprint_id IN (
                       'caebda72-7514-45da-96b4-9f0dc15e808a',
                       'e0ad7e13-f951-4603-87d6-3107c16b8fef',
                       '7515a5a8-1820-4744-a83c-5c143d9b7e63',
                       'c5469e5e-3a56-46f9-94cc-2b04e4e263fb',
                       '7ac0f5c2-c9c7-48e1-9bc0-5ec9682e17ba',
                       'b43a5d0c-fbe2-4d67-830f-7b650e031676',
                       '10dbe0d7-51a4-4f68-9a62-95af6637a979',
                       '17f32e78-729b-4dad-a644-eb7781f8616b'
);
UPDATE weapon_skin SET blueprint_id = 'c8cee447-30df-4c66-a779-7ed61098d768'
WHERE blueprint_id IN (
                       '3c6e4453-da0f-4caa-b71c-3069c294a4ad',
                       'fe8eb52c-c4b1-40c4-a6e1-df31aa56c7f3',
                       '599a4247-a95e-4144-88d0-78b590e536f9',
                       '2ad7ee86-0c9f-43af-9f4b-b6c75ff8de73',
                       '831cf381-16bf-4ad9-b418-68b24ef19fe7',
                       'c2d4ac29-1504-4d51-883b-4299f89db8cd',
                       'e2033c92-0f21-4909-b581-807fd90dc683',
                       'c47882d6-decc-42c1-afbc-5113ca42a9fc'
);
UPDATE weapon_skin SET blueprint_id = '65ff58c7-ec00-4695-81cc-8c04c167e6e6'
WHERE blueprint_id IN (
                       '3be26aa4-e7e6-4a7f-9d2e-b0a9f04329f6',
                       'bd603bc8-5fa6-4c3c-95b0-178ae1ddfb83',
                       '89c2487e-d816-463a-9c6e-e57bd53a17ce',
                       '904ae783-6b44-4d37-ac95-dcf7fa7370a6',
                       'e2857853-2fa4-428d-a484-e5ee8c8e0a67',
                       '8bec27c3-21a2-49aa-815e-e839c889627a',
                       '8eb2a7ec-450a-49f9-a695-b977c383eab9',
                       'b32ffd15-6376-44a3-8b31-472b70e7df3f'
);
-- endregion

-- region -- update mystery crates ------
UPDATE mystery_crate_blueprints SET blueprint_id = 'f6b4a628-78e5-47d1-9f75-c95d70b3472e'
WHERE blueprint_id IN (
                       '6bee4231-85d3-4b0d-91d3-dba759d8fb4f',
                       'a40d9c2c-812f-45f9-a059-f52dc7980982',
                       '22c26dc4-2107-4a0b-ab9b-c8da99e9abfb',
                       '561960ca-bb1f-4349-ad1e-33484be8dec0',
                       '2f2cbac4-8c40-43ef-ba9c-d91f91cd355a',
                       '9845bf88-f9b2-4c07-ac62-54accc1d4bc4',
                       '77cc363f-ac45-4545-af4a-3c9dfb704789',
                       '2e20cb8a-b40a-4b64-b21c-16fa94a9392e'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = 'e9886f77-1ed9-4aed-8f5a-abd3e6275464'
WHERE blueprint_id IN (
                       'a078b4d4-2e3d-49e4-86d2-d2b6a839534a',
                       '9bbcda1f-7c1e-4ea7-8de7-fa507bb0b39f',
                       'cdcde923-430e-4b24-9916-56b48150bbed',
                       '40740724-d891-469f-b74e-54e8b5685a19',
                       'b1b78eed-7371-4ad9-a3ea-6ef86d61e278',
                       '1cfe0387-33f3-4d80-843e-1998eb498875',
                       'aa67c963-9a4d-4d0d-bd01-d8075eaac56a',
                       '05c8c140-a073-4f7c-96a6-297ae99c86a9'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = 'f1ebf586-7ce1-4e5e-b6ba-113e96c4accd'
WHERE blueprint_id IN (
                       '4ad520a2-e033-4ca1-ba65-0ebbc3513697',
                       '97246188-09cd-4858-bbbb-49967c55cd38',
                       '7d5d8f82-27b0-4ef2-b44e-7476027a0371',
                       'd8e49315-62f8-47f4-9a8c-92e9ea4ebfa9',
                       'fe2b900f-5e56-4ce5-bf6c-31b498eb26af',
                       'e0427ba7-a4a4-46fc-820c-22d2f693d8eb',
                       '158189c0-9ac6-4323-8008-71ff210e2287',
                       '8f2bf410-3cda-459e-ad14-5e1271faf926'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '8c944ba8-bf7a-46b3-831f-641017430608'
WHERE blueprint_id IN (
                       '187c9bfb-d2e8-4af9-ac5a-9f8e86ee24ed',
                       '3bccd240-0378-4d13-a23e-b263be784178',
                       '8a9b48ed-48ef-44da-88a1-5f8e7988505d',
                       '61cf7d24-12dc-4993-88c3-037fc199da58',
                       'f83db0be-47a5-49d6-bd27-0315e3b0fd48',
                       'b38a5ab4-83e1-4183-ba81-c2834d4a6815',
                       '07b84501-67e3-42dc-92e1-e333f4504ae9',
                       'f90d8ee3-ac8c-46f2-85e8-9f522ce9b552'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '8c990899-35a3-4eb3-b710-79186b735fa4'
WHERE blueprint_id IN (
                       'e8b2be8f-04e8-48ee-a4cd-abe78fb717ca',
                       'eccff0f5-784d-4e87-a29d-b5a110b3af04',
                       '5c40580e-4ce6-4245-bc1d-a82e3c6f6d00',
                       '9303203d-1446-4a7a-83ca-f33f6b556534',
                       '7aaf7074-8acc-4594-a1ef-0d3b3252ebfa',
                       '3ffee0cc-b2bd-4044-9f0f-da79cc720acc',
                       '7ec4c431-1bce-42b0-b1a2-bfe4447f3ef8',
                       'c7e2c330-7241-46ae-bb8b-8fbe45295905'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = 'f5036def-a979-4d03-89de-0281be618d57'
WHERE blueprint_id IN (
                       '3a5f326d-e6cb-4e81-84aa-10633e07cff7',
                       'c0a4ae40-6dfa-486f-92b3-c377d07a7be6',
                       '589156dd-47fc-4d1b-a745-64d719993368',
                       '1770e130-d005-4ca8-846d-c3a4286ea55b',
                       '6c6b46dc-021a-4ba0-9d06-1b41e2524da4',
                       'afc1793f-da3f-42d2-a45f-b8524c92d420',
                       '38c91bc6-4bf6-4ac8-b631-63a8c059da76',
                       'd3c39a8e-85fa-499f-b346-d3aa459bcd13'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '3f7076cc-0cc0-4d2a-a79f-2462d2e02c20'
WHERE blueprint_id IN (
                       '53d0c32b-66f7-4635-8a23-dab98d60348b',
                       '4bffd3ef-6337-481b-8daa-4f519b82374e',
                       'd003120e-fd38-4d9c-af4a-eb54637e4bf5',
                       'e9c5b973-b8d8-4ee9-a47d-92a601c69126',
                       'a9b8516c-9be4-4ec8-9e8b-0ce155362ba6',
                       '54b5ec13-6468-4765-a97c-7063d0fda2f3',
                       '272635b1-8899-4904-9912-ebb14e0237a0',
                       '762a84f2-fc77-4522-b441-0d204274aa1c'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = 'ab8a5813-53cd-40f9-96db-65f3ee3f977d'
WHERE blueprint_id IN (
                       '971db084-ef1a-472d-96cb-c28a15264d0d',
                       '7875e080-6f37-429d-9c6c-ef0c97c6408a',
                       '5ddc9895-439f-48ad-a0cf-a5b841322c8e',
                       '1b42c687-57b8-4652-a4b0-d3abda6f473a',
                       '9b9b92cf-d07d-464f-941e-e934d9dc6997',
                       '26d276d2-0c91-4774-a10e-c4e358644ca3',
                       'f976202d-347a-49bb-b669-091705d1cb99',
                       '740cb444-68e1-4bf0-bdfd-f5b0883bcdf4'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '3dab2057-359d-422b-a0b3-b9714096fc00'
WHERE blueprint_id IN (
                       'a4d1500a-e879-4057-8b19-bbd1594a7c5f',
                       '4538e24e-65ef-4da2-a47a-30b675ffc714',
                       'e52ab94c-f710-4957-ab4b-cdfd61b0a8e5',
                       '463edf3b-72ed-4b89-9f8b-3068b2769d3c',
                       'dcce1019-867e-48a5-ad48-fcaa1980efc9',
                       'c12b704a-02da-4791-9d67-5c75a2213c5a',
                       '53a24b70-22db-4793-bc03-8a988782df19',
                       '27047b50-8c4e-453c-83f4-a42f7e7bc601'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '11faa37d-5c33-43ee-84ac-f01ba8ef8875'
WHERE blueprint_id IN (
                       'e62003dd-69aa-4c41-8a97-01d7f806015a',
                       '3fe2c440-aaf6-421f-9194-0991dd5120cd',
                       'fbba705b-775a-454d-9307-f558e677acc9',
                       'fa6eb8cb-0822-42a2-8b4c-edc0147cf18d',
                       '5599cfa8-9a16-47f7-aa5e-1581c362df8b',
                       '4d192e46-ecf6-4514-9ab1-e67d71003513',
                       'fdfe3a17-ec10-4cb6-a094-e7bd828f5e76',
                       '3b80c589-9eed-4082-baad-e7c8f237fa27'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '96c51cff-c2bb-49ae-a186-8477fd850d84'
WHERE blueprint_id IN (
                       'cd6151b8-bb2a-46a8-a965-a74c22578f9c',
                       '7df80cf3-7524-406a-ad18-4e32688c6f6c',
                       '92e7fdb6-1eee-4ba0-8d65-c58e81972578',
                       '7a4479cf-196f-4192-af6b-bf2792e23fc5',
                       'f63f5bd8-7520-401a-b362-dea2c44b830f',
                       '6ceb0b4a-73fb-4c07-949f-9ee3c67cec70',
                       '8a69404c-8436-4c2c-b70b-5bead79a76cb',
                       '1a50e776-32e9-4861-b2f4-d4417a36fa9d'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = 'f053db4f-cac9-4445-aa60-f75e691f4fe6'
WHERE blueprint_id IN (
                       '3f6b0357-f644-430c-a7df-5489537c046d',
                       'b992d386-b6aa-4f14-9bd1-37cff98bd98b',
                       '2b7f9c5d-1014-4f08-afdd-fdd8433e0dc3',
                       '74cb10fd-f958-4475-abbf-b026fc820785',
                       'e746a916-bcda-480f-9f89-2071b818c5df',
                       '590f895a-6bc9-46a4-933b-08c8f4ec6ffc',
                       '3de8313f-87df-4b18-b286-fcc5dff81fd5',
                       '587cac53-82af-4404-b8b7-7541797a8b9b'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = 'fb1a375e-4ce6-4e1f-b906-629198584c00'
WHERE blueprint_id IN (
                       '98c45848-677e-4a08-b890-b23ba09040fb',
                       'fdd79139-83aa-4657-83ba-db0e296c9ee0',
                       '0df33b12-ae2c-4529-9a8e-00d5da04eb53',
                       'd11f6c43-5707-4161-aea1-77d5dab23bff',
                       'a420f0b5-3286-4788-a71a-9782b54f6008',
                       '3e16f30e-a9ef-409c-807b-76f4dc556732',
                       'dd1e5ca9-3f18-4bdf-827b-ce3f90842374',
                       'c194a857-b558-4344-84aa-3f37d87a537d'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '7dc7a664-9bb1-49ba-ab3d-ed7a300bacfe'
WHERE blueprint_id IN (
                       'a5d0e25d-edfa-4f23-8d3a-e9d16e227048',
                       '53e3ce09-6a84-47bb-bb31-2e8cfd966cb7',
                       '8d3e1223-3527-406b-b310-aaa9bbd0c939',
                       '6592b54d-9ed4-4e2b-a368-4a84d5bf3987',
                       '1bf33888-cefd-42f4-b89b-3253a2b53e56',
                       '9d9cf7dc-e210-404c-b1e9-922b4e849514',
                       '09090725-a7fd-49ea-9a5c-6203344de741',
                       '13700be7-26fd-4e73-9c64-d3649a5a1f3a'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = 'f643bd16-e6df-4a1c-a2d6-283717843973'
WHERE blueprint_id IN (
                       '1109a7c8-03b4-4658-92ca-bd1c540b805f',
                       'ce8be4c0-c7b3-4d17-b3db-641c2a7faffa',
                       'eddc42ea-294e-40ba-8927-e28d5fd896bb',
                       'c5cc3968-25e1-443e-af7f-cb78f7754222',
                       'c5a1c6c0-ba9b-4509-9c26-37faad6a6986',
                       '2bfb4c29-5e1c-44f8-9867-0deedea59142',
                       'b74a7f4b-469b-4233-951e-cadea4376cd2',
                       '6689d421-a3b1-4b68-a438-562879aa9c01'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '889b7106-7a66-41b2-8d15-1e2960f2620b'
WHERE blueprint_id IN (
                       '9844b794-b802-4658-b778-3c1300998ef1',
                       '74727c69-55a4-4ecc-a90a-03eacfab7581',
                       '1ea1551c-b66a-46e9-b3ce-da0fb22c5bc5',
                       'b27f5dd7-393a-4e0a-a9d1-07520227ffa6',
                       '7f088002-5250-4447-a8f1-d39c3ca2ce0d',
                       '5193defd-40c5-4b13-83c2-03199677a0fb',
                       'f52f1ccf-59ba-46d4-9ed0-7ff39558d00b',
                       'c51a094b-3e5e-4b43-bb57-af0f82b38666'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '19895ef4-8951-4ed2-85a9-ef02d1c1efa0'
WHERE blueprint_id IN (
                       '6b61b82c-b5d9-4589-8303-57f418b0bb9d',
                       '0db51c94-95e7-4fa3-8f09-9ba9f60c2964',
                       '5f531492-82bc-434e-ab88-a1a88e002771',
                       '51528eb1-dd7a-41a4-a72a-e9490266407b',
                       'fa58696e-e410-40b7-a48c-16be320f9211',
                       '8afc0964-1573-4adb-aeff-2185d71c939d',
                       'a93f8b44-e149-4933-8f5a-4053c386580e'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '78a3b8a6-f4d6-497d-b621-87893950fd2b'
WHERE blueprint_id IN (
                       'e36ecf20-40e5-4ac0-ba52-757bf3372619',
                       'ae3722db-f5d8-410b-81b1-afcf601bd171',
                       '78349dbc-8197-476c-b84d-5d261034928e',
                       '772d0087-5409-4383-8c24-98d8f55e35bc',
                       'cb467e60-7305-4ba1-94a6-83229936bfaa',
                       'd79619ba-f897-49c5-ab40-0e879e9accaf',
                       'd1383c38-2047-46cc-b530-94e47d197cf1',
                       'b59e3e9c-77d3-4a63-be63-c4598e43eec1'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '23aabcd6-9e40-44b0-a0f1-b2cece19a54c'
WHERE blueprint_id IN (
                       '72852593-84b5-4883-b3ab-1496de44e7be',
                       '9d125cf1-5123-4bc8-a598-b3d0bc3890be',
                       'c1012df4-f8f2-4e04-8fe0-021f0d0cdc4f',
                       'f045ab10-f03b-4b1d-890b-9a1a3e973371',
                       '9c3b93ac-e1c5-45c9-bc72-ad49ec57c3a9',
                       '149fd8d0-459b-4516-9056-f8ffd9779d60',
                       'c0385acf-73f2-441f-b6ad-e1f622f4dd53',
                       '01259820-acd4-481a-9413-06aede2c4926'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '4fa842e9-9261-4153-a695-142c57155248'
WHERE blueprint_id IN (
                       '8e5dd2c1-b52e-48cd-8539-aed12feae5d8',
                       '6e369636-5ac3-4217-9074-0b1350d1ec6d',
                       'fffdd01f-5b14-432a-8f7e-6f86d557be8f',
                       '05b337b1-ce3b-4fcf-a3d2-b94aa76628a9',
                       'c6a8df72-896c-458f-b679-8162396ab2e0',
                       '6e897de9-f333-4a15-aeda-86c6eecc8a4e',
                       '244c979d-87de-4af7-a3a7-196a3d4c815a',
                       'f654b52b-ebea-4edf-abc9-b9431d7c299b'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = 'd2a5d45a-eca3-49c0-8c66-bda97626cd2d'
WHERE blueprint_id IN (
                       '98bc7460-2f0b-479b-befe-de663cc3a8f0',
                       '4e0e7f14-b237-484f-98bc-993ccbc3e887',
                       '4d939a57-ed63-47bf-a941-d654faf3b343',
                       '86acd9ff-f51f-433b-b58b-ce947d0e66e9',
                       'c0ced55a-1cc7-42f7-8749-9f616dd59f56',
                       'c4935402-4b1d-401f-996e-77876a97761c',
                       'e6769598-b69e-448b-abeb-20222208bb34',
                       'a91491fa-4bd0-49b1-8aff-fe35d1a3c20f'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '3029890f-5336-464b-ba0e-d701eb20a360'
WHERE blueprint_id IN (
                       'c5cf853c-758a-4169-ba05-866548b86a59',
                       'd4912e43-62c1-4891-a2f5-4cd1d0d0f6b3',
                       '5c766a22-7058-4456-abe6-97dcc189fd34',
                       '9f4255de-9d4b-4ffd-9ad3-bcfc16175497',
                       '1761bc66-5199-46be-a4e8-06c6d3e8566f',
                       '98899681-ba45-4fa6-a8c1-00f279531f1d',
                       '693adb11-2b9d-4921-ba04-93eb7aa744e3',
                       '7ff51d7a-8753-458c-953a-8b382fc98e51'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '9c20c27c-e0ed-406a-ad02-0ab05ba10364'
WHERE blueprint_id IN (
                       'c4ec30b1-3301-4bf6-ad94-a91c306df5ca',
                       '867daad5-1ce3-4ae0-93e2-6f9d0f465435',
                       '583fe88f-66e0-442a-96be-e2877fd251ca',
                       '1e38c064-08ff-4bac-a280-8325bab4762f',
                       'beb80df7-1344-41f7-934e-56272a1d2217',
                       '71b85a94-b35d-464a-8e37-0eed8f2e4c22',
                       '5b038029-6684-410b-8438-d0da29f929cb',
                       '66285093-fb4d-4bcc-a7b9-15b5c1fdb821'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '1eb3a17c-d157-4bb1-a6dc-7cd3fe7ab7b9'
WHERE blueprint_id IN (
                       '3dbe2090-064f-4430-bcdf-12d017689181',
                       '61d2184b-b4fd-4623-80e2-78ddf8fa8ebf',
                       '8303fa1c-ddb8-446a-8e0e-4d31709565fc',
                       '86e1200e-04af-4a4b-8a99-0844a69d2fa4',
                       'f6627fe6-10ef-4b37-8304-defffef98e6e',
                       'd4f65a17-991c-41e9-ac6e-9c28375660bc',
                       '53f57c4f-e8b5-45c7-b1c6-0ccce0e52ba1',
                       'b252c620-20c6-4465-a683-82956d41fec1'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '06e1894d-3872-49cd-a4cf-0c42ee4cedf0'
WHERE blueprint_id IN (
                       '555503c1-3086-42cf-a967-01e40b521897',
                       '4f7dff5f-a6e6-4d21-b638-f025b203ea9d',
                       '720e3668-2bce-4670-a0e9-7838e2929c89',
                       '9491926d-c234-4ebb-b96e-e7a8e50791fb',
                       'e44570ed-aa12-46a2-af26-a16d8692059b',
                       'ef7f0181-f36f-4622-9993-746ac90d50b7',
                       '6b4a8dfb-77dd-45ab-b2c3-bfcf01f54ce0',
                       '0d367f31-b3e2-4b53-960d-5200fe53bb22'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = 'c609283d-82bd-4a04-8e80-450ec5c64a56'
WHERE blueprint_id IN (
                       '8142a60a-ecb4-4f99-9a34-9936478370e4',
                       'bd9db924-a289-4039-84bf-2b9a9e236327',
                       '53192802-57e4-4fba-80d7-d5f0785b30b0',
                       '34083108-555d-490d-b4df-8dc3ee63c749',
                       'c6d3de5c-eeb6-4686-8a8a-f4ee69a6d86e',
                       '0cc56beb-b5bb-4a47-92bc-3a386ba0b9f1',
                       '6e72ff2d-4a23-4145-b8d9-8decbc9967b2',
                       'fbf3074c-e516-459d-af56-08149ff20615'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = 'fb11780c-c6f7-43a8-a5fe-62e77add9e9d'
WHERE blueprint_id IN (
                       '2b09d985-085b-440e-a0d1-752b1f5e0a2b',
                       'bf235fa4-df21-45e6-ac6d-c6981ea1c699',
                       '8de1f3d6-c41f-42cb-9c7c-f1eb895030e3',
                       '6e6eb9f9-a547-48cf-8681-5166e22785fd',
                       'dc3985e2-7689-4332-936c-7cd0e5a9f10c',
                       '1a388404-8811-47b8-8c67-f5c21fb2cbed',
                       '09ecb2e8-3783-4639-ae5d-8e584d346f48',
                       '393c003a-d23d-4f82-a93b-fbaa8a9d9dfe'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '3df394d1-3f01-45fb-a2c7-1bdecbbc90a6'
WHERE blueprint_id IN (
                       '492e22c6-ba3e-47d6-955e-53e362ea0f20',
                       '447af743-9b46-447a-aa5f-a683b5f90b28',
                       '33b32552-9639-4eef-97f5-0b570739976f',
                       '770bd969-6d3f-4236-94a7-325eaa4e998b',
                       'a481c00c-4c3e-47b3-ab8e-948da3150dcc',
                       'b0261840-d31a-491d-9591-9167f0f2720c',
                       '4e458fb4-01e8-4770-9574-45ea996acb95',
                       '2fafd818-95ad-4eaf-a744-c9a09f82535b'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '5c0b5d6b-1300-4727-af15-9c68c377554f'
WHERE blueprint_id IN (
                       '205b3a4e-e192-430c-b83f-7182c5e5aea8',
                       'f85c95e9-35f1-4e9d-8a6a-56b86af8344d',
                       '2fc63f40-1c70-4177-b0c3-fc59ac1a25a4',
                       '1cc07d02-c23c-4af4-a570-34478bea8185',
                       '043c84e9-de04-463f-9214-9c750d64bcf3',
                       '65b44826-d66a-456a-850b-348686bba714',
                       '805ab241-7f31-4673-b4cd-5cb26d4f440c',
                       'f7f9a1a9-1b5d-4207-a90b-89811ed58351'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '21eccaa7-fa81-4aab-837e-f17ce627b399'
WHERE blueprint_id IN (
                       'df77a278-0daf-46ad-8433-79e3efd63fff',
                       '22e9bf6b-e7da-4551-83fd-8c1583eebdf7',
                       'acc19db8-ea3e-4159-bdd5-397a95fb0662',
                       '79f76bea-efee-4046-8f79-e0b3cd29bd4e',
                       '5039d66d-a516-4ed7-8f92-cb4715a23285',
                       '858f7dac-8b53-47ca-96e5-d5ed2ce4e4d7',
                       '54d6ea9f-c896-4463-8def-ccc958f925c4',
                       'd61026b5-941c-4fa4-9b08-3ddfbdf1d131'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = 'c49ddbb9-ee57-4209-9956-7d605717f19f'
WHERE blueprint_id IN (
                       '2a5a359f-4159-4862-8ca6-16aa51f30e3d',
                       '8dec706b-e2a3-47bb-919b-bcd06c4c67bf',
                       'e58526c4-a7b6-4908-9550-3e1034cd93a6',
                       '4b3b9774-6d5c-4885-b494-637473a7ba89',
                       '0df47dba-bd6a-49cd-951f-9bb773e085b9',
                       '8370fe8c-6fa8-48df-a56e-0b716955b80f',
                       'caf52fae-438e-42d2-b616-36d74f375134',
                       '636437a4-8bba-4f92-9901-4cf9837e1cb3'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '03693e90-ece4-40f2-bb22-35dc32ed02ea'
WHERE blueprint_id IN (
                       'ea105045-cc90-4191-94f1-303d548fe7fb',
                       'eaccf9e5-cb47-435b-bbf7-9660deec5db0',
                       '9845213f-7d1d-4c8d-8729-df7acebf506d',
                       '0d584cc3-93c1-4d84-9549-e302e27249bb',
                       '991d2d86-1659-47a2-861c-ac6250fb7752',
                       '897aaf4d-2a6c-4ac8-882f-65b531c42d63',
                       '72ed0bb9-8aca-4746-9637-fe1a661483e6',
                       'a6944388-1cf1-4b42-b532-58e20307e75b'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '00fb66a4-c081-46cc-aa62-0bcc05f416e4'
WHERE blueprint_id IN (
                       '9ae03bae-d22c-419d-8968-cbb138d47754',
                       '6e246144-44a7-4ca0-95a7-0e5a43f2566b',
                       'b302967e-aa6c-4aec-90ef-7c09a9db0e6b',
                       '874db321-38e4-4aa9-aedc-b36feb1e31b5',
                       '706e2592-a959-41b7-bdd7-21e60c3651fb',
                       '2d37aa18-4405-491d-9dec-e5a2defa6a89',
                       '8fde9903-578e-401d-a9e2-7cca9f11a9bf',
                       '37bd7ab2-e045-4256-8928-e7759d61e62c'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = 'abac4eb4-95e5-4399-947f-54612024aefb'
WHERE blueprint_id IN (
                       'caebda72-7514-45da-96b4-9f0dc15e808a',
                       'e0ad7e13-f951-4603-87d6-3107c16b8fef',
                       '7515a5a8-1820-4744-a83c-5c143d9b7e63',
                       'c5469e5e-3a56-46f9-94cc-2b04e4e263fb',
                       '7ac0f5c2-c9c7-48e1-9bc0-5ec9682e17ba',
                       'b43a5d0c-fbe2-4d67-830f-7b650e031676',
                       '10dbe0d7-51a4-4f68-9a62-95af6637a979',
                       '17f32e78-729b-4dad-a644-eb7781f8616b'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = 'c8cee447-30df-4c66-a779-7ed61098d768'
WHERE blueprint_id IN (
                       '3c6e4453-da0f-4caa-b71c-3069c294a4ad',
                       'fe8eb52c-c4b1-40c4-a6e1-df31aa56c7f3',
                       '599a4247-a95e-4144-88d0-78b590e536f9',
                       '2ad7ee86-0c9f-43af-9f4b-b6c75ff8de73',
                       '831cf381-16bf-4ad9-b418-68b24ef19fe7',
                       'c2d4ac29-1504-4d51-883b-4299f89db8cd',
                       'e2033c92-0f21-4909-b581-807fd90dc683',
                       'c47882d6-decc-42c1-afbc-5113ca42a9fc'
    );
UPDATE mystery_crate_blueprints SET blueprint_id = '65ff58c7-ec00-4695-81cc-8c04c167e6e6'
WHERE blueprint_id IN (
                       '3be26aa4-e7e6-4a7f-9d2e-b0a9f04329f6',
                       'bd603bc8-5fa6-4c3c-95b0-178ae1ddfb83',
                       '89c2487e-d816-463a-9c6e-e57bd53a17ce',
                       '904ae783-6b44-4d37-ac95-dcf7fa7370a6',
                       'e2857853-2fa4-428d-a484-e5ee8c8e0a67',
                       '8bec27c3-21a2-49aa-815e-e839c889627a',
                       '8eb2a7ec-450a-49f9-a695-b977c383eab9',
                       'b32ffd15-6376-44a3-8b31-472b70e7df3f'
    );
--endregion

-- region -- update template_blueprints ------
UPDATE template_blueprints SET blueprint_id = 'f6b4a628-78e5-47d1-9f75-c95d70b3472e'
WHERE blueprint_id IN (
                       '6bee4231-85d3-4b0d-91d3-dba759d8fb4f',
                       'a40d9c2c-812f-45f9-a059-f52dc7980982',
                       '22c26dc4-2107-4a0b-ab9b-c8da99e9abfb',
                       '561960ca-bb1f-4349-ad1e-33484be8dec0',
                       '2f2cbac4-8c40-43ef-ba9c-d91f91cd355a',
                       '9845bf88-f9b2-4c07-ac62-54accc1d4bc4',
                       '77cc363f-ac45-4545-af4a-3c9dfb704789',
                       '2e20cb8a-b40a-4b64-b21c-16fa94a9392e'
    );
UPDATE template_blueprints SET blueprint_id = 'e9886f77-1ed9-4aed-8f5a-abd3e6275464'
WHERE blueprint_id IN (
                       'a078b4d4-2e3d-49e4-86d2-d2b6a839534a',
                       '9bbcda1f-7c1e-4ea7-8de7-fa507bb0b39f',
                       'cdcde923-430e-4b24-9916-56b48150bbed',
                       '40740724-d891-469f-b74e-54e8b5685a19',
                       'b1b78eed-7371-4ad9-a3ea-6ef86d61e278',
                       '1cfe0387-33f3-4d80-843e-1998eb498875',
                       'aa67c963-9a4d-4d0d-bd01-d8075eaac56a',
                       '05c8c140-a073-4f7c-96a6-297ae99c86a9'
    );
UPDATE template_blueprints SET blueprint_id = 'f1ebf586-7ce1-4e5e-b6ba-113e96c4accd'
WHERE blueprint_id IN (
                       '4ad520a2-e033-4ca1-ba65-0ebbc3513697',
                       '97246188-09cd-4858-bbbb-49967c55cd38',
                       '7d5d8f82-27b0-4ef2-b44e-7476027a0371',
                       'd8e49315-62f8-47f4-9a8c-92e9ea4ebfa9',
                       'fe2b900f-5e56-4ce5-bf6c-31b498eb26af',
                       'e0427ba7-a4a4-46fc-820c-22d2f693d8eb',
                       '158189c0-9ac6-4323-8008-71ff210e2287',
                       '8f2bf410-3cda-459e-ad14-5e1271faf926'
    );
UPDATE template_blueprints SET blueprint_id = '8c944ba8-bf7a-46b3-831f-641017430608'
WHERE blueprint_id IN (
                       '187c9bfb-d2e8-4af9-ac5a-9f8e86ee24ed',
                       '3bccd240-0378-4d13-a23e-b263be784178',
                       '8a9b48ed-48ef-44da-88a1-5f8e7988505d',
                       '61cf7d24-12dc-4993-88c3-037fc199da58',
                       'f83db0be-47a5-49d6-bd27-0315e3b0fd48',
                       'b38a5ab4-83e1-4183-ba81-c2834d4a6815',
                       '07b84501-67e3-42dc-92e1-e333f4504ae9',
                       'f90d8ee3-ac8c-46f2-85e8-9f522ce9b552'
    );
UPDATE template_blueprints SET blueprint_id = '8c990899-35a3-4eb3-b710-79186b735fa4'
WHERE blueprint_id IN (
                       'e8b2be8f-04e8-48ee-a4cd-abe78fb717ca',
                       'eccff0f5-784d-4e87-a29d-b5a110b3af04',
                       '5c40580e-4ce6-4245-bc1d-a82e3c6f6d00',
                       '9303203d-1446-4a7a-83ca-f33f6b556534',
                       '7aaf7074-8acc-4594-a1ef-0d3b3252ebfa',
                       '3ffee0cc-b2bd-4044-9f0f-da79cc720acc',
                       '7ec4c431-1bce-42b0-b1a2-bfe4447f3ef8',
                       'c7e2c330-7241-46ae-bb8b-8fbe45295905'
    );
UPDATE template_blueprints SET blueprint_id = 'f5036def-a979-4d03-89de-0281be618d57'
WHERE blueprint_id IN (
                       '3a5f326d-e6cb-4e81-84aa-10633e07cff7',
                       'c0a4ae40-6dfa-486f-92b3-c377d07a7be6',
                       '589156dd-47fc-4d1b-a745-64d719993368',
                       '1770e130-d005-4ca8-846d-c3a4286ea55b',
                       '6c6b46dc-021a-4ba0-9d06-1b41e2524da4',
                       'afc1793f-da3f-42d2-a45f-b8524c92d420',
                       '38c91bc6-4bf6-4ac8-b631-63a8c059da76',
                       'd3c39a8e-85fa-499f-b346-d3aa459bcd13'
    );
UPDATE template_blueprints SET blueprint_id = '3f7076cc-0cc0-4d2a-a79f-2462d2e02c20'
WHERE blueprint_id IN (
                       '53d0c32b-66f7-4635-8a23-dab98d60348b',
                       '4bffd3ef-6337-481b-8daa-4f519b82374e',
                       'd003120e-fd38-4d9c-af4a-eb54637e4bf5',
                       'e9c5b973-b8d8-4ee9-a47d-92a601c69126',
                       'a9b8516c-9be4-4ec8-9e8b-0ce155362ba6',
                       '54b5ec13-6468-4765-a97c-7063d0fda2f3',
                       '272635b1-8899-4904-9912-ebb14e0237a0',
                       '762a84f2-fc77-4522-b441-0d204274aa1c'
    );
UPDATE template_blueprints SET blueprint_id = 'ab8a5813-53cd-40f9-96db-65f3ee3f977d'
WHERE blueprint_id IN (
                       '971db084-ef1a-472d-96cb-c28a15264d0d',
                       '7875e080-6f37-429d-9c6c-ef0c97c6408a',
                       '5ddc9895-439f-48ad-a0cf-a5b841322c8e',
                       '1b42c687-57b8-4652-a4b0-d3abda6f473a',
                       '9b9b92cf-d07d-464f-941e-e934d9dc6997',
                       '26d276d2-0c91-4774-a10e-c4e358644ca3',
                       'f976202d-347a-49bb-b669-091705d1cb99',
                       '740cb444-68e1-4bf0-bdfd-f5b0883bcdf4'
    );
UPDATE template_blueprints SET blueprint_id = '3dab2057-359d-422b-a0b3-b9714096fc00'
WHERE blueprint_id IN (
                       'a4d1500a-e879-4057-8b19-bbd1594a7c5f',
                       '4538e24e-65ef-4da2-a47a-30b675ffc714',
                       'e52ab94c-f710-4957-ab4b-cdfd61b0a8e5',
                       '463edf3b-72ed-4b89-9f8b-3068b2769d3c',
                       'dcce1019-867e-48a5-ad48-fcaa1980efc9',
                       'c12b704a-02da-4791-9d67-5c75a2213c5a',
                       '53a24b70-22db-4793-bc03-8a988782df19',
                       '27047b50-8c4e-453c-83f4-a42f7e7bc601'
    );
UPDATE template_blueprints SET blueprint_id = '11faa37d-5c33-43ee-84ac-f01ba8ef8875'
WHERE blueprint_id IN (
                       'e62003dd-69aa-4c41-8a97-01d7f806015a',
                       '3fe2c440-aaf6-421f-9194-0991dd5120cd',
                       'fbba705b-775a-454d-9307-f558e677acc9',
                       'fa6eb8cb-0822-42a2-8b4c-edc0147cf18d',
                       '5599cfa8-9a16-47f7-aa5e-1581c362df8b',
                       '4d192e46-ecf6-4514-9ab1-e67d71003513',
                       'fdfe3a17-ec10-4cb6-a094-e7bd828f5e76',
                       '3b80c589-9eed-4082-baad-e7c8f237fa27'
    );
UPDATE template_blueprints SET blueprint_id = '96c51cff-c2bb-49ae-a186-8477fd850d84'
WHERE blueprint_id IN (
                       'cd6151b8-bb2a-46a8-a965-a74c22578f9c',
                       '7df80cf3-7524-406a-ad18-4e32688c6f6c',
                       '92e7fdb6-1eee-4ba0-8d65-c58e81972578',
                       '7a4479cf-196f-4192-af6b-bf2792e23fc5',
                       'f63f5bd8-7520-401a-b362-dea2c44b830f',
                       '6ceb0b4a-73fb-4c07-949f-9ee3c67cec70',
                       '8a69404c-8436-4c2c-b70b-5bead79a76cb',
                       '1a50e776-32e9-4861-b2f4-d4417a36fa9d'
    );
UPDATE template_blueprints SET blueprint_id = 'f053db4f-cac9-4445-aa60-f75e691f4fe6'
WHERE blueprint_id IN (
                       '3f6b0357-f644-430c-a7df-5489537c046d',
                       'b992d386-b6aa-4f14-9bd1-37cff98bd98b',
                       '2b7f9c5d-1014-4f08-afdd-fdd8433e0dc3',
                       '74cb10fd-f958-4475-abbf-b026fc820785',
                       'e746a916-bcda-480f-9f89-2071b818c5df',
                       '590f895a-6bc9-46a4-933b-08c8f4ec6ffc',
                       '3de8313f-87df-4b18-b286-fcc5dff81fd5',
                       '587cac53-82af-4404-b8b7-7541797a8b9b'
    );
UPDATE template_blueprints SET blueprint_id = 'fb1a375e-4ce6-4e1f-b906-629198584c00'
WHERE blueprint_id IN (
                       '98c45848-677e-4a08-b890-b23ba09040fb',
                       'fdd79139-83aa-4657-83ba-db0e296c9ee0',
                       '0df33b12-ae2c-4529-9a8e-00d5da04eb53',
                       'd11f6c43-5707-4161-aea1-77d5dab23bff',
                       'a420f0b5-3286-4788-a71a-9782b54f6008',
                       '3e16f30e-a9ef-409c-807b-76f4dc556732',
                       'dd1e5ca9-3f18-4bdf-827b-ce3f90842374',
                       'c194a857-b558-4344-84aa-3f37d87a537d'
    );
UPDATE template_blueprints SET blueprint_id = '7dc7a664-9bb1-49ba-ab3d-ed7a300bacfe'
WHERE blueprint_id IN (
                       'a5d0e25d-edfa-4f23-8d3a-e9d16e227048',
                       '53e3ce09-6a84-47bb-bb31-2e8cfd966cb7',
                       '8d3e1223-3527-406b-b310-aaa9bbd0c939',
                       '6592b54d-9ed4-4e2b-a368-4a84d5bf3987',
                       '1bf33888-cefd-42f4-b89b-3253a2b53e56',
                       '9d9cf7dc-e210-404c-b1e9-922b4e849514',
                       '09090725-a7fd-49ea-9a5c-6203344de741',
                       '13700be7-26fd-4e73-9c64-d3649a5a1f3a'
    );
UPDATE template_blueprints SET blueprint_id = 'f643bd16-e6df-4a1c-a2d6-283717843973'
WHERE blueprint_id IN (
                       '1109a7c8-03b4-4658-92ca-bd1c540b805f',
                       'ce8be4c0-c7b3-4d17-b3db-641c2a7faffa',
                       'eddc42ea-294e-40ba-8927-e28d5fd896bb',
                       'c5cc3968-25e1-443e-af7f-cb78f7754222',
                       'c5a1c6c0-ba9b-4509-9c26-37faad6a6986',
                       '2bfb4c29-5e1c-44f8-9867-0deedea59142',
                       'b74a7f4b-469b-4233-951e-cadea4376cd2',
                       '6689d421-a3b1-4b68-a438-562879aa9c01'
    );
UPDATE template_blueprints SET blueprint_id = '889b7106-7a66-41b2-8d15-1e2960f2620b'
WHERE blueprint_id IN (
                       '9844b794-b802-4658-b778-3c1300998ef1',
                       '74727c69-55a4-4ecc-a90a-03eacfab7581',
                       '1ea1551c-b66a-46e9-b3ce-da0fb22c5bc5',
                       'b27f5dd7-393a-4e0a-a9d1-07520227ffa6',
                       '7f088002-5250-4447-a8f1-d39c3ca2ce0d',
                       '5193defd-40c5-4b13-83c2-03199677a0fb',
                       'f52f1ccf-59ba-46d4-9ed0-7ff39558d00b',
                       'c51a094b-3e5e-4b43-bb57-af0f82b38666'
    );
UPDATE template_blueprints SET blueprint_id = '19895ef4-8951-4ed2-85a9-ef02d1c1efa0'
WHERE blueprint_id IN (
                       '6b61b82c-b5d9-4589-8303-57f418b0bb9d',
                       '0db51c94-95e7-4fa3-8f09-9ba9f60c2964',
                       '5f531492-82bc-434e-ab88-a1a88e002771',
                       '51528eb1-dd7a-41a4-a72a-e9490266407b',
                       'fa58696e-e410-40b7-a48c-16be320f9211',
                       '8afc0964-1573-4adb-aeff-2185d71c939d',
                       'a93f8b44-e149-4933-8f5a-4053c386580e'
    );
UPDATE template_blueprints SET blueprint_id = '78a3b8a6-f4d6-497d-b621-87893950fd2b'
WHERE blueprint_id IN (
                       'e36ecf20-40e5-4ac0-ba52-757bf3372619',
                       'ae3722db-f5d8-410b-81b1-afcf601bd171',
                       '78349dbc-8197-476c-b84d-5d261034928e',
                       '772d0087-5409-4383-8c24-98d8f55e35bc',
                       'cb467e60-7305-4ba1-94a6-83229936bfaa',
                       'd79619ba-f897-49c5-ab40-0e879e9accaf',
                       'd1383c38-2047-46cc-b530-94e47d197cf1',
                       'b59e3e9c-77d3-4a63-be63-c4598e43eec1'
    );
UPDATE template_blueprints SET blueprint_id = '23aabcd6-9e40-44b0-a0f1-b2cece19a54c'
WHERE blueprint_id IN (
                       '72852593-84b5-4883-b3ab-1496de44e7be',
                       '9d125cf1-5123-4bc8-a598-b3d0bc3890be',
                       'c1012df4-f8f2-4e04-8fe0-021f0d0cdc4f',
                       'f045ab10-f03b-4b1d-890b-9a1a3e973371',
                       '9c3b93ac-e1c5-45c9-bc72-ad49ec57c3a9',
                       '149fd8d0-459b-4516-9056-f8ffd9779d60',
                       'c0385acf-73f2-441f-b6ad-e1f622f4dd53',
                       '01259820-acd4-481a-9413-06aede2c4926'
    );
UPDATE template_blueprints SET blueprint_id = '4fa842e9-9261-4153-a695-142c57155248'
WHERE blueprint_id IN (
                       '8e5dd2c1-b52e-48cd-8539-aed12feae5d8',
                       '6e369636-5ac3-4217-9074-0b1350d1ec6d',
                       'fffdd01f-5b14-432a-8f7e-6f86d557be8f',
                       '05b337b1-ce3b-4fcf-a3d2-b94aa76628a9',
                       'c6a8df72-896c-458f-b679-8162396ab2e0',
                       '6e897de9-f333-4a15-aeda-86c6eecc8a4e',
                       '244c979d-87de-4af7-a3a7-196a3d4c815a',
                       'f654b52b-ebea-4edf-abc9-b9431d7c299b'
    );
UPDATE template_blueprints SET blueprint_id = 'd2a5d45a-eca3-49c0-8c66-bda97626cd2d'
WHERE blueprint_id IN (
                       '98bc7460-2f0b-479b-befe-de663cc3a8f0',
                       '4e0e7f14-b237-484f-98bc-993ccbc3e887',
                       '4d939a57-ed63-47bf-a941-d654faf3b343',
                       '86acd9ff-f51f-433b-b58b-ce947d0e66e9',
                       'c0ced55a-1cc7-42f7-8749-9f616dd59f56',
                       'c4935402-4b1d-401f-996e-77876a97761c',
                       'e6769598-b69e-448b-abeb-20222208bb34',
                       'a91491fa-4bd0-49b1-8aff-fe35d1a3c20f'
    );
UPDATE template_blueprints SET blueprint_id = '3029890f-5336-464b-ba0e-d701eb20a360'
WHERE blueprint_id IN (
                       'c5cf853c-758a-4169-ba05-866548b86a59',
                       'd4912e43-62c1-4891-a2f5-4cd1d0d0f6b3',
                       '5c766a22-7058-4456-abe6-97dcc189fd34',
                       '9f4255de-9d4b-4ffd-9ad3-bcfc16175497',
                       '1761bc66-5199-46be-a4e8-06c6d3e8566f',
                       '98899681-ba45-4fa6-a8c1-00f279531f1d',
                       '693adb11-2b9d-4921-ba04-93eb7aa744e3',
                       '7ff51d7a-8753-458c-953a-8b382fc98e51'
    );
UPDATE template_blueprints SET blueprint_id = '9c20c27c-e0ed-406a-ad02-0ab05ba10364'
WHERE blueprint_id IN (
                       'c4ec30b1-3301-4bf6-ad94-a91c306df5ca',
                       '867daad5-1ce3-4ae0-93e2-6f9d0f465435',
                       '583fe88f-66e0-442a-96be-e2877fd251ca',
                       '1e38c064-08ff-4bac-a280-8325bab4762f',
                       'beb80df7-1344-41f7-934e-56272a1d2217',
                       '71b85a94-b35d-464a-8e37-0eed8f2e4c22',
                       '5b038029-6684-410b-8438-d0da29f929cb',
                       '66285093-fb4d-4bcc-a7b9-15b5c1fdb821'
    );
UPDATE template_blueprints SET blueprint_id = '1eb3a17c-d157-4bb1-a6dc-7cd3fe7ab7b9'
WHERE blueprint_id IN (
                       '3dbe2090-064f-4430-bcdf-12d017689181',
                       '61d2184b-b4fd-4623-80e2-78ddf8fa8ebf',
                       '8303fa1c-ddb8-446a-8e0e-4d31709565fc',
                       '86e1200e-04af-4a4b-8a99-0844a69d2fa4',
                       'f6627fe6-10ef-4b37-8304-defffef98e6e',
                       'd4f65a17-991c-41e9-ac6e-9c28375660bc',
                       '53f57c4f-e8b5-45c7-b1c6-0ccce0e52ba1',
                       'b252c620-20c6-4465-a683-82956d41fec1'
    );
UPDATE template_blueprints SET blueprint_id = '06e1894d-3872-49cd-a4cf-0c42ee4cedf0'
WHERE blueprint_id IN (
                       '555503c1-3086-42cf-a967-01e40b521897',
                       '4f7dff5f-a6e6-4d21-b638-f025b203ea9d',
                       '720e3668-2bce-4670-a0e9-7838e2929c89',
                       '9491926d-c234-4ebb-b96e-e7a8e50791fb',
                       'e44570ed-aa12-46a2-af26-a16d8692059b',
                       'ef7f0181-f36f-4622-9993-746ac90d50b7',
                       '6b4a8dfb-77dd-45ab-b2c3-bfcf01f54ce0',
                       '0d367f31-b3e2-4b53-960d-5200fe53bb22'
    );
UPDATE template_blueprints SET blueprint_id = 'c609283d-82bd-4a04-8e80-450ec5c64a56'
WHERE blueprint_id IN (
                       '8142a60a-ecb4-4f99-9a34-9936478370e4',
                       'bd9db924-a289-4039-84bf-2b9a9e236327',
                       '53192802-57e4-4fba-80d7-d5f0785b30b0',
                       '34083108-555d-490d-b4df-8dc3ee63c749',
                       'c6d3de5c-eeb6-4686-8a8a-f4ee69a6d86e',
                       '0cc56beb-b5bb-4a47-92bc-3a386ba0b9f1',
                       '6e72ff2d-4a23-4145-b8d9-8decbc9967b2',
                       'fbf3074c-e516-459d-af56-08149ff20615'
    );
UPDATE template_blueprints SET blueprint_id = 'fb11780c-c6f7-43a8-a5fe-62e77add9e9d'
WHERE blueprint_id IN (
                       '2b09d985-085b-440e-a0d1-752b1f5e0a2b',
                       'bf235fa4-df21-45e6-ac6d-c6981ea1c699',
                       '8de1f3d6-c41f-42cb-9c7c-f1eb895030e3',
                       '6e6eb9f9-a547-48cf-8681-5166e22785fd',
                       'dc3985e2-7689-4332-936c-7cd0e5a9f10c',
                       '1a388404-8811-47b8-8c67-f5c21fb2cbed',
                       '09ecb2e8-3783-4639-ae5d-8e584d346f48',
                       '393c003a-d23d-4f82-a93b-fbaa8a9d9dfe'
    );
UPDATE template_blueprints SET blueprint_id = '3df394d1-3f01-45fb-a2c7-1bdecbbc90a6'
WHERE blueprint_id IN (
                       '492e22c6-ba3e-47d6-955e-53e362ea0f20',
                       '447af743-9b46-447a-aa5f-a683b5f90b28',
                       '33b32552-9639-4eef-97f5-0b570739976f',
                       '770bd969-6d3f-4236-94a7-325eaa4e998b',
                       'a481c00c-4c3e-47b3-ab8e-948da3150dcc',
                       'b0261840-d31a-491d-9591-9167f0f2720c',
                       '4e458fb4-01e8-4770-9574-45ea996acb95',
                       '2fafd818-95ad-4eaf-a744-c9a09f82535b'
    );
UPDATE template_blueprints SET blueprint_id = '5c0b5d6b-1300-4727-af15-9c68c377554f'
WHERE blueprint_id IN (
                       '205b3a4e-e192-430c-b83f-7182c5e5aea8',
                       'f85c95e9-35f1-4e9d-8a6a-56b86af8344d',
                       '2fc63f40-1c70-4177-b0c3-fc59ac1a25a4',
                       '1cc07d02-c23c-4af4-a570-34478bea8185',
                       '043c84e9-de04-463f-9214-9c750d64bcf3',
                       '65b44826-d66a-456a-850b-348686bba714',
                       '805ab241-7f31-4673-b4cd-5cb26d4f440c',
                       'f7f9a1a9-1b5d-4207-a90b-89811ed58351'
    );
UPDATE template_blueprints SET blueprint_id = '21eccaa7-fa81-4aab-837e-f17ce627b399'
WHERE blueprint_id IN (
                       'df77a278-0daf-46ad-8433-79e3efd63fff',
                       '22e9bf6b-e7da-4551-83fd-8c1583eebdf7',
                       'acc19db8-ea3e-4159-bdd5-397a95fb0662',
                       '79f76bea-efee-4046-8f79-e0b3cd29bd4e',
                       '5039d66d-a516-4ed7-8f92-cb4715a23285',
                       '858f7dac-8b53-47ca-96e5-d5ed2ce4e4d7',
                       '54d6ea9f-c896-4463-8def-ccc958f925c4',
                       'd61026b5-941c-4fa4-9b08-3ddfbdf1d131'
    );
UPDATE template_blueprints SET blueprint_id = 'c49ddbb9-ee57-4209-9956-7d605717f19f'
WHERE blueprint_id IN (
                       '2a5a359f-4159-4862-8ca6-16aa51f30e3d',
                       '8dec706b-e2a3-47bb-919b-bcd06c4c67bf',
                       'e58526c4-a7b6-4908-9550-3e1034cd93a6',
                       '4b3b9774-6d5c-4885-b494-637473a7ba89',
                       '0df47dba-bd6a-49cd-951f-9bb773e085b9',
                       '8370fe8c-6fa8-48df-a56e-0b716955b80f',
                       'caf52fae-438e-42d2-b616-36d74f375134',
                       '636437a4-8bba-4f92-9901-4cf9837e1cb3'
    );
UPDATE template_blueprints SET blueprint_id = '03693e90-ece4-40f2-bb22-35dc32ed02ea'
WHERE blueprint_id IN (
                       'ea105045-cc90-4191-94f1-303d548fe7fb',
                       'eaccf9e5-cb47-435b-bbf7-9660deec5db0',
                       '9845213f-7d1d-4c8d-8729-df7acebf506d',
                       '0d584cc3-93c1-4d84-9549-e302e27249bb',
                       '991d2d86-1659-47a2-861c-ac6250fb7752',
                       '897aaf4d-2a6c-4ac8-882f-65b531c42d63',
                       '72ed0bb9-8aca-4746-9637-fe1a661483e6',
                       'a6944388-1cf1-4b42-b532-58e20307e75b'
    );
UPDATE template_blueprints SET blueprint_id = '00fb66a4-c081-46cc-aa62-0bcc05f416e4'
WHERE blueprint_id IN (
                       '9ae03bae-d22c-419d-8968-cbb138d47754',
                       '6e246144-44a7-4ca0-95a7-0e5a43f2566b',
                       'b302967e-aa6c-4aec-90ef-7c09a9db0e6b',
                       '874db321-38e4-4aa9-aedc-b36feb1e31b5',
                       '706e2592-a959-41b7-bdd7-21e60c3651fb',
                       '2d37aa18-4405-491d-9dec-e5a2defa6a89',
                       '8fde9903-578e-401d-a9e2-7cca9f11a9bf',
                       '37bd7ab2-e045-4256-8928-e7759d61e62c'
    );
UPDATE template_blueprints SET blueprint_id = 'abac4eb4-95e5-4399-947f-54612024aefb'
WHERE blueprint_id IN (
                       'caebda72-7514-45da-96b4-9f0dc15e808a',
                       'e0ad7e13-f951-4603-87d6-3107c16b8fef',
                       '7515a5a8-1820-4744-a83c-5c143d9b7e63',
                       'c5469e5e-3a56-46f9-94cc-2b04e4e263fb',
                       '7ac0f5c2-c9c7-48e1-9bc0-5ec9682e17ba',
                       'b43a5d0c-fbe2-4d67-830f-7b650e031676',
                       '10dbe0d7-51a4-4f68-9a62-95af6637a979',
                       '17f32e78-729b-4dad-a644-eb7781f8616b'
    );
UPDATE template_blueprints SET blueprint_id = 'c8cee447-30df-4c66-a779-7ed61098d768'
WHERE blueprint_id IN (
                       '3c6e4453-da0f-4caa-b71c-3069c294a4ad',
                       'fe8eb52c-c4b1-40c4-a6e1-df31aa56c7f3',
                       '599a4247-a95e-4144-88d0-78b590e536f9',
                       '2ad7ee86-0c9f-43af-9f4b-b6c75ff8de73',
                       '831cf381-16bf-4ad9-b418-68b24ef19fe7',
                       'c2d4ac29-1504-4d51-883b-4299f89db8cd',
                       'e2033c92-0f21-4909-b581-807fd90dc683',
                       'c47882d6-decc-42c1-afbc-5113ca42a9fc'
    );
UPDATE template_blueprints SET blueprint_id = '65ff58c7-ec00-4695-81cc-8c04c167e6e6'
WHERE blueprint_id IN (
                       '3be26aa4-e7e6-4a7f-9d2e-b0a9f04329f6',
                       'bd603bc8-5fa6-4c3c-95b0-178ae1ddfb83',
                       '89c2487e-d816-463a-9c6e-e57bd53a17ce',
                       '904ae783-6b44-4d37-ac95-dcf7fa7370a6',
                       'e2857853-2fa4-428d-a484-e5ee8c8e0a67',
                       '8bec27c3-21a2-49aa-815e-e839c889627a',
                       '8eb2a7ec-450a-49f9-a695-b977c383eab9',
                       'b32ffd15-6376-44a3-8b31-472b70e7df3f'
    );
--endregion

-- create the new weapon skin objects
WITH wps AS (
--     this returns a weapon id and the default skin blueprint id for each weapon without an equipped skin (genesis weapons)
    SELECT _w.id, _wm.default_skin_id
    FROM weapons _w
    INNER JOIN collection_items _ci ON _ci.item_id = _w.id
    INNER JOIN blueprint_weapons _bpw ON _bpw.id = _w.blueprint_id
    INNER JOIN weapon_models _wm ON _wm.id = _bpw.weapon_model_id
    WHERE _w.equipped_weapon_skin_id is null
)
INSERT INTO weapon_skin(blueprint_id, equipped_on)
SELECT wps.default_skin_id, wps.id
FROM wps;

-- update weapons equipped_weapon_skin_id
UPDATE weapons w SET equipped_weapon_skin_id = (SELECT id FROM weapon_skin WHERE equipped_on = w.id) WHERE equipped_weapon_skin_id IS NULL;

-- insert the collection_items or these weapons
WITH wps_skin AS (
    SELECT 'weapon_skin' AS item_type, ws.id, bpws.tier, ci.owner_id
    FROM weapon_skin ws
    INNER JOIN blueprint_weapon_skin bpws ON bpws.id = ws.blueprint_id
    INNER JOIN weapons w ON w.id = ws.equipped_on
    INNER JOIN collection_items ci ON ci.item_id = w.id
    )
INSERT
INTO collection_items (token_id, item_type, item_id, tier, owner_id)
SELECT NEXTVAL('collection_general'),
       wps_skin.item_type::ITEM_TYPE,
       wps_skin.id,
       wps_skin.tier,
       wps_skin.owner_id
FROM wps_skin;

ALTER TABLE mechs
    ALTER COLUMN chassis_skin_id SET NOT NULL;

ALTER TABLE weapons
    ALTER COLUMN equipped_weapon_skin_id SET NOT NULL;
