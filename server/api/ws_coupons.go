package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpctypes"
	"server/xsyn_rpcclient"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/kevinms/leakybucket-go"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type CouponController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

func NewCouponsController(api *API) *CouponController {
	couponHub := &CouponController{
		API: api,
	}

	api.SecureUserFactionCommand(HubKeyCodeRedemption, couponHub.CodeRedemptionHandler)

	return couponHub
}

//retrieve code and redeem

type CodeRedemptionRequest struct {
	Payload struct {
		Code string `json:"code"`
	} `json:"payload"`
}

type Reward struct {
	Label       string              `json:"label"`
	ImageURL    null.String         `json:"image_url"`
	LockedUntil null.Time           `json:"locked_until"`
	Amount      decimal.NullDecimal `json:"amount"`
}

type CodeRedemptionResponse struct {
	Rewards []*Reward `json:"rewards"`
}

const HubKeyCodeRedemption = "CODE:REDEMPTION"

var bck = leakybucket.NewLeakyBucket(0.5, 1)

func (cc *CouponController) CodeRedemptionHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	v := bck.Add(1)
	if v == 0 {
		return terror.Error(fmt.Errorf("too many code redemption requests"), "Currently handling request, please try again.")
	}

	req := &CodeRedemptionRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	couponCode := req.Payload.Code

	rollbackRedeem := func() {
		_, err := boiler.NewQuery(
			qm.SQL(
				fmt.Sprintf(`
					UPDATE %s SET %s = false 
					WHERE %s = $1`,
					boiler.TableNames.Coupons,
					boiler.CouponColumns.Redeemed,
					boiler.CouponColumns.Code,
				),
				couponCode,
			),
		).Exec(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("coupon code: ", couponCode).Msg("handle")
			return
		}
	}

	coupon := &boiler.Coupon{}
	err = boiler.NewQuery(
		qm.SQL(
			fmt.Sprintf(`
					UPDATE %s SET %s = true 
					WHERE %s IS FALSE 
					AND %s = $1
					AND %s > NOW()
					RETURNING  %s, %s, %s, %s`,
				boiler.TableNames.Coupons,
				boiler.CouponColumns.Redeemed,
				boiler.CouponColumns.Redeemed,
				boiler.CouponColumns.Code,
				boiler.CouponColumns.ExpiryDate,
				boiler.CouponColumns.ID,
				boiler.CouponColumns.Code,
				boiler.CouponColumns.Redeemed,
				boiler.CouponColumns.ExpiryDate,
			),
			couponCode,
		),
	).QueryRow(gamedb.StdConn).Scan(&coupon.ID, &coupon.Code, &coupon.Redeemed, &coupon.ExpiryDate)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			rollbackRedeem()
			gamelog.L.Error().Err(err).Interface("coupon code: ", couponCode).Msg("failed to find coupon code")
			return terror.Error(err, "Issue finding coupon code, try again or contact support.")

		} else {
			return terror.Error(fmt.Errorf("unable to find unclaimed coupon"))
		}
	}

	err = coupon.L.LoadCouponItems(gamedb.StdConn, true, coupon,
		boiler.CouponItemWhere.Claimed.EQ(false),
	)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("coupon code: ", couponCode).Msg("failed to find coupon code loading items")
		return err
	}

	//get mech crates
	storeMechCrate, err := boiler.StorefrontMysteryCrates(
		boiler.StorefrontMysteryCrateWhere.MysteryCrateType.EQ(boiler.CrateTypeMECH),
		boiler.StorefrontMysteryCrateWhere.FactionID.EQ(factionID),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get mech crate for claim, please try again or contact support.")
	}
	//get mech crates
	storeWeaponCrate, err := boiler.StorefrontMysteryCrates(
		boiler.StorefrontMysteryCrateWhere.MysteryCrateType.EQ(boiler.CrateTypeWEAPON),
		boiler.StorefrontMysteryCrateWhere.FactionID.EQ(factionID),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get mech crate for claim, please try again or contact support.")
	}

	var rewards []*Reward

	for _, ci := range coupon.R.CouponItems {
		if ci.Claimed {
			continue
		}

		tx, err := gamedb.StdConn.Begin()
		defer tx.Rollback()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("unable to begin tx")
			rollbackRedeem()
			return terror.Error(err, "Issue claiming mystery crate, please try again or contact support.")
		}

		switch ci.ItemType {
		case boiler.CouponItemTypeSUPS:
			txID, err := transferSups(user.ID, ci.Amount.Decimal.String(), cc.API, req.Payload.Code)
			if err != nil {
				rollbackRedeem()
				return terror.Error(err, "Issue claiming $SUPS, please try again or contact support.")
			}
			reward := &Reward{
				Label:  "Sups",
				Amount: ci.Amount,
			}

			ci.TransactionID = null.StringFrom(txID)
			rewards = append(rewards, reward)
		case boiler.CouponItemTypeMECH_CRATE:
			assignedMechCrate, err := assignAndRegisterClaimCrate(user.ID, storeMechCrate, tx, cc.API)
			if err != nil {
				rollbackRedeem()
				return terror.Error(err, "Issue claiming mech crate, please try again or contact support.")
			}
			ci.ItemID = null.StringFrom(assignedMechCrate.ID)
			reward := &Reward{
				Label:       storeMechCrate.MysteryCrateType,
				ImageURL:    assignedMechCrate.ImageURL,
				LockedUntil: null.TimeFrom(assignedMechCrate.LockedUntil),
				Amount:      ci.Amount,
			}

			rewards = append(rewards, reward)
		case boiler.CouponItemTypeWEAPON_CRATE:
			assignedWeaponCrate, err := assignAndRegisterClaimCrate(user.ID, storeWeaponCrate, tx, cc.API)
			if err != nil {
				rollbackRedeem()
				return terror.Error(err, "Issue claiming weapon crate, please try again or contact support.")
			}
			ci.ItemID = null.StringFrom(assignedWeaponCrate.ID)
			reward := &Reward{
				Label:       storeWeaponCrate.MysteryCrateType,
				ImageURL:    assignedWeaponCrate.ImageURL,
				LockedUntil: null.TimeFrom(assignedWeaponCrate.LockedUntil),
				Amount:      ci.Amount,
			}
			rewards = append(rewards, reward)
		case boiler.CouponItemTypeGENESIS_MECH:
			//	TODO: genesis mech handle
			continue
		default:
			rollbackRedeem()
			return terror.Error(fmt.Errorf("invalid coupon item type: %s", ci.ItemType))
		}

		ci.Claimed = true
		_, err = ci.Update(tx, boil.Infer())
		if err != nil {
			rollbackRedeem()
			return terror.Error(err, "Issue claiming mystery crate, please try again or contact support.")
		}

		err = tx.Commit()
		if err != nil {
			rollbackRedeem()
			gamelog.L.Error().Err(err).Msg("failed to commit mystery crate transaction")
			return terror.Error(err, "Issue claiming mystery crate, please try again or contact support.")
		}

	}

	reply(CodeRedemptionResponse{
		Rewards: rewards,
	})
	return nil
}

