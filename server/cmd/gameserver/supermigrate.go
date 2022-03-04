package main

import (
	"errors"
	"fmt"
	"math/rand"
	"server/comms"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gosimple/slug"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type AutoGenerated struct {
	AssetPayload    []AssetPayload    `json:"AssetPayload"`
	MetadataPayload []MetadataPayload `json:"MetadataPayload"`
	StorePayload    []StorePayload    `json:"StorePayload"`
	UserPayload     []UserPayload     `json:"UserPayload"`
	FactionPayload  []FactionPayload  `json:"FactionPayload"`
}
type FactionPayload struct {
	ID          uuid.UUID `json:"id"`
	Label       string    `json:"label"`
	Description string    `json:"mintingSignature"`
}
type AssetPayload struct {
	UserID           string        `json:"userID"`
	TransferredInAt  time.Time     `json:"transferredInAt"`
	FrozenByID       interface{}   `json:"frozenByID"`
	LockedByID       interface{}   `json:"lockedByID"`
	FrozenAt         interface{}   `json:"frozenAt"`
	MintingSignature string        `json:"mintingSignature"`
	TxHistory        []interface{} `json:"txHistory"`
	ExternalTokenID  string        `json:"externalTokenID"`
	CollectionID     string        `json:"collectionID"`
	MetadataHash     string        `json:"metadataHash"`
	SignatureExpiry  string        `json:"signatureExpiry"`
}
type Attributes struct {
	Value       interface{} `json:"value"`
	TraitType   string      `json:"trait_type"`
	DisplayType string      `json:"display_type,omitempty"`
}
type MetadataPayload struct {
	Name               string       `json:"name"`
	CollectionID       string       `json:"collectionID"`
	GameObject         interface{}  `json:"gameObject"`
	Description        interface{}  `json:"description"`
	ExternalURL        string       `json:"externalURL"`
	Image              string       `json:"image"`
	AnimationURL       string       `json:"animationURL"`
	Durability         int          `json:"durability"`
	Attributes         []Attributes `json:"attributes"`
	AdditionalMetadata interface{}  `json:"additionalMetadata"`
	Keywords           interface{}  `json:"keywords"`
	DeletedAt          interface{}  `json:"deletedAt"`
	UpdatedAt          time.Time    `json:"updatedAt"`
	CreatedAt          time.Time    `json:"createdAt"`
	Minted             bool         `json:"minted"`
	Hash               string       `json:"hash"`
	ExternalTokenID    string       `json:"externalTokenID"`
}
type StorePayload struct {
	ID                 string       `json:"id"`
	FactionID          string       `json:"factionID"`
	Name               string       `json:"name"`
	CollectionID       string       `json:"collectionID"`
	Description        string       `json:"description"`
	Image              string       `json:"image"`
	AnimationURL       string       `json:"animationURL"`
	Attributes         []Attributes `json:"attributes"`
	AdditionalMetadata interface{}  `json:"additionalMetadata"`
	Keywords           string       `json:"keywords"`
	UsdCentCost        int          `json:"usdCentCost"`
	AmountSold         int          `json:"amountSold"`
	AmountAvailable    int          `json:"amountAvailable"`
	SoldAfter          time.Time    `json:"soldAfter"`
	SoldBefore         time.Time    `json:"soldBefore"`
	DeletedAt          interface{}  `json:"deletedAt"`
	CreatedAt          time.Time    `json:"createdAt"`
	UpdatedAt          time.Time    `json:"updatedAt"`
	Restriction        string       `json:"restriction"`
}
type Metadata struct {
}
type UserPayload struct {
	ID                               string      `json:"id"`
	Username                         string      `json:"username"`
	RoleID                           string      `json:"roleID"`
	AvatarID                         interface{} `json:"avatarID"`
	FacebookID                       interface{} `json:"facebookID"`
	GoogleID                         interface{} `json:"googleID"`
	TwitchID                         interface{} `json:"twitchID"`
	TwitterID                        interface{} `json:"twitterID"`
	DiscordID                        interface{} `json:"discordID"`
	FactionID                        interface{} `json:"factionID"`
	Email                            interface{} `json:"email"`
	FirstName                        string      `json:"firstName"`
	LastName                         string      `json:"lastName"`
	Verified                         bool        `json:"verified"`
	OldPasswordRequired              bool        `json:"oldPasswordRequired"`
	TwoFactorAuthenticationActivated bool        `json:"twoFactorAuthenticationActivated"`
	TwoFactorAuthenticationSecret    string      `json:"twoFactorAuthenticationSecret"`
	TwoFactorAuthenticationIsSet     bool        `json:"twoFactorAuthenticationIsSet"`
	Sups                             string      `json:"sups"`
	PublicAddress                    string      `json:"publicAddress"`
	PrivateAddress                   interface{} `json:"privateAddress"`
	Nonce                            interface{} `json:"nonce"`
	Keywords                         string      `json:"keywords"`
	DeletedAt                        interface{} `json:"deletedAt"`
	UpdatedAt                        time.Time   `json:"updatedAt"`
	CreatedAt                        time.Time   `json:"createdAt"`
	Metadata                         Metadata    `json:"metadata"`
}

