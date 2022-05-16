package battle

import (
	"fmt"
	"github.com/ninja-syndicate/ws"
	"net/http"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"sort"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/ninja-software/terror/v2"

	"github.com/ninja-software/tickle"
)

type PlayerRank string

const (
	PlayerRankNewRecruit PlayerRank = "NEW_RECRUIT"
	PlayerRankPrivate    PlayerRank = "PRIVATE"
	PlayerRankCorporal   PlayerRank = "CORPORAL"
	PlayerRankGeneral    PlayerRank = "GENERAL"
)
const HubKeyPlayerRankGet = "PLAYER:RANK:GET"

func (arena *Arena) PlayerRankUpdater() {
	// create a tickle to constantly update player ability kill and ranks
	updateTickle := tickle.New("Player rank and kill update", 30*60, func() (int, error) {
		// calculate player rank of each syndicate
		err := calcSyndicatePlayerRank(server.RedMountainFactionID)
		if err != nil {
			gamelog.L.Error().Str("faction id", server.RedMountainFactionID.String()).Err(err).Msg("Failed to re-calculate player rank in syndicate")
		}
		err = calcSyndicatePlayerRank(server.BostonCyberneticsFactionID)
		if err != nil {
			gamelog.L.Error().Str("faction id", server.BostonCyberneticsFactionID.String()).Err(err).Msg("Failed to re-calculate player rank in syndicate")
		}
		err = calcSyndicatePlayerRank(server.ZaibatsuFactionID)
		if err != nil {
			gamelog.L.Error().Str("faction id", server.ZaibatsuFactionID.String()).Err(err).Msg("Failed to re-calculate player rank in syndicate")
		}

		bus := arena.currentBattleUsersCopy()
		if bus != nil && len(bus) > 0 {
			// prepare user id list
			userIDs := []string{}
			for _, bu := range bus {
				userIDs = append(userIDs, bu.ID.String())
			}

			// query players' id and rank
			players, err := boiler.Players(
				qm.Select(
					boiler.PlayerColumns.ID,
					boiler.PlayerColumns.Rank,
				),
				boiler.PlayerWhere.ID.IN(userIDs),
			).All(gamedb.StdConn)
			if err != nil {
				return http.StatusInternalServerError, terror.Error(err, "Failed to get player from db")
			}

			// start broadcast user rank to current online users
			for _, bu := range bus {
				// find player from the list
				for _, player := range players {
					if player.ID == bu.ID.String() {
						// broadcast player rank to every player
						go func(bu *BattleUser, player *boiler.Player) {
							// broadcast stat
							ws.PublishMessage(fmt.Sprintf("/user/%s", player.ID), HubKeyPlayerRankGet, player.Rank)

							// broadcast user stat (player_last_seven_days_kills)
							us, err := db.UserStatsGet(player.ID)
							if err != nil {
								gamelog.L.Error().Err(err).Msg("failed to get user stat")
							}

							if us != nil {
								ws.PublishMessage(fmt.Sprintf("/user/%s", us.ID), HubKeyUserStatSubscribe, us)
							}
						}(bu, player)

						break
					}
				}
			}
		}

		return http.StatusOK, nil
	})

	updateTickle.Log = gamelog.L

	// start tickle
	updateTickle.Start()
}

func calcSyndicatePlayerRank(factionID server.FactionID) error {
	playerAbilityKills, err := db.GetPositivePlayerAbilityKillByFactionID(factionID)
	if err != nil {
		gamelog.L.Error().Str("faction id", factionID.String()).Err(err).Msg("Failed to get player ability kill from db")
		return terror.Error(err, "Failed to get player ability kill from db")
	}

	if len(playerAbilityKills) == 0 {
		return nil
	}

	// get top 20% players
	topTwentyPercentCount := len(playerAbilityKills) * 20 / 100
	if topTwentyPercentCount == 0 {
		topTwentyPercentCount = 1
	}

	// sort the slice
	sort.Slice(playerAbilityKills, func(i, j int) bool { return playerAbilityKills[i].KillCount > playerAbilityKills[j].KillCount })

	generalPlayerIDs := []string{}
	for i := 0; i < topTwentyPercentCount; i++ {
		generalPlayerIDs = append(generalPlayerIDs, playerAbilityKills[i].ID)
	}

	// update general players
	_, err = boiler.Players(
		boiler.PlayerWhere.ID.IN(generalPlayerIDs),
		boiler.PlayerWhere.FactionID.EQ(null.StringFrom(factionID.String())),
		boiler.PlayerWhere.CreatedAt.LT(time.Now().AddDate(0, 0, -1)), // should be created more than a day
	).UpdateAll(gamedb.StdConn, boiler.M{"rank": PlayerRankGeneral})
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to update general rank player")
		return terror.Error(err, "Failed to update general rank player")
	}

	// update corporal players
	_, err = boiler.Players(
		boiler.PlayerWhere.ID.NIN(generalPlayerIDs),
		boiler.PlayerWhere.FactionID.EQ(null.StringFrom(factionID.String())),
		boiler.PlayerWhere.CreatedAt.LT(time.Now().AddDate(0, 0, -1)),
		boiler.PlayerWhere.SentMessageCount.GT(0),
		qm.Where(
			fmt.Sprintf(
				"EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s > 0)",
				boiler.TableNames.UserStats,
				qm.Rels(boiler.TableNames.UserStats, boiler.UserStatColumns.ID),
				qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.ID),
				qm.Rels(boiler.TableNames.UserStats, boiler.UserStatColumns.AbilityKillCount),
			),
		),
	).UpdateAll(gamedb.StdConn, boiler.M{"rank": PlayerRankCorporal})
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to update corporal rank player")
		return terror.Error(err, "Failed to update corporal rank player")
	}

	// update private players
	_, err = boiler.Players(
		boiler.PlayerWhere.ID.NIN(generalPlayerIDs),
		boiler.PlayerWhere.FactionID.EQ(null.StringFrom(factionID.String())),
		boiler.PlayerWhere.CreatedAt.LT(time.Now().AddDate(0, 0, -1)),
		boiler.PlayerWhere.SentMessageCount.GT(0),
		qm.Where(
			fmt.Sprintf(
				"EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s <= 0)",
				boiler.TableNames.UserStats,
				qm.Rels(boiler.TableNames.UserStats, boiler.UserStatColumns.ID),
				qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.ID),
				qm.Rels(boiler.TableNames.UserStats, boiler.UserStatColumns.AbilityKillCount),
			),
		),
	).UpdateAll(gamedb.StdConn, boiler.M{"rank": PlayerRankPrivate})
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to update private rank player")
		return terror.Error(err, "Failed to update private rank player")
	}

	return nil
}
