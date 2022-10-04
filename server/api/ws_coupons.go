package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"server"
	"server/benchmark"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"sync"
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

type RedeemFailUser struct {
	Count         int
	DeleteAt      time.Time
	LockedUntilAt null.Time
}

type CouponController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API

	redeemedFailUsersMut sync.RWMutex
	redeemedFailUsers    map[string]RedeemFailUser
}

func NewCouponsController(api *API) *CouponController {
	couponHub := &CouponController{
		API:               api,
		redeemedFailUsers: make(map[string]RedeemFailUser),
	}

	api.SecureUserFactionCommand(HubKeyCodeRedemption, couponHub.CodeRedemptionHandler)

	go couponHub.RunRedeemFailUserGC()

	return couponHub
}

func (cc *CouponController) RunRedeemFailUserGC() {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the CouponController!", r)
		}
	}()

	mainTicker := time.NewTicker(1 * time.Minute)

	for range mainTicker.C {
		bm := benchmark.New()

		bm.Start("gc_coupon_redeem_fail_users")

		deleteKeys := []string{}

		cc.redeemedFailUsersMut.RLock()
		for userID, failData := range cc.redeemedFailUsers {
			if (!failData.LockedUntilAt.Valid && failData.DeleteAt.After(time.Now())) || (failData.LockedUntilAt.Valid && failData.LockedUntilAt.Time.After(time.Now())) {
				continue
			}
			deleteKeys = append(deleteKeys, userID)
		}
		cc.redeemedFailUsersMut.RUnlock()

		for _, userID := range deleteKeys {
			cc.redeemedFailUsersMut.Lock()
			delete(cc.redeemedFailUsers, userID)
			cc.redeemedFailUsersMut.Unlock()
		}

		bm.End("gc_coupon_redeem_fail_users")
	}
}

//retrieve code and redeem

type CodeRedemptionRequest struct {
	Payload struct {
		Code string `json:"code"`
	} `json:"payload"`
}

