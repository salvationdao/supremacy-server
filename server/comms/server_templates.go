package comms

import (
	"fmt"
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

	var mechIDs []string
	for _, m := range mechs {
		mechIDs = append(mechIDs, m.ID)
	}

	loadedMechs, err := db.Mechs(mechIDs...)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed loading mechs")
		return err
	}

	for _, m := range loadedMechs {
		m.CheckAndSetAsGenesisOrLimited()

		fmt.Println(m.CollectionSlug)
	}

	assets = append(assets, rpctypes.ServerMechsToXsynAsset(loadedMechs)...)
	if loadedMechs != nil && !loadedMechs[0].GenesisTokenID.Valid && !loadedMechs[0].LimitedReleaseTokenID.Valid {
		assets = append(assets, rpctypes.ServerMechAnimationsToXsynAsset(mechAnimations)...)
		assets = append(assets, rpctypes.ServerMechSkinsToXsynAsset(mechSkins)...)
		assets = append(assets, rpctypes.ServerPowerCoresToXsynAsset(powerCores)...)
		assets = append(assets, rpctypes.ServerWeaponsToXsynAsset(weapons)...)
		assets = append(assets, rpctypes.ServerUtilitiesToXsynAsset(utilities)...)
	}

	resp.Assets = assets
	return nil
}
