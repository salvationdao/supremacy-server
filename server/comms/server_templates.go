package comms

import (
	"server/db"
	"server/gamelog"
	"server/rpctypes"
)

func (s *S) TemplateRegisterHandler(req rpctypes.TemplateRegisterReq, resp *rpctypes.TemplateRegisterResp) error {
	gamelog.L.Debug().Msg("comms.TemplateRegisterHandler")

	mechs, mechAnimations, mechSkins, powerCores, weapons, utilities, err := db.TemplateRegister(req.TemplateID, req.OwnerID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to register template")
		return err
	}

	var assets []*rpctypes.XsynAsset

	assets = append(assets, rpctypes.ServerMechsToXsynAsset(mechs)...)
	assets = append(assets, rpctypes.ServerMechAnimationsToXsynAsset(mechAnimations)...)
	assets = append(assets, rpctypes.ServerMechSkinsToXsynAsset(mechSkins)...)
	assets = append(assets, rpctypes.ServerPowerCoresToXsynAsset(powerCores)...)
	assets = append(assets, rpctypes.ServerWeaponsToXsynAsset(weapons)...)
	assets = append(assets, rpctypes.ServerUtilitiesToXsynAsset(utilities)...)

	resp.Assets = assets
	return nil
}