var AssetProductMap = map[string]string{}

// func ProcessAsset(asset *AssetPayload, metadata *MetadataPayload) (*boiler.Mech, *boiler.Chassis, []*boiler.Weapon, []*boiler.Module, error) {
// 	exists, err := boiler.Mechs(qm.Where("hash = ?", asset.MetadataHash)).Exists(gamedb.StdConn)
// 	if err != nil {
// 		return nil, terror.Error(err)
// 	}
// 	if exists {
// 		gamelog.GameLog.Debug().Str("hash", asset.MetadataHash).Msg("mech exists, skipping...")
// 		return nil, nil
// 	}

// 	label := ""
// 	subModel := ""
// 	model := ""
// 	for _, att := range metadata.Attributes {
// 		if att.TraitType == "Name" {
// result.NameLabelAtt.Value.(string).(string)
// 		}
// 		if att.TraitType == "SubModel" {
// result.SubModelSubModelAtt.Value.(string).(string)
// 		}
// 		if att.TraitType == "Model" {
// result.ModelModelAtt.Value.(string).(string)
// 		}
// 	}

// 	mech := &boiler.Mech{
// 		OwnerID:    asset.UserID,
// 		TemplateID: "",
// 		Hash:       asset.MetadataHash,
// 		Label:      label,
// 		Skin:       subModel,
// 		Model:      model,
// 		BrandID:    "",
// 		Slug:       slug.Make(label),
// 		ChassisID:  "",
// 	}

// 	weapons, err := ProcessWeapons(metadata.Attributes)
// 	if err != nil {
// 		return nil, nil, nil, nil, err
// 	}
// 	modules, err := ProcessModules(metadata.Attributes)
// 	if err != nil {
// 		return nil, nil, nil, nil, err
// 	}
// 	chassis, err := ProcessChassis(metadata.Attributes)
// 	if err != nil {
// 		return nil, nil, nil, nil, err
// 	}

// 	return mech, nil
// }

// 	return result, nil
// }
// func ProcessWeapons(attributes []Attributes) ([]*boiler.Weapon, error) {
// 	return nil, nil
// }
// func ProcessModules(attributes []Attributes) ([]*boiler.Module, error) {
// 	return nil, nil
// }

type ParsedAttributes struct {
	Brand                 string
	Model                 string
	SubModel              string
	Rarity                string
	AssetType             string
	MaxStructureHitPoints int
	MaxShieldHitPoints    int
	Name                  string
	Speed                 int
	WeaponHardpoints      int
	TurretHardpoints      int
	UtilitySlots          int
	WeaponOne             string
	WeaponTwo             string
	TurretOne             string
	TurretTwo             string
	UtilityOne            string
	ShieldRechargeRate    int
}

