package battle

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/exp/slices"
	"math/rand"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"
)

type mechStat struct {
	health decimal.Decimal
	x      decimal.Decimal
	y      decimal.Decimal
}

func (btl *Battle) AIControl() {
	// load AI players
	ps, err := boiler.Players(
		boiler.PlayerWhere.IsAi.EQ(true),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load AI players")
		return
	}

	availableOCDistance := decimal.NewFromInt(100).Pow(decimal.NewFromInt(2))

	for _, wm := range btl.WarMachines {
		// skip, if the player is not AI
		index := slices.IndexFunc(ps, func(p *boiler.Player) bool { return p.ID == wm.OwnedByID })
		if index == -1 {
			continue
		}

		wm.Lock()
		wm.isAI = true
		wm.Unlock()

		go func(battle *Battle, warMachine *WarMachine, player *boiler.Player) {
			// get cooldown timer
			abilityCooldownSeconds := db.GetIntWithDefault(db.KeyMechAbilityCoolDownSeconds, 30)
			gameAbilityOC, err := boiler.GameAbilities(
				boiler.GameAbilityWhere.FactionID.EQ(warMachine.FactionID),
				boiler.GameAbilityWhere.Label.EQ("OVERCHARGE"),
			).One(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Str("faction id", warMachine.FactionID).Msg("Failed to load overcharge ability.")
				return
			}

			hasTriggeredRepair := false
			gameAbilityRepair, err := boiler.GameAbilities(
				boiler.GameAbilityWhere.FactionID.EQ(warMachine.FactionID),
				boiler.GameAbilityWhere.Label.EQ("REPAIR"),
			).One(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Str("faction id", warMachine.FactionID).Msg("Failed to load repair ability.")
				return
			}

			approachOpponentUntilTime := time.Now().Add(time.Duration(db.GetIntWithDefault(db.KeyApproachOpponentAfterSecond, 90)) * time.Second)

			commandTimer := time.NewTimer(1 * time.Second)

			for {
				<-commandTimer.C

				approachOpponent := time.Now().After(approachOpponentUntilTime)

				// exit if battle end
				if battle.stage.Load() != BattleStageStart {
					return
				}

				// check whether mech is still alive
				warMachine.RLock()
				health := warMachine.Health
				selfX := decimal.NewFromInt(int64(warMachine.Position.X))
				selfY := decimal.NewFromInt(int64(warMachine.Position.Y))
				shouldTriggerRepair := warMachine.MaxHealth > warMachine.Health*4
				warMachine.RUnlock()
				if health <= 0 {
					return
				}

				// trigger mech repair if needed
				if shouldTriggerRepair && !hasTriggeredRepair {
					hasTriggeredRepair = triggerRepair(battle, warMachine, gameAbilityRepair)
				}

				var alliesPositions []mechStat
				var playerOpponentPositions []mechStat
				var aiOpponentPositions []mechStat

				for _, wwm := range battle.WarMachines {
					// skip self and dead mechs
					if wwm.ID == warMachine.ID || wwm.Health <= 0 {
						continue
					}

					wwm.RLock()
					ms := mechStat{
						health: decimal.NewFromInt(int64(wwm.Health)),
						x:      decimal.NewFromInt(int64(wwm.Position.X)),
						y:      decimal.NewFromInt(int64(wwm.Position.Y)),
					}
					if wwm.FactionID == warMachine.FactionID {
						alliesPositions = append(alliesPositions, ms)
					} else if wwm.isAI {
						aiOpponentPositions = append(aiOpponentPositions, ms)
					} else {
						playerOpponentPositions = append(playerOpponentPositions, ms)
					}
					wwm.RUnlock()
				}

				closestDistant := decimal.NewFromInt(-1)
				closestPosition := mechStat{
					health: decimal.Zero,
					x:      decimal.Zero,
					y:      decimal.Zero,
				}

				// stick with allies
				if !approachOpponent {
					for _, allyPosition := range alliesPositions {
						newDistant := allyPosition.x.Sub(selfX).Pow(decimal.NewFromInt(2)).Add(allyPosition.y.Sub(selfY).Pow(decimal.NewFromInt(2)))
						if closestDistant.Equals(decimal.NewFromInt(-1)) || closestDistant.GreaterThan(newDistant) {
							closestDistant = newDistant
							closestPosition.x = allyPosition.x
							closestPosition.y = allyPosition.y
							continue
						}
					}
				} else {
					// approach the nearest player opponent
					if len(playerOpponentPositions) > 0 {
						for _, playerOpponent := range playerOpponentPositions {
							newDistant := playerOpponent.x.Sub(selfX).Pow(decimal.NewFromInt(2)).Add(playerOpponent.y.Sub(selfY).Pow(decimal.NewFromInt(2)))
							if closestDistant.Equals(decimal.NewFromInt(-1)) || closestDistant.GreaterThan(newDistant) {
								closestDistant = newDistant
								closestPosition.x = playerOpponent.x
								closestPosition.y = playerOpponent.y
								continue
							}
						}
					} else {
						// otherwise, go toward the nearest AI opponent
						for _, op := range aiOpponentPositions {
							newDistant := op.x.Sub(selfX).Pow(decimal.NewFromInt(2)).Add(op.y.Sub(selfY).Pow(decimal.NewFromInt(2)))
							if closestDistant.Equals(decimal.NewFromInt(-1)) || closestDistant.GreaterThan(newDistant) {
								closestDistant = newDistant
								closestPosition.x = op.x
								closestPosition.y = op.y
								continue
							}
						}
					}
				}

				if closestDistant.GreaterThanOrEqual(decimal.Zero) {
					// trigger mech move command
					go aiMechCommand(battle, warMachine, closestPosition)

					// trigger OC if available
					if approachOpponent && availableOCDistance.LessThanOrEqual(closestDistant) {
						go triggerOC(battle, warMachine, gameAbilityOC, abilityCooldownSeconds)
					}

				}

				commandTimer.Reset(time.Duration(rand.Intn(5)+1) * time.Second)
			}

		}(btl, wm, ps[index])
	}
}

