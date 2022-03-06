package comms

import (
	"fmt"
	"math/big"
	"net"
	"net/rpc"
	"server/gamelog"

	"github.com/sasha-s/go-deadlock"
)

// for sups trickle handler
type TickerPoolCache struct {
	deadlock.Mutex
	TricklingAmountMap map[string]*big.Int
}
type S struct {
}

func NewServer() *S {
	result := &S{}
	return result
}

func (c *C) listen(addrStr ...string) ([]net.Listener, error) {
	listeners := make([]net.Listener, len(addrStr))
	for i, a := range addrStr {
		gamelog.L.Info().Str("addr", a).Msg("registering RPC server")
		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%s", a))
		if err != nil {
			gamelog.L.Err(err).Str("addr", a).Msg("registering RPC server")
			return listeners, nil
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return listeners, err
		}

		listeners[i] = l
	}

	return listeners, nil
}

func Start(c *C) error {
	listeners, err := c.listen("10011", "10012", "10013", "10014", "10015", "10016")
	if err != nil {
		return err
	}
	for _, l := range listeners {
		s := rpc.NewServer()
		err = s.Register(c)
		if err != nil {
			return err
		}

		gamelog.L.Info().Str("addr", l.Addr().String()).Msg("starting up RPC server")
		go s.Accept(l)
	}

	return nil
}
