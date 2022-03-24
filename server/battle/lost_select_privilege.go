package battle

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/sasha-s/go-deadlock"
)

type LostSelectionPrivilege struct {
	deadlock.RWMutex
	InstanceMap map[uuid.UUID]*LSPInstance
}

type BanReason string

const (
	BanReasonTeamKill BanReason = "TEAM_KILL"
)

type LSPInstance struct {
	FactionID           uuid.UUID
	IntendToBenPlayerID uuid.UUID
	Reason              BanReason
	EndedAt             time.Time
	VoteChan            chan *LSPVote
}

type LSPVote struct {
	playerID uuid.UUID
	isAgreed bool
}

func NewLostSelectPrivilege() *LostSelectionPrivilege {
	lsp := &LostSelectionPrivilege{
		InstanceMap: make(map[uuid.UUID]*LSPInstance),
	}

	return lsp
}

// NewInstance add another instance
func (lsp *LostSelectionPrivilege) NewInstance(factionID uuid.UUID, intendBanPlayerID uuid.UUID, endedAt time.Time, reason BanReason) *LSPInstance {
	// new instance uuid
	lspUUID := uuid.Must(uuid.NewV4())

	instance := &LSPInstance{
		FactionID:           factionID,
		IntendToBenPlayerID: intendBanPlayerID,
		Reason:              reason,
		EndedAt:             endedAt,
		VoteChan:            make(chan *LSPVote),
	}

	lsp.InstanceMap[lspUUID] = instance

	go instance.Start()

	return instance
}

func (lspi *LSPInstance) Start() {
	agreePoint := int64(0)
	disagreePoint := int64(0)
	votedPlayer := make(map[uuid.UUID]bool)

	timer := time.NewTimer(lspi.EndedAt.Sub(time.Now()))

	for {
		select {
		case <-timer.C:
			// ban player if agree point higher than disagree point

			return
		case vote := <-lspi.VoteChan:
			// reject vote if time expired
			if time.Now().After(lspi.EndedAt) {
				continue
			}

			// skip, if the player has voted already
			if _, ok := votedPlayer[vote.playerID]; ok || lspi.IntendToBenPlayerID == vote.playerID {
				continue
			}

			// get user ability killed count

			if vote.isAgreed {
				agreePoint += 1
			} else {
				disagreePoint -= 1
			}

			votedPlayer[vote.playerID] = true
		}
	}
}
