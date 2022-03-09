package comms

import (
	"fmt"
	"net"
	"net/rpc"
	"server/gamelog"

	"github.com/ninja-software/terror/v2"
)

type S struct {
	*C
}

func NewServer(c *C) *S {
	result := &S{c}
	return result
}

func (s *S) listen(addrStr ...string) ([]net.Listener, error) {
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
			return listeners, terror.Error(err)
		}

		listeners[i] = l
	}

	return listeners, nil
}

func Start(s *S) error {
	listeners, err := s.listen("10011", "10012", "10013", "10014", "10015", "10016")
	if err != nil {
		return err
	}
	for _, l := range listeners {
		srv := rpc.NewServer()
		err = srv.Register(s)
		if err != nil {
			return terror.Error(err)
		}

		gamelog.L.Info().Str("addr", l.Addr().String()).Msg("starting up RPC server")
		go srv.Accept(l)
	}

	return nil
}
