package battle

import (
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/gofrs/uuid"
)

type AbilitiesSystem struct {
	battle    *Battle
	abilities map[uuid.UUID][]*GameAbility
}

func NewAbilitiesSystem(battle *Battle) *AbilitiesSystem {
	factionAbilities := map[uuid.UUID][]*GameAbility{}
	for factionID, _ := range battle.factions {
		initialAbilities, err := boiler.GameAbilities(qm.Where("faction_id = ?", factionID.String()), qm.And("battle_ability_id ISNULL")).All(gamedb.StdConn)
		factionAbilities[factionID] = []*GameAbility{}
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
			factionAbilities[factionID] = abilities
		}
	}
	return &AbilitiesSystem{battle, factionAbilities}
}
