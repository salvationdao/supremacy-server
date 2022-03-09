package comms

import (
	"fmt"
	"math/rand"
	"net/rpc"
	"server/gamelog"
	"time"

	"github.com/jpillora/backoff"
	"github.com/ninja-software/terror/v2"
	"go.uber.org/atomic"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type C struct {
	addrs   []string
	clients []*rpc.Client
	inc     *atomic.Int32
}

func NewClient(addrs ...string) (*C, error) {
	clients, err := connect(addrs...)
	if err != nil {
		return nil, terror.Error(err)
	}
	c := &C{addrs, clients, atomic.NewInt32(0)}
	return c, nil
}

func connect(addrs ...string) ([]*rpc.Client, error) {
	b := &backoff.Backoff{
		Min:    1 * time.Second,
		Max:    10 * time.Second,
		Factor: 2,
	}
	attempts := 0
	var clients []*rpc.Client
	for {
		attempts++
		gamelog.L.Info().Int("attempt", attempts).Msg("fetching battle queue from passport")
		clients = []*rpc.Client{}
		for _, addr := range addrs {
			gamelog.L.Info().Str("addr", addr).Msg("registering RPC client")
			client, err := rpc.Dial("tcp", addr)
			if err != nil {
				gamelog.L.Err(err).Str("addr", addr).Msg("registering RPC client")
				time.Sleep(b.Duration())
				continue
			}
			clients = append(clients, client)
		}

		break
	}
	return clients, nil
}

func (c *C) GoCall(serviceMethod string, args interface{}, reply interface{}, callback func(error)) {
	go func() {
		err := c.Call(serviceMethod, args, reply)
		if callback != nil {
			callback(err)
		}
	}()
}

func (c *C) Call(serviceMethod string, args interface{}, reply interface{}) error {
	if c == nil || c.clients == nil || len(c.clients) == 0 {
		clients, err := connect(c.addrs...)
		if err != nil {
			return terror.Error(err)
		}
		c.clients = clients
		return terror.Error(fmt.Errorf("rpc client not ready"))
	}
	gamelog.L.Debug().Str("fn", serviceMethod).Interface("args", args).Msg("rpc call")
	span := tracer.StartSpan("rpc.call", tracer.ResourceName(serviceMethod))
	defer span.Finish()
	c.inc.Add(1)
	i := c.inc.Load()
	if i >= int32(len(c.clients)-1) {
		c.inc.Store(0)
		i = 0
	}
	if len(c.clients) < int(i) {
		return terror.Error(fmt.Errorf("index out of range len = %d, index = %d", len(c.clients), int(i)))
	}
	client := c.clients[i]

	var err error
	var retryCall uint
	for {
		if client == nil {
			// keep redialing until 30 times
			client, err = dial(30, randAddr(c.addrs...))
			if err != nil {
				return terror.Error(err)
			}
		}

		// TODO not thread safe
		c.clients[i] = client

		err = client.Call(serviceMethod, args, reply)
		if err == nil {
			// done
			return nil
		}

		// clean up before retry
		if client != nil {
			// close first
			client.Close()
		}
		client = nil

		retryCall++
		if retryCall > 6 {
			return terror.Error(fmt.Errorf("call retry exceeded 6 times"))
		}
	}

	gamelog.L.Warn().Str("fn", "comms.Call").Msg("shouldnt reach here")
	return nil
}

// randAddr picks an address randomly
func randAddr(addrs ...string) string {
	rand.Seed(time.Now().UnixNano())
	c := rand.Intn(len(addrs))
	return addrs[c]
}

// dial is primitive rpc dialer, short and simple
// maxRetry -1 == unlimited
func dial(maxRetry int, addrAndPorts ...string) (client *rpc.Client, err error) {
	if len(addrAndPorts) <= 0 {
		return nil, terror.Error(fmt.Errorf("addr/port is zero length"))
	}
	addrAndPort := randAddr(addrAndPorts...)

	retry := 0
	err = fmt.Errorf("x")

	for err != nil {
		// rpc have own timeout probably 1~1.4 sec?
		client, err = rpc.Dial("tcp", addrAndPort)
		if err == nil {
			return
		}
		gamelog.L.Debug().Err(err).Str("fn", "comms.dial").Msgf("err: dial fail, retrying... %d", retry)

		// unlimited retry
		if maxRetry < 0 {
			continue
		}

		retry++
		// limited retry
		if retry > maxRetry {
			break
		}
	}

	return nil, terror.Error(fmt.Errorf("rpc dial failed after %d retries", maxRetry))
}
