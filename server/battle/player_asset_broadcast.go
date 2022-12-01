package battle

import (
	"fmt"
	"github.com/ninja-syndicate/ws"
	"golang.org/x/exp/slices"
	"server"
	"server/db"
	"server/gamelog"
	"time"
)

func (am *ArenaManager) PlayerAssetsDebounceBroadcaster() {
	interval := 200 * time.Millisecond

	// player mech broadcast objects
	var broadcastMechIDs []string
	mechBroadcastTimer := time.NewTicker(interval)

	for {
		select {

		// broadcast mech detail
		case mechIDs := <-am.MechDebounceBroadcastChan:
			for _, mechID := range mechIDs {
				if slices.Index(broadcastMechIDs, mechID) != -1 {
					continue
				}

				broadcastMechIDs = append(broadcastMechIDs, mechID)
			}

			mechBroadcastTimer.Reset(interval)
		case <-mechBroadcastTimer.C:
			if len(broadcastMechIDs) == 0 {
				continue
			}

			// NOTE: wrap broadcast in goroutine, so it won't block the channel
			go debounceBroadcastMechDetail(broadcastMechIDs)

			// clean up id list
			broadcastMechIDs = []string{}
		}

		// TODO: debounce broadcast other player assets
	}
}

// debounceBroadcastMechDetail broadcast mech update
func debounceBroadcastMechDetail(mechIDs []string) {
	l := gamelog.L.With().Strs("mech id list", mechIDs).Logger()

	mechs, err := db.LobbyMechsBrief("", mechIDs...)
	if err != nil {
		l.Error().Err(err).Msg("Failed to load mech detail")
		return
	}

	// build broadcast list
	type playerMechs struct {
		playerID string
		mechs    []*db.MechBrief
	}
	var pms []*playerMechs

	type factionStakedMechs struct {
		factionID string
		mechs     []*db.MechBrief
	}
	var fms []*factionStakedMechs

	for _, mech := range mechs {
		// append player mech
		index := slices.IndexFunc(pms, func(pm *playerMechs) bool { return pm.playerID == mech.OwnerID })
		if index == -1 {
			pms = append(pms, &playerMechs{
				playerID: mech.OwnerID,
				mechs:    []*db.MechBrief{},
			})

			index = len(pms) - 1
		}
		pms[index].mechs = append(pms[index].mechs, mech)

		if mech.FactionID.Valid {
			// update mech status
			ws.PublishMessage(fmt.Sprintf("/faction/%s/queue/%s", mech.FactionID.String, mech.ID), server.HubKeyPlayerAssetMechQueueSubscribe, server.MechArenaInfo{
				Status:              mech.Status,
				CanDeploy:           mech.CanDeploy,
				BattleLobbyIsLocked: mech.LobbyLockedAt.Valid,
			})

			// build data
			if mech.IsStaked {
				index = slices.IndexFunc(fms, func(fm *factionStakedMechs) bool { return fm.factionID == mech.FactionID.String })
				if index == -1 {
					fms = append(fms, &factionStakedMechs{
						factionID: mech.FactionID.String,
						mechs:     []*db.MechBrief{},
					})

					index = len(fms) - 1
				}
				fms[index].mechs = append(fms[index].mechs, mech)
			}
		}
	}

	// start broadcasting
	for _, pm := range pms {
		ws.PublishMessage(fmt.Sprintf("/secure/user/%s/owned_mechs", pm.playerID), server.HubKeyPlayerOwnedMechs, pm.mechs)
	}

	for _, fm := range fms {
		ws.PublishMessage(fmt.Sprintf("/faction/%s/staked_mechs", fm.factionID), server.HubKeyFactionStakedMechs, fm.mechs)
	}

	// free up memory
	pms = nil
	fms = nil
}
