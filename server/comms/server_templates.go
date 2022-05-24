package comms

import (
	"server/db"
	"server/gamelog"
	"server/rpctypes"
)

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
