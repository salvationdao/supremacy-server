package comms

import (
	"server"
	"server/db"

	"github.com/gofrs/uuid"
)

type S struct {
}

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

type TemplateReq struct {
	TemplateID uuid.UUID
}
type TemplateResp struct {
	Template *server.Template
}

func (s *S) Template(req TemplateReq, resp *TemplateResp) error {
	template, err := db.Template(req.TemplateID)
	if err != nil {
		return err
	}
	resp.Template = template
	return nil
}

type TemplatePurchasedCountReq struct {
	TemplateID uuid.UUID
}
type TemplatePurchasedCountResp struct {
	Count int
}

func (s *S) TemplatePurchasedCount(req TemplatePurchasedCountReq, resp *TemplatePurchasedCountResp) error {
	count, err := db.TemplatePurchasedCount(req.TemplateID)
	if err != nil {
		return err
	}
	resp.Count = count
	return nil
}

type TemplatesByFactionIDReq struct {
	FactionID uuid.UUID
}
type TemplatesByFactionIDResp struct {
	Templates []*server.Template
}

func (s *S) TemplatesByFactionID(req TemplatesByFactionIDReq, resp *TemplatesByFactionIDResp) error {
	templates, err := db.TemplatesByFactionID(req.FactionID)
	if err != nil {
		return err
	}
	resp.Templates = templates
	return nil
}
