package comms

import (
	"server/db"
	"server/db/boiler"
	"server/gamedb"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
)

type TemplatesReq struct {
}
type TemplatesResp struct {
	TemplateContainers []*TemplateContainer
}

// Templates is a heavy func, do not use on a running server
func (s *S) Templates(req TemplatesReq, resp *TemplatesResp) error {
	templates, err := boiler.Templates().All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err)
	}
	result := []*TemplateContainer{}
	for _, tpl := range templates {
		_, err := db.Template(uuid.Must(uuid.FromString(tpl.ID)))
		if err != nil {
			return terror.Error(err)
		}

		// TODO: convert mech object

		//result = append(result, template)

	}
	resp.TemplateContainers = result
	return nil
}

type TemplateReq struct {
	TemplateID uuid.UUID
}
type TemplateResp struct {
	TemplateContainer *TemplateContainer
}

func (s *S) Template(req TemplateReq, resp *TemplateResp) error {
	_, err := db.Template(req.TemplateID)
	if err != nil {
		return terror.Error(err)
	}

	// TODO: convert mech object

	//resp.TemplateContainer = template
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
		return terror.Error(err)
	}
	resp.Count = count
	return nil
}
