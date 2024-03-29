package battle

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"server/system_messages"
	"server/xsyn_rpcclient"
	"time"

	"github.com/gofrs/uuid"

	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/sasha-s/go-deadlock"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
)

func (am *ArenaManager) SendBattleQueueFunc(fn func() error) error {
	am.BattleQueueFuncMx.Lock()
	defer am.BattleQueueFuncMx.Unlock()
	return fn()
}

// DebounceSendBattleLobbiesUpdate debounce the lobby update sending
func (am *ArenaManager) DebounceSendBattleLobbiesUpdate() {
	duration := 250 * time.Millisecond

	timer := time.NewTimer(duration)
	impactedBattleLobbyIDs := []string{}

	for {
		select {
		case battleLobbyIDs := <-am.BattleLobbyDebounceBroadcastChan:
			for _, id := range battleLobbyIDs {
				// if id is not in the impacted lobby list
				if slices.Index(impactedBattleLobbyIDs, id) == -1 {
					impactedBattleLobbyIDs = append(impactedBattleLobbyIDs, id)
				}
			}

			// reset the timer duration
			timer.Reset(duration)

		case <-timer.C:
			// broadcast battle lobby update
			am.broadcastBattleLobbyUpdate(impactedBattleLobbyIDs...)

			// clean up battle lobby id list
			impactedBattleLobbyIDs = []string{}
		}
	}
}

// broadcastBattleLobbyUpdate broadcast the updated lobbies to each faction
func (am *ArenaManager) broadcastBattleLobbyUpdate(battleLobbyIDs ...string) {
	if battleLobbyIDs == nil || len(battleLobbyIDs) == 0 {
		return
	}

	bls, err := db.GetBattleLobbyViaIDs(battleLobbyIDs)
	if err != nil {
		gamelog.L.Error().Err(err).Strs("battle lobby id list", battleLobbyIDs).Msg("Failed to query battle lobby")
		return
	}

	// generate deleted lobbies
	deletedLobbies := []*server.BattleLobby{}
	for _, battleLobbyID := range battleLobbyIDs {
		if slices.IndexFunc(bls, func(bl *boiler.BattleLobby) bool { return bl.ID == battleLobbyID }) == -1 {
			deletedLobbies = append(deletedLobbies, &server.BattleLobby{
				BattleLobby: &boiler.BattleLobby{
					ID:        battleLobbyID,
					DeletedAt: null.TimeFrom(time.Now()),
				},
			})
		}
	}

	battleLobbies, err := server.BattleLobbiesFromBoiler(bls)
	if err != nil {
		return
	}

	type playerInvolveLobby struct {
		playerID  string
		factionID string
		bls       []*server.BattleLobby
	}

	var publicLobbies []*server.BattleLobby
	var privateLobbies []*server.BattleLobby
	var playersInvolvedLobbies []*playerInvolveLobby

	// broadcast to individual
	for _, bl := range battleLobbies {
		// build public/private lobby list
		if !bl.AccessCode.Valid {
			// append public lobbies
			publicLobbies = append(publicLobbies, bl)

		} else {
			// append private lobbies
			privateLobbies = append(privateLobbies, bl)

		}

		// broadcast individual lobby
		go ws.PublishMessage(fmt.Sprintf("/faction/%s/battle_lobby/%s", server.RedMountainFactionID, bl.ID), server.HubKeyBattleLobbyUpdate, server.BattleLobbyInfoFilter(bl, server.RedMountainFactionID, true))
		go ws.PublishMessage(fmt.Sprintf("/faction/%s/battle_lobby/%s", server.BostonCyberneticsFactionID, bl.ID), server.HubKeyBattleLobbyUpdate, server.BattleLobbyInfoFilter(bl, server.BostonCyberneticsFactionID, true))
		go ws.PublishMessage(fmt.Sprintf("/faction/%s/battle_lobby/%s", server.ZaibatsuFactionID, bl.ID), server.HubKeyBattleLobbyUpdate, server.BattleLobbyInfoFilter(bl, server.ZaibatsuFactionID, true))

		// build player involved lobby map
		if bl.HostBy != nil && bl.HostBy.FactionID.Valid {
			host := bl.HostBy
			// check host player
			index := slices.IndexFunc(playersInvolvedLobbies, func(pil *playerInvolveLobby) bool { return pil.playerID == host.ID })
			if index == -1 {
				playersInvolvedLobbies = append(playersInvolvedLobbies, &playerInvolveLobby{
					playerID:  host.ID,
					factionID: host.FactionID.String,
					bls:       []*server.BattleLobby{},
				})

				index = len(playersInvolvedLobbies) - 1
			}

			// skip, if the player already have the lobby on their list
			if slices.IndexFunc(playersInvolvedLobbies[index].bls, func(battleLobby *server.BattleLobby) bool { return battleLobby.ID == bl.ID }) != -1 {
				continue
			}

			playersInvolvedLobbies[index].bls = append(playersInvolvedLobbies[index].bls, bl)
		}

		// check joined players
		for _, blm := range bl.BattleLobbiesMechs {
			if blm.QueuedBy == nil || !blm.QueuedBy.FactionID.Valid {
				continue
			}

			queuedByID := blm.QueuedBy.ID
			factionID := blm.QueuedBy.FactionID.String

			// check host player
			index := slices.IndexFunc(playersInvolvedLobbies, func(pil *playerInvolveLobby) bool { return pil.playerID == queuedByID })
			if index == -1 {
				playersInvolvedLobbies = append(playersInvolvedLobbies, &playerInvolveLobby{
					playerID:  queuedByID,
					factionID: factionID,
					bls:       []*server.BattleLobby{},
				})

				index = len(playersInvolvedLobbies) - 1
			}

			// skip, if the player already have the lobby on their list
			if slices.IndexFunc(playersInvolvedLobbies[index].bls, func(battleLobby *server.BattleLobby) bool { return battleLobby.ID == bl.ID }) != -1 {
				continue
			}

			// otherwise, append to lobby to the player's list
			playersInvolvedLobbies[index].bls = append(playersInvolvedLobbies[index].bls, bl)
		}
	}

	// broadcast private lobbies individually
	for _, battleLobby := range privateLobbies {
		go ws.PublishMessage(fmt.Sprintf("/faction/%s/private_battle_lobby/%s", server.RedMountainFactionID, battleLobby.AccessCode.String), server.HubKeyPrivateBattleLobbyUpdate, server.BattleLobbyInfoFilter(battleLobby, server.RedMountainFactionID, true))
		go ws.PublishMessage(fmt.Sprintf("/faction/%s/private_battle_lobby/%s", server.BostonCyberneticsFactionID, battleLobby.AccessCode.String), server.HubKeyPrivateBattleLobbyUpdate, server.BattleLobbyInfoFilter(battleLobby, server.BostonCyberneticsFactionID, true))
		go ws.PublishMessage(fmt.Sprintf("/faction/%s/private_battle_lobby/%s", server.ZaibatsuFactionID, battleLobby.AccessCode.String), server.HubKeyPrivateBattleLobbyUpdate, server.BattleLobbyInfoFilter(battleLobby, server.ZaibatsuFactionID, true))
	}

	// broadcast public lobbies
	if len(publicLobbies) > 0 || len(deletedLobbies) > 0 {
		go ws.PublishMessage(fmt.Sprintf("/faction/%s/battle_lobbies", server.RedMountainFactionID), server.HubKeyBattleLobbyListUpdate, append(server.BattleLobbiesFactionFilter(publicLobbies, server.RedMountainFactionID, false), deletedLobbies...))
		go ws.PublishMessage(fmt.Sprintf("/faction/%s/battle_lobbies", server.BostonCyberneticsFactionID), server.HubKeyBattleLobbyListUpdate, append(server.BattleLobbiesFactionFilter(publicLobbies, server.BostonCyberneticsFactionID, false), deletedLobbies...))
		go ws.PublishMessage(fmt.Sprintf("/faction/%s/battle_lobbies", server.ZaibatsuFactionID), server.HubKeyBattleLobbyListUpdate, append(server.BattleLobbiesFactionFilter(publicLobbies, server.ZaibatsuFactionID, false), deletedLobbies...))
	}

	// broadcast the lobbies which players are involved in
	for _, pil := range playersInvolvedLobbies {
		ws.PublishMessage(fmt.Sprintf("/secure/user/%s/involved_battle_lobbies", pil.playerID), server.HubKeyInvolvedBattleLobbyListUpdate, server.BattleLobbiesFactionFilter(pil.bls, pil.factionID, true))
	}

	privateLobbies = nil
	publicLobbies = nil
	playersInvolvedLobbies = nil
	deletedLobbies = nil
}

