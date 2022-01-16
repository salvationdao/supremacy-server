package battle_arena

import "server"

func (ba *BattleArena) CurrentBattleID() server.BattleID {
	return ba.battle.ID
}
