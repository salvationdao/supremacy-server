package comms

import (
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type MechsReq struct {
}
type MechsResp struct {
	MechContainers []*server.MechContainer
}

type UserReq struct {
	ID uuid.UUID
}
type UserResp struct {
	ID            uuid.UUID
	Username      string
	FactionID     null.String
	PublicAddress common.Address
}

// Mechs is a heavy func, do not use on a running server
func (s *S) Mechs(req MechsReq, resp *MechsResp) error {
	gamelog.L.Debug().Msg("comms.Mechs")
	mechs, err := boiler.Mechs().All(gamedb.StdConn)
	if err != nil {
		return err
	}

	result := []*server.MechContainer{}
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
			spew.Dump(mech)
			spew.Dump(mechContainer)
			return terror.Error(fmt.Errorf("null ID"))
		}

		if mechContainer.Hash == "WQk0Qy80DJ" {
			mechContainer.ExternalTokenID = 6611
		}
		if mechContainer.Hash == "ewJ0GO0zYg" {
			mechContainer.ExternalTokenID = 6612
		}

		result = append(result, mechContainer)

	}
	resp.MechContainers = result
	return nil
}

type MechReq struct {
	MechID uuid.UUID
}

type MechResp struct {
	MechContainer *server.MechContainer
}

func (s *S) Mech(req MechReq, resp *MechResp) error {
	gamelog.L.Debug().Msg("comms.Mech")
	result, err := db.Mech(req.MechID)
	if err != nil {
		return terror.Error(err)
	}
	resp.MechContainer = result
	return nil
}

type MechsByOwnerIDReq struct {
	OwnerID uuid.UUID
}
type MechsByOwnerIDResp struct {
	MechContainers []*server.MechContainer
}

func (s *S) MechsByOwnerID(req MechsByOwnerIDReq, resp *MechsByOwnerIDResp) error {
	gamelog.L.Debug().Msg("comms.MechsByOwnerID")
	result, err := db.MechsByOwnerID(req.OwnerID)
	if err != nil {
		return terror.Error(err)
	}
	resp.MechContainers = result
	return nil
}

type MechRegisterReq struct {
	TemplateID uuid.UUID
	OwnerID    uuid.UUID
}
type MechRegisterResp struct {
	MechContainer *server.MechContainer
}

func (s *S) MechRegister(req MechRegisterReq, resp *MechRegisterResp) error {
	gamelog.L.Debug().Msg("comms.MechRegister")
	userResp := &UserResp{}
	err := s.PassportRPC.Call("S.User", &UserReq{ID: req.OwnerID}, userResp)
	if err != nil {
		return terror.Error(err)
	}
	player, err := boiler.FindPlayer(gamedb.StdConn, req.OwnerID.String())
	if err != nil {
		return terror.Error(err)
	}
	player.FactionID = null.StringFrom(userResp.FactionID.String)
	_, err = player.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerColumns.FactionID))
	if err != nil {
		return terror.Error(err)
	}

	mechID, err := db.MechRegister(req.TemplateID, req.OwnerID)
	if err != nil {
		return terror.Error(err)
	}
	mech, err := db.Mech(mechID)
	if err != nil {
		return terror.Error(err)
	}

	resp.MechContainer = mech
	return nil
}

type MechSetNameReq struct {
	MechID uuid.UUID
	Name   string
}
type MechSetNameResp struct {
	MechContainer *server.MechContainer
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
	resp.MechContainer = mech
	return nil
}

type MechSetOwnerReq struct {
	MechID  uuid.UUID
	OwnerID uuid.UUID
}
type MechSetOwnerResp struct {
	MechContainer *server.MechContainer
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
	resp.MechContainer = mech
	return nil
}
