package comms

import (
	"server"
	"server/db"

	"github.com/gofrs/uuid"
)

type MechReq struct {
	MechID uuid.UUID
}

type MechResp struct {
	Mech *server.Mech
}

func (s *S) Mech(req MechReq, resp *MechResp) error {
	result, err := db.Mech(req.MechID)
	if err != nil {
		return err
	}
	resp.Mech = result
	return nil
}

type MechsByOwnerIDReq struct {
	OwnerID uuid.UUID
}
type MechsByOwnerIDResp struct {
	Mechs []*server.Mech
}

func (s *S) MechsByOwnerID(req MechsByOwnerIDReq, resp *MechsByOwnerIDResp) error {
	result, err := db.MechsByOwnerID(req.OwnerID)
	if err != nil {
		return err
	}
	resp.Mechs = result
	return nil
}

type MechRegisterReq struct {
	TemplateID uuid.UUID
	OwnerID    uuid.UUID
}
type MechRegisterResp struct {
	Mech *server.Mech
}

func (s *S) MechRegister(req MechRegisterReq, resp *MechRegisterResp) error {
	mechID, err := db.MechRegister(req.TemplateID, req.OwnerID)
	if err != nil {
		return err
	}
	mech, err := db.Mech(mechID)
	if err != nil {
		return err
	}
	resp.Mech = mech
	return nil
}

type MechSetNameReq struct {
	MechID uuid.UUID
	Name   string
}
type MechSetNameResp struct {
	Mech *server.Mech
}

func (s *S) MechSetName(req MechSetNameReq, resp *MechSetNameResp) error {
	err := db.MechSetName(req.MechID, req.Name)
	if err != nil {
		return err
	}
	mech, err := db.Mech(req.MechID)
	if err != nil {
		return err
	}
	resp.Mech = mech
	return nil
}

type MechSetOwnerReq struct {
	MechID  uuid.UUID
	OwnerID uuid.UUID
}
type MechSetOwnerResp struct {
	Mech *server.Mech
}

func (s *S) MechSetOwner(req MechSetOwnerReq, resp *MechSetOwnerResp) error {
	err := db.MechSetOwner(req.MechID, req.OwnerID)
	if err != nil {
		return err
	}
	mech, err := db.Mech(req.MechID)
	if err != nil {
		return err
	}
	resp.Mech = mech
	return nil
}
