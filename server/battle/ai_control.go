package battle

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"golang.org/x/exp/slices"
	"math/rand"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"
)

func (btl *Battle) AIControl() {
	// load AI players
	ps, err := boiler.Players(
		boiler.PlayerWhere.IsAi.EQ(true),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load AI players")
		return
	}

	for _, wm := range btl.WarMachines {
		// skip, if the player is not AI
		index := slices.IndexFunc(ps, func(p *boiler.Player) bool { return p.ID == wm.OwnedByID })
		if index == -1 {
			continue
		}

		go func(battle *Battle, warMachine *WarMachine, player *boiler.Player) {
			commandTimer := time.NewTimer(120 * time.Second)
			<-commandTimer.C

			commandTimer.Reset(1 * time.Second)

			for {
				<-commandTimer.C

				// exit if battle end
				if battle.stage.Load() != BattleStageStart {
					return
				}

				// check whether mech is still alive
				warMachine.RLock()
				health := warMachine.Health
				selfX := decimal.NewFromInt(int64(warMachine.Position.X))
				selfY := decimal.NewFromInt(int64(warMachine.Position.Y))
				warMachine.RUnlock()
				if health <= 0 {
					return
				}

				var opponentPositions []struct {
					x decimal.Decimal
					y decimal.Decimal
				}

				for _, wwm := range battle.WarMachines {
					if wwm.FactionID == warMachine.FactionID || wwm.Health <= 0 {
						continue
					}

					wwm.RLock()
					opponentPositions = append(opponentPositions, struct {
						x decimal.Decimal
						y decimal.Decimal
					}{
						x: decimal.NewFromInt(int64(wwm.Position.X)),
						y: decimal.NewFromInt(int64(wwm.Position.Y)),
					})
					wwm.RUnlock()
				}

				closestDistant := decimal.NewFromInt(-1)
				closestPosition := struct {
					x decimal.Decimal
					y decimal.Decimal
				}{
					x: decimal.Zero,
					y: decimal.Zero,
				}
				for _, op := range opponentPositions {
					newDistant := op.x.Sub(selfX).Pow(decimal.NewFromInt(2)).Add(op.y.Sub(selfY).Pow(decimal.NewFromInt(2)))
					if closestDistant.Equals(decimal.NewFromInt(-1)) || closestDistant.GreaterThan(newDistant) {
						closestDistant = newDistant
						closestPosition.x = op.x
						closestPosition.y = op.y
						continue
					}
				}

				if closestDistant.GreaterThanOrEqual(decimal.Zero) {
					func() {
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
								ID:            btl.ID + warMachine.Hash,
								ArenaID:       battle.ArenaID,
								MechID:        warMachine.ID,
								TriggeredByID: player.ID,
								CellX:         int(cellXY.X.IntPart()),
								CellY:         int(cellXY.Y.IntPart()),
								BattleID:      battle.ID,
								CreatedAt:     time.Now(),
								IsMoving:      true,
							},
						}

						fmc := &FactionMechCommand{
							ID:         btl.ID + warMachine.Hash,
							BattleID:   btl.ID,
							IsEnded:    false,
							IsMiniMech: false,
							CellX:      int(cellXY.X.IntPart()),
							CellY:      int(cellXY.Y.IntPart()),
						}

						ws.PublishMessage(fmt.Sprintf("/mini_map/arena/%s/faction/%s/mech_command/%s", battle.ArenaID, warMachine.FactionID, warMachine.Hash), server.HubKeyMechCommandUpdateSubscribe, mmc)
						ws.PublishMessage(fmt.Sprintf("/mini_map/arena/%s/faction/%s/mech_commands", battle.ArenaID, warMachine.FactionID), server.HubKeyFactionMechCommandUpdateSubscribe, []*FactionMechCommand{fmc})
					}()
				}

				commandTimer.Reset(time.Duration(rand.Intn(5)+1) * time.Second)
			}

		}(btl, wm, ps[index])
	}
}