// SetDefaultPublicBattleLobbies ensure there are enough battle lobbies when server start
func (am *ArenaManager) SetDefaultPublicBattleLobbies() error {
	// check once when server start
	err := am.DefaultPublicLobbiesCheck()
	if err != nil {
		return err
	}

	go func() {
		publicLobbyTicker := time.NewTicker(1 * time.Minute)
		expireLobbyTicker := time.NewTicker(5 * time.Second)

		for {
			select {
			case <-publicLobbyTicker.C:
				err = am.DefaultPublicLobbiesCheck()
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to check default public lobbies.")
				}

			case <-expireLobbyTicker.C:
				err = am.ExpiredExhibitionLobbyCleanUp()
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to clean up expired exhibition lobbies")
				}
			}

		}
	}()

	return nil
}

type ExpiredLobbyMessage struct {
	PlayerID   string               `json:"player_id"`
	MechBriefs []*SystemMessageMech `json:"mech_briefs"`
}

func (am *ArenaManager) ExpiredExhibitionLobbyCleanUp() error {
	// lock queue func
	am.BattleQueueFuncMx.Lock()
	defer am.BattleQueueFuncMx.Unlock()

	// load the non-private lobbies which are expired and not ready
	bls, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.ExpiresAt.IsNotNull(),
		boiler.BattleLobbyWhere.ExpiresAt.LTE(null.TimeFrom(time.Now())),
		boiler.BattleLobbyWhere.ReadyAt.IsNull(),
		boiler.BattleLobbyWhere.GeneratedBySystem.EQ(false),
		boiler.BattleLobbyWhere.AccessCode.IsNull(),
		qm.Load(
			boiler.BattleLobbyRels.BattleLobbiesMechs,
			boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
			boiler.BattleLobbiesMechWhere.DeletedAt.IsNull(),
		),
		qm.Load(
			boiler.BattleLobbyRels.BattleLobbyExtraSupsRewards,
			boiler.BattleLobbyExtraSupsRewardWhere.RefundedTXID.IsNull(),
			boiler.BattleLobbyExtraSupsRewardWhere.DeletedAt.IsNull(),
		),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to load expired battle lobbies.")
	}

	// skip, if there is no expired battle lobbies
	if bls == nil || len(bls) == 0 {
		return nil
	}

	wg := deadlock.WaitGroup{}
	for _, bl := range bls {
		wg.Add(1)

		go func(battleLobby *boiler.BattleLobby) {
			l := gamelog.L.With().Str("func", "ExpiredExhibitionLobbyCleanUp").Interface("battle lobby", battleLobby).Logger()

			tx, err := gamedb.StdConn.Begin()
			if err != nil {
				l.Error().Err(err).Msg("Failed to start db transaction.")
				return
			}

			defer func() {
				tx.Rollback()
				wg.Done()
			}()

			battleLobby.EndedAt = null.TimeFrom(time.Now())
			_, err = battleLobby.Update(tx, boil.Whitelist(boiler.BattleLobbyColumns.EndedAt))
			if err != nil {
				l.Error().Err(err).Msg("Failed to soft delete battle lobby")
				return
			}

			var refundFns []func()
			refund := func() {
				for _, fn := range refundFns {
					fn()
				}
			}
			involvedPlayerMechs := []*ExpiredLobbyMessage{
				{
					PlayerID:   battleLobby.HostByID,
					MechBriefs: []*SystemMessageMech{},
				},
			}

			lobbyMechIDs := []string{}
			if battleLobby.R != nil {
				// refund battle lobby mechs' entry fee
				for _, battleLobbyMech := range battleLobby.R.BattleLobbiesMechs {
					// record involved player id
					index := slices.IndexFunc(involvedPlayerMechs, func(ip *ExpiredLobbyMessage) bool { return ip.PlayerID == battleLobbyMech.QueuedByID })
					if index == -1 {
						involvedPlayerMechs = append(involvedPlayerMechs, &ExpiredLobbyMessage{
							PlayerID:   battleLobbyMech.QueuedByID,
							MechBriefs: []*SystemMessageMech{},
						})
						index = len(involvedPlayerMechs) - 1
					}

					involvedPlayerMechs[index].MechBriefs = append(involvedPlayerMechs[index].MechBriefs, &SystemMessageMech{
						MechID:    battleLobbyMech.MechID,
						FactionID: battleLobbyMech.FactionID,
					})

					lobbyMechIDs = append(lobbyMechIDs, battleLobbyMech.MechID)
					battleLobbyMech.DeletedAt = null.TimeFrom(time.Now())
					updatedColumns := []string{
						boiler.BattleLobbiesMechColumns.DeletedAt,
					}

					if battleLobby.EntryFee.GreaterThan(decimal.Zero) && battleLobbyMech.PaidTXID.Valid {
						refundTxID, err := am.RPCClient.RefundSupsMessage(battleLobbyMech.PaidTXID.String)
						if err != nil {
							refund()
							l.Error().Err(err).Msg("Failed to refund entry fee.")
							return
						}

						battleLobbyMech.RefundTXID = null.StringFrom(refundTxID)
						updatedColumns = append(updatedColumns, boiler.BattleLobbiesMechColumns.RefundTXID)

						refundFns = append(refundFns, func() {
							_, err = am.RPCClient.RefundSupsMessage(refundTxID)
							if err != nil {
								l.Error().Err(err).Msg("Failed to refund refund entry fee")
							}
						})
					}

					_, err = battleLobbyMech.Update(tx, boil.Whitelist(updatedColumns...))
					if err != nil {
						refund()
						l.Error().Err(err).Msg("Failed to update battle lobby mech")
						return
					}
				}

				// refund any extra sups reward
				for _, esr := range battleLobby.R.BattleLobbyExtraSupsRewards {
					refundTxID, err := am.RPCClient.RefundSupsMessage(esr.PaidTXID)
					if err != nil {
						refund()
						l.Error().Err(err).Msg("Failed to refund entry fee.")
						return
					}

					refundFns = append(refundFns, func() {
						_, err = am.RPCClient.RefundSupsMessage(refundTxID)
						if err != nil {
							l.Error().Err(err).Msg("Failed to refund refund entry fee")
						}
					})

					esr.RefundedTXID = null.StringFrom(refundTxID)
					esr.DeletedAt = null.TimeFrom(time.Now())
					_, err = esr.Update(tx, boil.Whitelist(boiler.BattleLobbyExtraSupsRewardColumns.RefundedTXID, boiler.BattleLobbyExtraSupsRewardColumns.DeletedAt))
					if err != nil {
						refund()
						l.Error().Err(err).Interface("extra sups reward", esr).Msg("Failed to update extra sups reward")
						return
					}
				}
			}

			err = tx.Commit()
			if err != nil {
				refund()
				l.Error().Err(err).Msg("Failed to commit db transaction.")
				return
			}

			am.FactionStakedMechDashboardKeyChan <- []string{FactionStakedMechDashboardKeyQueue}

			// broadcast battle lobby
			am.BattleLobbyDebounceBroadcastChan <- []string{battleLobby.ID}

			// load mech data
			for _, playerMechs := range involvedPlayerMechs {

				go func(pms *ExpiredLobbyMessage) {
					ws.PublishMessage(fmt.Sprintf("/secure/user/%s/involved_battle_lobbies", pms.PlayerID), server.HubKeyInvolvedBattleLobbyListUpdate, []*boiler.BattleLobby{
						{
							ID:        battleLobby.ID,
							DeletedAt: null.TimeFrom(time.Now()),
						},
					})

					// send system message
					if len(pms.MechBriefs) > 0 {
						// collect mech brief data
						mechIDWhereIn := fmt.Sprintf("%s IN (", boiler.MechTableColumns.ID)
						for i, mb := range pms.MechBriefs {
							mechIDWhereIn += "'" + mb.MechID + "'"
							if i < len(pms.MechBriefs)-1 {
								mechIDWhereIn += ","
								continue
							}

							mechIDWhereIn += ")"
						}

						queries := []qm.QueryMod{
							qm.Select(
								boiler.MechTableColumns.ID,
								boiler.MechTableColumns.Name,
								boiler.BlueprintMechTableColumns.Label,
								boiler.BlueprintMechSkinTableColumns.Tier,
								boiler.MechModelSkinCompatibilityTableColumns.AvatarURL,
								boiler.BlueprintMechTableColumns.RepairBlocks,
								fmt.Sprintf(
									`COALESCE(
										(SELECT %s - %s FROM %s WHERE %s = %s AND %s ISNULL AND %s ISNULL),
										0
									) AS damaged_blocks`,
									boiler.RepairCaseTableColumns.BlocksRequiredRepair,
									boiler.RepairCaseTableColumns.BlocksRepaired,
									boiler.TableNames.RepairCases,
									boiler.RepairCaseTableColumns.MechID,
									boiler.MechTableColumns.ID,
									boiler.RepairCaseTableColumns.CompletedAt,
									boiler.RepairCaseTableColumns.DeletedAt,
								),
							),

							qm.From(fmt.Sprintf(
								`(SELECT * FROM %[1]s WHERE %[2]s) %[1]s`,
								boiler.TableNames.Mechs,
								mechIDWhereIn,
							)),

							// inner join blueprint mech
							qm.InnerJoin(fmt.Sprintf(
								"%s ON %s = %s",
								boiler.TableNames.BlueprintMechs,
								boiler.MechTableColumns.BlueprintID,
								boiler.BlueprintMechTableColumns.ID,
							)),

							// inner join mech skin
							qm.InnerJoin(fmt.Sprintf(
								"%s ON %s = %s",
								boiler.TableNames.MechSkin,
								boiler.MechTableColumns.ChassisSkinID,
								boiler.MechSkinTableColumns.ID,
							)),

							// inner join blueprint mech skin
							qm.InnerJoin(fmt.Sprintf(
								"%s ON %s = %s",
								boiler.TableNames.BlueprintMechSkin,
								boiler.MechSkinTableColumns.BlueprintID,
								boiler.BlueprintMechSkinTableColumns.ID,
							)),

							qm.InnerJoin(fmt.Sprintf(
								"%s ON %s = %s AND %s = %s",
								boiler.TableNames.MechModelSkinCompatibilities,
								boiler.MechModelSkinCompatibilityTableColumns.MechModelID,
								boiler.MechTableColumns.BlueprintID,
								boiler.MechModelSkinCompatibilityTableColumns.BlueprintMechSkinID,
								boiler.MechSkinTableColumns.BlueprintID,
							)),
						}

						rows, err := boiler.NewQuery(queries...).Query(gamedb.StdConn)
						if err != nil {
							gamelog.L.Error().Err(err).Interface("boiler query", queries).Msg("Failed to load mech data for system message")
							return
						}

						for rows.Next() {
							mechID := ""
							name := ""
							label := ""
							tier := ""
							avatarURL := ""
							totalBlocks := 0
							damagedBlocks := 0

							err = rows.Scan(&mechID, &name, &label, &tier, &avatarURL, &totalBlocks, &damagedBlocks)
							if err != nil {
								gamelog.L.Error().Err(err).Msg("Failed to scan mech detail for system message of expired lobby.")
								return
							}

							// fill the mech data
							index := slices.IndexFunc(pms.MechBriefs, func(mb *SystemMessageMech) bool { return mb.MechID == mechID })
							if index != -1 {
								pms.MechBriefs[index].Name = name
								pms.MechBriefs[index].TotalBlocks = totalBlocks
								pms.MechBriefs[index].DamagedBlocks = damagedBlocks
								pms.MechBriefs[index].Tier = tier
								pms.MechBriefs[index].ImageUrl = avatarURL
								if name == "" {
									pms.MechBriefs[index].Name = label
								}
							}
						}

						// send battle reward system message
						b, err := json.Marshal(pms.MechBriefs)
						if err != nil {
							gamelog.L.Error().Interface("player reward data", pms.MechBriefs).Err(err).Msg("Failed to marshal mech data into json.")
							return
						}

						sysMsg := boiler.SystemMessage{
							PlayerID: pms.PlayerID,
							SenderID: server.SupremacyBattleUserID,
							DataType: null.StringFrom(string(system_messages.SystemMessageDataTypeExpiredBattleLobby)),
							Title:    "Expired Battle Lobby",
							Message:  fmt.Sprintf("Unfortunately, the battle lobby '%s' you joined has expired.", battleLobby.Name),
							Data:     null.JSONFrom(b),
						}

						err = sysMsg.Insert(gamedb.StdConn, boil.Infer())
						if err != nil {
							gamelog.L.Error().Err(err).Interface("newSystemMessage", sysMsg).Msg("failed to insert new system message into db")
							return
						}
						ws.PublishMessage(fmt.Sprintf("/secure/user/%s/system_messages", pms.PlayerID), server.HubKeySystemMessageListUpdatedSubscribe, true)
					}

				}(playerMechs)
			}

			// broadcast the status changes of the lobby mechs
			am.MechDebounceBroadcastChan <- lobbyMechIDs

		}(bl)

		if !bl.AccessCode.Valid && !bl.IsAiDrivenMatch {
			// go am.DiscordSession.SendBattleLobbyEditMessage(bl.ID, "")
		}
	}

	wg.Wait()

	return nil
}

