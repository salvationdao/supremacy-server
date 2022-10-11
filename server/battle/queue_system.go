package battle

import (
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"time"
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
			broadcastBattleLobbyUpdate(impactedBattleLobbyIDs...)

			// clean up battle lobby id list
			impactedBattleLobbyIDs = []string{}
		}
	}
}

// broadcastBattleLobbyUpdate broadcast the updated lobbies to each faction
func broadcastBattleLobbyUpdate(battleLobbyIDs ...string) {
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
		factionID string
		bls       []*server.BattleLobby
	}

	playerInvolvedLobbiesMap := make(map[string]*playerInvolveLobby)
	// broadcast to individual
	for _, bl := range battleLobbies {

		if bl.HostBy != nil && bl.HostBy.FactionID.Valid {
			host := bl.HostBy
			// check host player
			_, ok := playerInvolvedLobbiesMap[host.ID]
			if !ok {
				playerInvolvedLobbiesMap[host.ID] = &playerInvolveLobby{
					factionID: host.FactionID.String,
					bls:       []*server.BattleLobby{},
				}
			}
			playerInvolvedLobbiesMap[host.ID].bls = append(playerInvolvedLobbiesMap[host.ID].bls, bl)
		}

		// check joined players
		for _, blm := range bl.BattleLobbiesMechs {
			if blm.Owner == nil || !blm.Owner.FactionID.Valid {
				continue
			}

			ownerID := blm.Owner.ID
			factionID := blm.Owner.FactionID.String

			_, ok := playerInvolvedLobbiesMap[ownerID]
			if !ok {
				playerInvolvedLobbiesMap[ownerID] = &playerInvolveLobby{
					factionID: factionID,
					bls:       []*server.BattleLobby{},
				}
			}

			// skip, if the player already have the lobby on their list
			if slices.IndexFunc(playerInvolvedLobbiesMap[ownerID].bls, func(battleLobby *server.BattleLobby) bool { return battleLobby.ID == bl.ID }) != -1 {
				continue
			}

			// otherwise, append to lobby to the player's list
			playerInvolvedLobbiesMap[ownerID].bls = append(playerInvolvedLobbiesMap[ownerID].bls, bl)
		}
	}
	// broadcast the lobbies which players are involved in
	for playerID, pil := range playerInvolvedLobbiesMap {
		go ws.PublishMessage(fmt.Sprintf("/secure/user/%s/involved_battle_lobbies", playerID), server.HubKeyInvolvedBattleLobbyListUpdate, server.BattleLobbiesFactionFilter(pil.bls, pil.factionID))
	}

	// separate public and private lobbies
	var publicLobbies []*server.BattleLobby

	for _, battleLobby := range battleLobbies {
		if !battleLobby.AccessCode.Valid {
			publicLobbies = append(publicLobbies, battleLobby)
			continue
		}
		// broadcast private lobby individually
		go ws.PublishMessage(fmt.Sprintf("/faction/%s/private_battle_lobby/%s", server.RedMountainFactionID, battleLobby.AccessCode.String), server.HubKeyPrivateBattleLobbyUpdate, server.BattleLobbyInfoFilter(battleLobby, server.RedMountainFactionID))
		go ws.PublishMessage(fmt.Sprintf("/faction/%s/private_battle_lobby/%s", server.BostonCyberneticsFactionID, battleLobby.AccessCode.String), server.HubKeyPrivateBattleLobbyUpdate, server.BattleLobbyInfoFilter(battleLobby, server.BostonCyberneticsFactionID))
		go ws.PublishMessage(fmt.Sprintf("/faction/%s/private_battle_lobby/%s", server.ZaibatsuFactionID, battleLobby.AccessCode.String), server.HubKeyPrivateBattleLobbyUpdate, server.BattleLobbyInfoFilter(battleLobby, server.ZaibatsuFactionID))
	}

	if len(publicLobbies) > 0 {
		go ws.PublishMessage(fmt.Sprintf("/faction/%s/battle_lobbies", server.RedMountainFactionID), server.HubKeyBattleLobbyListUpdate, append(server.BattleLobbiesFactionFilter(publicLobbies, server.RedMountainFactionID), deletedLobbies...))
		go ws.PublishMessage(fmt.Sprintf("/faction/%s/battle_lobbies", server.BostonCyberneticsFactionID), server.HubKeyBattleLobbyListUpdate, append(server.BattleLobbiesFactionFilter(publicLobbies, server.BostonCyberneticsFactionID), deletedLobbies...))
		go ws.PublishMessage(fmt.Sprintf("/faction/%s/battle_lobbies", server.ZaibatsuFactionID), server.HubKeyBattleLobbyListUpdate, append(server.BattleLobbiesFactionFilter(publicLobbies, server.ZaibatsuFactionID), deletedLobbies...))
	}
}