type Reward struct {
	Crate       *server.MysteryCrate `json:"mystery_crate,omitempty"`
	Label       string               `json:"label"`
	ImageURL    null.String          `json:"image_url"`
	LockedUntil null.Time            `json:"locked_until"`
	Amount      decimal.NullDecimal  `json:"amount"`
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

	cc.redeemedFailUsersMut.RLock()
	redeemFailUser, hasFailedBefore := cc.redeemedFailUsers[user.ID]
	cc.redeemedFailUsersMut.RUnlock()

	if hasFailedBefore {
		if redeemFailUser.LockedUntilAt.Valid && redeemFailUser.LockedUntilAt.Time.After(time.Now()) {
			return terror.Error(fmt.Errorf("too many invalid code redemption requests"), "Too many failed code redemption requests, please try again.")
		}
		cc.redeemedFailUsersMut.Lock()
		delete(cc.redeemedFailUsers, user.ID)
		cc.redeemedFailUsersMut.Unlock()
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
					UPDATE %s SET %s = false,
								  %s = null
					WHERE %s = $1`,
					boiler.TableNames.Coupons,
					boiler.CouponColumns.Redeemed,
					boiler.CouponColumns.RedeemedAt,
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
					UPDATE %s SET %s = true,
								  %s = $2,
								  %s = NOW()
					WHERE %s = FALSE 
					AND %s = $1
					AND %s > NOW()
					RETURNING  %s, %s, %s, %s`,
				boiler.TableNames.Coupons,
				boiler.CouponColumns.Redeemed,
				boiler.CouponColumns.RedeemedByID,
				boiler.CouponColumns.RedeemedAt,
				boiler.CouponColumns.Redeemed,
				boiler.CouponColumns.Code,
				boiler.CouponColumns.ExpiryDate,
				boiler.CouponColumns.ID,
				boiler.CouponColumns.Code,
				boiler.CouponColumns.Redeemed,
				boiler.CouponColumns.ExpiryDate,
			),
			couponCode,
			user.ID,
		),
	).QueryRow(gamedb.StdConn).Scan(&coupon.ID, &coupon.Code, &coupon.Redeemed, &coupon.ExpiryDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if !hasFailedBefore {
				redeemFailUser = RedeemFailUser{
					Count:         0,
					LockedUntilAt: null.Time{},
				}
			}
			redeemFailUser.DeleteAt = time.Now().Add(time.Minute * 5) // Only keep track of failure for 5 minutes (unless locked)
			redeemFailUser.Count++
			if redeemFailUser.Count >= 3 {
				redeemFailUser.LockedUntilAt = null.TimeFrom(time.Now().Add(time.Minute * 30))
			}

			cc.redeemedFailUsersMut.Lock()
			cc.redeemedFailUsers[user.ID] = redeemFailUser
			cc.redeemedFailUsersMut.Unlock()

			if redeemFailUser.LockedUntilAt.Valid && redeemFailUser.LockedUntilAt.Time.After(time.Now()) {
				return terror.Error(fmt.Errorf("too many invalid code redemption requests"), "Too many failed code redemption requests, please try again.")
			}

			return terror.Error(fmt.Errorf("unable to find unclaimed coupon"))
		}

		gamelog.L.Error().Err(err).Interface("coupon code: ", couponCode).Msg("failed to find coupon code")
		return terror.Error(err, "Issue finding coupon code, try again or contact support.")
	}

	if hasFailedBefore {
		cc.redeemedFailUsersMut.Lock()
		delete(cc.redeemedFailUsers, user.ID)
		cc.redeemedFailUsersMut.Unlock()
	}

	err = coupon.L.LoadCouponItems(gamedb.StdConn, true, coupon,
		boiler.CouponItemWhere.Claimed.EQ(false),
	)
	if err != nil {
		rollbackRedeem()
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
			assignedMechCrate, xa, err := assignAndRegisterPurchasedCrate(user.ID, storeMechCrate, tx, cc.API)
			if err != nil {
				rollbackRedeem()
				return terror.Error(err, "Issue claiming mech crate, please try again or contact support.")
			}

			err = cc.API.Passport.AssetRegister(xa)
			if err != nil {
				rollbackRedeem()
				gamelog.L.Error().Err(err).Interface("mystery crate", "").Msg("failed to register to XSYN")
				return terror.Error(err, "Failed to get mystery crate, please try again or contact support.")
			}
			ci.ItemID = null.StringFrom(assignedMechCrate.ID)
			reward := &Reward{
				Label:       storeMechCrate.MysteryCrateType,
				ImageURL:    assignedMechCrate.ImageURL,
				LockedUntil: null.TimeFrom(assignedMechCrate.LockedUntil),
				Amount:      ci.Amount,
			}

			serverMechCrate := server.StoreFrontMysteryCrateFromBoiler(storeMechCrate)

			ws.PublishMessage(fmt.Sprintf("/faction/%s/crate/%s", factionID, assignedMechCrate.ID), server.HubKeyMysteryCrateSubscribe, serverMechCrate)

			rewards = append(rewards, reward)
		case boiler.CouponItemTypeWEAPON_CRATE:
			assignedWeaponCrate, xa, err := assignAndRegisterPurchasedCrate(user.ID, storeWeaponCrate, tx, cc.API)
			if err != nil {
				rollbackRedeem()
				return terror.Error(err, "Issue claiming weapon crate, please try again or contact support.")
			}
			err = cc.API.Passport.AssetRegister(xa)
			if err != nil {
				rollbackRedeem()
				gamelog.L.Error().Err(err).Interface("mystery crate", "").Msg("failed to register to XSYN")
				return terror.Error(err, "Failed to get mystery crate, please try again or contact support.")
			}

			ci.ItemID = null.StringFrom(assignedWeaponCrate.ID)
			reward := &Reward{
				Label:       storeWeaponCrate.MysteryCrateType,
				ImageURL:    assignedWeaponCrate.ImageURL,
				LockedUntil: null.TimeFrom(assignedWeaponCrate.LockedUntil),
				Amount:      ci.Amount,
			}

			serverWeaponCrate := server.StoreFrontMysteryCrateFromBoiler(storeWeaponCrate)

			ws.PublishMessage(fmt.Sprintf("/faction/%s/crate/%s", factionID, assignedWeaponCrate.ID), server.HubKeyMysteryCrateSubscribe, serverWeaponCrate)

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
	})
	if err != nil {
		return "", terror.Error(err, "Could not transfer user SUPS from Supremacy Game User, try again or contact support")
	}
	return txID, nil
}
