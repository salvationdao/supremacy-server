package battle

import (
	"context"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/gofrs/uuid"

	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
)

type VotingSystem struct {
	phase  *Phase
	battle *Battle
}

type VoteState string

const (
	VoteStateHold             VoteState = "HOLD" // Waiting on signal
	VoteStateWaitMechIntro    VoteState = "WAIT_MECH_INTRO"
	VoteStateVoteCooldown     VoteState = "VOTE_COOLDOWN"
	VoteStateVoteAbilityRight VoteState = "VOTE_ABILITY_RIGHT"
	VoteStateNextVoteWin      VoteState = "NEXT_VOTE_WIN"
	VoteStateLocationSelect   VoteState = "LOCATION_SELECT"
)

type Phase struct {
	Phase   VoteState `json:"phase"`
	EndTime time.Time `json:"end_time"`
}

func NewVotingSystem(battle *Battle) *VotingSystem {
	for factionID, _ := range battle.factions {
		initialAbilities, err := boiler.GameAbilities(qm.Where("faction_id = ?", factionID.String()), qm.And("battle_ability_id ISNULL")).All(gamedb.StdConn)

		if factionID.String() == server.ZaibatsuFactionID.String() {
			if err != nil {
				gamelog.L.Error().Str("Battle ID", battle.ID.String()).Err(err).Msg("unable to retrieve game abilities")
			}

			//TODO: refactor mech abilities

			for _, ability := range initialAbilities {
				for i, wm := range battle.WarMachines {
					if wm.FactionID != factionID.String() || len(wm.Abilities) == 0 {
						continue
					}

					wmAbility := &GameAbility{
						ID:                  uuid.Must(uuid.FromString(ability.ID)), // generate a uuid for frontend to track sups contribution
						Identity:            wm.ID,
						GameClientAbilityID: byte(ability.GameClientAbilityID),
						ImageUrl:            ability.ImageURL,
						Description:         ability.Description,
						FactionID:           factionID,
						Label:               ability.Label,
						SupsCost:            ability.SupsCost,
						CurrentSups:         "0",
						WarMachineHash:      wm.Hash,
						ParticipantID:       &wm.ParticipantID,
						Title:               wm.Name,
					}
					// if it is zaibatsu faction ability set factionID back
					if ability.Label == "OVERCHARGE" {
						wmAbility.Colour = ability.Colour
						wmAbility.TextColour = ability.TextColour
					}
					//TODO: if more mech abilities shit will fail
					//mech abilities
					battle.WarMachines[i].Abilities = []*GameAbility{wmAbility}
				}
			}

		} else {
			abilities := make([]*GameAbility, len(initialAbilities))
			for i, ability := range initialAbilities {
				wmAbility := &GameAbility{
					ID:                  uuid.Must(uuid.FromString(ability.ID)), // generate a uuid for frontend to track sups contribution
					Identity:            ability.ID,
					GameClientAbilityID: byte(ability.GameClientAbilityID),
					ImageUrl:            ability.ImageURL,
					Description:         ability.Description,
					FactionID:           factionID,
					Label:               ability.Label,
					SupsCost:            ability.SupsCost,
					CurrentSups:         "0",
					Title:               "FACTION_WIDE",
				}
				abilities[i] = wmAbility
			}
		}
	}
	return &VotingSystem{
		&Phase{Phase: VoteStateHold},
		battle,
	}
}

func (v *VotingSystem) restart() {
	v.updateState(&Phase{
		Phase:   VoteStateHold,
		EndTime: time.Now().Add(1 * time.Hour),
	})
}

const HubKeyVoteStageUpdated hub.HubCommandKey = "VOTE:STAGE:UPDATED"

func (v *VotingSystem) updateState(state *Phase) {
	v.phase = state
	v.battle.arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeyVoteStageUpdated), v.phase)
}

func (v *VotingSystem) start() {
	introSeconds := len(v.battle.WarMachines)*3 + 7
	v.updateState(&Phase{
		Phase:   VoteStateWaitMechIntro,
		EndTime: time.Now().Add(time.Duration(introSeconds) * time.Second),
	})
}
