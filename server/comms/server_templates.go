package comms

import (
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/gofrs/uuid"
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
		return err
	}
	result := []*TemplateContainer{}
	for _, tpl := range templates {
		template, err := db.Template(uuid.Must(uuid.FromString(tpl.ID)))
		if err != nil {
			return err
		}

		result = append(result, ServerTemplateToApiTemplateV1(template))
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
	template, err := db.Template(req.TemplateID)
	if err != nil {
		return err
	}

	resp.TemplateContainer = ServerTemplateToApiTemplateV1(template)
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

func (s *S) TemplateRegister(req TemplateRegisterReq, resp *TemplateRegisterResp) error {
	gamelog.L.Debug().Msg("comms.TemplateRegister")

	//userResp, err := s.passportRPC.UserGet(server.UserID(req.OwnerID))
	//if err != nil {
	//	gamelog.L.Error().Err(err).Msg("Failed to get player")
	//
	//	return err
	//}
	//
	//player, err := boiler.FindPlayer(gamedb.StdConn, req.OwnerID.String())
	//if err != nil {
	//	gamelog.L.Error().Err(err).Msg("Failed to find player")
	//	return err
	//}
	//
	//player.FactionID = null.StringFrom(userResp.FactionID.String())
	//_, err = player.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerColumns.FactionID))
	//if err != nil {
	//	gamelog.L.Error().Err(err).Msg("Failed to update player")
	//	return err
	//}

	mechs, mechAnimations, mechSkins, powerCores, weapons, utilities, err := db.TemplateRegister(req.TemplateID, req.OwnerID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to register template")
		return err
	}

	var assets []*XsynAsset

	// convert into xsyn assets, maybe find a better way.... (generics? interfaces? change item schema?)
	assets = append(assets, ServerMechsToXsynAsset(mechs)...)
	assets = append(assets, ServerMechAnimationsToXsynAsset(mechAnimations)...)
	assets = append(assets, ServerMechSkinsToXsynAsset(mechSkins)...)
	assets = append(assets, ServerPowerCoresToXsynAsset(powerCores)...)
	assets = append(assets, ServerWeaponsToXsynAsset(weapons)...)
	assets = append(assets, ServerUtilitiesToXsynAsset(utilities)...)

	resp.Assets = assets
	return nil
}
