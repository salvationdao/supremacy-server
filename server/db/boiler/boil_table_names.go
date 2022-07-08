// Code generated by SQLBoiler 4.8.6 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package boiler

var TableNames = struct {
	Ammo                                               string
	BattleAbilities                                    string
	BattleAbilityTriggers                              string
	BattleContracts                                    string
	BattleContributions                                string
	BattleEvents                                       string
	BattleEventsGameAbility                            string
	BattleEventsState                                  string
	BattleEventsWarMachineDestroyed                    string
	BattleEventsWarMachineDestroyedAssistedWarMachines string
	BattleHistory                                      string
	BattleKills                                        string
	BattleMechs                                        string
	BattleQueue                                        string
	BattleQueueNotifications                           string
	BattleViewers                                      string
	BattleWarMachineQueues                             string
	BattleWins                                         string
	Battles                                            string
	Blobs                                              string
	BlockMarketplace                                   string
	BlueprintAmmo                                      string
	BlueprintKeycards                                  string
	BlueprintMechAnimation                             string
	BlueprintMechSkin                                  string
	BlueprintMechs                                     string
	BlueprintPlayerAbilities                           string
	BlueprintPowerCores                                string
	BlueprintUtility                                   string
	BlueprintUtilityAccelerator                        string
	BlueprintUtilityAntiMissile                        string
	BlueprintUtilityAttackDrone                        string
	BlueprintUtilityRepairDrone                        string
	BlueprintUtilityShield                             string
	BlueprintWeaponSkin                                string
	BlueprintWeapons                                   string
	Brands                                             string
	ChatBannedFingerprints                             string
	ChatHistory                                        string
	CollectionItems                                    string
	ConsumedAbilities                                  string
	CouponItems                                        string
	Coupons                                            string
	FactionStats                                       string
	Factions                                           string
	FailedPlayerKeycardsSync                           string
	FingerprintIps                                     string
	Fingerprints                                       string
	GameAbilities                                      string
	GameMaps                                           string
	GlobalAnnouncements                                string
	ItemKeycardSales                                   string
	ItemSales                                          string
	ItemSalesBidHistory                                string
	KV                                                 string
	Languages                                          string
	MarketplaceEvents                                  string
	MechAnimation                                      string
	MechModels                                         string
	MechMoveCommandLogs                                string
	MechRepair                                         string
	MechSkin                                           string
	MechStats                                          string
	MechUtility                                        string
	MechWeapons                                        string
	Mechs                                              string
	MechsOld                                           string
	Multipliers                                        string
	MysteryCrate                                       string
	MysteryCrateBlueprints                             string
	PlayerAbilities                                    string
	PlayerActiveLogs                                   string
	PlayerBans                                         string
	PlayerFingerprints                                 string
	PlayerIps                                          string
	PlayerKeycards                                     string
	PlayerKillLog                                      string
	PlayerLanguages                                    string
	PlayerMultipliers                                  string
	PlayerPreferences                                  string
	PlayerSettingsPreferences                          string
	PlayerSpoilsOfWar                                  string
	PlayerStats                                        string
	Players                                            string
	PlayersPunishVotes                                 string
	PowerCores                                         string
	Profanities                                        string
	PunishOptions                                      string
	PunishVoteInstantPassRecords                       string
	PunishVotes                                        string
	SalePlayerAbilities                                string
	SchemaMigrations                                   string
	SpoilsOfWar                                        string
	StorePurchaseHistory                               string
	StorefrontMysteryCrates                            string
	StreamList                                         string
	TelegramNotifications                              string
	TemplateBlueprints                                 string
	Templates                                          string
	TemplatesOld                                       string
	Utility                                            string
	UtilityAccelerator                                 string
	UtilityAntiMissile                                 string
	UtilityAttackDrone                                 string
	UtilityRepairDrone                                 string
	UtilityShield                                      string
	WeaponAmmo                                         string
	WeaponModels                                       string
	WeaponSkin                                         string
	Weapons                                            string
}{
	Ammo:                            "ammo",
	BattleAbilities:                 "battle_abilities",
	BattleAbilityTriggers:           "battle_ability_triggers",
	BattleContracts:                 "battle_contracts",
	BattleContributions:             "battle_contributions",
	BattleEvents:                    "battle_events",
	BattleEventsGameAbility:         "battle_events_game_ability",
	BattleEventsState:               "battle_events_state",
	BattleEventsWarMachineDestroyed: "battle_events_war_machine_destroyed",
	BattleEventsWarMachineDestroyedAssistedWarMachines: "battle_events_war_machine_destroyed_assisted_war_machines",
	BattleHistory:                "battle_history",
	BattleKills:                  "battle_kills",
	BattleMechs:                  "battle_mechs",
	BattleQueue:                  "battle_queue",
	BattleQueueNotifications:     "battle_queue_notifications",
	BattleViewers:                "battle_viewers",
	BattleWarMachineQueues:       "battle_war_machine_queues",
	BattleWins:                   "battle_wins",
	Battles:                      "battles",
	Blobs:                        "blobs",
	BlockMarketplace:             "block_marketplace",
	BlueprintAmmo:                "blueprint_ammo",
	BlueprintKeycards:            "blueprint_keycards",
	BlueprintMechAnimation:       "blueprint_mech_animation",
	BlueprintMechSkin:            "blueprint_mech_skin",
	BlueprintMechs:               "blueprint_mechs",
	BlueprintPlayerAbilities:     "blueprint_player_abilities",
	BlueprintPowerCores:          "blueprint_power_cores",
	BlueprintUtility:             "blueprint_utility",
	BlueprintUtilityAccelerator:  "blueprint_utility_accelerator",
	BlueprintUtilityAntiMissile:  "blueprint_utility_anti_missile",
	BlueprintUtilityAttackDrone:  "blueprint_utility_attack_drone",
	BlueprintUtilityRepairDrone:  "blueprint_utility_repair_drone",
	BlueprintUtilityShield:       "blueprint_utility_shield",
	BlueprintWeaponSkin:          "blueprint_weapon_skin",
	BlueprintWeapons:             "blueprint_weapons",
	Brands:                       "brands",
	ChatBannedFingerprints:       "chat_banned_fingerprints",
	ChatHistory:                  "chat_history",
	CollectionItems:              "collection_items",
	ConsumedAbilities:            "consumed_abilities",
	CouponItems:                  "coupon_items",
	Coupons:                      "coupons",
	FactionStats:                 "faction_stats",
	Factions:                     "factions",
	FailedPlayerKeycardsSync:     "failed_player_keycards_sync",
	FingerprintIps:               "fingerprint_ips",
	Fingerprints:                 "fingerprints",
	GameAbilities:                "game_abilities",
	GameMaps:                     "game_maps",
	GlobalAnnouncements:          "global_announcements",
	ItemKeycardSales:             "item_keycard_sales",
	ItemSales:                    "item_sales",
	ItemSalesBidHistory:          "item_sales_bid_history",
	KV:                           "kv",
	Languages:                    "languages",
	MarketplaceEvents:            "marketplace_events",
	MechAnimation:                "mech_animation",
	MechModels:                   "mech_models",
	MechMoveCommandLogs:          "mech_move_command_logs",
	MechRepair:                   "mech_repair",
	MechSkin:                     "mech_skin",
	MechStats:                    "mech_stats",
	MechUtility:                  "mech_utility",
	MechWeapons:                  "mech_weapons",
	Mechs:                        "mechs",
	MechsOld:                     "mechs_old",
	Multipliers:                  "multipliers",
	MysteryCrate:                 "mystery_crate",
	MysteryCrateBlueprints:       "mystery_crate_blueprints",
	PlayerAbilities:              "player_abilities",
	PlayerActiveLogs:             "player_active_logs",
	PlayerBans:                   "player_bans",
	PlayerFingerprints:           "player_fingerprints",
	PlayerIps:                    "player_ips",
	PlayerKeycards:               "player_keycards",
	PlayerKillLog:                "player_kill_log",
	PlayerLanguages:              "player_languages",
	PlayerMultipliers:            "player_multipliers",
	PlayerPreferences:            "player_preferences",
	PlayerSettingsPreferences:    "player_settings_preferences",
	PlayerSpoilsOfWar:            "player_spoils_of_war",
	PlayerStats:                  "player_stats",
	Players:                      "players",
	PlayersPunishVotes:           "players_punish_votes",
	PowerCores:                   "power_cores",
	Profanities:                  "profanities",
	PunishOptions:                "punish_options",
	PunishVoteInstantPassRecords: "punish_vote_instant_pass_records",
	PunishVotes:                  "punish_votes",
	SalePlayerAbilities:          "sale_player_abilities",
	SchemaMigrations:             "schema_migrations",
	SpoilsOfWar:                  "spoils_of_war",
	StorePurchaseHistory:         "store_purchase_history",
	StorefrontMysteryCrates:      "storefront_mystery_crates",
	StreamList:                   "stream_list",
	TelegramNotifications:        "telegram_notifications",
	TemplateBlueprints:           "template_blueprints",
	Templates:                    "templates",
	TemplatesOld:                 "templates_old",
	Utility:                      "utility",
	UtilityAccelerator:           "utility_accelerator",
	UtilityAntiMissile:           "utility_anti_missile",
	UtilityAttackDrone:           "utility_attack_drone",
	UtilityRepairDrone:           "utility_repair_drone",
	UtilityShield:                "utility_shield",
	WeaponAmmo:                   "weapon_ammo",
	WeaponModels:                 "weapon_models",
	WeaponSkin:                   "weapon_skin",
	Weapons:                      "weapons",
}
