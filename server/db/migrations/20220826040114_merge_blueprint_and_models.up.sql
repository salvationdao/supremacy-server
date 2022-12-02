ALTER TABLE mechs
    RENAME COLUMN weapon_hardpoints TO weapon_hardpoints_dont_use;
ALTER TABLE mechs
    RENAME COLUMN utility_slots TO utility_slots_dont_use;
ALTER TABLE mechs
    RENAME COLUMN speed TO speed_dont_use;
ALTER TABLE mechs
    RENAME COLUMN max_hitpoints TO max_hitpoints_dont_use;
ALTER TABLE mechs
    RENAME COLUMN power_core_size TO power_core_size_dont_use;

ALTER TABLE mech_skin
    ADD COLUMN level INT NOT NULL DEFAULT 0;

UPDATE mech_skin ms
SET level = (SELECT default_level FROM blueprint_mech_skin bms WHERE ms.blueprint_id = bms.id);

ALTER TABLE weapons
    RENAME COLUMN slug TO slug_dont_use;
ALTER TABLE weapons
    RENAME COLUMN damage TO damage_dont_use;
ALTER TABLE weapons
    RENAME COLUMN default_damage_type TO default_damage_type_dont_use;
ALTER TABLE weapons
    RENAME COLUMN damage_falloff TO damage_falloff_dont_use;
ALTER TABLE weapons
    RENAME COLUMN damage_falloff_rate TO damage_falloff_rate_dont_use;
ALTER TABLE weapons
    RENAME COLUMN spread TO spread_dont_use;
ALTER TABLE weapons
    RENAME COLUMN rate_of_fire TO rate_of_fire_dont_use;
ALTER TABLE weapons
    RENAME COLUMN projectile_speed TO projectile_speed_dont_use;
ALTER TABLE weapons
    RENAME COLUMN radius TO radius_dont_use;
ALTER TABLE weapons
    RENAME COLUMN radius_damage_falloff TO radius_damage_falloff_dont_use;
ALTER TABLE weapons
    RENAME COLUMN energy_cost TO energy_cost_dont_use;
ALTER TABLE weapons
    RENAME COLUMN is_melee TO is_melee_dont_use;
ALTER TABLE weapons
    RENAME COLUMN max_ammo TO max_ammo_dont_use;

-- update template blueprints because i am dumb and should have just used a FK
ALTER TABLE template_blueprints
    ADD COLUMN blueprint_id_old UUID;

-- weapons
UPDATE template_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '74005f3b-a6e2-4e4b-a59c-e07ff42cb800'
WHERE blueprint_id = 'daa6c1b0-e6ae-409a-a544-bfe7212d6f45';
UPDATE template_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '260370f6-aaef-4dcb-a75a-c4642ba89dfb'
WHERE blueprint_id = 'ba29ce67-4738-4a66-81dc-932a2ccf6cd7';
UPDATE template_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '8c3cbbc7-eae8-4ca8-9811-663de0e73eb6'
WHERE blueprint_id = '06216d51-e57f-4f60-adee-24d817a397ab';
UPDATE template_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'c1c78867-9de7-43d3-97e9-91381800f38e'
WHERE blueprint_id = '347cdf83-a245-4552-94b3-68faa88fbf79';
UPDATE template_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'd831dea7-e598-47af-b4e3-3c37b5eb969a'
WHERE blueprint_id = '26cccb14-5e61-4b3b-a522-b3b82b1ee511';
UPDATE template_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '0612fbda-3967-456b-86cb-1fb6eb03829d'
WHERE blueprint_id = '1b8a0178-b7ab-4016-b203-6ba557107a97';
UPDATE template_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '41099781-8586-4783-9d1c-b515a386fe9f'
WHERE blueprint_id = '17c32a72-0b7a-42bb-b144-7d7358509cde';
UPDATE template_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'e9fc2417-6a5b-489d-b82e-42942535af90'
WHERE blueprint_id = 'bf114d6b-bb1c-4de3-8abd-d2367715db52';

