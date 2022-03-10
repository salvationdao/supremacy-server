package comms

import (
	"fmt"
	"net"
	"net/rpc"
	"server/gamelog"
	"sync"

	"github.com/ninja-software/terror/v2"
)

type XrpcServer struct {
	PassportRPC XrpcClient     // common resource use by XrpcServer.clients
	isListening bool           // is server initialized and in use?
	listeners   []net.Listener // listening sockets
	mutex       sync.Mutex     // basic lock for listeners modification
}

type RPCListener int

func (s *XrpcServer) Listen(addrStrs ...string) error {
	s.listeners = make([]net.Listener, len(addrStrs))

	for i, a := range addrStrs {
		addy, err := net.ResolveTCPAddr("tcp", a)
		if err != nil {
			return terror.Error(err)
		}

		inbound, err := net.ListenTCP("tcp", addy)
		if err != nil {
			return terror.Error(err)
		}

		listener := new(RPCListener)
		rpc.Register(listener)
		s.mutex.Lock()
		s.listeners[i] = inbound
		s.mutex.Unlock()

		gamelog.L.Info().Str("addr", inbound.Addr().String()).Msg("starting up RPC server")

		// spun off and have running/blocking rpc listner
		go rpc.Accept(inbound)
	}

	s.isListening = true
	return nil
}

func (s *XrpcServer) Shutdown() error {
	var lastError error

	if !s.isListening {
		return terror.Error(fmt.Errorf("rpc server not yet started"))
	}

	for _, listener := range s.listeners {
		err := listener.Close()
		if err != nil {
			lastError = err
		}
	}

	s.isListening = false
	return lastError
}
