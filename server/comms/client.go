package comms

import (
	"net/rpc"

	"go.uber.org/atomic"
)

type C struct {
	clients []*rpc.Client
	inc     *atomic.Int32
}

func New(addrs ...string) (*C, error) {
	clients := []*rpc.Client{}
	for _, addr := range addrs {
		client, err := rpc.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)

	}
	c := &C{clients, atomic.NewInt32(0)}
	return c, nil
}

func (c *C) Call(serviceMethod string, args interface{}, reply interface{}) error {
	c.inc.Add(1)
	i := c.inc.Load()
	if i >= int32(len(c.clients)-1) {
		c.inc.Store(0)
		i = 0
	}
	client := c.clients[i]
	return client.Call(serviceMethod, args, reply)
}
