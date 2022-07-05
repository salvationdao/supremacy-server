package main

import (
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/volatiletech/null/v8"
	"log"
	"net/url"
	"os"
	"server/devtool/types"
)

type DevTool struct {
	db *sql.DB
}

func main() {
	if os.Getenv("GAMESERVER_ENVIRONMENT") == "production" {
		log.Fatal("Only works in dev and staging environment")
	}

	syncMech := flag.Bool("sync_mech", false, "Sync mech skins and models with staging data")

	flag.Parse()
	fmt.Println(*syncMech)

	params := url.Values{}
	params.Add("sslmode", "disable")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?%s",
		"gameserver",
		"dev",
		"localhost",
		"5437",
		"gameserver",
		params.Encode(),
	)
	cfg, err := pgx.ParseConfig(connString)
	if err != nil {
		log.Fatal(err)
	}
	conn := stdlib.OpenDB(*cfg)
	if err != nil {
		log.Fatal(err)
	}

	dt := DevTool{db: conn}

	if syncMech != nil && *syncMech {
		RemoveFKContraints(dt)
		SyncMechModels(dt)
		SyncMechSkins(dt)
		SyncBrands(dt)
		SyncMysteryCrates(dt)
		SyncWeaponModel(dt)
		SyncWeaponSkins(dt)
	}

	fmt.Println("Finish syncing static data")
}

func RemoveFKContraints(dt DevTool) error {
	_, err := dt.db.Exec(
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
			ALTER TABLE blueprint_mech_skin ADD CONSTRAINT blueprint_chassis_skin_mech_model_fkey FOREIGN KEY (mech_model) REFERENCES mech_models(id) ON UPDATE CASCADE;

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

func SyncMechModels(dt DevTool) error {
	f, err := os.OpenFile("./devtool/temp-sync/supremacy-static-data/mech_models.csv", os.O_RDONLY, 0755)
	if err != nil {
		log.Fatal("CANT OPEN FILE")
		return err
	}

	defer f.Close()

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

	for _, mechModel := range MechModels {
		_, err = dt.db.Exec(`UPDATE mech_models SET id=$1 WHERE label=$2 AND mech_type=$3`, mechModel.ID, mechModel.Label, mechModel.MechType)
		if err != nil {
			fmt.Println("ERROR: " + err.Error())
			continue
		}

		fmt.Println("UPDATED: " + mechModel.Label)
	}

	fmt.Println("Finish syncing mech models")

	return nil
}

func SyncMechSkins(dt DevTool) error {
	f, err := os.OpenFile("./devtool/temp-sync/supremacy-static-data/mech_skins.csv", os.O_RDONLY, 0755)
	if err != nil {
		log.Fatal("CANT OPEN FILE")
		return err
	}

	defer f.Close()

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
			ID:               record[0],
			Collection:       record[1],
			MechModel:        record[2],
			Label:            record[3],
			Tier:             record[4],
			ImageUrl:         null.StringFrom(record[5]),
			AnimationUrl:     null.StringFrom(record[6]),
			CardAnimationUrl: null.StringFrom(record[7]),
			LargeImageUrl:    null.StringFrom(record[8]),
			AvatarUrl:        null.StringFrom(record[9]),
			MechType:         record[13],
		}

		MechSkins = append(MechSkins, *mechModel)
	}

	for _, mechSkin := range MechSkins {
		_, err = dt.db.Exec(`UPDATE blueprint_mech_skin SET id=$1 WHERE label=$2 AND tier=$3 AND collection=$4 AND mech_model=$5 AND mech_type=$6`, mechSkin.ID, mechSkin.Label, mechSkin.Tier, mechSkin.Collection, mechSkin.MechModel, mechSkin.MechType)
		if err != nil {
			fmt.Println(err.Error()+mechSkin.ID, mechSkin.Label, mechSkin.Tier, mechSkin.Collection)
			continue
		}

		fmt.Println("UPDATED: "+mechSkin.ID, mechSkin.Label, mechSkin.Tier, mechSkin.Collection, mechSkin.MechModel)
	}

	fmt.Println("Finish syncing mech skins")

	return nil

}

