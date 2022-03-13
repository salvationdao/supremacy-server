package supermigrate

import (
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/volatiletech/null/v8"
)

type AutoGenerated struct {
	AssetPayload    []AssetPayload    `json:"asset_payload"`
	MetadataPayload []MetadataPayload `json:"metadata_payload"`
	StorePayload    []StorePayload    `json:"store_payload"`
	UserPayload     []UserPayload     `json:"user_payload"`
	FactionPayload  []FactionPayload  `json:"faction_payload"`
}
type FactionPayload struct {
	ID          uuid.UUID `json:"id"`
	Label       string    `json:"label"`
	Description string    `json:"minting_signature"`
}
type AssetPayload struct {
	UserID           string        `json:"user_id"`
	TransferredInAt  time.Time     `json:"transferred_in_at"`
	FrozenByID       interface{}   `json:"frozen_by_id"`
	LockedByID       interface{}   `json:"locked_by_id"`
	FrozenAt         interface{}   `json:"frozen_at"`
	MintingSignature string        `json:"minting_signature"`
	TxHistory        []interface{} `json:"tx_history"`
	ExternalTokenID  string        `json:"external_token_id"`
	CollectionID     string        `json:"collection_id"`
	MetadataHash     string        `json:"metadata_hash"`
	SignatureExpiry  string        `json:"signature_expiry"`
}
type Attributes struct {
	Value       interface{} `json:"value"`
	TraitType   string      `json:"trait_type"`
	DisplayType string      `json:"display_type,omitempty"`
}

type MetadataPayload struct {
	Name               string       `json:"name"`
	CollectionID       string       `json:"collection_id"`
	GameObject         interface{}  `json:"game_object"`
	Description        interface{}  `json:"description"`
	ExternalURL        string       `json:"external_url"`
	Image              string       `json:"image"`
	AnimationURL       string       `json:"animation_url"`
	Durability         int          `json:"durability"`
	Attributes         []Attributes `json:"attributes"`
	AdditionalMetadata interface{}  `json:"additional_metadata"`
	Keywords           interface{}  `json:"keywords"`
	DeletedAt          interface{}  `json:"deleted_at"`
	UpdatedAt          time.Time    `json:"updated_at"`
	CreatedAt          time.Time    `json:"created_at"`
	Minted             bool         `json:"minted"`
	Hash               string       `json:"hash"`
	ExternalTokenID    string       `json:"external_token_id"`
	ImageAvatar        string       `json:"image_avatar"`
}

type StorePayload struct {
	ID                 string       `json:"id"`
	FactionID          string       `json:"faction_id"`
	Name               string       `json:"name"`
	CollectionID       string       `json:"collection_id"`
	Description        string       `json:"description"`
	Image              string       `json:"image"`
	AnimationURL       string       `json:"animation_url"`
	Attributes         []Attributes `json:"attributes"`
	AdditionalMetadata interface{}  `json:"additional_metadata"`
	Keywords           string       `json:"keywords"`
	UsdCentCost        int          `json:"usd_cent_cost"`
	AmountSold         int          `json:"amount_sold"`
	AmountAvailable    int          `json:"amount_available"`
	SoldAfter          time.Time    `json:"sold_after"`
	SoldBefore         time.Time    `json:"sold_before"`
	DeletedAt          interface{}  `json:"deleted_at"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"`
	Restriction        string       `json:"restriction"`
}
type Metadata struct {
}
type UserPayload struct {
	ID                               string      `json:"id"`
	Username                         string      `json:"username"`
	RoleID                           string      `json:"role_id"`
	AvatarID                         interface{} `json:"avatar_id"`
	FacebookID                       interface{} `json:"facebook_id"`
	GoogleID                         interface{} `json:"google_id"`
	TwitchID                         interface{} `json:"twitch_id"`
	TwitterID                        interface{} `json:"twitter_id"`
	DiscordID                        interface{} `json:"discord_id"`
	FactionID                        null.String `json:"faction_id"`
	Email                            interface{} `json:"email"`
	FirstName                        string      `json:"first_name"`
	LastName                         string      `json:"last_name"`
	Verified                         bool        `json:"verified"`
	OldPasswordRequired              bool        `json:"old_password_required"`
	TwoFactorAuthenticationActivated bool        `json:"two_factor_authentication_activated"`
	TwoFactorAuthenticationSecret    string      `json:"two_factor_authentication_secret"`
	TwoFactorAuthenticationIsSet     bool        `json:"two_factor_authentication_is_set"`
	Sups                             string      `json:"sups"`
	PublicAddress                    string      `json:"public_address"`
	PrivateAddress                   interface{} `json:"private_address"`
	Nonce                            interface{} `json:"nonce"`
	Keywords                         string      `json:"keywords"`
	DeletedAt                        interface{} `json:"deleted_at"`
	UpdatedAt                        time.Time   `json:"updated_at"`
	CreatedAt                        time.Time   `json:"created_at"`
	Metadata                         Metadata    `json:"metadata"`
}