-- mechs
UPDATE template_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '5d3a973b-c62b-4438-b746-d3de2699d42a'
WHERE blueprint_id IN (
                       '8dd45a32-c201-41d4-a134-4b5a50419a6e',
                       '2b0df453-a9ac-4a43-967e-ed7a1cdaecf1',
                       'cc0de847-24a3-4764-bbbd-78304e79150e',
                       '74e3bd32-4b27-4038-8338-f4f60db4be71',
                       '8b98d84c-48ad-481b-bc86-d1d02822f16c',
                       '12ab349f-0831-4f3f-9aa1-f3aade13abb3',
                       '014371f7-bca3-4a1c-992e-f55ac5721de1',
                       'd041e6dc-0ed5-4294-8fdc-f49707e1a854',
                       '76a60a59-291a-49e2-b7d4-d7cbfa2a3feb',
                       '0dd254a6-edd6-481b-b3d7-36c57a9836de',
                       '11d09723-a9e2-4c57-bb01-5291be841cb7',
                       '3e710f5f-1290-428d-a6c1-1acc08f98837',
                       'd5c07ee1-8213-43bf-bd65-bbacb442d4ff',
                       '3e469099-7fba-4a17-a392-8fd850d35c28',
                       '88f4be7c-20f6-48a9-9d56-8f34ebc4607c',
                       'cfb96c42-873d-43e8-9060-ae107f18037a',
                       'ad124964-e062-4ab8-9cce-e0309fd6b31d'
    );
UPDATE template_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '625cd381-7c66-4e2f-9f69-f81589105730'
WHERE blueprint_id IN (
                       'e1dc2da7-78c4-400e-a302-a40b2d27a3c8',
                       'f9cfa25b-02bc-4a3a-9a31-9a8a4a073714',
                       '8a49a4ca-4a61-4f51-b296-b421c951c1f5',
                       '276f62c9-23f3-443c-8fad-90fc4d074009',
                       '4f6180a5-4d0c-44ca-9422-a4d28c3b863a',
                       '1d76a559-3c6d-4f26-a108-a3b5549e35ab',
                       'd0c291de-1edb-4b54-9559-ed3587034051',
                       'ac408b39-ba91-4a0e-9414-a37af0ea6314',
                       '80bf637e-9219-44fc-9907-1f58598d6800',
                       '9c4cd43d-35ed-4d3a-860c-d7ff00093dd6',
                       '17a1426a-b17e-4734-b845-588a87b6d8cd',
                       'f4da1432-0618-47c2-9337-9cda3e8125ba',
                       'bc19d28d-3fa4-45e5-86dd-1cf69bc3daee',
                       '0b297a84-0a07-4e09-87ba-ed2a877d7c4f',
                       '81f1fd1b-458c-499b-bbec-7560f8bea842',
                       'b2c43cd1-86ce-4154-b3d5-22776ed3a727'
    );
UPDATE template_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'ac27f3b9-753d-4ace-84a9-21c041195344'
WHERE blueprint_id IN (
                       'f9ca786e-730d-481e-a6c9-f04e5ed975d2',
                       'a5966250-4972-425f-b2d4-433e3820741a',
                       '24a4ebc1-e785-4b72-83e8-c8cf624801d0',
                       '93a65506-44aa-4d93-b184-7d72b8cf5d9b',
                       '9bc4611a-0da2-4b07-bab6-7f15b1c97f90',
                       'bd283770-0d77-45ea-b4d2-6286351eecc7',
                       '337b3c82-61ae-4959-bcf0-50f0985f8ed6',
                       '0a5b3678-f6c0-4f53-8a1c-e3c29f3a1632',
                       '5d213b5d-5d29-4e3d-bdf8-d9867c574d7f',
                       'beb40230-580e-4aa0-87c9-0aac27edcbb6',
                       'a8e507cd-f874-4e6a-a321-a4b04e171ba9',
                       '500128ca-7544-4250-aad4-aa1ae8a8ad1a',
                       '683c4461-a291-4245-9120-56c67f091fba',
                       '35deaa58-d05d-4a8c-8a84-178358c46647',
                       '1ca58e88-6280-427c-b6d2-25928e7bd292',
                       'a5806303-f397-44d5-b9d8-f435500362e6',
                       '30b40880-ffa3-4a17-883d-08de0bf1b479'
    );

-- update mystery crate blueprints because i am dumb and should have just used a FK
ALTER TABLE mystery_crate_blueprints
    ADD COLUMN blueprint_id_old UUID;

