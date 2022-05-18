package battle

import (
	"server"
)

func (arena *Arena) SecureUserCommand(key string, fn server.SecureCommandFunc) {
	arena.SecureUserCommander.Command(string(key), server.MustSecure(fn))
}

func (arena *Arena) SecureUserFactionCommand(key string, fn server.SecureFactionCommandFunc) {
	arena.SecureFactionCommander.Command(string(key), server.MustSecureFaction(fn))
}
