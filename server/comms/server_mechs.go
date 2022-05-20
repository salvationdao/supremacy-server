package comms

import (
	"fmt"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpctypes"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
)

// Mechs is a heavy func, do not use on a running server
func (s *S) Mechs(req rpctypes.MechsReq, resp *rpctypes.MechsResp) error {
	gamelog.L.Debug().Msg("comms.Mechs")
	mechs, err := boiler.Mechs().All(gamedb.StdConn)
	if err != nil {
		return err
	}

	result := []*rpctypes.Mech{}
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
		mechContainer, err := db.Mech(mech.ID)
		if err != nil {
			return err
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

		result = append(result, rpctypes.ServerMechToApiV1(mechContainer))

	}
	resp.MechContainers = result
	return nil
}

func (s *S) Mech(req rpctypes.MechReq, resp *rpctypes.MechResp) error {
	gamelog.L.Debug().Msg("comms.Mech")
	result, err := db.Mech(req.MechID.String())
	if err != nil {
		return err
	}

	resp.MechContainer = rpctypes.ServerMechToApiV1(result)
	return nil
}

func (s *S) MechsByOwnerID(req rpctypes.MechsByOwnerIDReq, resp *rpctypes.MechsByOwnerIDResp) error {
	gamelog.L.Debug().Msg("comms.MechsByOwnerID")
	result, err := db.MechsByOwnerID(req.OwnerID)
	if err != nil {
		return err
	}

	resp.MechContainers = rpctypes.ServerMechsToApiV1(result)
	return nil
}

func (s *S) MechSetName(req rpctypes.MechSetNameReq, resp *rpctypes.MechSetNameResp) error {
	gamelog.L.Debug().Msg("comms.MechSetName")
	err := db.MechSetName(req.MechID, req.Name)
	if err != nil {
		return err
	}
	mech, err := db.Mech(req.MechID.String())
	if err != nil {
		return err
	}

	resp.MechContainer = rpctypes.ServerMechToApiV1(mech)
	return nil
}

func (s *S) MechSetOwner(req rpctypes.MechSetOwnerReq, resp *rpctypes.MechSetOwnerResp) error {
	gamelog.L.Debug().Msg("comms.MechSetOwner")
	err := db.MechSetOwner(req.MechID, req.OwnerID)
	if err != nil {
		return err
	}
	mech, err := db.Mech(req.MechID.String())
	if err != nil {
		return err
	}

	resp.MechContainer = rpctypes.ServerMechToApiV1(mech)
	return nil
}