var AssetProductMap = map[string]string{}

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
			switch att.Value.(type) {
			case float64:
				result.MaxStructureHitPoints = int(att.Value.(float64))
			case int:
				result.MaxStructureHitPoints = att.Value.(int)
			case string:
				s, err := strconv.Atoi(att.Value.(string))
				if err != nil {
					result.MaxStructureHitPoints = 1000
				}
				result.MaxStructureHitPoints = s
			default:
				result.MaxStructureHitPoints = 1000
			}
		}
		if att.TraitType == "Max Shield Hit Points" {
			switch att.Value.(type) {
			case float64:
				result.MaxShieldHitPoints = int(att.Value.(float64))
			case int:
				result.MaxShieldHitPoints = att.Value.(int)
			case string:
				s, err := strconv.Atoi(att.Value.(string))
				if err != nil {
					result.MaxShieldHitPoints = 1000
				}
				result.MaxShieldHitPoints = s
			default:
				result.MaxShieldHitPoints = 1000
			}
		}
		if att.TraitType == "Name" {
			result.Name = att.Value.(string)
		}
		if att.TraitType == "Speed" {
			switch att.Value.(type) {
			case float64:
				result.Speed = int(att.Value.(float64))
			case int:
				result.Speed = att.Value.(int)
			case string:
				s, err := strconv.Atoi(att.Value.(string))
				if err != nil {
					result.Speed = 1750
				}
				result.Speed = s
			default:
				result.Speed = 1750
			}
		}
		if att.TraitType == "Weapon Hardpoints" {
			switch att.Value.(type) {
			case float64:
				result.WeaponHardpoints = int(att.Value.(float64))
			case int:
				result.WeaponHardpoints = att.Value.(int)
			case string:
				s, err := strconv.Atoi(att.Value.(string))
				if err != nil {
					result.WeaponHardpoints = 2
				}
				result.WeaponHardpoints = s
			default:
				result.WeaponHardpoints = 2
			}
		}
		if att.TraitType == "Turret Hardpoints" {
			switch att.Value.(type) {
			case float64:
				result.TurretHardpoints = int(att.Value.(float64))
			case int:
				result.TurretHardpoints = att.Value.(int)
			case string:
				s, err := strconv.Atoi(att.Value.(string))
				if err != nil {
					result.TurretHardpoints = 2
				}
				result.TurretHardpoints = s
			default:
				result.TurretHardpoints = 2
			}
		}
		if att.TraitType == "Utility Slots" {
			switch att.Value.(type) {
			case float64:
				result.UtilitySlots = int(att.Value.(float64))
			case int:
				result.UtilitySlots = att.Value.(int)
			case string:
				s, err := strconv.Atoi(att.Value.(string))
				if err != nil {
					result.UtilitySlots = 2
				}
				result.UtilitySlots = s
			default:
				result.UtilitySlots = 2
			}
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
			switch att.Value.(type) {
			case float64:
				result.ShieldRechargeRate = int(att.Value.(float64))
			case int:
				result.ShieldRechargeRate = att.Value.(int)
			case string:
				s, err := strconv.Atoi(att.Value.(string))
				if err != nil {
					result.ShieldRechargeRate = 80
				}
				result.ShieldRechargeRate = s
			default:
				result.ShieldRechargeRate = 80
			}
		}
	}
	return result
}
