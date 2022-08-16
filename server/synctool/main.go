package synctool

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"io"
	"log"
	"net/http"
	"os"
	"server/db/boiler"
	"server/synctool/types"
	"strconv"
	"time"
)

type StaticSyncTool struct {
	DB       *sql.DB
	FilePath string
}

func SyncTool(dt *StaticSyncTool) error {

	f, err := readFile(fmt.Sprintf("%sfactions.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncFactions(f, dt.DB)
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

	f, err = readFile(fmt.Sprintf("%smech_skins.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncMechSkins(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	f, err = readFile(fmt.Sprintf("%smech_models.csv", dt.FilePath))
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

	f, err = readFile(fmt.Sprintf("%sweapon_skins.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncWeaponSkins(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	f, err = readFile(fmt.Sprintf("%sweapon_models.csv", dt.FilePath))
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

	f, err = readFile(fmt.Sprintf("%spower_cores.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncPowerCores(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	f, err = readFile(fmt.Sprintf("%smechs.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncStaticMech(f, dt.DB)
	if err != nil {
		return err
	}
	f.Close()

	f, err = readFile(fmt.Sprintf("%sweapons.csv", dt.FilePath))
	if err != nil {
		return err
	}
	err = SyncStaticWeapon(f, dt.DB)
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
			ID:                   record[0],
			Label:                record[1],
			DefaultChassisSkinID: record[3],
			BrandID:              null.StringFrom(record[4]),
			MechType:             record[5],
		}

		MechModels = append(MechModels, *mechModel)
	}

	count := 0

	for _, mechModel := range MechModels {
		brandID := &mechModel.BrandID.String
		if mechModel.BrandID.String == "" || !mechModel.BrandID.Valid {
			brandID = nil
		}

		_, err = db.Exec(`
			INSERT INTO mech_models (id, label, default_chassis_skin_id, brand_id, mech_type)
			VALUES ($1,$2,$3,$4,$5)
			ON CONFLICT (id)
			DO
				UPDATE SET id=$1, label=$2, default_chassis_skin_id=$3, brand_id=$4, mech_type=$5;
		`, mechModel.ID, mechModel.Label, mechModel.DefaultChassisSkinID, brandID, mechModel.MechType)
		if err != nil {
			fmt.Println("ERROR: " + err.Error())
			continue
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
			continue
		}
		count++
		fmt.Printf("UPDATED: %s:%s \n", mechModelSkinCompat.MechSkinID, mechModelSkinCompat.MechModelID)
	}

	fmt.Println("Finish syncing mech_model_skin_compatibilities Count: " + strconv.Itoa(count))

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
			ID:           record[0],
			Collection:   record[1],
			Label:        record[2],
			Tier:         record[3],
			StatModifier: record[5],
		}

		MechSkins = append(MechSkins, *mechModel)
	}

	for _, mechSkin := range MechSkins {
		statModifier := null.NewString(mechSkin.StatModifier, mechSkin.StatModifier != "")

		_, err = db.Exec(`
			INSERT INTO blueprint_mech_skin(
			                                id,
			                                collection,
			                                label,
			                                tier,
			                                stat_modifier
			                                )
			VALUES ($1,$2,$3,$4,$5)
			ON CONFLICT (id)
			DO
			    UPDATE SET id=$1,
			               collection=$2,
			               label=$3,
			               tier=$4,
			               stat_modifier=$5;
		`,
			mechSkin.ID,
			mechSkin.Collection,
			mechSkin.Label,
			mechSkin.Tier,
			statModifier,
		)
		if err != nil {
			fmt.Println(err.Error()+mechSkin.ID, mechSkin.Label)
			continue
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
			ID:              record[0],
			ContractReward:  record[1],
			VotePrice:       record[2],
			Label:           record[3],
			GuildID:         record[4],
			DeletedAt:       record[5],
			UpdatedAt:       record[6],
			CreatedAt:       record[7],
			PrimaryColor:    record[8],
			SecondaryColor:  record[9],
			BackgroundColor: record[10],
			LogoURL:         record[11],
			BackgroundURL:   record[12],
			Description:     record[13],
		}

		Factions = append(Factions, *faction)
	}

	for _, faction := range Factions {
		guildID := &faction.GuildID
		if faction.GuildID == "" {
			guildID = nil
		}
		_, err = db.Exec(`
			INSERT INTO factions (id, label, guild_id, primary_color, secondary_color, background_color, logo_url, background_url, description)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			ON CONFLICT (id)
			DO
				UPDATE SET id=$1, label=$2, guild_id=$3, primary_color=$4, secondary_color=$5, background_color=$6, logo_url=$7, background_url=$8, description=$9;
		`, faction.ID, faction.Label, guildID, faction.PrimaryColor, faction.SecondaryColor, faction.BackgroundColor, faction.LogoURL, faction.BackgroundURL, faction.Description)
		if err != nil {
			fmt.Println(err.Error()+faction.ID, faction.Label)
			continue
		}

		fmt.Println("UPDATED: "+faction.ID, faction.Label, faction.GuildID, faction.PrimaryColor, faction.SecondaryColor, faction.BackgroundColor, faction.LogoURL, faction.BackgroundURL, faction.Description)
	}

	fmt.Println("Finish syncing Factions")
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
			continue
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
			continue
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

	var WeaponModels []types.WeaponModel
	for _, record := range records {
		weaponModel := &types.WeaponModel{
			ID:            record[0],
			BrandID:       record[1],
			Label:         record[2],
			WeaponType:    record[3],
			DefaultSkinID: record[4],
			DeletedAt:     record[5],
			UpdatedAt:     record[6],
		}

		WeaponModels = append(WeaponModels, *weaponModel)
	}

	for _, weaponModel := range WeaponModels {
		deletedAt := &weaponModel.DeletedAt
		if weaponModel.DeletedAt == "" {
			deletedAt = nil
		}

		brandID := &weaponModel.BrandID
		if weaponModel.BrandID == "" {
			brandID = nil
		}

		defaultSkinID := &weaponModel.DefaultSkinID
		if weaponModel.DefaultSkinID == "" {
			defaultSkinID = nil
		}

		_, err = db.Exec(`
			INSERT INTO weapon_models(id, brand_id, label, weapon_type, default_skin_id, deleted_at)
			VALUES ($1,$2,$3,$4,$5,$6)
			ON CONFLICT (id)
			DO 
			    UPDATE SET id=$1, brand_id=$2, label=$3, weapon_type=$4, default_skin_id=$5, deleted_at=$6;
		`, weaponModel.ID, brandID, weaponModel.Label, weaponModel.WeaponType, defaultSkinID, deletedAt)
		if err != nil {
			fmt.Println(err.Error()+weaponModel.ID, weaponModel.Label, weaponModel.WeaponType)
			continue
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
			ID:           record[0],
			Label:        record[1],
			Tier:         record[2],
			Collection:   record[4],
			StatModifier: record[5],
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
			                                  stat_modifier
			                                  )
			VALUES ($1,$2,$3,$4,$5)
			ON CONFLICT (id)
			DO 
			    UPDATE SET 
			               id=$1,
			               label=$2,
			               tier=$3,
			               collection=$4,
			               stat_modifier=$5;
		`,
			weaponSkin.ID,
			weaponSkin.Label,
			weaponSkin.Tier,
			weaponSkin.Collection,
			null.NewString(weaponSkin.StatModifier, weaponSkin.StatModifier != ""),
		)
		if err != nil {
			fmt.Println(err.Error()+weaponSkin.ID, weaponSkin.Label, weaponSkin.Tier)
			continue
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
			fmt.Println(weaponModelSkinCompat.WeaponSkinID)
			fmt.Println(weaponModelSkinCompat.WeaponModelID)
			fmt.Println(weaponModelSkinCompat.ImageUrl)
			fmt.Println(weaponModelSkinCompat.AnimationUrl)
			fmt.Println(weaponModelSkinCompat.CardAnimationUrl)
			fmt.Println(weaponModelSkinCompat.LargeImageUrl)
			fmt.Println(weaponModelSkinCompat.AvatarUrl)
			fmt.Println(weaponModelSkinCompat.BackgroundColor)
			fmt.Println(weaponModelSkinCompat.YoutubeUrl)
			fmt.Println("ERROR: " + err.Error())
			continue
		}
		count++
		fmt.Printf("UPDATED: %s:%s \n", weaponModelSkinCompat.WeaponSkinID, weaponModelSkinCompat.WeaponModelID)
	}

	fmt.Println("Finish syncing mech_model_skin_compatibilities Count: " + strconv.Itoa(count))

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

	var BattleAbilities []types.BattleAbility
	for _, record := range records {
		battleAbility := &types.BattleAbility{
			ID:               record[0],
			Label:            record[1],
			CoolDownDuration: record[2],
			Description:      record[3],
		}

		BattleAbilities = append(BattleAbilities, *battleAbility)
	}

	for _, battleAbility := range BattleAbilities {
		_, err = db.Exec(`
			INSERT INTO battle_abilities(id, label, cooldown_duration_second, description)
			VALUES ($1,$2,$3,$4)
			ON CONFLICT (id)
			DO 
			    UPDATE SET id=$1, label=$2, cooldown_duration_second=$3, description=$4;
		`, battleAbility.ID, battleAbility.Label, battleAbility.CoolDownDuration, battleAbility.Description)
		if err != nil {
			fmt.Println(err.Error()+battleAbility.ID, battleAbility.Label, battleAbility.CoolDownDuration, battleAbility.Description)
			continue
		}
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

	for _, record := range records {
		ga := &types.GameAbility{
			ID:          record[0],
			FactionID:   record[2],
			Label:       record[4],
			Colour:      record[5],
			ImageUrl:    record[6],
			SupsCost:    record[7],
			Description: record[8],
			TextColour:  record[9],
			CurrentSups: record[10],
			Level:       record[11],
		}
		gcID, err := strconv.Atoi(record[1])
		if err == nil {
			ga.GameClientAbilityID = gcID
		}

		if record[3] != "" {
			ga.BattleAbilityID = &record[3]
		}

		_, err = db.Exec(`
			INSERT INTO game_abilities (id, game_client_ability_id, faction_id, battle_ability_id, label, colour, image_url, sups_cost, description, text_colour, current_sups, level)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			ON CONFLICT (id)
			DO 
			    UPDATE SET id=$1, game_client_ability_id=$2, faction_id=$3, battle_ability_id=$4, label=$5, colour=$6, image_url=$7, sups_cost=$8, description=$9, text_colour=$10, current_sups=$11, level=$12;
		`,
			ga.ID,
			ga.GameClientAbilityID,
			ga.FactionID,
			ga.BattleAbilityID,
			ga.Label,
			ga.Colour,
			ga.ImageUrl,
			ga.SupsCost,
			ga.Description,
			ga.TextColour,
			ga.CurrentSups,
			ga.Level,
		)
		if err != nil {
			fmt.Println(err.Error(), ga.ID, ga.Label)
			continue
		}

		fmt.Println("UPDATED: "+ga.ID, ga.GameClientAbilityID, ga.Label)
	}

	fmt.Println("Finish syncing battle abilities")

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
			INSERT INTO blueprint_power_cores(id, collection, label, size, capacity, max_draw_rate, recharge_rate, armour, max_hitpoints, tier, image_url, card_animation_url, avatar_url, large_image_url, background_color, animation_url, youtube_url)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
			ON CONFLICT (id)
			DO 
			    UPDATE SET id=$1, collection=$2, label=$3, size=$4, capacity=$5, max_draw_rate=$6, recharge_rate=$7, armour=$8, max_hitpoints=$9, tier=$10, image_url=$11, card_animation_url=$12, avatar_url=$13, large_image_url=$14, background_color=$15, animation_url=$16, youtube_url=$17;
		`, powerCore.ID, powerCore.Collection, powerCore.Label, powerCore.Size, powerCore.Capacity, powerCore.MaxDrawRate, powerCore.RechargeRate, powerCore.Armour, powerCore.MaxHitpoints, powerCore.Tier, imageURL, cardAnimationURL, avatarURL, largeImageURL, backgroundColor, animationURL, youtubeURL)
		if err != nil {
			fmt.Println(err.Error()+powerCore.ID, powerCore.Collection, powerCore.Label)
			continue
		}

		fmt.Println("UPDATED: "+powerCore.ID, powerCore.Collection, powerCore.Label)
	}

	fmt.Println("Finish syncing power cores")

	return nil
}

func SyncStaticMech(f io.Reader, db *sql.DB) error {
	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	var BlueprintMechs []types.BlueprintMechs
	for _, record := range records {
		blueprintMechs := &types.BlueprintMechs{
			ID:               record[0],
			Label:            record[1],
			Slug:             record[2],
			WeaponHardpoints: record[3],
			UtilitySlots:     record[4],
			Speed:            record[5],
			MaxHitpoints:     record[6],
			DeletedAt:        record[7],
			UpdatedAt:        record[8],
			CreatedAt:        record[9],
			ModelID:          record[10],
			Collection:       record[11],
			PowerCoreSize:    record[12],
			Tier:             record[13],
		}

		BlueprintMechs = append(BlueprintMechs, *blueprintMechs)
	}

	for _, blueprintMech := range BlueprintMechs {
		deletedAt := &blueprintMech.DeletedAt
		if blueprintMech.DeletedAt == "" {
			deletedAt = nil
		}

		_, err = db.Exec(`
			INSERT INTO blueprint_mechs(
			                            id, 
			                            label, 
			                            slug, 
			                            weapon_hardpoints, 
			                            utility_slots, 
			                            speed, 
			                            max_hitpoints, 
			                            deleted_at, 
			                            updated_at, 
			                            created_at, 
			                            model_id, 
			                            collection, 
			                            power_core_size, 
			                            tier
			                            )
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
			ON CONFLICT (id)
			DO 
			    UPDATE SET id=$1, 
			               label=$2, 
			               slug=$3, 
			               weapon_hardpoints=$4, 
			               utility_slots=$5, 
			               speed=$6, 
			               max_hitpoints=$7, 
			               deleted_at=$8, 
			               updated_at=$9, 
			               created_at=$10, 
			               model_id=$11, 
			               collection=$12, 
			               power_core_size=$13, 
			               tier=$14;
		`,
			blueprintMech.ID,
			blueprintMech.Label,
			blueprintMech.Slug,
			blueprintMech.WeaponHardpoints,
			blueprintMech.UtilitySlots,
			blueprintMech.Speed,
			blueprintMech.MaxHitpoints,
			deletedAt,
			blueprintMech.UpdatedAt,
			blueprintMech.CreatedAt,
			blueprintMech.ModelID,
			blueprintMech.Collection,
			blueprintMech.PowerCoreSize,
			blueprintMech.Tier)
		if err != nil {
			fmt.Println(err.Error()+blueprintMech.ID, blueprintMech.Label)
			continue
		}

		fmt.Println("UPDATED: "+blueprintMech.ID, blueprintMech.Label)
	}

	fmt.Println("Finish syncing static mech")

	return nil
}

func SyncStaticWeapon(f io.Reader, db *sql.DB) error {
	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	var BlueprintWeapons []types.BlueprintWeapons
	for _, record := range records {
		blueprintWeapon := &types.BlueprintWeapons{
			ID:                  record[0],
			Label:               record[1],
			Slug:                record[2],
			Damage:              record[3],
			DeletedAt:           record[4],
			UpdatedAt:           record[5],
			CreatedAt:           record[6],
			GameClientWeaponID:  record[7],
			WeaponType:          record[8],
			Collection:          record[9],
			DefaultDamageType:   record[10],
			DamageFalloff:       record[11],
			DamageFalloffRate:   record[12],
			Radius:              record[13],
			RadiusDamageFalloff: record[14],
			Spread:              record[15],
			RateOfFire:          record[16],
			ProjectileSpeed:     record[17],
			MaxAmmo:             record[18],
			IsMelee:             record[19],
			Tier:                record[20],
			EnergyCost:          record[21],
			WeaponModelID:       record[22],
		}

		BlueprintWeapons = append(BlueprintWeapons, *blueprintWeapon)
	}

	for _, blueprintWeapon := range BlueprintWeapons {
		deletedAt := &blueprintWeapon.DeletedAt
		if blueprintWeapon.DeletedAt == "" {
			deletedAt = nil
		}

		gameClientID := &blueprintWeapon.GameClientWeaponID
		if blueprintWeapon.GameClientWeaponID == "" {
			gameClientID = nil
		}

		_, err = db.Exec(`
			INSERT INTO blueprint_weapons(
			                              id, 
			                              label, 
			                              slug, 
			                              damage, 
			                              deleted_at, 
			                              updated_at, 
			                              created_at, 
			                              game_client_weapon_id, 
			                              weapon_type, 
			                              collection, 
			                              default_damage_type, 
			                              damage_falloff, 
			                              damage_falloff_rate, 
			                              radius, 
			                              radius_damage_falloff, 
			                              spread, 
			                              rate_of_fire, 
			                              projectile_speed, 
			                              max_ammo, 
			                              is_melee, 
			                              tier, 
			                              energy_cost, 
			                              weapon_model_id)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)
			ON CONFLICT (id)
			DO 
			    UPDATE SET id=$1,
			               label=$2,
			               slug=$3,
			               damage=$4,
			               deleted_at=$5,
			               updated_at=$6,
			               created_at=$7,
			               game_client_weapon_id=$8,
			               weapon_type=$9,
			               collection=$10,
			               default_damage_type=$11,
			               damage_falloff=$12,
			               damage_falloff_rate=$13,
			               radius=$14,
			               radius_damage_falloff=$15,
			               spread=$16,
			               rate_of_fire=$17,
			               projectile_speed=$18,
			               max_ammo=$19,
			               is_melee=$20,
			               tier=$21,
			               energy_cost=$22,
			               weapon_model_id=$23;
		`, blueprintWeapon.ID, blueprintWeapon.Label, blueprintWeapon.Slug, blueprintWeapon.Damage, deletedAt, blueprintWeapon.UpdatedAt, blueprintWeapon.CreatedAt, gameClientID, blueprintWeapon.WeaponType, blueprintWeapon.Collection, blueprintWeapon.DefaultDamageType, blueprintWeapon.DamageFalloff, blueprintWeapon.DamageFalloffRate, blueprintWeapon.Radius, blueprintWeapon.RadiusDamageFalloff, blueprintWeapon.Spread, blueprintWeapon.RateOfFire, blueprintWeapon.ProjectileSpeed, blueprintWeapon.MaxAmmo, blueprintWeapon.IsMelee, blueprintWeapon.Tier, blueprintWeapon.EnergyCost, blueprintWeapon.WeaponModelID)
		if err != nil {
			fmt.Println(err.Error()+blueprintWeapon.ID, blueprintWeapon.Label)
			continue
		}

		fmt.Println("UPDATED: "+blueprintWeapon.ID, blueprintWeapon.Label)
	}

	fmt.Println("Finish syncing weapon")

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
			continue
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
			continue
		}

		fmt.Println("UPDATED: "+blueprintQuest.ID, blueprintQuest.Name)

	}

	fmt.Println("Finish syncing static quest")

	return nil
}
