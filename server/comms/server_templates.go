package comms

import (
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"

	"github.com/gofrs/uuid"
)

type TemplatesReq struct {
}
type TemplatesResp struct {
	TemplateContainers []*server.TemplateContainer
}

// Templates is a heavy func, do not use on a running server
func (s *S) Templates(req TemplatesReq, resp *TemplatesResp) error {
	templates, err := boiler.Templates().All(gamedb.StdConn)
	if err != nil {
		return err
	}
	result := []*server.TemplateContainer{}
	for _, tpl := range templates {
		template, err := db.Template(uuid.Must(uuid.FromString(tpl.ID)))
		if err != nil {
			return err
		}
		result = append(result, template)

	}
	resp.TemplateContainers = result
	return nil
}

type TemplateReq struct {
	TemplateID uuid.UUID
}
type TemplateResp struct {
	TemplateContainer *server.TemplateContainer
}

func (s *S) Template(req TemplateReq, resp *TemplateResp) error {
	template, err := db.Template(req.TemplateID)
	if err != nil {
		return err
	}
	resp.TemplateContainer = template
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
