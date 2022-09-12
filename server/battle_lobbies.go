package server

import (
	"github.com/volatiletech/null/v8"
	"server/db/boiler"
)

type BattleLobby struct {
	*boiler.BattleLobby
	HostBy    *boiler.Player              `json:"host_by"`
	GameMap   *boiler.GameMap             `json:"game_map"`
	Mechs     []*boiler.BattleLobbiesMech `json:"mechs"`
	IsPrivate bool                        `json:"is_private"`
}

func BattleLobbiesFromBoiler(bls []*boiler.BattleLobby) []*BattleLobby {
	resp := []*BattleLobby{}

	for _, bl := range bls {
		copiedBattleLobby := *bl
		sbl := &BattleLobby{
			BattleLobby: &copiedBattleLobby,
			Mechs:       []*boiler.BattleLobbiesMech{},
			IsPrivate:   copiedBattleLobby.Password.Valid,
		}

		// omit password
		sbl.Password = null.StringFromPtr(nil)

		if bl.R != nil {
			for _, blm := range bl.R.BattleLobbiesMechs {
				sbl.Mechs = append(sbl.Mechs, blm)
			}

			if bl.R.HostBy != nil {
				host := bl.R.HostBy
				// trim info
				sbl.HostBy = &boiler.Player{
					ID:        host.ID,
					Username:  host.Username,
					FactionID: host.FactionID,
					Gid:       host.Gid,
					Rank:      host.Rank,
				}
			}

			if bl.R.GameMap != nil {
				sbl.GameMap = bl.R.GameMap
			}
		}

		resp = append(resp, sbl)
	}

	return resp
}
