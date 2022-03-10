package comms

import (
	"fmt"
	"net"
	"net/rpc"
	"server/gamelog"

	"github.com/ninja-software/terror/v2"
)

type XrpcServer struct {
	PassportRPC XrpcClient
}

func listen(addrStr ...string) ([]net.Listener, error) {
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

func (s *XrpcServer) Listen(addrStr ...string) error {
	listeners, err := listen(addrStr...)
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
