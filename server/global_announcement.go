package server

type Severity string

const (
	SeveritySuccess Severity = "SUCCESS"
	SeverityWarning Severity = "WARNING"
	SeverityInfo    Severity = "INFO"
	SeverityDanger  Severity = "DANGER"
)

type GlobalAnnouncement struct {
	ID                    string   `json:"id"`
	Title                 string   `json:"title"`
	Severity              Severity `json:"severity"`
	Message               string   `json:"message"`
	ShowFromBattleNumber  *int     `json:"show_from_battle_number,omitempty"`  // the battle number this announcement wil show
	ShowUntilBattleNumber *int     `json:"show_until_battle_number,omitempty"` // the battle number this announcement will be deleted
}

func (s Severity) IsValid() bool {
	switch s {
	case
		SeveritySuccess,
		SeverityWarning,
		SeverityInfo,
		SeverityDanger:
		return true
	}
	return false
}
