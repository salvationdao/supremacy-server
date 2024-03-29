package api

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
	"net/http"
	"server"
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
	bonusSupPerWinner := db.GetDecimalWithDefault(db.KeyBattleSupsRewardBonus, decimal.New(45, 18))

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

		deployedMechs := []*boiler.BattleLobbiesMech{}

		for _, bl := range bls {
			impactedLobbyIDs = append(impactedLobbyIDs, bl.ID)

			needRM := bl.EachFactionMechAmount
			needBC := bl.EachFactionMechAmount
			needZai := bl.EachFactionMechAmount

			if bl.R != nil {
				for _, blm := range bl.R.BattleLobbiesMechs {
					if slices.IndexFunc(deployedMechs, func(bm *boiler.BattleLobbiesMech) bool { return bm.MechID == blm.ID }) == -1 {
						deployedMechs = append(deployedMechs, blm)
					}
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
					QueuedByID:    server.BostonCyberneticsPlayerID,
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
					QueuedByID:    server.RedMountainPlayerID,
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
					QueuedByID:    server.ZaibatsuPlayerID,
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
					Name:                  helpers.GenerateAdjectiveName(),
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
			api.ArenaManager.BattleLobbyDebounceBroadcastChan <- impactedLobbyIDs
		}

		for _, blm := range deployedMechs {
			ws.PublishMessage(fmt.Sprintf("/faction/%s/queue/%s", blm.FactionID, blm.MechID), server.HubKeyPlayerAssetMechQueueSubscribe, &server.MechArenaInfo{
				Status:              server.MechArenaStatusQueue,
				CanDeploy:           false,
				BattleLobbyIsLocked: true,
			})
		}
		return nil
	})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	go api.ArenaManager.KickIdleArenas()

	return http.StatusOK, nil
}