// DefaultPublicLobbiesCheck check there are enough public lobbies
func (am *ArenaManager) DefaultPublicLobbiesCheck() error {
	// load default public lobbies amount
	publicLobbiesCount := db.GetIntWithDefault(db.KeyDefaultPublicLobbyCount, 1)

	// lock queue func
	am.BattleQueueFuncMx.Lock()
	defer am.BattleQueueFuncMx.Unlock()

	bls, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.ReadyAt.IsNull(),
		boiler.BattleLobbyWhere.GeneratedBySystem.EQ(true),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load active public battle lobbies.")
		return terror.Error(err, "Failed to load active public battle lobbies.")
	}

	count := len(bls)

	if count >= publicLobbiesCount {
		return nil
	}

	// fill up battle lobbies
	for i := 0; i < publicLobbiesCount-count; i++ {
		bl := &boiler.BattleLobby{
			Name:                  helpers.GenerateAdjectiveName(),
			HostByID:              server.SupremacyBattleUserID,
			EntryFee:              decimal.Zero, // free to join
			FirstFactionCut:       decimal.NewFromFloat(0.75),
			SecondFactionCut:      decimal.NewFromFloat(0.25),
			ThirdFactionCut:       decimal.Zero,
			EachFactionMechAmount: 3,
			GeneratedBySystem:     true,
		}

		err = bl.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to insert public battle lobbies.")
			return terror.Error(err, "Failed to insert public battle lobbies.")
		}

		amount := db.GetDecimalWithDefault(db.KeySystemLobbyDefaultExtraReward, decimal.New(100, 18))

		if amount.GreaterThan(decimal.Zero) {
			paidTXID, err := am.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
				FromUserID:           uuid.UUID(server.XsynTreasuryUserID),
				ToUserID:             uuid.FromStringOrNil(server.SupremacyBattleUserID),
				Amount:               amount.StringFixed(0),
				TransactionReference: server.TransactionReference(fmt.Sprintf("top_up_system_lobby_default_reward|%s|%d", bl.ID, time.Now().UnixNano())),
				Group:                string(server.TransactionGroupSupremacy),
				SubGroup:             string(server.TransactionGroupBattle),
				Description:          fmt.Sprintf("top up system lobby default reward %s.", bl.ID),
			})
			if err != nil {
				return terror.Error(err, "Failed to top up reward.")
			}

			blr := &boiler.BattleLobbyExtraSupsReward{
				BattleLobbyID: bl.ID,
				OfferedByID:   server.SupremacyBattleUserID,
				Amount:        amount,
				PaidTXID:      paidTXID,
			}

			err = blr.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Interface("battle lobby reward", blr).Msg("Failed to add battle lobby reward.")
				continue
			}
		}
	}

	am.KickFactionLobbyChecker()

	return nil
}