// SetDefaultPublicBattleLobbies ensure there are enough battle lobbies when server start
func (am *ArenaManager) SetDefaultPublicBattleLobbies() error {
	// check once when server start
	err := am.DefaultPublicLobbiesCheck()
	if err != nil {
		return err
	}

	// check every minutes
	go func() {
		ticker := time.NewTicker(1 * time.Minute)

		for {
			<-ticker.C
			err = am.DefaultPublicLobbiesCheck()
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to check default public lobbies.")
			}
		}
	}()

	return nil
}

// DefaultPublicLobbiesCheck check there are enough public lobbies
func (am *ArenaManager) DefaultPublicLobbiesCheck() error {
	// load default public lobbies amount
	publicLobbiesCount := db.GetIntWithDefault(db.KeyDefaultPublicLobbyCount, 20)

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
	}

	return nil
}

func BroadcastMechQueueStatus(playerID string, mechIDs ...string) {
	if len(mechIDs) == 0 {
		return
	}

	mechInfo, err := db.OwnedMechsBrief(playerID, mechIDs...)
	if err != nil {
		return
	}

	ws.PublishMessage(fmt.Sprintf("/secure/user/%s/owned_mechs", playerID), server.HubKeyPlayerMechsBrief, mechInfo)
}

func BroadcastPlayerQueueStatus(playerID string) {
	resp := &server.PlayerQueueStatus{
		TotalQueued: 0,
		QueueLimit:  db.GetIntWithDefault(db.KeyPlayerQueueLimit, 10),
	}

	blms, err := boiler.BattleLobbiesMechs(
		boiler.BattleLobbiesMechWhere.OwnerID.EQ(playerID),
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

	// get default mechs
	rows, err := boiler.NewQuery(
		qm.Select(
			boiler.MechTableColumns.ID,
			boiler.CollectionItemTableColumns.OwnerID,
			fmt.Sprintf(
				"(SELECT %s FROM %s WHERE %s = %s)",
				boiler.PlayerTableColumns.FactionID,
				boiler.TableNames.Players,
				boiler.PlayerTableColumns.ID,
				boiler.CollectionItemTableColumns.OwnerID,
			),
		),

		qm.From(fmt.Sprintf(
			"(SELECT %s FROM %s WHERE %s = TRUE) %s",
			boiler.MechTableColumns.ID,
			boiler.TableNames.Mechs,
			boiler.MechTableColumns.IsDefault,
			boiler.TableNames.Mechs,
		)),

		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.CollectionItems,
			boiler.CollectionItemTableColumns.ItemID,
			boiler.MechTableColumns.ID,
		)),
	).Query(tx)
	if err != nil {
		l.Error().Err(err).Msg("Failed to load default mechs.")
		return nil, terror.Error(err, "Failed to load default mechs.")
	}

	// control insert mech amount
	rmCount := 0
	bcCount := 0
	zaiCount := 0

	var blms []*boiler.BattleLobbiesMech
	for rows.Next() {
		mechID := ""
		ownerID := ""
		factionID := ""

		err = rows.Scan(&mechID, &ownerID, &factionID)
		if err != nil {
			l.Error().Err(err).Msg("Failed to scan mech info.")
			return nil, terror.Error(err, "Failed to scan mech info.")
		}

		switch factionID {
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
			MechID:        mechID,
			OwnerID:       ownerID,
			FactionID:     factionID,
			LockedAt:      bl.ReadyAt,
		})
	}

	if rmCount < bl.EachFactionMechAmount || bcCount < bl.EachFactionMechAmount || zaiCount < bl.EachFactionMechAmount {
		l.Error().Err(err).Msg("Not enough mech to generate AI driven match.")
		return nil, terror.Error(err, "Not enough mech to generate AI driven match.")
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
