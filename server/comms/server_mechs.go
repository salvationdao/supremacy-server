package comms

import (
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
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
	fmt.Println("s.Mechs")

	templates, err := boiler.Mechs().All(gamedb.StdConn)
	if err != nil {
		return err
	}
	result := []*server.MechContainer{}
	for _, tpl := range templates {
		template, err := db.Mech(uuid.Must(uuid.FromString(tpl.ID)))
		if err != nil {
			return err
		}
		result = append(result, template)

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
	fmt.Println("s.Mech")
	result, err := db.Mech(req.MechID)
	if err != nil {
		return err
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
	fmt.Println("s.MechsByOwnerID")
	result, err := db.MechsByOwnerID(req.OwnerID)
	if err != nil {
		return err
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
	fmt.Println("s.MechRegister")
	userResp := &UserResp{}
	err := s.C.Call("S.User", &UserReq{ID: req.OwnerID}, userResp)
	if err != nil {
		return fmt.Errorf("refresh player: %w", err)
	}

	player, err := boiler.FindPlayer(gamedb.StdConn, req.OwnerID.String())
	if err != nil {
		return fmt.Errorf("get player: %w", err)
	}
	player.FactionID = null.StringFrom(userResp.FactionID.String)
	_, err = player.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerColumns.FactionID))
	if err != nil {
		return fmt.Errorf("update player: %w", err)
	}
	spew.Dump(player)
	mechID, err := db.MechRegister(req.TemplateID, req.OwnerID)
	if err != nil {
		return fmt.Errorf("mech register: %w", err)
	}
	mech, err := db.Mech(mechID)
	if err != nil {
		return fmt.Errorf("get created mech: %w", err)
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
	fmt.Println("s.MechSetName")
	err := db.MechSetName(req.MechID, req.Name)
	if err != nil {
		return err
	}
	mech, err := db.Mech(req.MechID)
	if err != nil {
		return err
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
	fmt.Println("s.MechSetOwner")
	err := db.MechSetOwner(req.MechID, req.OwnerID)
	if err != nil {
		return err
	}
	mech, err := db.Mech(req.MechID)
	if err != nil {
		return err
	}
	resp.MechContainer = mech
	return nil
}
