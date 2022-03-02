package battle_arena

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"server"
	"server/db"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
	"github.com/sasha-s/go-deadlock"
)

type WarMachineQueue struct {
	RedMountain *FactionQueue
	Boston      *FactionQueue
	Zaibatsu    *FactionQueue
	log         *zerolog.Logger
}

type ContractReward struct {
	deadlock.RWMutex // lock for query
	Amount           *big.Int
}

type FactionQueue struct {
	ID               server.FactionID
	deadlock.RWMutex // lock for query
	Conn             *pgxpool.Pool

	ContractReward     *ContractReward
	QueuingWarMachines []*server.WarMachineMetadata

	InGameWarMachines  []*server.WarMachineMetadata
	defaultWarMachines []*server.WarMachineMetadata
	log                *zerolog.Logger
}

func NewWarMachineQueue(factions []*server.Faction, conn *pgxpool.Pool, log *zerolog.Logger, ba *BattleArena) (*WarMachineQueue, error) {
	var err error
	wmq := &WarMachineQueue{
		RedMountain: &FactionQueue{server.RedMountainFactionID, deadlock.RWMutex{}, conn, &ContractReward{deadlock.RWMutex{}, big.NewInt(0)}, []*server.WarMachineMetadata{}, []*server.WarMachineMetadata{}, []*server.WarMachineMetadata{}, log_helpers.NamedLogger(log, "Red Mountain queue")},
		Boston:      &FactionQueue{server.BostonCyberneticsFactionID, deadlock.RWMutex{}, conn, &ContractReward{deadlock.RWMutex{}, big.NewInt(0)}, []*server.WarMachineMetadata{}, []*server.WarMachineMetadata{}, []*server.WarMachineMetadata{}, log_helpers.NamedLogger(log, "Boston queue")},
		Zaibatsu:    &FactionQueue{server.ZaibatsuFactionID, deadlock.RWMutex{}, conn, &ContractReward{deadlock.RWMutex{}, big.NewInt(0)}, []*server.WarMachineMetadata{}, []*server.WarMachineMetadata{}, []*server.WarMachineMetadata{}, log_helpers.NamedLogger(log, "Zaibatsu queue")},
		log:         log_helpers.NamedLogger(log, "war machine queue"),
	}

	for _, faction := range factions {
		switch faction.ID {

		// initialise Red Mountain war machine queue
		case server.RedMountainFactionID:
			wmq.RedMountain.defaultWarMachines = ba.DefaultWarMachinesGet(faction.ID)
			err = wmq.RedMountain.Init(faction)
			if err != nil {
				return nil, terror.Error(err)
			}

			// initialise Boston war machine queue
		case server.BostonCyberneticsFactionID:
			wmq.Boston.defaultWarMachines = ba.DefaultWarMachinesGet(faction.ID)
			err = wmq.Boston.Init(faction)
			if err != nil {
				return nil, terror.Error(err)
			}

			// initialise Zaibatsu war machine queue
		case server.ZaibatsuFactionID:
			wmq.Zaibatsu.defaultWarMachines = ba.DefaultWarMachinesGet(faction.ID)
			err = wmq.Zaibatsu.Init(faction)
			if err != nil {
				return nil, terror.Error(err)
			}
		default:
			return nil, terror.Error(fmt.Errorf("faction switch fallthrough: %s", faction.ID))
		}
	}

	return wmq, nil
}

//
func (ba *BattleArena) DefaultWarMachinesGet(factionID server.FactionID) []*server.WarMachineMetadata {
	warMachines := []*server.WarMachineMetadata{}
	// add default war machine to meet the total amount
	wg := deadlock.WaitGroup{}
	wg.Add(1)
	ba.passport.GetDefaultWarMachines(context.Background(), factionID, func(wms []*server.WarMachineMetadata) {
		defer wg.Done()
		warMachines = append(warMachines, wms...)
	})
	wg.Wait()
	return warMachines
}

var RedMountainFaction = &server.Faction{
	ID:    server.RedMountainFactionID,
	Label: "Red Mountain Offworld Mining Corporation",
	Theme: &server.FactionTheme{
		Primary:    "#C24242",
		Secondary:  "#FFFFFF",
		Background: "#120E0E",
	},
}

