package server

type VoiceStreamResp struct {
	ListenURL          string `json:"listen_url,omitempty"`
	SendURL            string `json:"send_url,omitempty"`
	IsFactionCommander bool   `json:"is_faction_commander"`
}
