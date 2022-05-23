package comms

import (
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpctypes"

	"github.com/gofrs/uuid"
)

type TemplatesReq struct {
}
type TemplatesResp struct {
	TemplateContainers []*rpctypes.TemplateContainer
}

// Templates is a heavy func, do not use on a running server
func (s *S) Templates(req TemplatesReq, resp *TemplatesResp) error {
	templates, err := boiler.Templates().All(gamedb.StdConn)
	if err != nil {
		return err
	}
	result := []*rpctypes.TemplateContainer{}
	for _, tpl := range templates {
		template, err := db.Template(uuid.Must(uuid.FromString(tpl.ID)))
		if err != nil {
			return err
		}

		result = append(result, rpctypes.ServerTemplateToApiTemplateV1(template))
	}
	resp.TemplateContainers = result
	return nil
}

type TemplateReq struct {
	TemplateID uuid.UUID
}
type TemplateResp struct {
	TemplateContainer *rpctypes.TemplateContainer
}

func (s *S) Template(req TemplateReq, resp *TemplateResp) error {
	template, err := db.Template(req.TemplateID)
	if err != nil {
		return err
	}

	resp.TemplateContainer = rpctypes.ServerTemplateToApiTemplateV1(template)
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

func (s *S) TemplateRegister(req rpctypes.TemplateRegisterReq, resp *rpctypes.TemplateRegisterResp) error {
	gamelog.L.Debug().Msg("comms.TemplateRegister")

	mechs, mechAnimations, mechSkins, powerCores, weapons, utilities, err := db.TemplateRegister(req.TemplateID, req.OwnerID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to register template")
		return err
	}

	var assets []*rpctypes.XsynAsset

	// TODO: HERE! add images for mechs and weapons
	// convert into xsyn assets, maybe find a better way.... (generics? interfaces? change item schema?)
	assets = append(assets, rpctypes.ServerMechsToXsynAsset(mechs)...)
	assets = append(assets, rpctypes.ServerMechAnimationsToXsynAsset(mechAnimations)...)
	assets = append(assets, rpctypes.ServerMechSkinsToXsynAsset(mechSkins)...)
	assets = append(assets, rpctypes.ServerPowerCoresToXsynAsset(powerCores)...)
	assets = append(assets, rpctypes.ServerWeaponsToXsynAsset(weapons)...)
	assets = append(assets, rpctypes.ServerUtilitiesToXsynAsset(utilities)...)

	resp.Assets = assets
	return nil
}