-- weapons
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '74005f3b-a6e2-4e4b-a59c-e07ff42cb800'
WHERE blueprint_id = 'daa6c1b0-e6ae-409a-a544-bfe7212d6f45';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '260370f6-aaef-4dcb-a75a-c4642ba89dfb'
WHERE blueprint_id = 'ba29ce67-4738-4a66-81dc-932a2ccf6cd7';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '8c3cbbc7-eae8-4ca8-9811-663de0e73eb6'
WHERE blueprint_id = '06216d51-e57f-4f60-adee-24d817a397ab';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'c1c78867-9de7-43d3-97e9-91381800f38e'
WHERE blueprint_id = '347cdf83-a245-4552-94b3-68faa88fbf79';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'd831dea7-e598-47af-b4e3-3c37b5eb969a'
WHERE blueprint_id = '26cccb14-5e61-4b3b-a522-b3b82b1ee511';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '0612fbda-3967-456b-86cb-1fb6eb03829d'
WHERE blueprint_id = '1b8a0178-b7ab-4016-b203-6ba557107a97';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '4b81474e-bc04-412f-9544-0c7104c9ab3d'
WHERE blueprint_id = '012aa14e-c6ad-4c14-bbe2-62f6591b71eb';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '59d2ff64-99ba-4634-bf2d-fb519be08430'
WHERE blueprint_id = 'e107fcbf-7235-4763-b64f-463b52bbc2e8';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'ef763255-6f39-4670-b1e4-087bb703b2e5'
WHERE blueprint_id = '4a8d5b4d-e776-41b2-a7d4-05f35949c121';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'e0208799-b2e3-416a-abfc-c2c236202a22'
WHERE blueprint_id = 'a98ed407-2222-4983-9db2-507b51bd3fb7';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'c534a594-ec9d-41dc-976d-f701c0500d07'
WHERE blueprint_id = '836eaff1-4177-4365-b002-1e33c52fc072';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '7ce877bf-fdc0-4169-a8f0-dd18ebfedf8a'
WHERE blueprint_id = '2aac545a-ffd3-45fb-b90c-f37efd71a455';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '7d8ae538-cdd5-493b-80b3-c078f01bec17'
WHERE blueprint_id = 'a04dc624-c380-4995-a7de-465cab8f04b4';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '833a15d9-1121-4ddd-a9a9-bf71bf8bde61'
WHERE blueprint_id = '5afb451a-25d9-48d0-bdc7-2568b82d76dd';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'ac78fbc7-ee23-497e-a3e1-7d748e2df5d6'
WHERE blueprint_id = 'f7eb3fff-e4e1-4ac8-b4d5-97bdf83410a5';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'fd0d47dc-4288-42da-b049-0ffbabb9153e'
WHERE blueprint_id = '3222705d-b944-49ab-a4cd-40a7e0a37770';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '2eeca329-23d3-4ede-85d6-4db716238853'
WHERE blueprint_id = '75bd39d5-3f22-417f-9d13-fe6f976f31be';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '4e8f98e9-80a5-43f7-a93d-faceb08bb824'
WHERE blueprint_id = '605d3154-0ebc-488a-98e8-34cc96ac299b';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '555ba722-9411-4ff8-9317-4a8bfcffb956'
WHERE blueprint_id = 'ffddffa1-34df-4f6e-848b-bfcd6ff51e70';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '1fe9d9f5-00df-4669-b95f-d8f1978bc24f'
WHERE blueprint_id = 'e2dfe8d3-6ae9-4c74-bdb5-69a89abd0a0f';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '8b532eef-8f75-45f2-ba5c-f3ea87000135'
WHERE blueprint_id = '6643553d-028b-4a2a-83a1-b05ba3d1de44';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '574e96d6-1f56-4458-82e7-1edb3d625df5'
WHERE blueprint_id = '4efb8c69-6791-4654-be38-d53e1bf8b946';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '01b7aa4e-6678-49d2-81f2-ea095ef8cb1c'
WHERE blueprint_id = '59fdfdbc-b76c-4985-989d-28ebabece510';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '536831f0-799b-4fd1-ba70-023a98d53668'
WHERE blueprint_id = '8bf79c2b-3965-4c6a-9603-499c34b83c18';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '52c0d775-ac19-4bef-ad4b-3280cc524d68'
WHERE blueprint_id = '5516fca1-7cba-41c7-a346-c291c06004ad';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'e557697f-3955-43c4-8020-eeb55a95a2ae'
WHERE blueprint_id = '9d16c59c-42ef-48fe-b940-235d79ab038c';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '2ed3ae32-1d78-4fee-b51d-54bce67f023f'
WHERE blueprint_id = '5d03834f-48d4-43c2-ae3a-d6928f57eb72';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '29b3e2f7-9e1b-4bed-9b26-edcd9d978b1e'
WHERE blueprint_id = 'fbc06608-0d87-4122-915c-da6943f80b6c';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'd61b6aa1-5ddb-42b9-b8af-1eeb5336134d'
WHERE blueprint_id = 'a072f906-a430-49a8-93b2-a8ad9abc007f';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '189424f3-11cb-4026-8a8b-6567a24108e0'
WHERE blueprint_id = 'cb7a1b76-840a-4165-a23f-6840a5f6a403';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '38a8d642-a449-4fc3-a1d4-4a6e3ac01c4e'
WHERE blueprint_id = '860b3c6c-9922-4857-a4f5-f5aa17d4b2a8';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'a16fc997-023e-4a99-8912-a2e62b44d710'
WHERE blueprint_id = '6bee5e8d-6831-4379-9f3b-a8632563113f';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '1301d52d-cef4-4956-8f31-161b1c8b4a63'
WHERE blueprint_id = 'ef859228-c087-4193-8afc-c0ac9d02724d';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '41099781-8586-4783-9d1c-b515a386fe9f'
WHERE blueprint_id = '17c32a72-0b7a-42bb-b144-7d7358509cde';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'e9fc2417-6a5b-489d-b82e-42942535af90'
WHERE blueprint_id = 'bf114d6b-bb1c-4de3-8abd-d2367715db52';