var BostonFaction = &server.Faction{
	ID:    server.BostonCyberneticsFactionID,
	Label: "Boston Cybernetics",
	Theme: &server.FactionTheme{
		Primary:    "#428EC1",
		Secondary:  "#FFFFFF",
		Background: "#080C12",
	},
}

var ZaibatsuFaction = &server.Faction{
	ID:    server.ZaibatsuFactionID,
	Label: "Zaibatsu Heavy Industries",
	Theme: &server.FactionTheme{
		Primary:    "#FFFFFF",
		Secondary:  "#000000",
		Background: "#0D0D0D",
	},
}

// Init read war machine list from db and set up the list
func (fq *FactionQueue) Init(faction *server.Faction) error {
	// read war machine queue from db
	wms, err := db.BattleQueueGetByFactionID(context.Background(), fq.Conn, faction.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return terror.Error(err, "failed to read battle queue list from db")
	}

	if wms == nil {
		wms = []*server.WarMachineMetadata{}
	}
	// set up contract reward
	crStr, err := db.FactionContractRewardGet(context.Background(), fq.Conn, faction.ID)
	if err != nil {
		return terror.Error(err, "failed to get contract reward")
	}

	contractReward := big.NewInt(0)
	cr, ok := contractReward.SetString(crStr, 10)
	if !ok {
		return terror.Error(fmt.Errorf("Failed to convert contract reward to big int"))
	}

	// chuck war machines into list
	fq.QueuingWarMachines = wms

	// set up war machines' faction detail
	for _, wm := range fq.QueuingWarMachines {
		wm.Faction = faction
	}

	// set up faction contract reward
	fq.ContractReward.Amount.Add(fq.ContractReward.Amount, cr)

	return nil
}