func SyncBrands(dt DevTool) error {
	f, err := os.OpenFile("./devtool/temp-sync/supremacy-static-data/brands.csv", os.O_RDONLY, 0755)
	if err != nil {
		log.Fatal("CANT OPEN FILE")
		return err
	}

	defer f.Close()

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
		_, err = dt.db.Exec(`UPDATE brands SET id=$1 WHERE label=$2 AND faction_id=$3 `, brand.ID, brand.Label, brand.FactionID)
		if err != nil {
			fmt.Println(err.Error()+brand.ID, brand.FactionID, brand.Label)
			continue
		}

		fmt.Println("UPDATED: "+brand.ID, brand.FactionID, brand.Label)
	}

	fmt.Println("Finish syncing brands")
	return nil
}

func SyncMysteryCrates(dt DevTool) error {
	f, err := os.OpenFile("./devtool/temp-sync/supremacy-static-data/mystery_crates.csv", os.O_RDONLY, 0755)
	if err != nil {
		log.Fatal("CANT OPEN FILE")
		return err
	}

	defer f.Close()

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
			Label:            record[9],
		}

		MysteryCrates = append(MysteryCrates, *mysteryCrate)
	}

	for _, mysteryCrate := range MysteryCrates {
		_, err = dt.db.Exec(`UPDATE storefront_mystery_crates SET id=$1 WHERE label=$2 AND mystery_crate_type=$3 `, mysteryCrate.ID, mysteryCrate.Label, mysteryCrate.MysteryCrateType)
		if err != nil {
			fmt.Println(err.Error()+mysteryCrate.ID, mysteryCrate.Label, mysteryCrate.MysteryCrateType)
			continue
		}

		fmt.Println("UPDATED: "+mysteryCrate.ID, mysteryCrate.Label, mysteryCrate.MysteryCrateType)
	}

	fmt.Println("Finish syncing crates")
	return nil
}

func SyncWeaponModel(dt DevTool) error {
	f, err := os.OpenFile("./devtool/temp-sync/supremacy-static-data/weapon_models.csv", os.O_RDONLY, 0755)
	if err != nil {
		log.Fatal("CANT OPEN FILE")
		return err
	}

	defer f.Close()

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
			ID:         record[0],
			Label:      record[2],
			WeaponType: record[3],
		}

		WeaponModels = append(WeaponModels, *weaponModel)
	}

	for _, weaponModel := range WeaponModels {
		_, err = dt.db.Exec(`UPDATE weapon_models SET id=$1 WHERE label=$2 AND weapon_type=$3 `, weaponModel.ID, weaponModel.Label, weaponModel.WeaponType)
		if err != nil {
			fmt.Println(err.Error()+weaponModel.ID, weaponModel.Label, weaponModel.WeaponType)
			continue
		}

		fmt.Println("UPDATED: "+weaponModel.ID, weaponModel.Label, weaponModel.WeaponType)
	}

	fmt.Println("Finish syncing weapon models")
	return nil
}

func SyncWeaponSkins(dt DevTool) error {
	f, err := os.OpenFile("./devtool/temp-sync/supremacy-static-data/weapon_skins.csv", os.O_RDONLY, 0755)
	if err != nil {
		log.Fatal("CANT OPEN FILE")
		return err
	}

	defer f.Close()

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
			ID:            record[0],
			Label:         record[1],
			WeaponType:    record[2],
			Tier:          record[3],
			WeaponModelID: record[13],
		}

		WeaponSkins = append(WeaponSkins, *weaponSkin)
	}

	for _, weaponSkin := range WeaponSkins {
		_, err = dt.db.Exec(`UPDATE blueprint_weapon_skin SET id=$1 WHERE label=$2 AND weapon_type=$3 AND tier=$4 AND weapon_model_id=$5 `, weaponSkin.ID, weaponSkin.Label, weaponSkin.WeaponType, weaponSkin.Tier, weaponSkin.WeaponModelID)
		if err != nil {
			fmt.Println(err.Error()+weaponSkin.ID, weaponSkin.Label, weaponSkin.WeaponType, weaponSkin.Tier, weaponSkin.WeaponModelID)
			continue
		}

		fmt.Println("UPDATED: "+weaponSkin.ID, weaponSkin.Label, weaponSkin.WeaponType, weaponSkin.Tier, weaponSkin.WeaponModelID)
	}

	fmt.Println("Finish syncing weapon skins")
	return nil
}