func GetAttributes(attributes []Attributes) *ParsedAttributes {
	result := &ParsedAttributes{}
	for _, att := range attributes {
		if att.TraitType == "Brand" {
			result.Brand = att.Value.(string)
		}
		if att.TraitType == "Model" {
			result.Model = att.Value.(string)
		}
		if att.TraitType == "SubModel" {
			result.SubModel = att.Value.(string)
		}
		if att.TraitType == "Rarity" {
			result.Rarity = att.Value.(string)
		}
		if att.TraitType == "Asset Type" {
			result.AssetType = att.Value.(string)
		}
		if att.TraitType == "Max Structure Hit Points" {
			result.MaxStructureHitPoints = int(att.Value.(float64))
		}
		if att.TraitType == "Max Shield Hit Points" {
			result.MaxShieldHitPoints = int(att.Value.(float64))
		}
		if att.TraitType == "Name" {
			result.Name = att.Value.(string)
		}
		if att.TraitType == "Speed" {
			result.Speed = int(att.Value.(float64))
		}
		if att.TraitType == "Weapon Hardpoints" {
			result.WeaponHardpoints = int(att.Value.(float64))
		}
		if att.TraitType == "Turret Hardpoints" {
			result.TurretHardpoints = int(att.Value.(float64))
		}
		if att.TraitType == "Utility Slots" {
			result.UtilitySlots = int(att.Value.(float64))
		}
		if att.TraitType == "Weapon One" {
			result.WeaponOne = att.Value.(string)
		}
		if att.TraitType == "Weapon Two" {
			result.WeaponTwo = att.Value.(string)
		}
		if att.TraitType == "Turret One" {
			result.TurretOne = att.Value.(string)
		}
		if att.TraitType == "Turret Two" {
			result.TurretTwo = att.Value.(string)
		}
		if att.TraitType == "Utility One" {
			result.UtilityOne = att.Value.(string)
		}
		if att.TraitType == "Shield Recharge Rate" {
			result.ShieldRechargeRate = int(att.Value.(float64))
		}
	}
	return result
}

func ProcessChassis(brand *boiler.Brand, attributes []Attributes) (*boiler.Chassis, error) {
	att := GetAttributes(attributes)
	label := fmt.Sprintf("%s %s %s %s Chassis", att.Brand, att.Model, att.SubModel, att.Name)
	result := &boiler.Chassis{
		ID:                 uuid.Must(uuid.NewV4()).String(),
		ShieldRechargeRate: att.ShieldRechargeRate,
		BrandID:            brand.ID,
		MaxShield:          att.MaxShieldHitPoints,
		Label:              label,
		Slug:               slug.Make(fmt.Sprintf("%s#%d", label, 1000+rand.Intn(8999))),
		HealthRemaining:    att.MaxStructureHitPoints,
		WeaponHardpoints:   att.WeaponHardpoints,
		TurretHardpoints:   att.TurretHardpoints,
		UtilitySlots:       att.UtilitySlots,
		Speed:              att.Speed,
		MaxHitpoints:       att.MaxStructureHitPoints,
	}
	return result, nil
}
func ProcessBlueprintChassis(brand *boiler.Brand, attributes []Attributes) (*boiler.BlueprintChassis, error) {
	att := GetAttributes(attributes)
	label := fmt.Sprintf("%s %s %s %s Chassis", att.Brand, att.Model, att.SubModel, att.Name)
	result := &boiler.BlueprintChassis{
		ID:                 uuid.Must(uuid.NewV4()).String(),
		ShieldRechargeRate: att.ShieldRechargeRate,
		BrandID:            brand.ID,
		MaxShield:          att.MaxShieldHitPoints,
		Label:              label,
		Slug:               slug.Make(fmt.Sprintf("%s#%d", label, 1000+rand.Intn(8999))),
		WeaponHardpoints:   att.WeaponHardpoints,
		TurretHardpoints:   att.TurretHardpoints,
		UtilitySlots:       att.UtilitySlots,
		Speed:              att.Speed,
		MaxHitpoints:       att.MaxStructureHitPoints,
	}
	return result, nil
}
func ProcessModule(brand *boiler.Brand, attributes []Attributes) (*boiler.Module, error) {
	att := GetAttributes(attributes)
	label := att.UtilityOne
	result := &boiler.Module{
		ID:               uuid.Must(uuid.NewV4()).String(),
		BrandID:          brand.ID,
		Label:            att.UtilityOne,
		Slug:             slug.Make(fmt.Sprintf("%s#%d", label, 1000+rand.Intn(8999))),
		HitpointModifier: 100,
		ShieldModifier:   100,
	}
	return result, nil
}
func ProcessBlueprintModule(brand *boiler.Brand, attributes []Attributes) (*boiler.BlueprintModule, error) {
	att := GetAttributes(attributes)
	label := att.UtilityOne
	result := &boiler.BlueprintModule{
		ID:               uuid.Must(uuid.NewV4()).String(),
		BrandID:          brand.ID,
		Label:            att.UtilityOne,
		Slug:             slug.Make(fmt.Sprintf("%s#%d", label, 1000+rand.Intn(8999))),
		HitpointModifier: 100,
		ShieldModifier:   100,
	}
	return result, nil
}
func ProcessBlueprintWeapon(weaponType string, index int, brand *boiler.Brand, attributes []Attributes) (*boiler.BlueprintWeapon, error) {
	att := GetAttributes(attributes)
	label := ""
	weapslug := ""
	if weaponType == "TURRET" {
		if att.TurretHardpoints == 0 {
			return nil, nil
		}
		if index == 1 {
			label = att.TurretOne
			weapslug = slug.Make(fmt.Sprintf("%s#%d", label, 1000+rand.Intn(8999)))
		}
		if index == 2 {
			label = att.TurretTwo
			weapslug = slug.Make(fmt.Sprintf("%s#%d", label, 1000+rand.Intn(8999)))
		}
	}

	if weaponType == "ARM" {
		if att.WeaponHardpoints == 0 {
			return nil, nil
		}
		if index == 1 {
			label = att.WeaponOne
			weapslug = slug.Make(fmt.Sprintf("%s#%d", label, 1000+rand.Intn(8999)))
		}
		if index == 2 {
			label = att.WeaponTwo
			weapslug = slug.Make(fmt.Sprintf("%s#%d", label, 1000+rand.Intn(8999)))
		}
	}

	if label == "" || weapslug == "" {
		gamelog.GameLog.Debug().Interface("att", att).Str("weapon_type", weaponType).Int("index", index).Msg("attributes")
		return nil, errors.New("could not find label, weapon or type")
	}
	result := &boiler.BlueprintWeapon{
		ID:         uuid.Must(uuid.NewV4()).String(),
		BrandID:    brand.ID,
		Label:      label,
		Slug:       weapslug,
		Damage:     -1,
		WeaponType: weaponType,
	}
	return result, nil
}

