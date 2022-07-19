package devtool

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"github.com/volatiletech/null/v8"
	"log"
	"os"
	"server/devtool/types"
	"strconv"
)

type DevTool struct {
	DB       *sql.DB
	FilePath string
}

func SyncTool(dt *DevTool) error {
	err := SyncFactions(dt)
	if err != nil {
		return err
	}
	err = SyncBrands(dt)
	if err != nil {
		return err
	}
	err = SyncMechSkins(dt)
	if err != nil {
		return err
	}
	err = SyncMechModels(dt)
	if err != nil {
		return err
	}

	//err = SyncMysteryCrates(dt)
	//if err != nil {
	//	return err
	//}

	err = SyncWeaponSkins(dt)
	if err != nil {
		return err
	}
	err = SyncWeaponModel(dt)
	if err != nil {
		return err
	}

	err = SyncBattleAbilities(dt)
	if err != nil {
		return err
	}

	err = SyncPowerCores(dt)
	if err != nil {
		return err
	}

	return nil
}

func RemoveFKContraints(dt DevTool) error {
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

func SyncMechModels(dt *DevTool) error {
	f, err := os.OpenFile(fmt.Sprintf("%smech_models.csv", dt.FilePath), os.O_RDONLY, 0755)
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

	count := 0

	for _, mechModel := range MechModels {
		brandID := &mechModel.BrandID.String
		if mechModel.BrandID.String == "" || !mechModel.BrandID.Valid {
			brandID = nil
		}

		_, err = dt.DB.Exec(`
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

func SyncMechSkins(dt *DevTool) error {
	f, err := os.OpenFile(fmt.Sprintf("%smech_skins.csv", dt.FilePath), os.O_RDONLY, 0755)
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
			BackgroundColor:  record[11],
			YoutubeURL:       record[12],
			MechType:         record[13],
			StatModifier:     record[14],
		}

		MechSkins = append(MechSkins, *mechModel)
	}

	for _, mechSkin := range MechSkins {
		imageURl := &mechSkin.ImageUrl.String
		if !mechSkin.ImageUrl.Valid || mechSkin.ImageUrl.String == "" {
			imageURl = nil
		}

		animationURL := &mechSkin.AnimationUrl.String
		if !mechSkin.AnimationUrl.Valid || mechSkin.AnimationUrl.String == "" {
			animationURL = nil
		}

		cardAnimationURL := &mechSkin.CardAnimationUrl.String
		if !mechSkin.CardAnimationUrl.Valid || mechSkin.CardAnimationUrl.String == "" {
			cardAnimationURL = nil
		}

		largeImageURL := &mechSkin.LargeImageUrl.String
		if !mechSkin.LargeImageUrl.Valid || mechSkin.LargeImageUrl.String == "" {
			largeImageURL = nil
		}

		avatarURL := &mechSkin.AvatarUrl.String
		if !mechSkin.AvatarUrl.Valid || mechSkin.AvatarUrl.String == "" {
			avatarURL = nil
		}

		statModifier := &mechSkin.StatModifier
		if mechSkin.StatModifier == "" {
			statModifier = nil
		}

		backgroundColor := &mechSkin.BackgroundColor
		if mechSkin.BackgroundColor == "" {
			backgroundColor = nil
		}

		youtubeURL := &mechSkin.YoutubeURL
		if mechSkin.YoutubeURL == "" {
			youtubeURL = nil
		}

		_, err = dt.DB.Exec(`
			INSERT INTO blueprint_mech_skin(id,collection, mech_model, label, tier, image_url, animation_url, card_animation_url, large_image_url, avatar_url,background_color, youtube_url, mech_type, stat_modifier)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
			ON CONFLICT (id)
			DO
			    UPDATE SET id=$1,collection=$2, mech_model=$3, label=$4, tier=$5, image_url=$6, animation_url=$7, card_animation_url=$8, large_image_url=$9, avatar_url=$10,background_color=$11, youtube_url=$12, mech_type=$13, stat_modifier=$14;
		`, mechSkin.ID, mechSkin.Collection, mechSkin.MechModel, mechSkin.Label, mechSkin.Tier, imageURl, animationURL, cardAnimationURL, largeImageURL, avatarURL, backgroundColor, youtubeURL, mechSkin.MechType, statModifier)
		if err != nil {
			fmt.Println(err.Error()+mechSkin.ID, mechSkin.MechModel)
			continue
		}

		fmt.Println("UPDATED: "+mechSkin.ID, mechSkin.Label, mechSkin.Tier, mechSkin.Collection, mechSkin.MechModel)
	}

	fmt.Println("Finish syncing mech skins")

	return nil

}

func SyncFactions(dt *DevTool) error {
	f, err := os.OpenFile(fmt.Sprintf("%sfactions.csv", dt.FilePath), os.O_RDONLY, 0755)
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
		_, err = dt.DB.Exec(`
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

	return nil
}

func SyncBrands(dt *DevTool) error {
	f, err := os.OpenFile(fmt.Sprintf("%sbrands.csv", dt.FilePath), os.O_RDONLY, 0755)
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
		_, err = dt.DB.Exec(`
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

func SyncMysteryCrates(dt *DevTool) error {
	f, err := os.OpenFile(fmt.Sprintf("%smystery_crates.csv", dt.FilePath), os.O_RDONLY, 0755)
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
		_, err = dt.DB.Exec(`
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

func SyncWeaponModel(dt *DevTool) error {
	f, err := os.OpenFile(fmt.Sprintf("%sweapon_models.csv", dt.FilePath), os.O_RDONLY, 0755)
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

		_, err = dt.DB.Exec(`
			INSERT INTO weapon_models(id, brand_id, label, weapon_type, default_skin_id, deleted_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
			ON CONFLICT (id)
			DO 
			    UPDATE SET id=$1, brand_id=$2, label=$3, weapon_type=$4, default_skin_id=$5, deleted_at=$6, updated_at=$7;
		`, weaponModel.ID, brandID, weaponModel.Label, weaponModel.WeaponType, defaultSkinID, deletedAt, weaponModel.UpdatedAt)
		if err != nil {
			fmt.Println(err.Error()+weaponModel.ID, weaponModel.Label, weaponModel.WeaponType)
			continue
		}

		fmt.Println("UPDATED: "+weaponModel.ID, weaponModel.Label, weaponModel.WeaponType)
	}

	fmt.Println("Finish syncing weapon models")
	return nil
}

func SyncWeaponSkins(dt *DevTool) error {
	f, err := os.OpenFile(fmt.Sprintf("%sweapon_skins.csv", dt.FilePath), os.O_RDONLY, 0755)
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
			ID:               record[0],
			Label:            record[1],
			WeaponType:       record[2],
			Tier:             record[3],
			ImageUrl:         record[4],
			CardAnimationUrl: record[5],
			AvatarUrl:        record[6],
			LargeImageUrl:    record[7],
			BackgroundColor:  record[8],
			AnimationUrl:     record[9],
			YoutubeUrl:       record[10],
			Collection:       record[11],
			WeaponModelID:    record[13],
			StatModifier:     record[14],
		}

		WeaponSkins = append(WeaponSkins, *weaponSkin)
	}

	for _, weaponSkin := range WeaponSkins {
		imageURL := &weaponSkin.ImageUrl
		if weaponSkin.ImageUrl == "" {
			imageURL = nil
		}

		cardAnimationURL := &weaponSkin.CardAnimationUrl
		if weaponSkin.CardAnimationUrl == "" {
			cardAnimationURL = nil
		}

		avatarURL := &weaponSkin.AvatarUrl
		if weaponSkin.AvatarUrl == "" {
			avatarURL = nil
		}

		largeImageURL := &weaponSkin.LargeImageUrl
		if weaponSkin.LargeImageUrl == "" {
			largeImageURL = nil
		}

		backgroundColor := &weaponSkin.BackgroundColor
		if weaponSkin.BackgroundColor == "" {
			backgroundColor = nil
		}

		animationURL := &weaponSkin.AnimationUrl
		if weaponSkin.AnimationUrl == "" {
			animationURL = nil
		}

		youtubeURL := &weaponSkin.YoutubeUrl
		if weaponSkin.YoutubeUrl == "" {
			youtubeURL = nil
		}

		statModifier := &weaponSkin.StatModifier
		if weaponSkin.StatModifier == "" {
			statModifier = nil
		}

		_, err = dt.DB.Exec(`
			INSERT INTO blueprint_weapon_skin(id, label, weapon_type, tier, image_url, card_animation_url, avatar_url, large_image_url, background_color, animation_url, youtube_url, collection, weapon_model_id, stat_modifier)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
			ON CONFLICT (id)
			DO 
			    UPDATE SET id=$1, label=$2, weapon_type=$3, tier=$4, image_url=$5, card_animation_url=$6, avatar_url=$7, large_image_url=$8, background_color=$9, animation_url=$10, youtube_url=$11, collection=$12, weapon_model_id=$13, stat_modifier=$14;
		`, weaponSkin.ID, weaponSkin.Label, weaponSkin.WeaponType, weaponSkin.Tier, imageURL, cardAnimationURL, avatarURL, largeImageURL, backgroundColor, animationURL, youtubeURL, weaponSkin.Collection, weaponSkin.WeaponModelID, statModifier)
		if err != nil {
			fmt.Println(err.Error()+weaponSkin.ID, weaponSkin.Label, weaponSkin.WeaponType, weaponSkin.Tier, weaponSkin.WeaponModelID)
			continue
		}

		fmt.Println("UPDATED: "+weaponSkin.ID, weaponSkin.Label, weaponSkin.WeaponType, weaponSkin.Tier, weaponSkin.WeaponModelID)
	}

	fmt.Println("Finish syncing weapon skins")
	return nil
}

func SyncBattleAbilities(dt *DevTool) error {
	f, err := os.OpenFile(fmt.Sprintf("%sbattle_abilities.csv", dt.FilePath), os.O_RDONLY, 0755)
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
		_, err = dt.DB.Exec(`
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

func SyncPowerCores(dt *DevTool) error {
	f, err := os.OpenFile(fmt.Sprintf("%spower_cores.csv", dt.FilePath), os.O_RDONLY, 0755)
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

		_, err = dt.DB.Exec(`
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