-- mechs
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '02ba91b7-55dc-450a-9fbd-e7337ae97a2b'
WHERE blueprint_id = '7b31fdbe-d8b3-4c0a-92f4-d85b434388fa';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '0639ebde-fbba-498b-88ac-f7122ead9c90'
WHERE blueprint_id = '1827934d-36dd-4648-84b8-2541532d1706';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '3dc5888b-f5ff-4d08-a520-26fd3681707f'
WHERE blueprint_id = '91e344be-d3da-41e4-bb9a-5a6e5a09ece5';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '5d3a973b-c62b-4438-b746-d3de2699d42a'
WHERE blueprint_id = '8dd45a32-c201-41d4-a134-4b5a50419a6e';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '5d3a973b-c62b-4438-b746-d3de2699d42a'
WHERE blueprint_id = '2b0df453-a9ac-4a43-967e-ed7a1cdaecf1';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '5d3a973b-c62b-4438-b746-d3de2699d42a'
WHERE blueprint_id = 'cc0de847-24a3-4764-bbbd-78304e79150e';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '5d3a973b-c62b-4438-b746-d3de2699d42a'
WHERE blueprint_id = '74e3bd32-4b27-4038-8338-f4f60db4be71';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '5d3a973b-c62b-4438-b746-d3de2699d42a'
WHERE blueprint_id = '8b98d84c-48ad-481b-bc86-d1d02822f16c';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '5d3a973b-c62b-4438-b746-d3de2699d42a'
WHERE blueprint_id = '12ab349f-0831-4f3f-9aa1-f3aade13abb3';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '5d3a973b-c62b-4438-b746-d3de2699d42a'
WHERE blueprint_id = '014371f7-bca3-4a1c-992e-f55ac5721de1';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '5d3a973b-c62b-4438-b746-d3de2699d42a'
WHERE blueprint_id = 'd041e6dc-0ed5-4294-8fdc-f49707e1a854';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '5d3a973b-c62b-4438-b746-d3de2699d42a'
WHERE blueprint_id = '76a60a59-291a-49e2-b7d4-d7cbfa2a3feb';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '5d3a973b-c62b-4438-b746-d3de2699d42a'
WHERE blueprint_id = '0dd254a6-edd6-481b-b3d7-36c57a9836de';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '5d3a973b-c62b-4438-b746-d3de2699d42a'
WHERE blueprint_id = '11d09723-a9e2-4c57-bb01-5291be841cb7';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '5d3a973b-c62b-4438-b746-d3de2699d42a'
WHERE blueprint_id = '3e710f5f-1290-428d-a6c1-1acc08f98837';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '5d3a973b-c62b-4438-b746-d3de2699d42a'
WHERE blueprint_id = 'd5c07ee1-8213-43bf-bd65-bbacb442d4ff';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '5d3a973b-c62b-4438-b746-d3de2699d42a'
WHERE blueprint_id = '3e469099-7fba-4a17-a392-8fd850d35c28';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '5d3a973b-c62b-4438-b746-d3de2699d42a'
WHERE blueprint_id = '88f4be7c-20f6-48a9-9d56-8f34ebc4607c';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '5d3a973b-c62b-4438-b746-d3de2699d42a'
WHERE blueprint_id = 'cfb96c42-873d-43e8-9060-ae107f18037a';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '5d3a973b-c62b-4438-b746-d3de2699d42a'
WHERE blueprint_id = 'ad124964-e062-4ab8-9cce-e0309fd6b31d';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '625cd381-7c66-4e2f-9f69-f81589105730'
WHERE blueprint_id = 'e1dc2da7-78c4-400e-a302-a40b2d27a3c8';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '625cd381-7c66-4e2f-9f69-f81589105730'
WHERE blueprint_id = 'f9cfa25b-02bc-4a3a-9a31-9a8a4a073714';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '625cd381-7c66-4e2f-9f69-f81589105730'
WHERE blueprint_id = '8a49a4ca-4a61-4f51-b296-b421c951c1f5';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '625cd381-7c66-4e2f-9f69-f81589105730'
WHERE blueprint_id = '276f62c9-23f3-443c-8fad-90fc4d074009';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '625cd381-7c66-4e2f-9f69-f81589105730'
WHERE blueprint_id = '4f6180a5-4d0c-44ca-9422-a4d28c3b863a';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '625cd381-7c66-4e2f-9f69-f81589105730'
WHERE blueprint_id = '1d76a559-3c6d-4f26-a108-a3b5549e35ab';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '625cd381-7c66-4e2f-9f69-f81589105730'
WHERE blueprint_id = 'd0c291de-1edb-4b54-9559-ed3587034051';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '625cd381-7c66-4e2f-9f69-f81589105730'
WHERE blueprint_id = 'ac408b39-ba91-4a0e-9414-a37af0ea6314';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '625cd381-7c66-4e2f-9f69-f81589105730'
WHERE blueprint_id = '80bf637e-9219-44fc-9907-1f58598d6800';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '625cd381-7c66-4e2f-9f69-f81589105730'
WHERE blueprint_id = '9c4cd43d-35ed-4d3a-860c-d7ff00093dd6';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '625cd381-7c66-4e2f-9f69-f81589105730'
WHERE blueprint_id = '17a1426a-b17e-4734-b845-588a87b6d8cd';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '625cd381-7c66-4e2f-9f69-f81589105730'
WHERE blueprint_id = 'f4da1432-0618-47c2-9337-9cda3e8125ba';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '625cd381-7c66-4e2f-9f69-f81589105730'
WHERE blueprint_id = 'bc19d28d-3fa4-45e5-86dd-1cf69bc3daee';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '625cd381-7c66-4e2f-9f69-f81589105730'
WHERE blueprint_id = '0b297a84-0a07-4e09-87ba-ed2a877d7c4f';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '625cd381-7c66-4e2f-9f69-f81589105730'
WHERE blueprint_id = '81f1fd1b-458c-499b-bbec-7560f8bea842';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '625cd381-7c66-4e2f-9f69-f81589105730'
WHERE blueprint_id = 'b2c43cd1-86ce-4154-b3d5-22776ed3a727';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = '7068ab3e-89dc-4ac1-bcbb-1089096a5eda'
WHERE blueprint_id = 'b93f249a-683f-4d04-aebd-7026e10f9a67';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'ac27f3b9-753d-4ace-84a9-21c041195344'
WHERE blueprint_id = 'f9ca786e-730d-481e-a6c9-f04e5ed975d2';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'ac27f3b9-753d-4ace-84a9-21c041195344'
WHERE blueprint_id = 'a5966250-4972-425f-b2d4-433e3820741a';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'ac27f3b9-753d-4ace-84a9-21c041195344'
WHERE blueprint_id = '24a4ebc1-e785-4b72-83e8-c8cf624801d0';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'ac27f3b9-753d-4ace-84a9-21c041195344'
WHERE blueprint_id = '93a65506-44aa-4d93-b184-7d72b8cf5d9b';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'ac27f3b9-753d-4ace-84a9-21c041195344'
WHERE blueprint_id = '9bc4611a-0da2-4b07-bab6-7f15b1c97f90';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'ac27f3b9-753d-4ace-84a9-21c041195344'
WHERE blueprint_id = 'bd283770-0d77-45ea-b4d2-6286351eecc7';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'ac27f3b9-753d-4ace-84a9-21c041195344'
WHERE blueprint_id = '337b3c82-61ae-4959-bcf0-50f0985f8ed6';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'ac27f3b9-753d-4ace-84a9-21c041195344'
WHERE blueprint_id = '0a5b3678-f6c0-4f53-8a1c-e3c29f3a1632';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'ac27f3b9-753d-4ace-84a9-21c041195344'
WHERE blueprint_id = '5d213b5d-5d29-4e3d-bdf8-d9867c574d7f';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'ac27f3b9-753d-4ace-84a9-21c041195344'
WHERE blueprint_id = 'beb40230-580e-4aa0-87c9-0aac27edcbb6';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'ac27f3b9-753d-4ace-84a9-21c041195344'
WHERE blueprint_id = 'a8e507cd-f874-4e6a-a321-a4b04e171ba9';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'ac27f3b9-753d-4ace-84a9-21c041195344'
WHERE blueprint_id = '500128ca-7544-4250-aad4-aa1ae8a8ad1a';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'ac27f3b9-753d-4ace-84a9-21c041195344'
WHERE blueprint_id = '683c4461-a291-4245-9120-56c67f091fba';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'ac27f3b9-753d-4ace-84a9-21c041195344'
WHERE blueprint_id = '35deaa58-d05d-4a8c-8a84-178358c46647';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'ac27f3b9-753d-4ace-84a9-21c041195344'
WHERE blueprint_id = '1ca58e88-6280-427c-b6d2-25928e7bd292';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id,  blueprint_id = 'ac27f3b9-753d-4ace-84a9-21c041195344'
WHERE blueprint_id = 'a5806303-f397-44d5-b9d8-f435500362e6';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'ac27f3b9-753d-4ace-84a9-21c041195344'
WHERE blueprint_id = '30b40880-ffa3-4a17-883d-08de0bf1b479';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'df1ac803-0a90-4631-b9e0-b62a44bdadff'
WHERE blueprint_id = '5938ce12-6965-4977-8700-98b0909df1fe';
UPDATE mystery_crate_blueprints
SET blueprint_id_old = blueprint_id, blueprint_id = 'fc9546d0-9682-468e-af1f-24eb1735315b'
WHERE blueprint_id = '6d2ecfdd-abd2-4147-abd8-67429ea0e37f';