// EmptySystemLobbyRemover delete any empty lobby and left one available
func (am *ArenaManager) EmptySystemLobbyRemover() {
	// load default public lobbies amount
	publicLobbiesCount := db.GetIntWithDefault(db.KeyDefaultPublicLobbyCount, 1)

	// lock queue func
	am.BattleQueueFuncMx.Lock()
	defer am.BattleQueueFuncMx.Unlock()

	now := time.Now()

	bls, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.GeneratedBySystem.EQ(true),
		boiler.BattleLobbyWhere.ReadyAt.IsNull(),
		qm.Where(fmt.Sprintf(
			"NOT EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s ISNULL AND %s ISNULL)",
			boiler.TableNames.BattleLobbiesMechs,
			boiler.BattleLobbiesMechTableColumns.BattleLobbyID,
			boiler.BattleLobbyTableColumns.ID,
			boiler.BattleLobbiesMechTableColumns.RefundTXID,
			boiler.BattleLobbiesMechTableColumns.DeletedAt,
		)),
		qm.OrderBy(boiler.BattleLobbyTableColumns.CreatedAt),

		qm.Load(boiler.BattleLobbyRels.BattleLobbyExtraSupsRewards),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load empty battle lobbies.")
		return
	}

	if len(bls) <= publicLobbiesCount {
		return
	}

	// preserve lobbies
	bls = slices.Delete(bls, 0, publicLobbiesCount)

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to begin db transaction.")
		return
	}

	defer tx.Rollback()

	// soft delete the rest of the system lobbies
	_, err = bls.UpdateAll(tx, boiler.M{
		boiler.BattleLobbyColumns.EndedAt:   null.TimeFrom(now),
		boiler.BattleLobbyColumns.DeletedAt: null.TimeFrom(now),
	})
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to soft delete redundant system lobbies.")
		return
	}

	// generate deleted lobbies
	var deletedLobbyIDs []string

	var refundFns []func()
	refund := func() {
		for _, fn := range refundFns {
			fn()
		}
	}

	// refund extra reward and
	for _, bl := range bls {
		if bl.R != nil {
			for _, er := range bl.R.BattleLobbyExtraSupsRewards {
				// refund the payment
				refundTXID, err := am.RPCClient.RefundSupsMessage(er.PaidTXID)
				if err != nil {
					refund()
					gamelog.L.Error().Err(err).Msg("Failed to refund lobby extra sups")
					return
				}

				// append refund function
				refundFns = append(refundFns, func() {
					txID, err := am.RPCClient.RefundSupsMessage(refundTXID)
					if err != nil {
						gamelog.L.Error().Err(err).Str("tx id", txID).Msg("Failed to refund the refund")
						return
					}
				})

				// update refund
				er.RefundedTXID = null.StringFrom(refundTXID)
				_, err = er.Update(tx, boil.Whitelist(boiler.BattleLobbyExtraSupsRewardColumns.RefundedTXID))
				if err != nil {
					refund()
					gamelog.L.Error().Err(err).Interface("extra reward", er).Msg("Failed to update refund tx id of lobby extra reward.")
					return
				}

			}
		}

		deletedLobbyIDs = append(deletedLobbyIDs, bl.ID)
	}

	err = tx.Commit()
	if err != nil {
		refund()
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
		return
	}

	// broadcast deleted lobbies
	am.BattleLobbyDebounceBroadcastChan <- deletedLobbyIDs

	// free up lobbies
	bls = nil
	refundFns = nil
	refund = nil
	deletedLobbyIDs = nil
}

