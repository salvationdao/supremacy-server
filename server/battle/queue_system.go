package battle

import (
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"
)

func (am *ArenaManager) SendBattleQueueFunc(fn func() error) error {
	am.BattleQueueFuncMx.Lock()
	defer am.BattleQueueFuncMx.Unlock()
	return fn()
}

func BroadcastBattleLobbyUpdate(battleLobbyIDs ...string) {
	if battleLobbyIDs == nil || len(battleLobbyIDs) == 0 {
		return
	}

	bls, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.ID.IN(battleLobbyIDs),
		qm.Load(boiler.BattleLobbyRels.HostBy),
		qm.Load(boiler.BattleLobbyRels.GameMap),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Strs("battle lobby id list", battleLobbyIDs).Msg("Failed to query battle lobby")
		return
	}

	if bls == nil {
		return
	}

	resp, err := server.BattleLobbiesFromBoiler(bls)
	if err != nil {
		return
	}

	ws.PublishMessage("/secure/battle_lobbies", server.HubKeyBattleLobbyListUpdate, resp)
}

// BroadcastLobbyInfoForQueueLockedMechs
// mechs which are in "READY" battle lobbies, will receive the new position of the battle lobbies
func BroadcastLobbyInfoForQueueLockedMechs() {
	bls, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.ReadyAt.IsNotNull(),
		boiler.BattleLobbyWhere.AssignedToBattleID.IsNull(),
		boiler.BattleLobbyWhere.EndedAt.IsNull(),
		qm.OrderBy(boiler.BattleLobbyColumns.ReadyAt+" DESC"),
		qm.Load(boiler.BattleLobbyRels.BattleLobbiesMechs),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load battle lobbies")
		return
	}

	for i, bl := range bls {
		if bl.R == nil || bl.R.BattleLobbiesMechs == nil {
			continue
		}
		mai := &server.MechArenaInfo{
			Status:                   server.MechArenaStatusQueue,
			CanDeploy:                false,
			BattleLobbyNumber:        null.IntFrom(bl.Number),
			BattleLobbyQueuePosition: null.IntFrom(i + 1),
		}
		for _, blm := range bl.R.BattleLobbiesMechs {
			ws.PublishMessage(fmt.Sprintf("/faction/%s/queue/%s", blm.FactionID, blm.MechID), server.HubKeyPlayerAssetMechQueueSubscribe, mai)
		}
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

func BroadcastBattleBountiesUpdate(battleBountyIDs ...string) {
	bbs, err := boiler.BattleBounties(
		qm.Select(
			boiler.BattleBountyColumns.ID,
			boiler.BattleBountyColumns.BattleLobbyID,
			boiler.BattleBountyColumns.TargetedMechID,
			boiler.BattleBountyColumns.Amount,
			boiler.BattleBountyColumns.OfferedByID,
			boiler.BattleBountyColumns.PayoutTXID,
			boiler.BattleBountyColumns.RefundTXID,
		),
		boiler.BattleBountyWhere.ID.IN(battleBountyIDs),
		qm.Load(boiler.BattleBountyRels.OfferedBy),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Strs("battle bounty id list", battleBountyIDs).Msg("Failed to load battle bounties")
		return
	}

	ws.PublishMessage("/secure/battle_bounties", server.HubKeyBattleBountyListUpdate, server.BattleBountiesFromBoiler(bbs))
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