-- update existing mechs
ALTER TABLE mechs
    DROP CONSTRAINT IF EXISTS chassis_blueprint_id_fkey,
    ADD COLUMN blueprint_id_old UUID;


WITH toupdate AS (SELECT m.id
                  FROM mechs m
                           INNER JOIN blueprint_mechs_old bpmo ON bpmo.id = m.blueprint_id)
UPDATE mechs m
SET blueprint_id_old = blueprint_id, blueprint_id = (SELECT bpmo.model_id
                                                     FROM blueprint_mechs_old bpmo
                                                     WHERE bpmo.id = m.blueprint_id)
FROM toupdate
WHERE m.id = toupdate.id;

ALTER TABLE mechs
    ADD CONSTRAINT chassis_blueprint_id_fkey FOREIGN KEY (blueprint_id) REFERENCES blueprint_mechs (id);

-- update existing weapons
ALTER TABLE weapons
    DROP CONSTRAINT IF EXISTS weapons_blueprint_id_fkey,
    ADD COLUMN blueprint_id_old UUID;

WITH toupdate AS (SELECT w.id
                  FROM weapons w
                           INNER JOIN blueprint_weapons_old bwo ON bwo.id = w.blueprint_id)
UPDATE weapons w
SET blueprint_id_old = w.blueprint_id, blueprint_id = (SELECT bpwo.weapon_model_id
                                                       FROM blueprint_weapons_old bpwo
                                                       WHERE bpwo.id = w.blueprint_id)
