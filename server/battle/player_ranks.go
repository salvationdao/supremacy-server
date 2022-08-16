package battle

import (
	"fmt"
	"net/http"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"sort"
	"sync"
	"time"

	"github.com/ninja-syndicate/ws"

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

func (am *ArenaManager) PlayerRankUpdater() {
	// create a tickle to constantly update player ability kill and ranks
	updateTickle := tickle.New("Player rank and kill update", 30*60, func() (int, error) {
		// calculate player rank of each syndicate
		err := calcSyndicatePlayerRank(server.RedMountainFactionID)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Str("faction id", server.RedMountainFactionID).Err(err).Msg("Failed to re-calculate player rank in syndicate")
		}
		err = calcSyndicatePlayerRank(server.BostonCyberneticsFactionID)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Str("faction id", server.BostonCyberneticsFactionID).Err(err).Msg("Failed to re-calculate player rank in syndicate")
		}
		err = calcSyndicatePlayerRank(server.ZaibatsuFactionID)
		if err != nil {
			gamelog.L.Error().Str("log_name", "battle arena").Str("faction id", server.ZaibatsuFactionID).Err(err).Msg("Failed to re-calculate player rank in syndicate")
		}

		connectedUserIDs := ws.TrackedIdents()
		if len(connectedUserIDs) > 0 {
			// query players' id and rank
			players, err := boiler.Players(
				qm.Select(
					boiler.PlayerColumns.ID,
					boiler.PlayerColumns.Rank,
				),
				boiler.PlayerWhere.ID.IN(connectedUserIDs),
			).All(gamedb.StdConn)
			if err != nil {
				return http.StatusInternalServerError, terror.Error(err, "Failed to get player from db")
			}

			// find player from the list
			wg := sync.WaitGroup{}
			for _, player := range players {
				wg.Add(1)
				// broadcast player rank to every player
				go func(player *boiler.Player) {
					defer wg.Done()

					// broadcast stat
					ws.PublishMessage(fmt.Sprintf("/secure/user/%s", player.ID), server.HubKeyPlayerRankGet, player.Rank)

					// broadcast user stat (player_last_seven_days_kills)
					us, err := db.UserStatsGet(player.ID)
					if err != nil {
						gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("failed to get user stat")
					}

					if us != nil {
						ws.PublishMessage(fmt.Sprintf("/secure/user/%s/stat", us.ID), server.HubKeyUserStatSubscribe, us)
					}
				}(player)
			}
			wg.Wait()
		}

		return http.StatusOK, nil
	})

	updateTickle.Log = gamelog.L

	// start tickle
	updateTickle.Start()
}

func calcSyndicatePlayerRank(factionID string) error {
	playerAbilityKills, err := db.GetPositivePlayerAbilityKillByFactionID(factionID)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("faction id", factionID).Err(err).Msg("Failed to get player ability kill from db")
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
		boiler.PlayerWhere.FactionID.EQ(null.StringFrom(factionID)),
		boiler.PlayerWhere.CreatedAt.LT(time.Now().AddDate(0, 0, -1)), // should be created more than a day
	).UpdateAll(gamedb.StdConn, boiler.M{"rank": PlayerRankGeneral})
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to update general rank player")
		return terror.Error(err, "Failed to update general rank player")
	}

	// update corporal players
	_, err = boiler.Players(
		boiler.PlayerWhere.ID.NIN(generalPlayerIDs),
		boiler.PlayerWhere.FactionID.EQ(null.StringFrom(factionID)),
		boiler.PlayerWhere.CreatedAt.LT(time.Now().AddDate(0, 0, -1)),
		boiler.PlayerWhere.SentMessageCount.GT(0),
		qm.Where(
			fmt.Sprintf(
				"EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s > 0)",
				boiler.TableNames.PlayerStats,
				qm.Rels(boiler.TableNames.PlayerStats, boiler.PlayerStatColumns.ID),
				qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.ID),
				qm.Rels(boiler.TableNames.PlayerStats, boiler.PlayerStatColumns.AbilityKillCount),
			),
		),
	).UpdateAll(gamedb.StdConn, boiler.M{"rank": PlayerRankCorporal})
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to update corporal rank player")
		return terror.Error(err, "Failed to update corporal rank player")
	}

	// update private players
	_, err = boiler.Players(
		boiler.PlayerWhere.ID.NIN(generalPlayerIDs),
		boiler.PlayerWhere.FactionID.EQ(null.StringFrom(factionID)),
		boiler.PlayerWhere.CreatedAt.LT(time.Now().AddDate(0, 0, -1)),
		boiler.PlayerWhere.SentMessageCount.GT(0),
		qm.Where(
			fmt.Sprintf(
				"EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s <= 0)",
				boiler.TableNames.PlayerStats,
				qm.Rels(boiler.TableNames.PlayerStats, boiler.PlayerStatColumns.ID),
				qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.ID),
				qm.Rels(boiler.TableNames.PlayerStats, boiler.PlayerStatColumns.AbilityKillCount),
			),
		),
	).UpdateAll(gamedb.StdConn, boiler.M{"rank": PlayerRankPrivate})
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Err(err).Msg("Failed to update private rank player")
		return terror.Error(err, "Failed to update private rank player")
	}

	return nil
}
