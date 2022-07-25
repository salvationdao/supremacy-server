package battle

import (
	"server/db/boiler"
	"server/xsyn_rpcclient"
	"sync"
)

type RepairSystem struct {
	passport *xsyn_rpcclient.XsynXrpcClient
	caseMap  map[string]*RepairCase
	sync.RWMutex
}

func New(pp *xsyn_rpcclient.XsynXrpcClient) *RepairSystem {
	rs := &RepairSystem{
		passport: pp,
		caseMap:  make(map[string]*RepairCase),
	}

	return rs
}

// register repair case
func (rs *RepairSystem) RegisterMechRepairCase(mechID string, maxHealth string, remainHealth string) {
	if remainHealth == maxHealth {
		return
	}

	// get full instant pay price from kv

	// calculate instant pay

}

type RepairCase struct {
	*boiler.RepairOffer
	isClosed bool // protected by channel

	repairAgentRegisterChan chan *boiler.Player
	repairAgentCancelChan   chan *boiler.Player
	repairAgentCompleteChan chan string
	instantRepairChan       chan bool
	onClose                 func()
}

func (rc *RepairCase) start() {
	defer rc.onClose()

	for {
		if rc.isClosed {
			return
		}

		select {
		case <-rc.instantRepairChan:
			// calculate remain time

			rc.isClosed = true
		}

	}
}