func BroadcastPlayerQueueStatus(playerID string) {
	resp := &server.PlayerQueueStatus{
		TotalQueued: 0,
		QueueLimit:  db.GetIntWithDefault(db.KeyPlayerQueueLimit, 10),
	}

	blms, err := boiler.BattleLobbiesMechs(
		boiler.BattleLobbiesMechWhere.QueuedByID.EQ(playerID),
		boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
		boiler.BattleLobbiesMechWhere.EndedAt.IsNull(),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("player id", playerID).Err(err).Msg("Failed to load player battle queue mechs")
		return
	}

	if blms != nil {
		resp.TotalQueued = len(blms)
	}

	ws.PublishMessage(fmt.Sprintf("/secure/user/%s/queue_status", playerID), server.HubKeyPlayerQueueStatus, resp)
}

// GenerateAIDrivenBattle load mechs from mech staking pool, and fill with AI mechs if no enough
func GenerateAIDrivenBattle() (*boiler.BattleLobby, error) {
	l := gamelog.L.With().Str("func", "GenerateAIDrivenBattle").Logger()

	now := time.Now()

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		l.Error().Err(err).Msg("Failed to start db transaction.")
		return nil, terror.Error(err, "Failed to start db transaction.")
	}

	defer tx.Rollback()

	bl := &boiler.BattleLobby{
		HostByID:              server.SupremacyBattleUserID,
		Name:                  helpers.GenerateAdjectiveName(),
		EntryFee:              decimal.Zero, // free to join
		FirstFactionCut:       decimal.NewFromFloat(0.75),
		SecondFactionCut:      decimal.NewFromFloat(0.25),
		ThirdFactionCut:       decimal.Zero,
		EachFactionMechAmount: 3,
		GeneratedBySystem:     true,
		IsAiDrivenMatch:       true, // is AI driven match
		ReadyAt:               null.TimeFrom(now),
	}

	err = bl.Insert(tx, boil.Infer())
	if err != nil {
		l.Error().Err(err).Msg("Failed to insert AI driven battle.")
		return nil, terror.Error(err, "Failed to insert AI driven battle.")
	}

	// get mechs from staked pool (not damaged and not AI)
	sms, err := boiler.StakedMechs(
		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s AND %s = FALSE",
			boiler.TableNames.Players,
			boiler.StakedMechTableColumns.OwnerID,
			boiler.PlayerTableColumns.ID,
			boiler.PlayerTableColumns.IsAi,
		)),
		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.Mechs,
			boiler.MechTableColumns.ID,
			boiler.StakedMechTableColumns.MechID,
		)),
		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.BlueprintMechs,
			boiler.BlueprintMechTableColumns.ID,
			boiler.MechTableColumns.BlueprintID,
		)),
		qm.Where(fmt.Sprintf(
			"NOT EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s ISNULL)",
			boiler.TableNames.BattleLobbiesMechs,
			boiler.BattleLobbiesMechTableColumns.MechID,
			boiler.StakedMechTableColumns.MechID,
			boiler.BattleLobbiesMechTableColumns.EndedAt,
		)),
		qm.Where(fmt.Sprintf(
			`NOT EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s ISNULL AND %s < (%s - %s)*2)`,
			boiler.TableNames.RepairCases,
			boiler.RepairCaseTableColumns.MechID,
			boiler.StakedMechTableColumns.MechID,
			boiler.RepairCaseTableColumns.CompletedAt,
			boiler.BlueprintMechTableColumns.RepairBlocks,
			boiler.RepairCaseTableColumns.BlocksRequiredRepair,
			boiler.RepairCaseTableColumns.BlocksRepaired,
		)),
	).All(tx)
	if err != nil {
		l.Error().Err(err).Msg("Failed to load staked mechs")
		return nil, terror.Error(err, "Failed to load staked mechs.")
	}

	// shuffle list
	rand.Seed(time.Now().UnixNano())
	for i := range sms {
		j := rand.Intn(i + 1)
		sms[i], sms[j] = sms[j], sms[i]
	}

	// control insert mech amount
	rmCount := 0
	bcCount := 0
	zaiCount := 0

	var blms []*boiler.BattleLobbiesMech
	for _, sm := range sms {
		queuedByID := ""
		switch sm.FactionID {
		case server.RedMountainFactionID:
			if rmCount == bl.EachFactionMechAmount {
				continue
			}
			queuedByID = server.RedMountainPlayerID
			rmCount++
		case server.BostonCyberneticsFactionID:
			if bcCount == bl.EachFactionMechAmount {
				continue
			}
			queuedByID = server.BostonCyberneticsPlayerID
			bcCount++
		case server.ZaibatsuFactionID:
			if zaiCount == bl.EachFactionMechAmount {
				continue
			}
			queuedByID = server.ZaibatsuPlayerID
			zaiCount++
		}

		blms = append(blms, &boiler.BattleLobbiesMech{
			BattleLobbyID: bl.ID,
			MechID:        sm.MechID,
			QueuedByID:    queuedByID,
			FactionID:     sm.FactionID,
			LockedAt:      bl.ReadyAt,
		})
	}

	// if not enough player mechs, insert AI mechs
	if rmCount < bl.EachFactionMechAmount || bcCount < bl.EachFactionMechAmount || zaiCount < bl.EachFactionMechAmount {
		// load AI mechs
		stakedAIMechs, err := boiler.StakedMechs(
			qm.Where(fmt.Sprintf(
				`EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s = TRUE)`,
				boiler.TableNames.Players,
				boiler.PlayerTableColumns.ID,
				boiler.StakedMechTableColumns.OwnerID,
				boiler.PlayerTableColumns.IsAi,
			)),
		).All(gamedb.StdConn)
		if err != nil {
			l.Error().Err(err).Msg("Failed to load AI mechs.")
			return nil, terror.Error(err, "Failed to load AI mechs.")
		}

		for _, sm := range stakedAIMechs {
			switch sm.FactionID {
			case server.RedMountainFactionID:
				if rmCount == bl.EachFactionMechAmount {
					continue
				}
				rmCount++
			case server.BostonCyberneticsFactionID:
				if bcCount == bl.EachFactionMechAmount {
					continue
				}
				bcCount++
			case server.ZaibatsuFactionID:
				if zaiCount == bl.EachFactionMechAmount {
					continue
				}
				zaiCount++
			}

			blms = append(blms, &boiler.BattleLobbiesMech{
				BattleLobbyID: bl.ID,
				MechID:        sm.MechID,
				QueuedByID:    sm.OwnerID,
				FactionID:     sm.FactionID,
				LockedAt:      bl.ReadyAt,
			})
		}
	}

	for _, blm := range blms {
		err = blm.Insert(tx, boil.Infer())
		if err != nil {
			l.Error().Err(err).Msg("Failed to insert assign mech to battle lobby.")
			return nil, terror.Error(err, "Failed to insert assign mech to battle lobby.")
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, terror.Error(err, "Failed to commit db transaction.")
	}

	return bl, nil
}