FROM toupdate
WHERE w.id = toupdate.id;

ALTER TABLE weapons
    ADD CONSTRAINT weapons_blueprint_id_fkey FOREIGN KEY (blueprint_id) REFERENCES blueprint_weapons (id);

ALTER TABLE utility
    ADD COLUMN blueprint_id_old UUID;

-- rm faction id 98bf7bb3-1a7c-4f21-8843-458d62884060
-- RM 0551d044-b8ff-47ac-917e-80c3fce37378
-- BC d429be75-6f98-4231-8315-a86db8477d05
-- bc faction id 7c6dde21-b067-46cf-9e56-155c88a520e2

UPDATE blueprint_utility_shield_old
SET deleted_at = NOW()
WHERE blueprint_utility_id NOT IN (
                                   'd429be75-6f98-4231-8315-a86db8477d05',
                                   '1e9a8bd4-b6c3-4a46-86e9-4c68a95f09b8',
                                   '0551d044-b8ff-47ac-917e-80c3fce37378'
    );

UPDATE blueprint_utility
SET deleted_at = NOW()
WHERE id NOT IN (
                 'd429be75-6f98-4231-8315-a86db8477d05',
                 '1e9a8bd4-b6c3-4a46-86e9-4c68a95f09b8',
                 '0551d044-b8ff-47ac-917e-80c3fce37378'
    );