func transferSups(userID string, amount string, api *API, code string) (string, error) {
	txID, err := api.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		Amount:               amount,
		FromUserID:           uuid.FromStringOrNil(server.SupremacyGameUserID),
		ToUserID:             uuid.FromStringOrNil(userID),
		TransactionReference: server.TransactionReference(fmt.Sprintf("coupon_redemption_code %s, by: %s |%d", code, userID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             "Coupon Redemption",
		Description:          fmt.Sprintf("Coupon redemption code: %s", code),
		NotSafe:              true,
	})
	if err != nil {
		return "", terror.Error(err, "Could not transfer user SUPS from Supremacy Game User, try again or contact support")
	}
	return txID, nil
}

func assignAndRegisterClaimCrate(userID string, storeCrate *boiler.StorefrontMysteryCrate, tx *sql.Tx, api *API) (*server.MysteryCrate, error) {
	assignedCrate, err := boiler.MysteryCrates(
		boiler.MysteryCrateWhere.FactionID.EQ(storeCrate.FactionID),
		boiler.MysteryCrateWhere.Type.EQ(storeCrate.MysteryCrateType),
		qm.Load(boiler.MysteryCrateRels.MysteryCrateBlueprints),
		qm.OrderBy("RANDOM()"),
	).One(tx)
	if err != nil {
		return nil, terror.Error(err, "Failed to get available crates, please try again or contact support.")
	}
	faction, err := boiler.FindFaction(tx, storeCrate.FactionID)
	if err != nil {
		return nil, terror.Error(err, "Failed to find faction, please try again or contact support.")
	}

	//copy assigned crate
	copiedMC := &boiler.MysteryCrate{
		Type:        assignedCrate.Type,
		FactionID:   assignedCrate.FactionID,
		Label:       assignedCrate.Label,
		Opened:      assignedCrate.Opened,
		LockedUntil: assignedCrate.LockedUntil,
		Purchased:   true,
		UpdatedAt:   time.Now(),
		CreatedAt:   time.Now(),
		Description: "",
	}
	err = copiedMC.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err, "Failed to copy mystery crate, please try again or contact support.")
	}

	for _, bpmc := range assignedCrate.R.MysteryCrateBlueprints {
		copiedBPMC := &boiler.MysteryCrateBlueprint{
			MysteryCrateID: copiedMC.ID,
			BlueprintType:  bpmc.BlueprintType,
			BlueprintID:    bpmc.BlueprintID,
			UpdatedAt:      time.Now(),
			CreatedAt:      time.Now(),
		}

		err = copiedBPMC.Insert(tx, boil.Infer())
		if err != nil {
			return nil, terror.Error(err, "Failed to copy mystery crate blueprints, please try again or contact support.")
		}
	}

	collectionItem, err := db.InsertNewCollectionItem(tx,
		"supremacy-general",
		boiler.ItemTypeMysteryCrate,
		copiedMC.ID,
		"",
		userID,
		storeCrate.ImageURL,
		storeCrate.CardAnimationURL,
		storeCrate.AvatarURL,
		storeCrate.LargeImageURL,
		storeCrate.BackgroundColor,
		storeCrate.AnimationURL,
		storeCrate.YoutubeURL,
	)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("mystery crate", copiedMC).Msg("failed to insert into collection items")
		return nil, terror.Error(err, "Failed to purchase mystery crate, please try again or contact support.")
	}

	//register
	assignedCrateServer := server.MysteryCrateFromBoiler(copiedMC, collectionItem)
	xsynAsset := rpctypes.ServerMysteryCrateToXsynAsset(assignedCrateServer, faction.Label)

	err = api.Passport.AssetRegister(xsynAsset)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("mystery crate", copiedMC).Msg("failed to register to XSYN")
		return nil, terror.Error(err, "Failed to get mystery crate, please try again or contact support.")
	}

	return assignedCrateServer, nil
}
