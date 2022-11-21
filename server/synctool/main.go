package synctool

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"github.com/shopspring/decimal"
	"io"
	"log"
	"net/http"
	"os"
	"server/db/boiler"
	"server/synctool/types"
	"strconv"
	"strings"
	"time"

	"github.com/volatiletech/null/v8"

	"github.com/volatiletech/sqlboiler/v4/boil"
)

type StaticSyncTool struct {
	DB       *sql.DB
	FilePath string
}

func SyncTool(dt *StaticSyncTool) error {

	f, err := readFile(fmt.Sprintf("%sbattle_arena.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncBattleArenas(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	f, err = readFile(fmt.Sprintf("%sfactions.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncFactions(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	f, err = readFile(fmt.Sprintf("%sfaction_palettes.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncFactionPalettes(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	f, err = readFile(fmt.Sprintf("%sfaction_passes.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncFactionPasses(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	f, err = readFile(fmt.Sprintf("%sbrands.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncBrands(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	f, err = readFile(fmt.Sprintf("%sshield_types.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncShieldTypes(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	f, err = readFile(fmt.Sprintf("%sweapon_skins.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncWeaponSkins(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	f, err = readFile(fmt.Sprintf("%smech_skins.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncMechSkins(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	f, err = readFile(fmt.Sprintf("%smechs.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncMechModels(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	f, err = readFile(fmt.Sprintf("%smech_model_skin_compatibilities.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncMechModelSkinCompatibilities(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	//err = SyncMysteryCrates(f, dt.DB)
	//if err != nil {
	//	return err
	//}

	f, err = readFile(fmt.Sprintf("%sweapons.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncWeaponModel(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	f, err = readFile(fmt.Sprintf("%sweapon_model_skin_compatibilities.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncWeaponModelSkinCompatibilities(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	f, err = readFile(fmt.Sprintf("%sbattle_abilities.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncBattleAbilities(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	f, err = readFile(fmt.Sprintf("%sgame_abilities.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncGameAbilities(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	f, err = readFile(fmt.Sprintf("%splayer_abilities.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncPlayerAbilities(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	f, err = readFile(fmt.Sprintf("%spower_cores.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncPowerCores(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	f, err = readFile(fmt.Sprintf("%squests.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncStaticQuest(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	return nil
}

func DownloadFile(ctx context.Context, url string, timout time.Duration) (io.Reader, error) {
	// Set up client
	client := &http.Client{
		Timeout: timout,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Writer the body to file
	readr := bufio.NewReader(resp.Body)

	return readr, nil
}

func readFile(fileName string) (*os.File, error) {
	f, err := os.OpenFile(fileName, os.O_RDONLY, os.ModeAppend)
	if err != nil {
		log.Fatalf("CANT OPEN FILE: %s", fileName)
		return nil, err
	}
	return f, nil
}

func RemoveFKContraints(dt StaticSyncTool) error {
	_, err := dt.DB.Exec(
		`
			ALTER TABLE mech_models DROP CONSTRAINT mech_model_default_chassis_skin_id_fkey;
			ALTER TABLE mech_models ADD CONSTRAINT mech_model_default_chassis_skin_id_fkey FOREIGN KEY (default_chassis_skin_id) REFERENCES blueprint_mech_skin(id) ON UPDATE CASCADE;

			ALTER TABLE mech_models DROP CONSTRAINT fk_brands;
			ALTER TABLE mech_models ADD CONSTRAINT fk_brands FOREIGN KEY (brand_id) REFERENCES brands(id) ON UPDATE CASCADE;

			ALTER TABLE blueprint_weapons DROP CONSTRAINT blueprint_weapons_brand_id_fkey;
			ALTER TABLE blueprint_weapons ADD CONSTRAINT blueprint_weapons_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES brands(id) ON UPDATE CASCADE ;

			ALTER TABLE weapon_models DROP CONSTRAINT weapon_models_brand_id_fkey;
			ALTER TABLE weapon_models ADD CONSTRAINT weapon_models_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES brands(id) ON UPDATE CASCADE ;

			ALTER TABLE blueprint_mech_skin DROP CONSTRAINT blueprint_chassis_skin_mech_model_fkey;

			ALTER TABLE mech_skin DROP CONSTRAINT chassis_skin_blueprint_id_fkey;
			ALTER TABLE mech_skin ADD CONSTRAINT chassis_skin_blueprint_id_fkey FOREIGN KEY (blueprint_id) REFERENCES blueprint_mech_skin(id) ON UPDATE CASCADE;

			ALTER TABLE mech_skin DROP CONSTRAINT chassis_skin_mech_model_fkey;
			ALTER TABLE mech_skin ADD CONSTRAINT chassis_skin_mech_model_fkey FOREIGN KEY (mech_model) REFERENCES mech_models(id) ON UPDATE CASCADE;

			ALTER TABLE mechs DROP CONSTRAINT chassis_model_id_fkey;
			ALTER TABLE mechs ADD CONSTRAINT chassis_model_id_fkey FOREIGN KEY (model_id) REFERENCES mech_models(id) ON UPDATE CASCADE;

			ALTER TABLE blueprint_mechs DROP CONSTRAINT blueprint_chassis_model_id_fkey;
			ALTER TABLE blueprint_mechs ADD CONSTRAINT blueprint_chassis_model_id_fkey FOREIGN KEY (model_id) REFERENCES mech_models(id) ON UPDATE CASCADE;

			ALTER TABLE blueprint_mechs DROP CONSTRAINT blueprint_chassis_brand_id_fkey;
			ALTER TABLE blueprint_mechs ADD CONSTRAINT blueprint_chassis_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES brands(id) ON UPDATE CASCADE;

			ALTER TABLE mech_skin DROP CONSTRAINT chassis_skin_blueprint_id_fkey;
			ALTER TABLE mech_skin ADD CONSTRAINT chassis_skin_blueprint_id_fkey FOREIGN KEY (blueprint_id) REFERENCES blueprint_mech_skin(id) ON UPDATE CASCADE;

			ALTER TABLE blueprint_weapon_skin DROP CONSTRAINT blueprint_weapon_skin_weapon_model_id_fkey;
			ALTER TABLE blueprint_weapon_skin ADD CONSTRAINT blueprint_weapon_skin_weapon_model_id_fkey FOREIGN KEY (weapon_model_id) REFERENCES weapon_models(id) ON UPDATE CASCADE;

			ALTER TABLE blueprint_weapons DROP CONSTRAINT blueprint_weapons_weapon_model_id_fkey;
			ALTER TABLE blueprint_weapons ADD CONSTRAINT  blueprint_weapons_weapon_model_id_fkey FOREIGN KEY (weapon_model_id) REFERENCES weapon_models(id) ON UPDATE CASCADE;

			ALTER TABLE weapons DROP CONSTRAINT fk_weapon_models;
			ALTER TABLE weapons ADD CONSTRAINT fk_weapon_models FOREIGN KEY (weapon_model_id) REFERENCES weapon_models(id) ON UPDATE CASCADE;

			ALTER TABLE weapon_models DROP CONSTRAINT fk_weapon_model_default_skin;
			ALTER TABLE weapon_models ADD CONSTRAINT fk_weapon_model_default_skin FOREIGN KEY (default_skin_id) REFERENCES blueprint_weapon_skin(id) ON UPDATE CASCADE;
			`,
	)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("Finished removing constraints")

	return nil
}

func SyncBattleArenas(f io.Reader, db *sql.DB) error {

	r := csv.NewReader(f)
	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	for _, record := range records {
		battleArena := &boiler.BattleArena{
			ID: record[0],
		}

		if record[1] != "" {
			battleArena.DeletedAt = null.TimeFrom(time.Now())
		}

		err = battleArena.Upsert(db, false, []string{boiler.BattleArenaColumns.ID}, boil.Whitelist(boiler.BattleArenaColumns.DeletedAt), boil.Infer())
		if err != nil {
			fmt.Println(err.Error(), battleArena.ID)
			return err
		}

		fmt.Println("UPDATED: " + battleArena.ID)
	}

	fmt.Println("Finish syncing battle arenas")

	return nil
}

func SyncMechModels(f io.Reader, db *sql.DB) error {

	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	var MechModels []types.MechModel
	for _, record := range records {
		mechModel := &types.MechModel{
			ID:                      record[0],
			Label:                   record[1],
			DefaultChassisSkinID:    record[3],
			BrandID:                 record[4],
			MechType:                record[5],
			BoostStat:               record[6],
			WeaponHardpoints:        record[7],
			UtilitySlots:            record[8],
			Speed:                   record[9],
			MaxHitpoints:            record[10],
			PowerCoreSize:           record[11],
			Collection:              record[12],
			AvailabilityID:          null.NewString(record[13], record[13] != ""),
			ShieldMax:               record[14],
			ShieldRechargeRate:      record[15],
			ShieldRechargePowerCost: record[16],
			ShieldTypeID:            record[17],
			ShieldRechargeDelay:     record[18],
			HeightMeters:            record[19],
			WalkSpeedModifier:       record[20],
			SprintSpreadModifier:    record[21],
			IdleDrain:               record[22],
			WalkDrain:               record[23],
			RunDrain:                record[24],
		}

		MechModels = append(MechModels, *mechModel)
	}

	count := 0

	for _, mechModel := range MechModels {
		_, err = db.Exec(`
			INSERT INTO blueprint_mechs (
			                                   id, 
			                                   label, 
			                                   default_chassis_skin_id, 
			                                   brand_id, 
			                                   mech_type,
												boost_stat,
												weapon_hardpoints,
												utility_slots,
												speed,
												max_hitpoints,
												power_core_size,
			                             		collection,
			                             		availability_id,
												shield_max,
												shield_recharge_rate,
												shield_recharge_power_cost,
			                             		shield_type_id,
			                             		shield_recharge_delay,
			                             		height_meters,
												walk_speed_modifier,
												sprint_spread_modifier,
												idle_drain,
												walk_drain,
												run_drain
			                                   )
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)
			ON CONFLICT (id)
			DO
				UPDATE SET 
				           id=$1, 
				           label=$2, 
				           default_chassis_skin_id=$3, 
				           brand_id=$4, 
				           mech_type=$5,
							boost_stat=$6,
							weapon_hardpoints=$7,
							utility_slots=$8,
							speed=$9,
							max_hitpoints=$10,
							power_core_size=$11,
							collection=$12,
							availability_id=$13,
							shield_max=$14,
							shield_recharge_rate=$15,
							shield_recharge_power_cost=$16,
							shield_type_id=$17,
							shield_recharge_delay=$18,
							height_meters=$19,
							walk_speed_modifier=$20,
							sprint_spread_modifier=$21,
							idle_drain=$22,
							walk_drain=$23,
							run_drain=$24;
		`,
			mechModel.ID,
			mechModel.Label,
			mechModel.DefaultChassisSkinID,
			mechModel.BrandID,
			mechModel.MechType,
			mechModel.BoostStat,
			mechModel.WeaponHardpoints,
			mechModel.UtilitySlots,
			mechModel.Speed,
			mechModel.MaxHitpoints,
			mechModel.PowerCoreSize,
			mechModel.Collection,
			mechModel.AvailabilityID,
			mechModel.ShieldMax,
			mechModel.ShieldRechargeRate,
			mechModel.ShieldRechargePowerCost,
			mechModel.ShieldTypeID,
			mechModel.ShieldRechargeDelay,
			mechModel.HeightMeters,
			mechModel.WalkSpeedModifier,
			mechModel.SprintSpreadModifier,
			mechModel.IdleDrain,
			mechModel.WalkDrain,
			mechModel.RunDrain,
		)
		if err != nil {
			fmt.Println("ERROR: " + err.Error())
			return err
		}
		count++
		fmt.Println("UPDATED: " + mechModel.Label)
	}

	fmt.Println("Finish syncing mech models Count: " + strconv.Itoa(count))

	return nil
}

func SyncMechModelSkinCompatibilities(f io.Reader, db *sql.DB) error {

	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	var mechModelSkinCompatibities []*types.MechModelSkinCompatibility
	for _, record := range records {
		mechModelSkinCompatibity := &types.MechModelSkinCompatibility{
			MechSkinID:       record[0],
			MechModelID:      record[1],
			ImageUrl:         record[2],
			AnimationUrl:     record[3],
			CardAnimationUrl: record[4],
			LargeImageUrl:    record[5],
			AvatarUrl:        record[6],
			BackgroundColor:  record[8],
			YoutubeUrl:       record[9],
		}

		mechModelSkinCompatibities = append(mechModelSkinCompatibities, mechModelSkinCompatibity)
	}

	count := 0

	for _, mechModelSkinCompat := range mechModelSkinCompatibities {
		_, err = db.Exec(`
			INSERT INTO mech_model_skin_compatibilities (
												blueprint_mech_skin_id,
												mech_model_id,
												image_url,
											 	animation_url,
												card_animation_url,
											 	large_image_url,
												avatar_url,
												background_color,
												youtube_url
			)
			VALUES ($1,$2,$3,$4,$5, $6, $7, $8, $9)
			ON CONFLICT (blueprint_mech_skin_id, mech_model_id)
			DO
				UPDATE SET 	blueprint_mech_skin_id = $1,
							mech_model_id = $2,
							image_url = $3,
							animation_url = $4,
							card_animation_url = $5,
							large_image_url = $6,
							avatar_url = $7,
							background_color = $8,
							youtube_url = $9;
		`,
			mechModelSkinCompat.MechSkinID,
			mechModelSkinCompat.MechModelID,
			null.NewString(mechModelSkinCompat.ImageUrl, mechModelSkinCompat.ImageUrl != ""),
			null.NewString(mechModelSkinCompat.AnimationUrl, mechModelSkinCompat.AnimationUrl != ""),
			null.NewString(mechModelSkinCompat.CardAnimationUrl, mechModelSkinCompat.CardAnimationUrl != ""),
			null.NewString(mechModelSkinCompat.LargeImageUrl, mechModelSkinCompat.LargeImageUrl != ""),
			null.NewString(mechModelSkinCompat.AvatarUrl, mechModelSkinCompat.AvatarUrl != ""),
			null.NewString(mechModelSkinCompat.BackgroundColor, mechModelSkinCompat.BackgroundColor != ""),
			null.NewString(mechModelSkinCompat.YoutubeUrl, mechModelSkinCompat.YoutubeUrl != ""),
		)
		if err != nil {
			fmt.Println("ERROR: " + err.Error())
			return err
		}
		count++
		fmt.Printf("UPDATED: %s:%s \n", mechModelSkinCompat.MechSkinID, mechModelSkinCompat.MechModelID)
	}

	fmt.Println("Finish syncing mech_model_skin_compatibilities Count: " + strconv.Itoa(count))

	return nil
}

func SyncShieldTypes(f io.Reader, db *sql.DB) error {
	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	for _, record := range records {
		blueprintShield := &boiler.BlueprintShieldType{
			ID:          record[0],
			Label:       record[1],
			Description: record[2],
		}

		// upsert blueprint quest
		err = blueprintShield.Upsert(
			db,
			true,
			[]string{
				boiler.BlueprintShieldTypeColumns.ID,
			},
			boil.Whitelist(
				boiler.BlueprintShieldTypeColumns.Label,
				boiler.BlueprintShieldTypeColumns.Description,
			),
			boil.Infer(),
		)
		if err != nil {
			fmt.Printf("%s: %s %s", err.Error(), blueprintShield.ID, blueprintShield.Label)
			return err
		}

		fmt.Printf("UPDATED: %s\n", blueprintShield.Label)
	}

	fmt.Println("Finish syncing blueprint shield types")

	return nil

}

func SyncMechSkins(f io.Reader, db *sql.DB) error {
	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	var MechSkins []types.MechSkin
	for _, record := range records {
		mechModel := &types.MechSkin{
			ID:                    record[0],
			Collection:            record[1],
			Label:                 record[2],
			Tier:                  record[3],
			DefaultLevel:          record[5],
			ImageUrl:              null.NewString(record[6], record[6] != ""),
			AnimationUrl:          null.NewString(record[7], record[7] != ""),
			CardAnimationUrl:      null.NewString(record[8], record[8] != ""),
			LargeImageUrl:         null.NewString(record[9], record[9] != ""),
			AvatarUrl:             null.NewString(record[10], record[10] != ""),
			BackgroundColor:       null.NewString(record[11], record[11] != ""),
			YoutubeUrl:            null.NewString(record[12], record[12] != ""),
			BlueprintWeaponSkinID: null.NewString(record[13], record[13] != ""),
		}

		MechSkins = append(MechSkins, *mechModel)
	}

	for _, mechSkin := range MechSkins {
		defaultLevel := null.NewString(mechSkin.DefaultLevel, mechSkin.DefaultLevel != "")

		_, err = db.Exec(`
			INSERT INTO blueprint_mech_skin(
			                                id,
			                                collection,
			                                label,
			                                tier,
			                                default_level,
			                                image_url,
			                                animation_url,
			                                card_animation_url,
			                                large_image_url,
			                                avatar_url,
			                                background_color,
			                                youtube_url,
											blueprint_weapon_skin_id
			                                )
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
			ON CONFLICT (id)
			DO
			    UPDATE SET id=$1,
			               collection=$2,
			               label=$3,
			               tier=$4,
			               default_level=$5,
			               image_url=$6,
			               animation_url=$7,
			               card_animation_url=$8,
			               large_image_url=$9,
			               avatar_url=$10,
			               background_color=$11,
			               youtube_url=$12,
						   blueprint_weapon_skin_id=$13;
		`,
			mechSkin.ID,
			mechSkin.Collection,
			mechSkin.Label,
			mechSkin.Tier,
			defaultLevel,
			mechSkin.ImageUrl,
			mechSkin.AnimationUrl,
			mechSkin.CardAnimationUrl,
			mechSkin.LargeImageUrl,
			mechSkin.AvatarUrl,
			mechSkin.BackgroundColor,
			mechSkin.YoutubeUrl,
			mechSkin.BlueprintWeaponSkinID,
		)
		if err != nil {
			fmt.Println(err.Error()+mechSkin.ID, mechSkin.Label)
			return err
		}

		fmt.Println("UPDATED: "+mechSkin.ID, mechSkin.Label, mechSkin.Tier, mechSkin.Collection, mechSkin.Label)
	}

	fmt.Println("Finish syncing mech skins")

	return nil

}

func SyncFactions(f io.Reader, db *sql.DB) error {
	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	var Factions []types.Faction
	for _, record := range records {
		faction := &types.Faction{
			ID:             record[0],
			ContractReward: record[1],
			VotePrice:      record[2],
			Label:          record[3],
			GuildID:        record[4],
			DeletedAt:      record[5],
			UpdatedAt:      record[6],
			CreatedAt:      record[7],
			LogoURL:        record[8],
			BackgroundURL:  record[9],
			Description:    record[10],
		}

		Factions = append(Factions, *faction)
	}

	for _, faction := range Factions {
		guildID := &faction.GuildID
		if faction.GuildID == "" {
			guildID = nil
		}
		_, err = db.Exec(`
			INSERT INTO factions (id, label, guild_id, logo_url, background_url, description)
			VALUES ($1,$2,$3,$4,$5,$6)
			ON CONFLICT (id)
			DO
				UPDATE SET id=$1, label=$2, guild_id=$3, logo_url=$4, background_url=$5, description=$6;
		`, faction.ID, faction.Label, guildID, faction.LogoURL, faction.BackgroundURL, faction.Description)
		if err != nil {
			fmt.Println(err.Error()+faction.ID, faction.Label)
			return err
		}

		fmt.Println("UPDATED: "+faction.ID, faction.Label)
	}

	fmt.Println("Finish syncing Factions")
	return nil
}

func SyncFactionPalettes(f io.Reader, db *sql.DB) error {
	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	var factionPalettes []types.FactionPalette
	for _, record := range records {
		fp := &types.FactionPalette{
			FactionID:  record[0],
			Primary:    record[1],
			Text:       record[2],
			Background: record[3],
			S100:       record[4],
			S200:       record[5],
			S300:       record[6],
			S400:       record[7],
			S500:       record[8],
			S600:       record[9],
			S700:       record[10],
			S800:       record[11],
			S900:       record[12],
		}

		factionPalettes = append(factionPalettes, *fp)
	}

	for _, fp := range factionPalettes {
		_, err = db.Exec(`
		INSERT INTO faction_palettes (faction_id, "primary", "text", background, s100, s200, s300, s400, s500, s600, s700, s800, s900)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (faction_id)
			DO UPDATE SET
				faction_id = $1,
				"primary" = $2,
				"text" = $3,
				background = $4,
				s100 = $5,
				s200 = $6,
				s300 = $7,
				s400 = $8,
				s500 = $9,
				s600 = $10,
				s700 = $11,
				s800 = $12,
				s900 = $13;	
		`,
			fp.FactionID,  // $1
			fp.Primary,    // $2
			fp.Text,       // $3
			fp.Background, // $4
			fp.S100,       // $5
			fp.S200,       // $6
			fp.S300,       // $7
			fp.S400,       // $8
			fp.S500,       // $9
			fp.S600,       // $10
			fp.S700,       // $11
			fp.S800,       // $12
			fp.S900,       // $13
		)
		if err != nil {
			fmt.Println(err.Error()+fp.FactionID, "color palette")
			return err
		}

		fmt.Println("UPDATED: "+fp.FactionID, "color palette")
	}

	fmt.Println("Finish syncing Faction palettes")
	return nil
}

func SyncFactionPasses(f io.Reader, db *sql.DB) error {
	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	for _, record := range records {
		factionPass := &types.FactionPass{
			ID:        record[0],
			Label:     record[1],
			DeletedAt: null.Time{Valid: record[5] != "", Time: time.Now()},
		}

		factionPass.LastForDays, err = strconv.Atoi(record[2])
		if err != nil {
			return fmt.Errorf("invalid last for days for faction pass")
		}

		factionPass.SupsCost, err = decimal.NewFromString(record[3])
		if err != nil {
			return fmt.Errorf("invalid sups cost for faction pass")
		}

		factionPass.SupsDiscountPercentage, err = decimal.NewFromString(record[4])
		if err != nil {
			return fmt.Errorf("invalid sups cost for faction pass")
		}

		_, err = db.Exec(`
			INSERT INTO faction_passes (id, label, last_for_days, sups_cost, sups_discount_percentage, deleted_at) 
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id)
			DO UPDATE SET label=$2, last_for_days=$3, sups_cost=$4, sups_discount_percentage=$5, deleted_at=$6;
		`,
			factionPass.ID,
			factionPass.Label,
			factionPass.LastForDays,
			factionPass.SupsCost,
			factionPass.SupsDiscountPercentage,
			factionPass.DeletedAt,
		)
		if err != nil {
			fmt.Println(err.Error()+factionPass.ID, factionPass.Label)
			return err
		}

	}
	fmt.Println("Finish syncing Faction Passes")
	return nil
}

func SyncBrands(f io.Reader, db *sql.DB) error {
	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	var Brands []types.Brands
	for _, record := range records {
		brand := &types.Brands{
			ID:        record[0],
			FactionID: record[1],
			Label:     record[2],
		}

		Brands = append(Brands, *brand)
	}

	for _, brand := range Brands {
		_, err = db.Exec(`
			INSERT INTO brands(id, label, faction_id)
			VALUES ($1,$2,$3)
			ON CONFLICT (id)
			DO
				UPDATE SET id=$1, label=$2, faction_id=$3;
		`, brand.ID, brand.Label, brand.FactionID)
		if err != nil {
			fmt.Println(err.Error()+brand.ID, brand.Label, brand.FactionID)
			return err
		}

		fmt.Println("UPDATED: "+brand.ID, brand.Label, brand.FactionID)
	}

	fmt.Println("Finish syncing brands")
	return nil
}

func SyncMysteryCrates(f io.Reader, db *sql.DB) error {
	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	var MysteryCrates []types.MysteryCrate
	for _, record := range records {
		mysteryCrate := &types.MysteryCrate{
			ID:               record[0],
			MysteryCrateType: record[1],
			FactionID:        record[5],
			Label:            record[9],
			Description:      record[10],
			ImageUrl:         record[11],
			CardAnimationUrl: record[12],
			AvatarUrl:        record[13],
			LargeImageUrl:    record[14],
			BackgroundColor:  record[15],
			AnimationUrl:     record[16],
			YoutubeUrl:       record[17],
		}

		MysteryCrates = append(MysteryCrates, *mysteryCrate)
	}

	for _, mysteryCrate := range MysteryCrates {

		imageURL := &mysteryCrate.ImageUrl
		if mysteryCrate.ImageUrl == "" {
			imageURL = nil
		}

		animationURL := &mysteryCrate.AnimationUrl
		if mysteryCrate.AnimationUrl == "" {
			animationURL = nil
		}

		cardAnimationURL := &mysteryCrate.CardAnimationUrl
		if mysteryCrate.CardAnimationUrl == "" {
			cardAnimationURL = nil
		}

		largeImageURL := &mysteryCrate.LargeImageUrl
		if mysteryCrate.LargeImageUrl == "" {
			largeImageURL = nil
		}

		avatarURL := &mysteryCrate.AvatarUrl
		if mysteryCrate.AvatarUrl == "" {
			avatarURL = nil
		}

		youtubeURL := &mysteryCrate.YoutubeUrl
		if mysteryCrate.YoutubeUrl == "" {
			youtubeURL = nil
		}
		_, err = db.Exec(`
			INSERT INTO storefront_mystery_crates (id,mystery_crate_type,faction_id, label, description, image_url, card_animation_url, avatar_url, large_image_url, background_color, animation_url, youtube_url)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			ON CONFLICT (id)
			DO 
			    UPDATE SET id=$1,mystery_crate_type=$2,faction_id=$3, label=$4, description=$5, image_url=$6, card_animation_url=$7, avatar_url=$8, large_image_url=$9, background_color=$10, animation_url=$11, youtube_url=$12;
		`, mysteryCrate.ID, mysteryCrate.MysteryCrateType, mysteryCrate.FactionID, mysteryCrate.Label, mysteryCrate.Description, imageURL, cardAnimationURL, avatarURL, largeImageURL, mysteryCrate.BackgroundColor, animationURL, youtubeURL)
		if err != nil {
			fmt.Println(err.Error()+mysteryCrate.ID, mysteryCrate.Label, mysteryCrate.MysteryCrateType)
			return err
		}

		fmt.Println("UPDATED: "+mysteryCrate.ID, mysteryCrate.Label, mysteryCrate.MysteryCrateType)
	}

	fmt.Println("Finish syncing crates")
	return nil
}

func SyncWeaponModel(f io.Reader, db *sql.DB) error {
	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	var WeaponModels []types.Weapon
	for _, record := range records {
		weaponModel := &types.Weapon{
			ID:                  record[0],
			BrandID:             record[1],
			Label:               record[2],
			WeaponType:          record[3],
			DefaultSkinID:       record[4],
			Damage:              record[5],
			DamageFallOff:       record[6],
			DamageFalloffRate:   record[7],
			Radius:              record[8],
			RadiusDamageFalloff: record[9],
			Spread:              record[10],
			RateOfFire:          record[11],
			ProjectileSpeed:     record[12],
			MaxAmmo:             record[13],
			IsMelee:             record[14],
			PowerCost:           record[15],
			GameClientWeaponID:  null.NewString(record[16], record[16] != ""),
			Collection:          record[17],
			DefaultDamageType:   record[18],
			ProjectileAmount:    record[19],
			DotTickDamage:       record[20],
			DotMaxTicks:         record[21],
			IsArced:             record[22],
			ChargeTimeSeconds:   record[23],
			BurstRateOfFire:     record[24],
			PowerInstantDrain:   record[25],
			DotTickDuration:     record[26],
			ProjectileLifeSpan:  record[27],
			RecoilForce:         record[28],
			IdlePowerCost:       record[29],
		}

		WeaponModels = append(WeaponModels, *weaponModel)
	}

	for _, weaponModel := range WeaponModels {

		_, err = db.Exec(`
			INSERT INTO blueprint_weapons(
										id, 
										brand_id, 
										label, 
										weapon_type, 
										default_skin_id, 
										damage,
										damage_falloff,
										damage_falloff_rate,
										radius,
										radius_damage_falloff,
										spread,
										rate_of_fire,
										projectile_speed,
										max_ammo,
										is_melee,
										power_cost,
										game_client_weapon_id,
										collection,
										default_damage_type,
										projectile_amount,
										dot_tick_damage,
										dot_max_ticks,
										is_arced,
										charge_time_seconds,
										burst_rate_of_fire,
										power_instant_drain,
										dot_tick_duration,
										projectile_life_span,
										recoil_force,
										idle_power_cost
			                          )
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30)
			ON CONFLICT (id)
			DO 
			    UPDATE SET 
			               id=$1, 
			               brand_id=$2, 
			               label=$3, 
			               weapon_type=$4, 
			               default_skin_id=$5, 
							damage=$6,
							damage_falloff=$7,
							damage_falloff_rate=$8,
							radius=$9,
							radius_damage_falloff=$10,
							spread=$11,
							rate_of_fire=$12,
							projectile_speed=$13,
							max_ammo=$14,
							is_melee=$15,
							power_cost=$16,
							game_client_weapon_id=$17,
							collection=$18,
							default_damage_type=$19,
							projectile_amount=$20,
							dot_tick_damage=$21,
							dot_max_ticks=$22,
							is_arced=$23,
							charge_time_seconds=$24,
							burst_rate_of_fire=$25,
							power_instant_drain=$26,
							dot_tick_duration=$27,
							projectile_life_span=$28,
							recoil_force=$29,
							idle_power_cost=$30
							;
		`,
			weaponModel.ID,
			weaponModel.BrandID,
			weaponModel.Label,
			weaponModel.WeaponType,
			weaponModel.DefaultSkinID,
			weaponModel.Damage,
			weaponModel.DamageFallOff,
			weaponModel.DamageFalloffRate,
			weaponModel.Radius,
			weaponModel.RadiusDamageFalloff,
			weaponModel.Spread,
			weaponModel.RateOfFire,
			weaponModel.ProjectileSpeed,
			weaponModel.MaxAmmo,
			weaponModel.IsMelee,
			weaponModel.PowerCost,
			weaponModel.GameClientWeaponID,
			weaponModel.Collection,
			weaponModel.DefaultDamageType,
			weaponModel.ProjectileAmount,
			weaponModel.DotTickDamage,
			weaponModel.DotMaxTicks,
			weaponModel.IsArced,
			weaponModel.ChargeTimeSeconds,
			weaponModel.BurstRateOfFire,
			weaponModel.PowerInstantDrain,
			weaponModel.DotTickDuration,
			weaponModel.ProjectileLifeSpan,
			weaponModel.RecoilForce,
			weaponModel.IdlePowerCost,
		)
		if err != nil {
			fmt.Println(err.Error()+weaponModel.ID, weaponModel.Label, weaponModel.WeaponType)
			return err
		}

		fmt.Println("UPDATED: "+weaponModel.ID, weaponModel.Label, weaponModel.WeaponType)
	}

	fmt.Println("Finish syncing weapon models")
	return nil
}

func SyncWeaponSkins(f io.Reader, db *sql.DB) error {
	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	var WeaponSkins []types.WeaponSkin
	for _, record := range records {
		weaponSkin := &types.WeaponSkin{
			ID:               record[0],
			Label:            record[1],
			Tier:             record[2],
			Collection:       record[4],
			StatModifier:     record[5],
			ImageUrl:         null.NewString(record[6], record[6] != ""),
			AnimationUrl:     null.NewString(record[7], record[7] != ""),
			CardAnimationUrl: null.NewString(record[8], record[8] != ""),
			LargeImageUrl:    null.NewString(record[9], record[9] != ""),
			AvatarUrl:        null.NewString(record[10], record[10] != ""),
			BackgroundColor:  null.NewString(record[11], record[11] != ""),
			YoutubeUrl:       null.NewString(record[12], record[12] != ""),
		}

		WeaponSkins = append(WeaponSkins, *weaponSkin)
	}

	for _, weaponSkin := range WeaponSkins {
		_, err = db.Exec(`
			INSERT INTO blueprint_weapon_skin(
			                                  id,
			                                  label,
			                                  tier,
			                                  collection,
			                                  stat_modifier,
			                                  image_url,
			                                  animation_url,
			                                  card_animation_url,
			                                  large_image_url,
			                                  avatar_url,
			                                  background_color,
			                                  youtube_url
			                                  )
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			ON CONFLICT (id)
			DO 
			    UPDATE SET 
			               id=$1,
			               label=$2,
			               tier=$3,
			               collection=$4,
			               stat_modifier=$5,
			               image_url=$6,
			               animation_url=$7,
			               card_animation_url=$8,
			               large_image_url=$9,
			               avatar_url=$10,
			               background_color=$11,
			               youtube_url=$12;
		`,
			weaponSkin.ID,
			weaponSkin.Label,
			weaponSkin.Tier,
			weaponSkin.Collection,
			null.NewString(weaponSkin.StatModifier, weaponSkin.StatModifier != ""),
			weaponSkin.ImageUrl,
			weaponSkin.AnimationUrl,
			weaponSkin.CardAnimationUrl,
			weaponSkin.LargeImageUrl,
			weaponSkin.AvatarUrl,
			weaponSkin.BackgroundColor,
			weaponSkin.YoutubeUrl,
		)
		if err != nil {
			fmt.Println(err.Error()+weaponSkin.ID, weaponSkin.Label, weaponSkin.Tier)
			return err
		}

		fmt.Println("UPDATED: "+weaponSkin.ID, weaponSkin.Label, weaponSkin.Tier)
	}

	fmt.Println("Finish syncing weapon skins")
	return nil
}

func SyncWeaponModelSkinCompatibilities(f io.Reader, db *sql.DB) error {

	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	var weaponModelSkinCompatibities []*types.WeaponModelSkinCompatibility
	for _, record := range records {
		weaponModelSkinCompatibity := &types.WeaponModelSkinCompatibility{
			WeaponSkinID:     record[0],
			WeaponModelID:    record[1],
			ImageUrl:         record[2],
			CardAnimationUrl: record[3],
			AvatarUrl:        record[4],
			LargeImageUrl:    record[5],
			BackgroundColor:  record[6],
			AnimationUrl:     record[7],
			YoutubeUrl:       record[8],
		}

		weaponModelSkinCompatibities = append(weaponModelSkinCompatibities, weaponModelSkinCompatibity)
	}

	count := 0
	for _, weaponModelSkinCompat := range weaponModelSkinCompatibities {
		_, err = db.Exec(`
			INSERT INTO weapon_model_skin_compatibilities (
												blueprint_weapon_skin_id,
												weapon_model_id,
												image_url,
												card_animation_url,
												avatar_url,
												large_image_url,
												background_color,
												animation_url,
												youtube_url
			)
			VALUES ($1,$2,$3,$4,$5, $6, $7, $8, $9)
			ON CONFLICT (blueprint_weapon_skin_id, weapon_model_id)
			DO
				UPDATE SET 	blueprint_weapon_skin_id = $1,
							weapon_model_id = $2,
							image_url = $3,
							card_animation_url = $4,
							avatar_url = $5,
							large_image_url = $6,
							background_color = $7,
							animation_url = $8,
							youtube_url = $9;
		`,
			weaponModelSkinCompat.WeaponSkinID,
			weaponModelSkinCompat.WeaponModelID,
			null.NewString(weaponModelSkinCompat.ImageUrl, weaponModelSkinCompat.ImageUrl != ""),
			null.NewString(weaponModelSkinCompat.CardAnimationUrl, weaponModelSkinCompat.CardAnimationUrl != ""),
			null.NewString(weaponModelSkinCompat.AvatarUrl, weaponModelSkinCompat.AvatarUrl != ""),
			null.NewString(weaponModelSkinCompat.LargeImageUrl, weaponModelSkinCompat.LargeImageUrl != ""),
			null.NewString(weaponModelSkinCompat.BackgroundColor, weaponModelSkinCompat.BackgroundColor != ""),
			null.NewString(weaponModelSkinCompat.AnimationUrl, weaponModelSkinCompat.AnimationUrl != ""),
			null.NewString(weaponModelSkinCompat.YoutubeUrl, weaponModelSkinCompat.YoutubeUrl != ""),
		)
		if err != nil {
			fmt.Printf("ERROR WITH %s - %s: %s\n", weaponModelSkinCompat.WeaponSkinID, weaponModelSkinCompat.WeaponModelID, err.Error())
			return err
		}
		count++
		fmt.Printf("UPDATED: %s:%s \n", weaponModelSkinCompat.WeaponSkinID, weaponModelSkinCompat.WeaponModelID)
	}

	fmt.Println("Finish syncing weapon_model_skin_compatibilities Count: " + strconv.Itoa(count))

	return nil
}

func SyncBattleAbilities(f io.Reader, db *sql.DB) error {
	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	ids := []string{}
	for _, record := range records {
		battleAbility := &boiler.BattleAbility{
			ID:                record[0],
			Label:             record[1],
			Description:       record[3],
			KillingPowerLevel: record[4],
		}

		battleAbility.CooldownDurationSecond, err = strconv.Atoi(record[2])
		if err != nil {
			fmt.Println(err.Error()+battleAbility.ID, battleAbility.Label, battleAbility.Description)
			continue
		}
		battleAbility.MaximumCommanderCount, err = strconv.Atoi(record[5])
		if err != nil {
			fmt.Println(err.Error()+battleAbility.ID, battleAbility.Label, battleAbility.Description)
			continue
		}

		// upsert blueprint quest
		err = battleAbility.Upsert(
			db,
			true,
			[]string{
				boiler.BattleAbilityColumns.ID,
			},
			boil.Whitelist(
				boiler.BattleAbilityColumns.Label,
				boiler.BattleAbilityColumns.Description,
				boiler.BattleAbilityColumns.KillingPowerLevel,
				boiler.BattleAbilityColumns.CooldownDurationSecond,
				boiler.BattleAbilityColumns.MaximumCommanderCount,
			),
			boil.Infer(),
		)
		if err != nil {
			fmt.Println(err.Error()+battleAbility.ID, battleAbility.Label, battleAbility.Description)
			return err
		}

		fmt.Println("UPDATED: "+battleAbility.ID, battleAbility.Label)

		// record id list
		ids = append(ids, battleAbility.ID)
	}

	// soft delete any row that is not on the list
	_, err = boiler.BattleAbilities(
		boiler.BattleAbilityWhere.ID.NIN(ids),
	).UpdateAll(db, boiler.M{boiler.BattleAbilityColumns.DeletedAt: null.TimeFrom(time.Now())})
	if err != nil {
		fmt.Println(err.Error(), "Failed to archive rows that are not in the static battle abilities data.")
	}

	fmt.Println("Finish syncing battle abilities")

	return nil
}

func SyncGameAbilities(f io.Reader, db *sql.DB) error {
	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	ids := []string{}
	for _, record := range records {
		gameAbility := &boiler.GameAbility{
			ID:                       record[0],
			FactionID:                record[2],
			BattleAbilityID:          null.NewString(record[3], record[3] != ""),
			Label:                    record[4],
			Colour:                   record[5],
			ImageURL:                 record[6],
			Description:              record[7],
			TextColour:               record[8],
			Level:                    record[9],
			LocationSelectType:       record[10],
			DisplayOnMiniMap:         strings.ToLower(record[13]) == "true",
			MiniMapDisplayEffectType: record[14],
			MechDisplayEffectType:    record[15],
			ShouldCheckTeamKill:      strings.ToLower(record[17]) == "true",
			IgnoreSelfKill:           strings.ToLower(record[19]) == "true",
		}

		gameAbility.GameClientAbilityID, err = strconv.Atoi(record[1])
		if err != nil {
			fmt.Println(err.Error()+gameAbility.ID, gameAbility.Label, gameAbility.Description)
			continue
		}

		if record[11] != "" {
			gameAbility.DeletedAt = null.TimeFrom(time.Now())
		}

		gameAbility.LaunchingDelaySeconds, err = strconv.Atoi(record[12])
		if err != nil {
			fmt.Println(err.Error()+gameAbility.ID, gameAbility.Label, gameAbility.Description)
			continue
		}

		gameAbility.AnimationDurationSeconds, err = strconv.Atoi(record[16])
		if err != nil {
			fmt.Println(err.Error()+gameAbility.ID, gameAbility.Label, gameAbility.Description)
			continue
		}

		gameAbility.MaximumTeamKillTolerantCount, err = strconv.Atoi(record[18])
		if err != nil {
			fmt.Println(err.Error()+gameAbility.ID, gameAbility.Label, gameAbility.Description)
			continue
		}

		gameAbility.CountPerBattle, err = strconv.Atoi(record[20])
		if err != nil {
			fmt.Println(err.Error()+gameAbility.ID, gameAbility.Label, gameAbility.Description)
			continue
		}

		// upsert game ability
		err = gameAbility.Upsert(
			db,
			true,
			[]string{
				boiler.GameAbilityColumns.ID,
			},
			boil.Whitelist(
				boiler.GameAbilityColumns.GameClientAbilityID,
				boiler.GameAbilityColumns.FactionID,
				boiler.GameAbilityColumns.BattleAbilityID,
				boiler.GameAbilityColumns.Label,
				boiler.GameAbilityColumns.Colour,
				boiler.GameAbilityColumns.ImageURL,
				boiler.GameAbilityColumns.Description,
				boiler.GameAbilityColumns.TextColour,
				boiler.GameAbilityColumns.Level,
				boiler.GameAbilityColumns.LocationSelectType,
				boiler.GameAbilityColumns.DeletedAt,
				boiler.GameAbilityColumns.LaunchingDelaySeconds,
				boiler.GameAbilityColumns.DisplayOnMiniMap,
				boiler.GameAbilityColumns.MiniMapDisplayEffectType,
				boiler.GameAbilityColumns.MechDisplayEffectType,
				boiler.GameAbilityColumns.AnimationDurationSeconds,
				boiler.GameAbilityColumns.ShouldCheckTeamKill,
				boiler.GameAbilityColumns.MaximumTeamKillTolerantCount,
				boiler.GameAbilityColumns.IgnoreSelfKill,
				boiler.GameAbilityColumns.CountPerBattle,
			),
			boil.Infer(),
		)
		if err != nil {
			fmt.Println(err.Error()+gameAbility.ID, gameAbility.Label, gameAbility.Description)
			return err
		}

		fmt.Println("UPDATED: "+gameAbility.ID, gameAbility.GameClientAbilityID, gameAbility.Label)

		// record id list
		ids = append(ids, gameAbility.ID)
	}

	// soft delete any row that is not on the list
	_, err = boiler.GameAbilities(
		boiler.GameAbilityWhere.ID.NIN(ids),
	).UpdateAll(db, boiler.M{boiler.GameAbilityColumns.DeletedAt: null.TimeFrom(time.Now())})
	if err != nil {
		fmt.Println(err.Error(), "Failed to archive rows that are not in the static game abilities data.")
	}

	fmt.Println("Finish syncing game abilities")

	return nil
}

func SyncPlayerAbilities(f io.Reader, db *sql.DB) error {
	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	for _, record := range records {
		playerAbility := &boiler.BlueprintPlayerAbility{
			ID:                       record[0],
			Label:                    record[2],
			Colour:                   record[3],
			ImageURL:                 record[4],
			Description:              record[5],
			TextColour:               record[6],
			LocationSelectType:       record[7],
			DisplayOnMiniMap:         strings.ToLower(record[12]) == "true",
			MiniMapDisplayEffectType: record[14],
			MechDisplayEffectType:    record[15],
		}

		playerAbility.GameClientAbilityID, err = strconv.Atoi(record[1])
		if err != nil {
			fmt.Println(err.Error()+playerAbility.ID, playerAbility.Label, playerAbility.Description)
			continue
		}

		playerAbility.RarityWeight, err = strconv.Atoi(record[9])
		if err != nil {
			fmt.Println(err.Error()+playerAbility.ID, playerAbility.Label, playerAbility.Description)
			continue
		}

		playerAbility.InventoryLimit, err = strconv.Atoi(record[10])
		if err != nil {
			fmt.Println(err.Error()+playerAbility.ID, playerAbility.Label, playerAbility.Description)
			continue
		}

		playerAbility.CooldownSeconds, err = strconv.Atoi(record[11])
		if err != nil {
			fmt.Println(err.Error()+playerAbility.ID, playerAbility.Label, playerAbility.Description)
			continue
		}

		playerAbility.LaunchingDelaySeconds, err = strconv.Atoi(record[13])
		if err != nil {
			fmt.Println(err.Error()+playerAbility.ID, playerAbility.Label, playerAbility.Description)
			continue
		}

		playerAbility.AnimationDurationSeconds, err = strconv.Atoi(record[16])
		if err != nil {
			fmt.Println(err.Error()+playerAbility.ID, playerAbility.Label, playerAbility.Description)
			continue
		}

		// upsert game ability
		err = playerAbility.Upsert(
			db,
			true,
			[]string{
				boiler.BlueprintPlayerAbilityColumns.ID,
			},
			boil.Whitelist(
				boiler.BlueprintPlayerAbilityColumns.GameClientAbilityID,
				boiler.BlueprintPlayerAbilityColumns.Label,
				boiler.BlueprintPlayerAbilityColumns.Colour,
				boiler.BlueprintPlayerAbilityColumns.ImageURL,
				boiler.BlueprintPlayerAbilityColumns.Description,
				boiler.BlueprintPlayerAbilityColumns.TextColour,
				boiler.BlueprintPlayerAbilityColumns.LocationSelectType,
				boiler.BlueprintPlayerAbilityColumns.RarityWeight,
				boiler.BlueprintPlayerAbilityColumns.InventoryLimit,
				boiler.BlueprintPlayerAbilityColumns.CooldownSeconds,
				boiler.BlueprintPlayerAbilityColumns.DisplayOnMiniMap,
				boiler.BlueprintPlayerAbilityColumns.LaunchingDelaySeconds,
				boiler.BlueprintPlayerAbilityColumns.MiniMapDisplayEffectType,
				boiler.BlueprintPlayerAbilityColumns.MechDisplayEffectType,
				boiler.BlueprintPlayerAbilityColumns.AnimationDurationSeconds,
			),
			boil.Infer(),
		)
		if err != nil {
			fmt.Println(err.Error()+playerAbility.ID, playerAbility.Label, playerAbility.Description)
			return err
		}

		fmt.Println("UPDATED: "+playerAbility.ID, playerAbility.GameClientAbilityID, playerAbility.Label)

		// get all the existing blueprint player abilities and update their details
		existingPlayerAbilities, err := boiler.BlueprintPlayerAbilities(
			boiler.BlueprintPlayerAbilityWhere.GameClientAbilityID.EQ(playerAbility.GameClientAbilityID),
			boiler.BlueprintPlayerAbilityWhere.ID.NEQ(playerAbility.ID),
		).All(db)
		if err != nil {
			fmt.Println(err.Error()+playerAbility.ID, playerAbility.Label, playerAbility.Description)
		}

		for _, bpa := range existingPlayerAbilities {
			// swap id
			playerAbility.ID = bpa.ID

			// update everything
			_, err := playerAbility.Update(db, boil.Whitelist(
				boiler.BlueprintPlayerAbilityColumns.GameClientAbilityID,
				boiler.BlueprintPlayerAbilityColumns.Label,
				boiler.BlueprintPlayerAbilityColumns.Colour,
				boiler.BlueprintPlayerAbilityColumns.ImageURL,
				boiler.BlueprintPlayerAbilityColumns.Description,
				boiler.BlueprintPlayerAbilityColumns.TextColour,
				boiler.BlueprintPlayerAbilityColumns.LocationSelectType,
				boiler.BlueprintPlayerAbilityColumns.RarityWeight,
				boiler.BlueprintPlayerAbilityColumns.InventoryLimit,
				boiler.BlueprintPlayerAbilityColumns.CooldownSeconds,
				boiler.BlueprintPlayerAbilityColumns.DisplayOnMiniMap,
				boiler.BlueprintPlayerAbilityColumns.LaunchingDelaySeconds,
				boiler.BlueprintPlayerAbilityColumns.MiniMapDisplayEffectType,
				boiler.BlueprintPlayerAbilityColumns.MechDisplayEffectType,
				boiler.BlueprintPlayerAbilityColumns.AnimationDurationSeconds,
			))
			if err != nil {
				fmt.Println(err.Error()+playerAbility.ID, playerAbility.Label, playerAbility.Description)
				break
			}
		}
	}

	fmt.Println("Finish syncing game abilities")

	return nil
}

func SyncPowerCores(f io.Reader, db *sql.DB) error {
	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	var PowerCores []types.PowerCores
	for _, record := range records {
		powerCore := &types.PowerCores{
			ID:               record[0],
			Collection:       record[1],
			Label:            record[2],
			Size:             record[3],
			Capacity:         record[4],
			MaxDrawRate:      record[5],
			RechargeRate:     record[6],
			Armour:           record[7],
			MaxHitpoints:     record[8],
			Tier:             record[9],
			ImageUrl:         record[11],
			CardAnimationUrl: record[12],
			AvatarUrl:        record[13],
			LargeImageUrl:    record[14],
			BackgroundColor:  record[15],
			AnimationUrl:     record[16],
			YoutubeUrl:       record[17],
			WeaponShare:      record[18],
			MovementShare:    record[19],
			UtilityShare:     record[20],
		}

		PowerCores = append(PowerCores, *powerCore)
	}

	for _, powerCore := range PowerCores {
		imageURL := &powerCore.ImageUrl
		if powerCore.ImageUrl == "" {
			imageURL = nil
		}

		cardAnimationURL := &powerCore.CardAnimationUrl
		if powerCore.CardAnimationUrl == "" {
			cardAnimationURL = nil
		}

		avatarURL := &powerCore.AvatarUrl
		if powerCore.AvatarUrl == "" {
			avatarURL = nil
		}

		largeImageURL := &powerCore.LargeImageUrl
		if powerCore.LargeImageUrl == "" {
			largeImageURL = nil
		}

		backgroundColor := &powerCore.BackgroundColor
		if powerCore.BackgroundColor == "" {
			backgroundColor = nil
		}

		animationURL := &powerCore.AnimationUrl
		if powerCore.AnimationUrl == "" {
			animationURL = nil
		}

		youtubeURL := &powerCore.YoutubeUrl
		if powerCore.YoutubeUrl == "" {
			youtubeURL = nil
		}

		_, err = db.Exec(`
			INSERT INTO blueprint_power_cores(id, collection, label, size, capacity, max_draw_rate, recharge_rate, armour, max_hitpoints, tier, image_url, card_animation_url, avatar_url, large_image_url, background_color, animation_url, youtube_url, weapon_share, movement_share, utility_share)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
			ON CONFLICT (id)
			DO 
			    UPDATE SET id=$1, collection=$2, label=$3, size=$4, capacity=$5, max_draw_rate=$6, recharge_rate=$7, armour=$8, max_hitpoints=$9, tier=$10, image_url=$11, card_animation_url=$12, avatar_url=$13, large_image_url=$14, background_color=$15, animation_url=$16, youtube_url=$17, weapon_share=$18, movement_share=$19, utility_share=$20;
		`, powerCore.ID, powerCore.Collection, powerCore.Label, powerCore.Size, powerCore.Capacity, powerCore.MaxDrawRate, powerCore.RechargeRate, powerCore.Armour, powerCore.MaxHitpoints, powerCore.Tier, imageURL, cardAnimationURL, avatarURL, largeImageURL, backgroundColor, animationURL, youtubeURL, powerCore.WeaponShare, powerCore.MovementShare, powerCore.UtilityShare)
		if err != nil {
			fmt.Println(err.Error()+powerCore.ID, powerCore.Collection, powerCore.Label)
			return err
		}

		fmt.Println("UPDATED: "+powerCore.ID, powerCore.Collection, powerCore.Label)
	}

	fmt.Println("Finish syncing power cores")

	return nil
}

func SyncStaticQuest(f io.Reader, db *sql.DB) error {
	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	for _, record := range records {
		blueprintQuest := &boiler.BlueprintQuest{
			ID:             record[0],
			QuestEventType: record[1],
			Key:            record[2],
			Name:           record[3],
			Description:    record[4],
		}

		// convert request amount
		blueprintQuest.RequestAmount, err = strconv.Atoi(record[5])
		if err != nil {
			fmt.Println(err.Error(), blueprintQuest.ID, blueprintQuest.Name)
			return err
		}

		// upsert blueprint quest
		err = blueprintQuest.Upsert(
			db,
			true,
			[]string{
				boiler.BlueprintQuestColumns.ID,
			},
			boil.Whitelist(
				boiler.BlueprintQuestColumns.QuestEventType,
				boiler.BlueprintQuestColumns.Key,
				boiler.BlueprintQuestColumns.Name,
				boiler.BlueprintQuestColumns.Description,
				boiler.BlueprintQuestColumns.RequestAmount,
			),
			boil.Infer(),
		)
		if err != nil {
			fmt.Println(err.Error(), blueprintQuest.ID, blueprintQuest.Name)
			return err
		}

		fmt.Println("UPDATED: "+blueprintQuest.ID, blueprintQuest.Name)

	}

	fmt.Println("Finish syncing static quest")

	return nil
}
