package pubsub

import (
	"encoding/json"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/sasha-s/go-deadlock"
	"github.com/valyala/fastjson"
	"go.uber.org/atomic"
	"net/http"
	"server/gamelog"
	"sync"
)

type ReplyFunc func(interface{})

type CommandFunc func(key string, payload []byte, reply ReplyFunc) error

var (
	commands *CommandTree
	commOnce sync.Once
	commLock deadlock.RWMutex
)

func Endpoint(key string, fn CommandFunc) {
	commOnce.Do(func() {
		commands = NewCommandTree()
	})

	commLock.Lock()
	commands.Insert(key, fn)
	commLock.Unlock()
}

type CommandServer struct{}

func (csrv *CommandServer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, rw)
	if err != nil {
		return
	}

	w := wsutil.NewWriter(conn, ws.StateServerSide, ws.OpText)
	j := json.NewEncoder(w)

	go func() {
		defer conn.Close()
		shouldClose := atomic.NewBool(false)
		for {
			if shouldClose.Load() {
				return
			}
			msg, _, err := wsutil.ReadClientData(conn)
			if err != nil {
				return
			}
			key := fastjson.GetString(msg, "key")
			txid := fastjson.GetString(msg, "key")
			commLock.RLock()
			prefix, fn, ok := commands.LongestPrefix(key)
			if ok {
				reply := func(payload interface{}) {
					if txid == "" {
						//this could do with a warning
						return
					}
					resp := struct {
						Key           string      `json:"key"`
						TransactionID string      `json:"transaction_id"`
						Success       bool        `json:"success"`
						Payload       interface{} `json:"payload"`
					}{
						Key:           key,
						TransactionID: txid,
						Success:       true,
						Payload:       payload,
					}

					b, err := json.Marshal(resp)
					if err != nil {
						gamelog.L.Err(err).Msgf("marshalling json in reply for %s has failed", key)
					}
					err = wsutil.WriteServerMessage(conn, ws.OpText, b)
					if err := j.Encode(&resp); err != nil {
						return
					}
					if err = w.Flush(); err != nil {
						shouldClose.Store(true)
						return
					}
				}
				err = fn(prefix, msg, reply)
				if err != nil {
					//hmm
				}
			}
			commLock.RUnlock()
		}
	}()
}
