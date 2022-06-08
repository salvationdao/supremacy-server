package comms

import (
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/sasha-s/go-deadlock"
	"net"
	"net/rpc"
	"server/gamelog"
	"server/xsyn_rpcclient"
)

// S holds all the listeners together
type XrpcServer struct {
	isListening bool           // is server initialized and in use?
	listeners   []net.Listener // listening sockets
	mutex       deadlock.Mutex // basic lock for listeners modification
}

// S holds all the RPC answer functions, remote rpc caller must use same naming.
// Keep seperate from XrpcServer so it wont cause issue and complain about Listen and Shutdown being invalid length and trigger by remotely
type S struct {
	passportRPC *xsyn_rpcclient.XsynXrpcClient // rpc client to call passport server
}

func (s *XrpcServer) Listen(passportRPC *xsyn_rpcclient.XsynXrpcClient, startPort, numPorts int) error {
	if passportRPC == nil {
		return terror.Error(fmt.Errorf("passportRPC is nil"))
	}
	addrStrs := make([]string, numPorts)
	for i := 0; i < numPorts; i++ {
		addrStrs[i] = fmt.Sprintf(":%d", i+startPort)
	}
	if len(addrStrs) == 0 {
		return terror.Error(fmt.Errorf("no rpc listen given, minimum of 1"))
	}
	s.listeners = make([]net.Listener, len(addrStrs))

	for i, a := range addrStrs {
		addy, err := net.ResolveTCPAddr("tcp", a)
		if err != nil {
			return err
		}

		inbound, err := net.ListenTCP("tcp", addy)
		if err != nil {
			return err
		}

		listener := new(S)
		listener.passportRPC = passportRPC
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

// Ping to make sure it works and healthy
func (s *S) Ping(req bool, resp *string) error {
	*resp = "PONG from GAMESERVER"
	return nil
}
