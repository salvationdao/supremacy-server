package comms

import (
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// Mechs is a heavy func, do not use on a running server
func (s *S) Mechs(req MechsReq, resp *MechsResp) error {
	gamelog.L.Debug().Msg("comms.Mechs")
	mechs, err := boiler.Mechs().All(gamedb.StdConn)
	if err != nil {
		return err
	}

	result := []*Mech{}
	for _, mech := range mechs {
		// // Refresh player faction
		// ownerID := uuid.Must(uuid.FromString(mech.OwnerID))
		// userResp := &UserResp{}
		// err := s.C.Call("S.User", &UserReq{ID: ownerID}, userResp)
		// if err != nil {
		// 	return fmt.Errorf("refresh player: %w", err)
		// }
		// player, err := boiler.FindPlayer(gamedb.StdConn, ownerID.String())
		// if err != nil {
		// 	return fmt.Errorf("get player: %w", err)
		// }
		// player.FactionID = null.StringFrom(userResp.FactionID.String)
		// _, err = player.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerColumns.FactionID))
		// if err != nil {
		// 	return fmt.Errorf("update player: %w", err)
		// }

		// Get mech after refreshing player faction
		gamelog.L.Debug().Str("id", mech.ID).Msg("fetch mech")
		mechContainer, err := db.Mech(uuid.Must(uuid.FromString(mech.ID)))
		if err != nil {
			return terror.Error(err)
		}
		if mechContainer.ID == "" || mechContainer.ID == uuid.Nil.String() {
			return terror.Error(fmt.Errorf("null ID"))
		}

		if mechContainer.Hash == "WQk0Qy80DJ" {
			mechContainer.TokenID = 6611
		}
		if mechContainer.Hash == "ewJ0GO0zYg" {
			mechContainer.TokenID = 6612
		}

		result = append(result, ServerMechToApiV1(mechContainer))

	}
	resp.MechContainers = result
	return nil
}

func (s *S) Mech(req MechReq, resp *MechResp) error {
	gamelog.L.Debug().Msg("comms.Mech")
	result, err := db.Mech(req.MechID)
	if err != nil {
		return terror.Error(err)
	}

	resp.MechContainer = ServerMechToApiV1(result)
	return nil
}

func (s *S) MechsByOwnerID(req MechsByOwnerIDReq, resp *MechsByOwnerIDResp) error {
	gamelog.L.Debug().Msg("comms.MechsByOwnerID")
	result, err := db.MechsByOwnerID(req.OwnerID)
	if err != nil {
		return terror.Error(err)
	}

	resp.MechContainers = ServerMechsToApiV1(result)
	return nil
}

func (s *S) TemplateRegister(req TemplateRegisterReq, resp *TemplateRegisterResp) error {
	gamelog.L.Debug().Msg("comms.TemplateRegister")

	userResp, err := s.passportRPC.UserGet(server.UserID(req.OwnerID))
	if err != nil {
		return terror.Error(err)
	}

	player, err := boiler.FindPlayer(gamedb.StdConn, req.OwnerID.String())
	if err != nil {
		return terror.Error(err)
	}

	player.FactionID = null.StringFrom(userResp.FactionID.String())
	_, err = player.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerColumns.FactionID))
	if err != nil {
		return terror.Error(err)
	}

	mechs, mechAnimations, mechSkins, powerCores, weapons, utilities, err := db.TemplateRegister(req.TemplateID, req.OwnerID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to register template")
		return terror.Error(err)
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

func (s *S) MechSetName(req MechSetNameReq, resp *MechSetNameResp) error {
	gamelog.L.Debug().Msg("comms.MechSetName")
	err := db.MechSetName(req.MechID, req.Name)
	if err != nil {
		return terror.Error(err)
	}
	mech, err := db.Mech(req.MechID)
	if err != nil {
		return terror.Error(err)
	}

	resp.MechContainer = ServerMechToApiV1(mech)
	return nil
}

func (s *S) MechSetOwner(req MechSetOwnerReq, resp *MechSetOwnerResp) error {
	gamelog.L.Debug().Msg("comms.MechSetOwner")
	err := db.MechSetOwner(req.MechID, req.OwnerID)
	if err != nil {
		return terror.Error(err)
	}
	mech, err := db.Mech(req.MechID)
	if err != nil {
		return terror.Error(err)
	}

	resp.MechContainer = ServerMechToApiV1(mech)
	return nil
}
