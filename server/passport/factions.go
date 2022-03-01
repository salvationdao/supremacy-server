package passport

import (
	"server"

	"github.com/ninja-syndicate/hub"
)

type FactionAllReq struct{}

type FactionAllResp struct {
	Factions []*server.Faction `json:"factions"`
}

// FactionAll get all the factions from passport server
func (pp *Passport) FactionAll(callback func(factions []*server.Faction)) {
	resp := &FactionAllResp{}
	err := pp.Comms.Call("C.SupremacyFactionAllHandler", FactionAllReq{}, resp)
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyFactionAllHandler").Msg("rpc error")
	}
	callback(resp.Factions)
}

type FactionStatSendReq struct {
	FactionStatSends []*FactionStatSend `json:"factionStatSends"`
}

type FactionStatSend struct {
	FactionStat     *server.FactionStat `json:"factionStat"`
	ToUserID        *server.UserID      `json:"toUserID,omitempty"`
	ToUserSessionID *hub.SessionID      `json:"toUserSessionID,omitempty"`
}

type FactionStatSendResp struct{}

// FactionStatsSend send faction stat to passport serer
func (pp *Passport) FactionStatsSend(factionStatSends []*FactionStatSend) {
	err := pp.Comms.Call("C.SupremacyFactionStatSendHandler", FactionStatSendReq{factionStatSends}, &FactionStatSendResp{})
	if err != nil {
		pp.Log.Err(err).Str("method", "SupremacyFactionStatSendHandler").Msg("rpc error")
	}
}
