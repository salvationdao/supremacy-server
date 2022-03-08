package comms

import (
	"server/db"
	"server/db/boiler"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
)

type PlayerReq struct {
	ID            uuid.UUID
	Username      string
	FactionID     uuid.UUID
	PublicAddress common.Address
}
type PlayerResp struct {
	Player *boiler.Player
}

func (s *S) PlayerRegister(req PlayerReq, resp *PlayerResp) error {
	result, err := db.PlayerRegister(req.ID, req.Username, req.FactionID, req.PublicAddress)
	if err != nil {
		return err
	}
	resp.Player = result
	return nil
}