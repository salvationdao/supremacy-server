package battle

import (
	"net/http"
	"nhooyr.io/websocket"
)

type Arena struct {
}

func (arena *Arena) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		gamelog
	}
}