func ProcessTemplate(data *StorePayload) error {
	att := GetAttributes(data.Attributes)

	brand, err := boiler.Brands(qm.Where("label = ?", att.Brand)).One(gamedb.StdConn)
	if err != nil {
		return err
	}
	chassis, err := ProcessBlueprintChassis(brand, data.Attributes)
	if err != nil {
		return err
	}

	weapon1, err := ProcessBlueprintWeapon("ARM", 1, brand, data.Attributes)
	if err != nil {
		return err
	}

	weapon2, err := ProcessBlueprintWeapon("ARM", 2, brand, data.Attributes)
	if err != nil {
		return err
	}

	turret1, err := ProcessBlueprintWeapon("TURRET", 1, brand, data.Attributes)
	if err != nil {
		return err
	}

	turret2, err := ProcessBlueprintWeapon("TURRET", 2, brand, data.Attributes)
	if err != nil {
		return err
	}

	module, err := ProcessBlueprintModule(brand, data.Attributes)
	if err != nil {
		return err
	}
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	gamelog.GameLog.Debug().Msg("inserting chassis")
	err = chassis.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return err
	}

	newTemplate := &boiler.Template{
		ID:                 data.ID,
		BlueprintChassisID: chassis.ID,
		Label:              data.Name,
	}

	err = newTemplate.Insert(tx, boil.Infer())
	if err != nil {
		return err
	}

	if weapon1 != nil {
		gamelog.GameLog.Debug().Msg("inserting weapon 1")
		err = weapon1.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return err
		}

	}
	if weapon2 != nil {
		gamelog.GameLog.Debug().Msg("inserting weapon 2")
		err = weapon2.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return err
		}
	}
	if turret1 != nil {
		gamelog.GameLog.Debug().Msg("inserting turret 1")
		err = turret1.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return err
		}
	}
	if turret2 != nil {
		gamelog.GameLog.Debug().Msg("inserting turret 2")
		err = turret2.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return err
		}
	}
	gamelog.GameLog.Debug().Msg("inserting module")
	err = module.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return err
	}
	tx.Commit()

	return nil
}

