package comms

import (
	"errors"
	"net/rpc"
	"time"

	"github.com/rs/zerolog"
	"go.uber.org/atomic"
)

type C struct {
	addrs   []string
	clients []*rpc.Client
	inc     *atomic.Int32
	log     *zerolog.Logger
}

func New(log *zerolog.Logger, addrs ...string) (*C, error) {
	clients, err := connect(log, addrs...)
	if err != nil {
		return nil, err
	}
	c := &C{addrs, clients, atomic.NewInt32(0), log}
	return c, nil
}
func connect(log *zerolog.Logger, addrs ...string) ([]*rpc.Client, error) {
	clients := []*rpc.Client{}
	for _, addr := range addrs {
		log.Info().Str("addr", addr).Msg("registering RPC client")
		client, err := rpc.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)

	}
	return clients, nil
}

func (c *C) Call(serviceMethod string, args interface{}, reply interface{}) error {
	c.inc.Add(1)
	i := c.inc.Load()
	if i >= int32(len(c.clients)-1) {
		c.inc.Store(0)
		i = 0
	}
	client := c.clients[i]
	err := client.Call(serviceMethod, args, reply)
	if err != nil && errors.Is(err, rpc.ErrShutdown) {
		newClients, err := connect(c.log, c.addrs...)
		if err != nil {
			time.Sleep(5 * time.Second)
			return c.Call(serviceMethod, args, reply)
		}
		c.clients = newClients
		return c.Call(serviceMethod, args, reply)
	}
	if err != nil {
		return err
	}
	return nil
}
