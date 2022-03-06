package comms

import (
	"server"
	"server/db"

	"github.com/gofrs/uuid"
)

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