// UpdateContractReward update contract reward when battle end
func (fq *FactionQueue) UpdateContractReward(winningFactionID server.FactionID) error {
	fq.ContractReward.Lock()
	defer fq.ContractReward.Unlock()
	if winningFactionID == fq.ID {
		// decrease 2.5% if win a battle
		fq.ContractReward.Amount.Mul(fq.ContractReward.Amount, big.NewInt(975))
		fq.ContractReward.Amount.Div(fq.ContractReward.Amount, big.NewInt(1000))
	} else {
		// increase 2.5% if loss a battle
		fq.ContractReward.Amount.Mul(fq.ContractReward.Amount, big.NewInt(1025))
		fq.ContractReward.Amount.Div(fq.ContractReward.Amount, big.NewInt(1000))
	}

	// store contract reward into
	err := db.FactionContractRewardUpdate(context.Background(), fq.Conn, fq.ID, fq.ContractReward.Amount.String())
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// GetContractReward return contract reward for current faction queue
func (fq *FactionQueue) GetContractReward() string {
	fq.ContractReward.RLock()
	defer fq.ContractReward.RUnlock()
	return fq.ContractReward.Amount.String()
}

// return the length of current queuing list
func (fq *FactionQueue) QueuingLength() int {
	return len(fq.QueuingWarMachines)
}

// Join check war machines' faction and join them into their faction queue
func (wmq *WarMachineQueue) Join(wmm *server.WarMachineMetadata, isInsured bool) error {
	// check faction id
	switch wmm.FactionID {
	case server.RedMountainFactionID:
		err := wmq.RedMountain.Join(wmm, isInsured, RedMountainFaction)
		if err != nil {
			return terror.Error(err)
		}
		return nil
	case server.BostonCyberneticsFactionID:
		err := wmq.Boston.Join(wmm, isInsured, BostonFaction)
		if err != nil {
			return terror.Error(err)
		}
		return nil
	case server.ZaibatsuFactionID:
		err := wmq.Zaibatsu.Join(wmm, isInsured, ZaibatsuFaction)
		if err != nil {
			return terror.Error(err)
		}
		return nil
	default:
		return terror.Error(fmt.Errorf("No faction war machine"), "NON-FACTION WAR MACHINE IS NOT ALLOWED!!!!!!!!!!!!!!!!!!!")
	}
}

func (fq *FactionQueue) Join(wmm *server.WarMachineMetadata, isInsured bool, faction *server.Faction) error {
	// reject queue if already in the queue
	if index := checkWarMachineExist(fq.QueuingWarMachines, wmm.Hash); index != -1 {
		return terror.Error(fmt.Errorf("war machine is already in the queue"), "war machine "+wmm.Hash+" is already in queue")
	}

	// reject if already in the game
	if index := checkWarMachineExist(fq.InGameWarMachines, wmm.Hash); index != -1 {
		return terror.Error(fmt.Errorf("war machine is currently in game"), "war machine "+wmm.Hash+" is currently in game")
	}

	contractReward := fq.ContractReward.Amount.String()

	// insert war machine into db
	err := db.BattleQueueInsert(context.Background(), fq.Conn, wmm, contractReward, isInsured)
	if err != nil {
		return terror.Error(err, "Failed to insert a copy of queue in db, token id:"+wmm.Hash)
	}

	// join war machine to queue
	fq.Lock()
	wmm.Faction = faction
	wmm.ContractReward = &contractReward
	fq.QueuingWarMachines = append(fq.QueuingWarMachines, wmm)
	fq.Unlock()

	return nil
}

func (fq *FactionQueue) GetWarMachineForEnterGame(desireAmount int) []*server.WarMachineMetadata {
	newList := []*server.WarMachineMetadata{}
	fq.Lock()
	defer fq.Unlock()

	if len(fq.QueuingWarMachines) < desireAmount {
		newList = append(newList, fq.QueuingWarMachines...)

		// fill mech with
		newList = append(newList, fq.defaultWarMachines[:desireAmount-len(newList)]...)

		// clean up the queuing list
		fq.QueuingWarMachines = []*server.WarMachineMetadata{}

		// set the in game war machine list
		fq.InGameWarMachines = newList
		return newList
	}

	newList = append(newList, fq.QueuingWarMachines[:desireAmount]...)

	// set the in game war machine list
	fq.InGameWarMachines = newList

	// clear up the war machine list
	fq.QueuingWarMachines = fq.QueuingWarMachines[desireAmount:]

	return newList
}

// checkWarMachineExist return true if war machine already exist in the list
func checkWarMachineExist(list []*server.WarMachineMetadata, hash string) int {
	for i, wm := range list {
		if wm.Hash == hash {
			return i
		}
	}
	return -1
}

func (wmq *WarMachineQueue) GetWarMachineQueue(factionID server.FactionID, hash string) (*int, *string) {
	// check faction id
	switch factionID {
	case server.RedMountainFactionID:
		return wmq.RedMountain.WarMachineQueuePosition(hash)
	case server.BostonCyberneticsFactionID:
		return wmq.Boston.WarMachineQueuePosition(hash)
	case server.ZaibatsuFactionID:
		return wmq.Zaibatsu.WarMachineQueuePosition(hash)
	}

	return nil, nil
}

func (fq *FactionQueue) WarMachineQueuePosition(hash string) (*int, *string) {
	fq.RLock()
	defer fq.RUnlock()
	for i, wm := range fq.QueuingWarMachines {
		if wm.Hash != hash {
			continue
		}

		position := i + 1

		return &position, wm.ContractReward
	}

	for _, wm := range fq.InGameWarMachines {
		if wm.Hash != hash {
			continue
		}

		position := -1

		return &position, wm.ContractReward
	}

	return nil, nil
}

func (fq *FactionQueue) GetFirstFiveQueuingWarMachines() []*server.WarMachineBrief {
	result := []*server.WarMachineBrief{}
	fq.RLock()
	defer fq.RUnlock()
	if len(fq.QueuingWarMachines) < 5 {
		for _, wm := range fq.QueuingWarMachines {
			result = append(result, wm.Brief())
		}

		return result
	}

	for i := 0; i < 5; i++ {
		result = append(result, fq.QueuingWarMachines[i].Brief())
	}

	return result
}