-- below are tables that are completely not used
DROP TABLE IF EXISTS blueprint_utility_accelerator;
DROP TABLE IF EXISTS blueprint_utility_anti_missile;
DROP TABLE IF EXISTS blueprint_utility_attack_drone;
DROP TABLE IF EXISTS blueprint_utility_repair_drone;
DROP TABLE IF EXISTS utility_accelerator;
DROP TABLE IF EXISTS utility_anti_missile;
DROP TABLE IF EXISTS utility_attack_drone;
DROP TABLE IF EXISTS utility_repair_drone;

WITH toupdate AS (SELECT tbp.id, t.label
                  FROM templates t
                           INNER JOIN template_blueprints tbp ON tbp.template_id = t.id
                  WHERE tbp.type = 'UTILITY'
                    AND t.label ILIKE '%Boston%')
UPDATE template_blueprints tbp
SET blueprint_id_old = blueprint_id, blueprint_id = 'd429be75-6f98-4231-8315-a86db8477d05'
FROM toupdate
WHERE tbp.id = toupdate.id;

WITH toupdate AS (SELECT tbp.id, t.label
                  FROM templates t
                           INNER JOIN template_blueprints tbp ON tbp.template_id = t.id
                  WHERE tbp.type = 'UTILITY'
                    AND t.label ILIKE '%red%')
UPDATE template_blueprints tbp
SET blueprint_id_old = blueprint_id, blueprint_id = '0551d044-b8ff-47ac-917e-80c3fce37378'
FROM toupdate
WHERE tbp.id = toupdate.id;

WITH toupdate AS (
    SELECT tbp.id, t.label
    FROM templates t
             INNER JOIN template_blueprints tbp ON tbp.template_id = t.id
    WHERE tbp.type = 'UTILITY'
      AND t.label ILIKE '%zai%'
)
UPDATE template_blueprints tbp
SET blueprint_id_old = blueprint_id, blueprint_id = '1e9a8bd4-b6c3-4a46-86e9-4c68a95f09b8'
FROM toupdate
WHERE tbp.id = toupdate.id;

ALTER TABLE utility_shield
    RENAME TO utility_shield_dont_use;

ALTER TABLE utility
    RENAME COLUMN label TO label_dont_use;
ALTER TABLE utility
    RENAME COLUMN brand_id TO brand_dont_use;