// AIMechFillingProcess fill up the lobby with AI mechs
// IMPORTANT: this function MUST NOT be wrapped inside the "ArenaManager.SendBattleQueueFunc()" function
func (am *ArenaManager) AIMechFillingProcess(battleLobbyID string) {
	am.BattleQueueFuncMx.Lock()
	defer am.BattleQueueFuncMx.Unlock()

	// load AI mechs
	cis, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
		qm.Where(fmt.Sprintf(
			"EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s = TRUE)",
			boiler.TableNames.Players,
			boiler.PlayerTableColumns.ID,
			boiler.CollectionItemTableColumns.OwnerID,
			boiler.PlayerTableColumns.IsAi,
		)),
		qm.Load(boiler.CollectionItemRels.Owner),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load AI mechs.")
		return
	}

	// load the battle lobby and its queued mechs
	bl, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.ID.EQ(battleLobbyID),
		qm.Load(boiler.BattleLobbyRels.BattleLobbiesMechs),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("battle lobby id", battleLobbyID).Msg("Failed to load battle lobby")
		return
	}

	// terminate the process if the lobby is not exists
	if bl == nil {
		return
	}

	// terminate the process if there is no mech queued in the lobby
	if bl.R == nil || bl.R.BattleLobbiesMechs == nil || len(bl.R.BattleLobbiesMechs) == 0 {
		return
	}

	// terminate the process if the lobby is full
	if len(bl.R.BattleLobbiesMechs) == bl.EachFactionMechAmount*3 {
		return
	}

	// fill the lobby with staked mechs
	type factionAvailableSlots struct {
		factionID      string
		availableSlots int
	}
	factionSlots := []*factionAvailableSlots{
		{server.RedMountainFactionID, bl.EachFactionMechAmount},
		{server.BostonCyberneticsFactionID, bl.EachFactionMechAmount},
		{server.ZaibatsuFactionID, bl.EachFactionMechAmount},
	}

	lobbyMechIDs := []string{}
	for _, blm := range bl.R.BattleLobbiesMechs {
		lobbyMechIDs = append(lobbyMechIDs, blm.MechID)

		// find faction slot
		index := slices.IndexFunc(factionSlots, func(fs *factionAvailableSlots) bool { return fs.factionID == blm.FactionID })

		// should never happen, but just in case.
		if index == -1 {
			gamelog.L.Error().Str("faction id", blm.FactionID).Msg("Detect a faction id that is not exist in the system!!!")
			return
		}

		factionSlots[index].availableSlots -= 1
	}

	var insertRows []string
	for _, factionSlot := range factionSlots {
		if factionSlot.availableSlots <= 0 {
			continue
		}

		// queued by faction AI player
		queuedByID := ""
		switch factionSlot.factionID {
		case server.RedMountainFactionID:
			queuedByID = server.RedMountainPlayerID
		case server.BostonCyberneticsFactionID:
			queuedByID = server.BostonCyberneticsPlayerID
		case server.ZaibatsuFactionID:
			queuedByID = server.ZaibatsuPlayerID
		}

		// load available staked mechs
		sms, err := boiler.StakedMechs(
			boiler.StakedMechWhere.FactionID.EQ(factionSlot.factionID),
			qm.Where(fmt.Sprintf(
				"NOT EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s ISNULL AND %s ISNULL AND %s ISNULL)",
				boiler.TableNames.BattleLobbiesMechs,
				boiler.BattleLobbiesMechTableColumns.MechID,
				boiler.StakedMechTableColumns.MechID,
				boiler.BattleLobbiesMechTableColumns.EndedAt,
				boiler.BattleLobbiesMechTableColumns.RefundTXID,
				boiler.BattleLobbiesMechTableColumns.DeletedAt,
			)),
			// no AI mechs
			qm.Where(fmt.Sprintf(
				"EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s = FALSE)",
				boiler.TableNames.Players,
				boiler.PlayerTableColumns.ID,
				boiler.StakedMechTableColumns.OwnerID,
				boiler.PlayerTableColumns.IsAi,
			)),
			// no damaged mech
			qm.Where(fmt.Sprintf(
				"NOT EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s ISNULL)",
				boiler.TableNames.RepairCases,
				boiler.RepairCaseTableColumns.MechID,
				boiler.StakedMechTableColumns.MechID,
				boiler.RepairCaseTableColumns.CompletedAt,
			)),
			qm.Limit(factionSlot.availableSlots),
		).All(gamedb.StdConn)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to load staked mech.")
		}

		var mechIDs []string
		for _, sm := range sms {
			mechIDs = append(mechIDs, sm.MechID)
		}
		for _, ci := range cis {
			// skip, if the faction id not match
			if ci.R == nil || ci.R.Owner == nil || ci.R.Owner.FactionID.String != factionSlot.factionID {
				continue
			}
			mechIDs = append(mechIDs, ci.ItemID)
		}

		// generate insert rows
		for _, mechID := range mechIDs {
			// fill AI mechs into the slots
			insertRows = append(insertRows, fmt.Sprintf(
				"('%s', '%s', '%s', '%s')",
				battleLobbyID,
				mechID,
				queuedByID,
				factionSlot.factionID,
			))

			factionSlot.availableSlots -= 1

			// break, no available slot
			if factionSlot.availableSlots == 0 {
				break
			}
		}
	}

	if len(insertRows) == 0 {
		return
	}

	// insert AI mech into the lobby
	q := fmt.Sprintf(
		"INSERT INTO %s (%s, %s, %s, %s)  VALUES ",
		boiler.TableNames.BattleLobbiesMechs,
		boiler.BattleLobbiesMechColumns.BattleLobbyID,
		boiler.BattleLobbiesMechColumns.MechID,
		boiler.BattleLobbiesMechColumns.QueuedByID,
		boiler.BattleLobbiesMechColumns.FactionID,
	)

	for i, insertRow := range insertRows {
		q += insertRow
		if i < len(insertRows)-1 {
			q += ","
			continue
		}
		q += ";"
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to start db transaction")
		return
	}

	defer tx.Rollback()

	_, err = tx.Exec(q)
	if err != nil {
		gamelog.L.Error().Err(err).Str("query", q).Msg("Failed to insert battle lobby mechs")
		return
	}

	bl.ReadyAt = null.TimeFrom(time.Now())
	bl.AccessCode = null.StringFromPtr(nil)
	_, err = bl.Update(tx, boil.Whitelist(boiler.BattleLobbyColumns.ReadyAt, boiler.BattleLobbyColumns.AccessCode))
	if err != nil {
		gamelog.L.Error().Interface("battle lobby", bl).Err(err).Msg("Failed to update battle lobby.")
		return
	}

	_, err = bl.BattleLobbiesMechs(
		boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
	).UpdateAll(tx, boiler.M{boiler.BattleLobbiesMechColumns.LockedAt: bl.ReadyAt})
	if err != nil {
		gamelog.L.Error().Interface("battle lobby", bl).Err(err).Msg("Failed to lock battle lobby mechs.")
		return
	}

	err = tx.Commit()
	if err != nil {
		return
	}

	// broadcast battle lobby
	am.BattleLobbyDebounceBroadcastChan <- []string{bl.ID}

	// broadcast the status changes of the lobby mechs
	am.MechDebounceBroadcastChan <- lobbyMechIDs

	// update faction staked mech queue status
	am.FactionStakedMechDashboardKeyChan <- []string{FactionStakedMechDashboardKeyQueue}

}

