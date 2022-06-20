package comms

import (
	"server/db"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
)

type PlayerRegisterReq struct {
	ID            uuid.UUID
	Username      string
	FactionID     uuid.UUID
	PublicAddress common.Address
}
type PlayerRegisterResp struct {
	ID uuid.UUID
}

func (s *S) PlayerRegisterHandler(req PlayerRegisterReq, resp *PlayerRegisterResp) error {
	result, err := db.PlayerRegister(req.ID, req.Username, req.FactionID, req.PublicAddress)
	if err != nil {
		return err
	}

	resp.ID = uuid.Must(uuid.FromString(result.ID))

	return nil
}
