package comms

import (
	"server/db"
	"server/db/boiler"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
)

type PlayerRegisterReq struct {
	ID            uuid.UUID
	Username      string
	FactionID     uuid.UUID
	PublicAddress common.Address
}
type PlayerRegisterResp struct {
	Player *boiler.Player
}

func (s *S) PlayerRegister(req PlayerRegisterReq, resp *PlayerRegisterResp) error {
	result, err := db.PlayerRegister(req.ID, req.Username, req.FactionID, req.PublicAddress)
	if err != nil {
		return terror.Error(err)
	}
	resp.Player = result
	return nil
}