func aiMechCommand(battle *Battle, warMachine *WarMachine, closestPosition mechStat) {
	// register channel
	eventID := uuid.Must(uuid.NewV4())
	ch := make(chan bool)
	battle.arena.MechCommandCheckMap.Register(eventID.String(), ch)

	worldCoordination := &server.GameLocation{
		X: int(closestPosition.x.IntPart()),
		Y: int(closestPosition.y.IntPart()),
	}
	cellXY := battle.getCellCoordinatesFromGameWorldXY(worldCoordination)

	event := &server.GameAbilityEvent{
		IsTriggered:         true,
		GameClientAbilityID: MechMoveCommandCreateGameAbilityID, // 8
		WarMachineHash:      &warMachine.Hash,
		ParticipantID:       &warMachine.ParticipantID, // trigger on war machine
		EventID:             eventID,
		GameLocation:        worldCoordination,
	}

	// check mech command
	battle.arena.Message("BATTLE:ABILITY", event)

	// wait for the message to come back
	select {
	case isValidLocation := <-ch:
		// remove channel after complete
		battle.arena.MechCommandCheckMap.Remove(eventID.String())
		if !isValidLocation {
			gamelog.L.Error().Msg("position not available")
			return
		}
	case <-time.After(1500 * time.Millisecond):
		// remove channel after 1.5s timeout
		battle.arena.MechCommandCheckMap.Remove(eventID.String())

		gamelog.L.Error().Msg("spot check timeout")
		return
	}

	mmc := &MechMoveCommandResponse{
		MechMoveCommandLog: &boiler.MechMoveCommandLog{
			ID:            battle.ID + warMachine.Hash,
			ArenaID:       battle.ArenaID,
			MechID:        warMachine.ID,
			TriggeredByID: warMachine.OwnedByID,
			CellX:         int(cellXY.X.IntPart()),
			CellY:         int(cellXY.Y.IntPart()),
			BattleID:      battle.ID,
			CreatedAt:     time.Now(),
			IsMoving:      true,
		},
	}

	fmc := &FactionMechCommand{
		ID:         battle.ID + warMachine.Hash,
		BattleID:   battle.ID,
		IsEnded:    false,
		IsMiniMech: false,
		CellX:      int(cellXY.X.IntPart()),
		CellY:      int(cellXY.Y.IntPart()),
	}

	ws.PublishMessage(fmt.Sprintf("/mini_map/arena/%s/faction/%s/mech_command/%s", battle.ArenaID, warMachine.FactionID, warMachine.Hash), server.HubKeyMechCommandUpdateSubscribe, mmc)
	ws.PublishMessage(fmt.Sprintf("/mini_map/arena/%s/faction/%s/mech_commands", battle.ArenaID, warMachine.FactionID), server.HubKeyFactionMechCommandUpdateSubscribe, []*FactionMechCommand{fmc})

}