type UserBattleMechAlert struct {
	userID           string
	battleLobbyAlert *server.BattleLobbyMechsAlert
}

func broadcastBattleMechAlert(lobbyID string) {
	mib, err := db.GetMechsInLobby(lobbyID)
	if err != nil {
		return
	}

	var ubs []*UserBattleMechAlert
	for _, data := range mib {
		// queued by mech data
		queuedByIndex := slices.IndexFunc(ubs, func(ub *UserBattleMechAlert) bool { return ub.userID == data.QueuedByID })
		if queuedByIndex == -1 {
			ubs = append(ubs, &UserBattleMechAlert{data.QueuedByID, &server.BattleLobbyMechsAlert{
				ArenaID:    data.ArenaID,
				MechAlerts: []*server.MechAlert{},
			}})
			queuedByIndex = len(ubs) - 1
		}

		if slices.IndexFunc(ubs[queuedByIndex].battleLobbyAlert.MechAlerts, func(ma *server.MechAlert) bool { return ma.ID == data.MechID }) == -1 {
			ma := &server.MechAlert{
				ID:   data.MechID,
				Name: data.MechName,
			}

			if ma.Name == "" {
				ma.Name = data.MechLabel
			}

			ubs[queuedByIndex].battleLobbyAlert.MechAlerts = append(ubs[queuedByIndex].battleLobbyAlert.MechAlerts, ma)
		}

		// add alert for staked mech owners
		if data.StakedMechOwnerID.Valid {
			queuedByIndex = slices.IndexFunc(ubs, func(ub *UserBattleMechAlert) bool { return ub.userID == data.StakedMechOwnerID.String })
			if queuedByIndex == -1 {
				ubs = append(ubs, &UserBattleMechAlert{
					data.StakedMechOwnerID.String,
					&server.BattleLobbyMechsAlert{
						ArenaID:    data.ArenaID,
						MechAlerts: []*server.MechAlert{},
					}})
				queuedByIndex = len(ubs) - 1
			}

			if slices.IndexFunc(ubs[queuedByIndex].battleLobbyAlert.MechAlerts, func(ma *server.MechAlert) bool { return ma.ID == data.MechID }) == -1 {
				ma := &server.MechAlert{
					ID:   data.MechID,
					Name: data.MechName,
				}

				if ma.Name == "" {
					ma.Name = data.MechLabel
				}

				ubs[queuedByIndex].battleLobbyAlert.MechAlerts = append(ubs[queuedByIndex].battleLobbyAlert.MechAlerts, ma)
			}
		}
	}

	for _, ub := range ubs {
		ws.PublishMessage(fmt.Sprintf("/secure/user/%s/browser_alert", ub.userID), server.HubKeyPlayerBrowserAlert, &server.PlayerBrowserAlertStruct{
			Title: "MECH_IN_BATTLE",
			Data:  ub.battleLobbyAlert,
		})
	}
}