func SuperMigrate(passportRPC *comms.C) error {
	result := &comms.GetAll{}
	err := passportRPC.Call("C.SuperMigrate", comms.GetAllReq{}, result)
	if err != nil {
		return terror.Error(err)
	}
	metadataPayload := []*MetadataPayload{}
	err = result.MetadataPayload.Unmarshal(&metadataPayload)
	if err != nil {
		return terror.Error(err)
	}
	assetPayload := []*AssetPayload{}
	err = result.AssetPayload.Unmarshal(&assetPayload)
	if err != nil {
		return terror.Error(err)
	}
	storePayload := []*StorePayload{}
	err = result.StorePayload.Unmarshal(&storePayload)
	if err != nil {
		return terror.Error(err)
	}
	factionPayload := []*FactionPayload{}
	err = result.FactionPayload.Unmarshal(&factionPayload)
	if err != nil {
		return terror.Error(err)
	}
	// Begin processing

	for _, syndicate := range factionPayload {
		// Process syndicates
		exists, err := boiler.Syndicates(qm.Where("label = ?", syndicate.Label)).Exists(gamedb.StdConn)
		if err != nil {
			return fmt.Errorf("check syndicate exists: %w", err)
		}
		if exists {
			gamelog.GameLog.Debug().Str("label", syndicate.Label).Msg("syndicate exists, skipping")
			continue
		}
		gamelog.GameLog.Debug().Str("label", syndicate.Label).Msg("inserting syndicate")
		newSyndicate := &boiler.Syndicate{
			ID:          syndicate.ID.String(),
			Description: syndicate.Description,
			Label:       syndicate.Label,
		}
		err = newSyndicate.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return fmt.Errorf("insert syndicate: %w", err)
		}

		// Process brands (basically syndicates)
		exists, err = boiler.Brands(qm.Where("label = ?", syndicate.Label)).Exists(gamedb.StdConn)
		if err != nil {
			return fmt.Errorf("check exist brand: %w", err)
		}
		if exists {
			gamelog.GameLog.Debug().Str("label", syndicate.Label).Msg("brand exists, skipping")
			continue
		}
		existingSyndicate, err := boiler.Syndicates(qm.Where("label = ?", syndicate.Label)).One(gamedb.StdConn)
		if err != nil {
			return fmt.Errorf("get existing syndicate: %w", err)
		}
		gamelog.GameLog.Debug().Str("label", syndicate.Label).Msg("inserting brand")
		newBrand := &boiler.Brand{
			ID:          syndicate.ID.String(),
			SyndicateID: existingSyndicate.ID,
			Label:       syndicate.Label,
		}
		err = newBrand.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return fmt.Errorf("insert brand: %w", err)
		}
	}

	for _, item := range storePayload {
		err = ProcessTemplate(item)
		if err != nil {
			return fmt.Errorf("process template: %w", err)
		}
	}

	// // Check that every asset has a matching metadata
	// for _, asset := range assetPayload {
	// 	for _, meta := range metadataPayload {
	// 		if meta.Hash == asset.MetadataHash {
	// 			continue
	// 		}
	// 		return errors.New("no matching hash found between asset and metadata")
	// 	}
	// }

	// for _, asset := range assetPayload {
	// 	for _, meta := range metadataPayload {
	// 		if meta.Hash != asset.MetadataHash {
	// 			continue
	// 		}
	// 		syndicate, brand, mech, weapons, modules, err := ProcessAsset(asset, meta)
	// 		if err != nil {
	// 			return terror.Error(err)
	// 		}

	// 	}
	// }

	// userPayload := []*UserPayload{}
	// err = result.UserPayload.Unmarshal(&userPayload)
	// if err != nil {
	// 	return terror.Error(err)
	// }
	// for _, user := range userPayload {
	// 	fmt.Println(user)
	// }
	return nil
}