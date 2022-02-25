package battle_arena

import "server"

func (ba *BattleArena) CurrentBattleID() server.BattleID {
	return ba.battle.ID
}
func (ba *BattleArena) CurrentBattleIdentifier() int64 {
	return ba.battle.Identifier
}

func (ba *BattleArena) WarMachineDestroyedRecord(participantID byte) *server.WarMachineDestroyedRecord {
	record, ok := ba.battle.WarMachineDestroyedRecordMap[participantID]
	if !ok {
		return nil
	}

	return record
}

func (ba *BattleArena) InGameWarMachines() []*server.WarMachineMetadata {
	return ba.battle.WarMachines
}

func (ba *BattleArena) GetWarMachine(hash string) *server.WarMachineMetadata {
	for _, wm := range ba.battle.WarMachines {
		if wm.Hash == hash {
			return wm
		}
	}

	return nil
}
