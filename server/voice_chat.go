package server

import "github.com/volatiletech/null/v8"

type VoiceStreamResp struct {
	ListenURL          string      `json:"listen_url,omitempty"`
	SendURL            string      `json:"send_url,omitempty"`
	IsFactionCommander bool        `json:"is_faction_commander"`
	Username           null.String `json:"username"`
	UserGID            int         `json:"user_gid"`
	CurrentKickVote    int         `json:"current_kick_vote"`
}
