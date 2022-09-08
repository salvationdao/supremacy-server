package server

import (
	"github.com/volatiletech/null/v8"
	"server/db/boiler"
)

type BattleLobby struct {
	*boiler.BattleLobby
	Mechs     []*boiler.BattleLobbiesMech `json:"mechs"`
	IsPrivate bool                        `json:"is_private"`
}

func BattleLobbiesFromBoiler(bls []*boiler.BattleLobby) []*BattleLobby {
	resp := []*BattleLobby{}

	for _, bl := range bls {
		copiedBL := *bl
		sbl := &BattleLobby{
			BattleLobby: &copiedBL,
			Mechs:       []*boiler.BattleLobbiesMech{},
			IsPrivate:   copiedBL.Password.Valid,
		}

		// omit password
		sbl.Password = null.StringFromPtr(nil)

		if bl.R != nil {
			for _, blm := range bl.R.BattleLobbiesMechs {
				sbl.Mechs = append(sbl.Mechs, blm)
			}
		}

		resp = append(resp, sbl)
	}

	return resp
}
