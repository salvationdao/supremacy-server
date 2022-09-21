package api

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"net/http"
	"server"
	"server/battle"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/helpers"
	"time"
)

func BattleRouter(api *API) chi.Router {
	r := chi.NewRouter()
	r.Get("/mech/{id}/destroyed_detail", WithError(api.MechDestroyedDetail))
	r.Get("/challenge_fund_amount", WithError(api.ChallengeFundAmount))

	if server.IsDevelopmentEnv() {
		r.Get("/fill_up_incomplete_lobbies", WithError(api.FillUpIncompleteLobbies))
	}

	return r
}

// MechDestroyedDetail return mech destroyed record if exists
func (api *API) MechDestroyedDetail(w http.ResponseWriter, r *http.Request) (int, error) {
	mechID := chi.URLParam(r, "id")

	if destroyedRecord := api.ArenaManager.WarMachineDestroyedDetail(mechID); destroyedRecord != nil {
		return helpers.EncodeJSON(w, destroyedRecord)
	}

	return http.StatusOK, nil
}

func (api *API) ChallengeFundAmount(w http.ResponseWriter, r *http.Request) (int, error) {
	challengeFundBalance := api.Passport.UserBalanceGet(uuid.FromStringOrNil(server.SupremacyChallengeFundUserID))
	bonusSupPerWinner := db.GetDecimalWithDefault(db.KeyBattleSupsRewardBonus, decimal.New(330, 18))

	return helpers.EncodeJSON(w, struct {
		ChallengeFundBalance decimal.Decimal `json:"challenge_fund_balance"`
		BonusSupsPerWinner   decimal.Decimal `json:"bonus_sups_per_winner"`
	}{
		challengeFundBalance,
		bonusSupPerWinner,
	})
}

func (api *API) FillUpIncompleteLobbies(w http.ResponseWriter, r *http.Request) (int, error) {
	// load default mechs
	cis, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
		qm.Where(
			fmt.Sprintf("EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s = ?)",
				boiler.TableNames.Players,
				boiler.PlayerTableColumns.ID,
				boiler.CollectionItemTableColumns.OwnerID,
				boiler.PlayerTableColumns.IsAi,
			),
			true,
		),
		qm.Load(boiler.CollectionItemRels.Owner),
	).All(gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	var bcMechs []*boiler.CollectionItem
	var rmMechs []*boiler.CollectionItem
	var zaiMechs []*boiler.CollectionItem

	for _, ci := range cis {
		if ci.R == nil || ci.R.Owner == nil {
			continue
		}

		switch ci.R.Owner.FactionID.String {
		case server.BostonCyberneticsFactionID:
			bcMechs = append(bcMechs, ci)
		case server.RedMountainFactionID:
			rmMechs = append(rmMechs, ci)
		case server.ZaibatsuFactionID:
			zaiMechs = append(zaiMechs, ci)
		}

	}

	err = api.ArenaManager.SendBattleQueueFunc(func() error {
		// load all the incomplete lobbies
		var bls boiler.BattleLobbySlice
		bls, err = boiler.BattleLobbies(
			boiler.BattleLobbyWhere.ReadyAt.IsNull(),
			qm.Load(
				boiler.BattleLobbyRels.BattleLobbiesMechs,
				boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
				boiler.BattleLobbiesMechWhere.DeletedAt.IsNull(),
			),
		).All(gamedb.StdConn)
		if err != nil {
			return err
		}

		var impactedLobbyIDs []string

		for _, bl := range bls {
			impactedLobbyIDs = append(impactedLobbyIDs, bl.ID)

			needRM := bl.EachFactionMechAmount
			needBC := bl.EachFactionMechAmount
			needZai := bl.EachFactionMechAmount

			if bl.R != nil {
				for _, blm := range bl.R.BattleLobbiesMechs {
					switch blm.FactionID {
					case server.BostonCyberneticsFactionID:
						needBC -= 1
					case server.RedMountainFactionID:
						needRM -= 1
					case server.ZaibatsuFactionID:
						needZai -= 1
					}
				}
			}

			// queue bc mechs
			for i := 0; i < needBC; i++ {
				blm := &boiler.BattleLobbiesMech{
					BattleLobbyID: bl.ID,
					MechID:        bcMechs[i].ItemID,
					FactionID:     server.BostonCyberneticsFactionID,
					OwnerID:       server.BostonCyberneticsPlayerID,
				}

				err = blm.Insert(gamedb.StdConn, boil.Infer())
				if err != nil {
					return err
				}
			}

			// queue rm mechs
			for i := 0; i < needRM; i++ {
				blm := &boiler.BattleLobbiesMech{
					BattleLobbyID: bl.ID,
					MechID:        rmMechs[i].ItemID,
					FactionID:     server.RedMountainFactionID,
					OwnerID:       server.RedMountainPlayerID,
				}

				err = blm.Insert(gamedb.StdConn, boil.Infer())
				if err != nil {
					return err
				}
			}

			// queue zai mechs
			for i := 0; i < needZai; i++ {
				blm := &boiler.BattleLobbiesMech{
					BattleLobbyID: bl.ID,
					MechID:        zaiMechs[i].ItemID,
					FactionID:     server.ZaibatsuFactionID,
					OwnerID:       server.ZaibatsuPlayerID,
				}

				err = blm.Insert(gamedb.StdConn, boil.Infer())
				if err != nil {
					return err
				}
			}

			now := time.Now()

			bl.ReadyAt = null.TimeFrom(now)
			_, err = bl.Update(gamedb.StdConn, boil.Whitelist(boiler.BattleLobbyColumns.ReadyAt))
			if err != nil {
				return err
			}

			_, err = boiler.BattleLobbiesMechs(
				boiler.BattleLobbiesMechWhere.BattleLobbyID.EQ(bl.ID),
			).UpdateAll(gamedb.StdConn, boiler.M{
				boiler.BattleLobbiesMechColumns.LockedAt: null.TimeFrom(now),
			})
			if err != nil {
				return err
			}

			if bl.GeneratedBySystem {
				newLobby := &boiler.BattleLobby{
					HostByID:              server.SupremacyBattleUserID,
					EntryFee:              decimal.Zero, // free to join
					FirstFactionCut:       decimal.NewFromFloat(0.75),
					SecondFactionCut:      decimal.NewFromFloat(0.25),
					ThirdFactionCut:       decimal.Zero,
					EachFactionMechAmount: 3,
					GeneratedBySystem:     true,
				}

				err = newLobby.Insert(gamedb.StdConn, boil.Infer())
				if err != nil {
					return err
				}

				impactedLobbyIDs = append(impactedLobbyIDs, newLobby.ID)
			}
		}

		if impactedLobbyIDs != nil {
			go battle.BroadcastBattleLobbyUpdate(impactedLobbyIDs...)
		}
		return nil
	})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	api.ArenaManager.KickIdleArenas()

	return http.StatusOK, nil
}