func triggerOC(battle *Battle, warMachine *WarMachine, gameAbilityOC *boiler.GameAbility, abilityCooldownSeconds int) {
	now := time.Now()

	// get ability from db
	lastTrigger, err := boiler.BattleAbilityTriggers(
		boiler.BattleAbilityTriggerWhere.OnMechID.EQ(null.StringFrom(warMachine.ID)),
		boiler.BattleAbilityTriggerWhere.GameAbilityID.EQ(gameAbilityOC.ID),
		boiler.BattleAbilityTriggerWhere.TriggeredAt.GT(now.Add(time.Duration(-abilityCooldownSeconds)*time.Second)),
		boiler.BattleAbilityTriggerWhere.TriggerType.EQ(boiler.AbilityTriggerTypeMECH_ABILITY),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return
	}

	// skip, if the ability is still cooling down
	if lastTrigger != nil {
		return
	}

	offeringID := uuid.Must(uuid.NewV4())

	if gameAbilityOC.DisplayOnMiniMap {
		mma := &MiniMapAbilityContent{
			OfferingID:               offeringID.String(),
			LocationSelectType:       gameAbilityOC.LocationSelectType,
			ImageUrl:                 gameAbilityOC.ImageURL,
			Colour:                   gameAbilityOC.Colour,
			MiniMapDisplayEffectType: gameAbilityOC.MiniMapDisplayEffectType,
			MechDisplayEffectType:    gameAbilityOC.MechDisplayEffectType,
			MechID:                   warMachine.ID,
		}

		ws.PublishMessage(
			fmt.Sprintf("/mini_map/arena/%s/public/mini_map_ability_display_list", battle.ArenaID),
			server.HubKeyMiniMapAbilityContentSubscribe,
			battle.MiniMapAbilityDisplayList.Add(offeringID.String(), mma),
		)

		// cancel ability after animation end
		if gameAbilityOC.AnimationDurationSeconds > 0 {
			go func() {
				time.Sleep(time.Duration(gameAbilityOC.AnimationDurationSeconds) * time.Second)
				ws.PublishMessage(
					fmt.Sprintf("/mini_map/arena/%s/public/mini_map_ability_display_list", battle.ArenaID),
					server.HubKeyMiniMapAbilityContentSubscribe,
					battle.MiniMapAbilityDisplayList.Remove(offeringID.String()),
				)
			}()
		}
	}

	// log battle ability trigger
	mat := &boiler.BattleAbilityTrigger{
		OnMechID:          null.StringFrom(warMachine.ID),
		PlayerID:          null.StringFrom(warMachine.OwnedByID),
		GameAbilityID:     gameAbilityOC.ID,
		AbilityLabel:      gameAbilityOC.Label,
		IsAllSyndicates:   false,
		FactionID:         warMachine.FactionID,
		BattleID:          battle.BattleID,
		AbilityOfferingID: offeringID.String(),
		TriggeredAt:       now,
		TriggerType:       boiler.AbilityTriggerTypeMECH_ABILITY,
	}

	err = mat.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Interface("mech ability trigger", mat).Err(err).Msg("Failed to insert mech ability trigger.")
		return
	}

	// trigger the ability
	event := &server.GameAbilityEvent{
		IsTriggered:         true,
		GameClientAbilityID: byte(gameAbilityOC.GameClientAbilityID),
		WarMachineHash:      &warMachine.Hash,
		ParticipantID:       &warMachine.ParticipantID,
		EventID:             offeringID,
		FactionID:           &warMachine.FactionID,
	}

	// fire mech command
	battle.arena.Message("BATTLE:ABILITY", event)
}

func triggerRepair(battle *Battle, warMachine *WarMachine, gameAbilityRepair *boiler.GameAbility) bool {
	// get ability from db
	lastTrigger, err := boiler.BattleAbilityTriggers(
		boiler.BattleAbilityTriggerWhere.OnMechID.EQ(null.StringFrom(warMachine.ID)),
		boiler.BattleAbilityTriggerWhere.GameAbilityID.EQ(gameAbilityRepair.ID),
		boiler.BattleAbilityTriggerWhere.BattleID.EQ(battle.ID),
		boiler.BattleAbilityTriggerWhere.TriggerType.EQ(boiler.AbilityTriggerTypeMECH_ABILITY),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false
	}

	if lastTrigger != nil {
		return true
	}

	now := time.Now()
	offeringID := uuid.Must(uuid.NewV4())

	// log mech move command
	mat := &boiler.BattleAbilityTrigger{
		OnMechID:          null.StringFrom(warMachine.ID),
		PlayerID:          null.StringFrom(warMachine.OwnedByID),
		GameAbilityID:     gameAbilityRepair.ID,
		AbilityLabel:      gameAbilityRepair.Label,
		IsAllSyndicates:   false,
		FactionID:         warMachine.FactionID,
		BattleID:          battle.ID,
		AbilityOfferingID: offeringID.String(),
		TriggeredAt:       now,
		TriggerType:       boiler.AbilityTriggerTypeMECH_ABILITY,
	}

	err = mat.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Interface("mech ability trigger", mat).Err(err).Msg("Failed to insert mech ability trigger.")
		return false
	}

	// trigger the ability
	event := &server.GameAbilityEvent{
		IsTriggered:         true,
		GameClientAbilityID: byte(gameAbilityRepair.GameClientAbilityID),
		WarMachineHash:      &warMachine.Hash,
		ParticipantID:       &warMachine.ParticipantID,
		EventID:             offeringID,
		FactionID:           &warMachine.FactionID,
	}

	// fire mech command
	battle.arena.Message("BATTLE:ABILITY", event)

	return true
}
