package xsyn_rpcclient

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/rpc"
	"server/gamelog"
	"strings"
	"sync"
	"time"

	"go.uber.org/atomic"

	"github.com/ninja-software/terror/v2"
)

// XrpcClient is a basic RPC client with retry function and also support multiple addresses for increase through-put
type XrpcClient struct {
	Addrs   []string      // list of rpc addresses available to use
	clients []*rpc.Client // holds rpc clients, same len/pos as the Addrs
	counter atomic.Int64  // counter for cycling address/clients
	mutex   sync.Mutex    // lock and unlocks clients slice editing
}

type XsynXrpcClient struct {
	*XrpcClient
	ApiKey string `json:"apiKey"`
}

func NewXsynXrpcClient(apiKey string, hostname string, startPort, numPorts int) *XsynXrpcClient {
	addrs := make([]string, numPorts)
	for i := 0; i < numPorts; i++ {
		addrs[i] = fmt.Sprintf("%s:%d", hostname, i+startPort)
	}
	new := &XsynXrpcClient{
		ApiKey:     apiKey,
		XrpcClient: &XrpcClient{Addrs: addrs},
	}

	return new
}

// Gocall plan to deprecate if possible
// func (c *XrpcClient) GoCall(serviceMethod string, args interface{}, reply interface{}, callback func(error)) {
// 	go func() {
// 		err := c.Call(serviceMethod, args, reply)
// 		if callback != nil {
// 			callback(err)
// 		}
// 	}()
// }

// Call calls RPC server and retry, also initialise if it is the first time
func (c *XrpcClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	// used for the first time, initialise
	if c.clients == nil {
		gamelog.L.Debug().Msg("comms.Call init first time")
		if len(c.Addrs) <= 0 {
			log.Fatal("no rpc address set")
		}

		c.mutex.Lock()
		for i := 0; i < len(c.Addrs); i++ {
			c.clients = append(c.clients, nil)
		}
		c.mutex.Unlock()
	}

	gamelog.L.Trace().Str("fn", serviceMethod).Interface("args", args).Msg("rpc call")

	// count up, and use the next client/address
	c.counter.Add(1)
	i := int(c.counter.Load()) % len(c.Addrs)
	client := c.clients[i]

	var err error
	var retryCall uint
	for {
		if client == nil {
			// keep redialing until rpc server comes back online
			client, err = dial(-1, c.Addrs[i])
			if err != nil {
				return err
			}
			c.mutex.Lock()
			c.clients[i] = client
			c.mutex.Unlock()
		}

		err = client.Call(serviceMethod, args, reply)
		if err == nil {
			// done
			break
		}

		if !errors.Is(err, rpc.ErrShutdown) {
			// TODO: create a error type to check
			if !strings.Contains(err.Error(), "not enough funds") && !strings.Contains(err.Error(), "sql: no rows in result set") {
				gamelog.L.Error().Err(err).Msg("RPC call has failed.")
			}
			return err
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

	return nil
}

// dial is primitive rpc dialer, short and simple
// maxRetry -1 == unlimited
func dial(maxRetry int, addrAndPort string) (client *rpc.Client, err error) {
	retry := 0
	err = fmt.Errorf("x")
	sleepTime := time.Millisecond * 1000 // 1 second
	sleepTimeMax := time.Second * 10     // 10 seconds
	sleepTimeScale := 1.01               // power scale, takes 12 retries to reach max 10 sec sleep, which is total of ~40 seconds later

	for err != nil {
		client, err = rpc.Dial("tcp", addrAndPort)
		if err == nil {
			break
		}
		gamelog.L.Debug().Err(err).Str("fn", "comms.dial").Msgf("err: dial fail, retrying... %d", retry)

		// increase timeout each time upto max
		time.Sleep(sleepTime)
		sleepTime = time.Duration(math.Pow(float64(sleepTime), sleepTimeScale))
		if sleepTime > sleepTimeMax {
			sleepTime = sleepTimeMax
		}

		retry++
		// unlimited retry
		if maxRetry < 0 {
			continue
		}

		// limited retry
		if retry > maxRetry {
			return nil, terror.Error(fmt.Errorf("rpc dial failed after %d retries", retry))
		}
	}

	return client, nil
}
