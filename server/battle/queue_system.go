package battle

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
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

func (am *ArenaManager) BroadcastBattleLobbyUpdate(battleLobbyID string) {
	bl, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.ID.EQ(battleLobbyID),
		qm.Load(
			boiler.BattleLobbyRels.BattleLobbiesMechs,
			boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
			boiler.BattleLobbiesMechWhere.DeletedAt.IsNull(),
		),
		qm.Load(boiler.BattleLobbyRels.HostBy),
		qm.Load(boiler.BattleLobbyRels.GameMap),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("battle lobby id", battleLobbyID).Msg("Failed to query battle lobby")
		return
	}

	if bl == nil {
		return
	}

	ws.PublishMessage("/secure/battle_lobbies", server.HubKeyBattleLobbyListUpdate, server.BattleLobbiesFromBoiler([]*boiler.BattleLobby{bl}))
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

	// load game maps
	gameMaps, err := boiler.GameMaps().All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load game map.")
		return terror.Error(err, "Failed to load game map.")
	}

	// fill up battle lobbies
	for i := 0; i < publicLobbiesCount-count; i++ {
		gameMap := gameMaps[i%len(gameMaps)]

		bl := &boiler.BattleLobby{
			HostByID:              server.SupremacyBattleUserID,
			EntryFee:              decimal.Zero, // free to join
			FirstFactionCut:       decimal.NewFromInt(75),
			SecondFactionCut:      decimal.NewFromInt(25),
			ThirdFactionCut:       decimal.NewFromInt(0),
			EachFactionMechAmount: 3,
			GameMapID:             gameMap.ID,
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

// CheckMechOwnership return error if any mechs is not available to queue
func (am *ArenaManager) CheckMechOwnership(playerID string, mechIDs []string) error {
	// ownership check
	mcis, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.IN(mechIDs),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Strs("mech ids", mechIDs).Err(err).Msg("unable to retrieve mech collection item from hash")
		return err
	}

	if len(mcis) != len(mechIDs) {
		return terror.Error(fmt.Errorf("contain non-mech assest"), "The list contains non-mech asset.")
	}

	for _, mci := range mcis {
		if mci.XsynLocked {
			err := fmt.Errorf("mech is locked to xsyn locked")
			gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mci.ItemID).Err(err).Msg("war machine is xsyn locked")
			return err
		}

		if mci.LockedToMarketplace {
			err := fmt.Errorf("mech is listed in marketplace")
			gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mci.ItemID).Err(err).Msg("war machine is listed in marketplace")
			return err
		}

		battleReady, err := db.MechBattleReady(mci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to load battle ready status")
			return err
		}

		if !battleReady {
			gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mci.ItemID).Msg("war machine is not available for queuing")
			return fmt.Errorf("mech is cannot be used")
		}

		if mci.OwnerID != playerID {
			return terror.Error(fmt.Errorf("does not own the mech"), "This mech is not owned by you")
		}
	}

	return nil
}

// FilterCanDeployMechIDs return the list of mech which are able to deploy
func (am *ArenaManager) FilterCanDeployMechIDs(mechIDs []string) ([]string, error) {
	// check mech is still in repair
	rcs, err := boiler.RepairCases(
		boiler.RepairCaseWhere.MechID.IN(mechIDs),
		boiler.RepairCaseWhere.CompletedAt.IsNull(),
		qm.Load(boiler.RepairCaseRels.RepairOffers, boiler.RepairOfferWhere.ClosedAt.IsNull()),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Strs("mech ids", mechIDs).Msg("Failed to get repair case")
		return nil, terror.Error(err, "Failed to queue mech.")
	}

	if rcs == nil {
		return mechIDs, nil
	}

	canDeployRatio := db.GetDecimalWithDefault(db.KeyCanDeployDamagedRatio, decimal.NewFromFloat(0.5))

	canDeployedMechIDs := []string{}
	for _, mechID := range mechIDs {
		index := slices.IndexFunc(rcs, func(rc *boiler.RepairCase) bool { return rc.MechID == mechID })

		// if the mech does not have repair case
		if index == -1 {
			canDeployedMechIDs = append(canDeployedMechIDs, mechID)
			continue
		}

		// append mech id, if the damaged ratio of the mech is not above the can deploy ratio
		rc := rcs[index]
		if decimal.NewFromInt(int64(rc.BlocksRequiredRepair - rc.BlocksRepaired)).Div(decimal.NewFromInt(int64(db.TotalRepairBlocks(rc.MechID)))).LessThanOrEqual(canDeployRatio) {
			canDeployedMechIDs = append(canDeployedMechIDs, mechID)
			continue
		}
	}

	return canDeployedMechIDs, nil
}

// FilterOutMechAlreadyInQueue return error if mech is already in queue
func (am *ArenaManager) FilterOutMechAlreadyInQueue(mechIDs []string) ([]string, error) {
	// check any mechs is already queued
	blm, err := boiler.BattleLobbiesMechs(
		boiler.BattleLobbiesMechWhere.MechID.IN(mechIDs),
		boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
		qm.Where(
			fmt.Sprintf(
				"EXISTS ( SELECT 1 FROM %s WHERE %s = %s AND %s ISNULL)",
				boiler.TableNames.BattleLobbies,
				boiler.BattleLobbyTableColumns.ID,
				boiler.BattleLobbiesMechTableColumns.BattleLobbyID,
				boiler.BattleLobbyTableColumns.FinishedAt,
			),
		),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Strs("mech id list", mechIDs).Msg("Failed to check mech queue.")
		return nil, terror.Error(err, "Failed to check mech queue.")
	}

	if blm == nil {
		return mechIDs, nil
	}

	remainMechIDs := []string{}
	for _, mechID := range mechIDs {
		// skip, if mech is already queued
		if slices.IndexFunc(blm, func(bl *boiler.BattleLobbiesMech) bool { return bl.MechID == mechID }) != -1 {
			continue
		}

		// append mechs to the list
		remainMechIDs = append(remainMechIDs, mechID)
	}

	return remainMechIDs, nil
}
