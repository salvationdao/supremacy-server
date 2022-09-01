package battle

func (am *ArenaManager) SendBattleLobbyFunc(fn func() error) error {
	am.BattleLobbyFuncMx.Lock()
	defer am.BattleLobbyFuncMx.Unlock()
	return fn()
}