func (am *ArenaManager) KickFactionLobbyChecker() {
	go am.FactionBattleLobbyMechsChecker(server.RedMountainFactionID)
	go am.FactionBattleLobbyMechsChecker(server.BostonCyberneticsFactionID)
	go am.FactionBattleLobbyMechsChecker(server.ZaibatsuFactionID)
}

// FactionBattleLobbyMechsChecker check the faction of the
func (am *ArenaManager) FactionBattleLobbyMechsChecker(factionID string) {
	am.BattleQueueFuncMx.Lock()
	defer am.BattleQueueFuncMx.Unlock()

	queries := []qm.QueryMod{
		qm.Select(boiler.BattleLobbyTableColumns.ID),
		qm.From(fmt.Sprintf(
			"(SELECT * FROM %s WHERE %s = TRUE AND %s ISNULL AND %s ISNULL) %s",
			boiler.TableNames.BattleLobbies,
			boiler.BattleLobbyTableColumns.GeneratedBySystem,
			boiler.BattleLobbyTableColumns.ReadyAt,
			boiler.BattleLobbyTableColumns.DeletedAt,
			boiler.TableNames.BattleLobbies,
		)),

		qm.LeftOuterJoin(fmt.Sprintf(
			"%s ON %s = %s AND %s = '%s' AND %s ISNULL AND %s ISNULL",
			boiler.TableNames.BattleLobbiesMechs,
			boiler.BattleLobbiesMechTableColumns.BattleLobbyID,
			boiler.BattleLobbyTableColumns.ID,
			boiler.BattleLobbiesMechTableColumns.FactionID,
			factionID,
			boiler.BattleLobbiesMechTableColumns.RefundTXID,
			boiler.BattleLobbiesMechTableColumns.DeletedAt,
		)),

		qm.GroupBy(boiler.BattleLobbyTableColumns.ID + "," + boiler.BattleLobbyTableColumns.EachFactionMechAmount),
		qm.Having(fmt.Sprintf("COUNT(%s) < %s", boiler.BattleLobbiesMechTableColumns.ID, boiler.BattleLobbyTableColumns.EachFactionMechAmount)),
		qm.Limit(1),
	}
	availableLobbyID := ""
	err := boiler.NewQuery(queries...).QueryRow(gamedb.StdConn).Scan(&availableLobbyID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msg("Failed to check faction queue lobby")
		return
	}

	// skip, if there still lobby available
	if availableLobbyID != "" {
		return
	}

	// generate a new system lobby

	bl := &boiler.BattleLobby{
		Name:                  helpers.GenerateAdjectiveName(),
		HostByID:              server.SupremacyBattleUserID,
		EntryFee:              decimal.Zero, // free to join
		FirstFactionCut:       decimal.NewFromFloat(0.75),
		SecondFactionCut:      decimal.NewFromFloat(0.25),
		ThirdFactionCut:       decimal.Zero,
		EachFactionMechAmount: 3,
		GeneratedBySystem:     true,
	}

	err = bl.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to insert public battle lobbies.")
		return
	}

	amount := db.GetDecimalWithDefault(db.KeySystemLobbyDefaultExtraReward, decimal.New(100, 18))

	if amount.GreaterThan(decimal.Zero) {
		paidTXID, err := am.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
			FromUserID:           uuid.UUID(server.XsynTreasuryUserID),
			ToUserID:             uuid.FromStringOrNil(server.SupremacyBattleUserID),
			Amount:               amount.StringFixed(0),
			TransactionReference: server.TransactionReference(fmt.Sprintf("top_up_system_lobby_default_reward|%s|%d", bl.ID, time.Now().UnixNano())),
			Group:                string(server.TransactionGroupSupremacy),
			SubGroup:             string(server.TransactionGroupBattle),
			Description:          fmt.Sprintf("top up system lobby default reward %s.", bl.ID),
		})
		if err != nil {
			return
		}

		blr := &boiler.BattleLobbyExtraSupsReward{
			BattleLobbyID: bl.ID,
			OfferedByID:   server.SupremacyBattleUserID,
			Amount:        amount,
			PaidTXID:      paidTXID,
		}

		err = blr.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Interface("battle lobby reward", blr).Msg("Failed to add battle lobby reward.")
			return
		}
	}

	// broadcast battle lobby
	am.BattleLobbyDebounceBroadcastChan <- []string{bl.ID}
}
