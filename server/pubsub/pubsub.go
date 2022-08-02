package pubsub

import (
	"encoding/json"
	"errors"
	"github.com/sasha-s/go-deadlock"
	"net"
	"net/http"
	"server/gamelog"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"go.uber.org/atomic"
)

type Subscribers struct {
	pools map[string]*pools
	deadlock.RWMutex
}

var (
	subscribers *Subscribers
	once        sync.Once
	l           deadlock.RWMutex
)

func Sub(URI string, c *Client) {
	l.Lock()
	pl, ok := subscribers.pools[URI]
	if !ok {
		pl = &pools{
			p: make([]*pool, 10),
		}
		pl.lastPool = pl.p[0]
		subscribers.pools[URI] = pl
	}
	pl.register(c)
}

type Server struct {
}

type Client struct {
	conn      net.Conn
	sendQueue chan *message
	close     chan bool
	onClose   func()
}

type message struct {
	op  ws.OpCode
	msg []byte
}

type pool struct {
	connections *[10]*Client
	sender      chan *message
	inserter    chan *Client
	remover     chan int32
	gone        int
	next        *atomic.Int32
	full        *atomic.Bool
}

func (cpool *pool) insert(c *Client) error {
	if cpool.full.Load() {
		return ErrPoolFull
	}
	cpool.inserter <- c
	return nil
}

func (cpool *pool) close() {
	cpool.connections = nil
}

var ErrPoolFull = errors.New("pool is full")

func (cpool *pool) send(m *message) {
	select {
	case cpool.sender <- m:
		return
	default:
		gamelog.L.Error().Msg("unable to send to pool as sender is full")
	}
}

func (cpool *pool) run() {
	for {
		select {
		case m := <-cpool.sender:
			for _, c := range cpool.connections {
				c.send(m)
			}
		case c := <-cpool.inserter:
			next := cpool.next.Load()
			c.onClose = func() {
				cpool.remover <- next
			}
			cpool.connections[next] = c
			next++
			cpool.next.Store(next)
			if next > 9 {
				cpool.full.Store(true)
			}
		case i := <-cpool.remover:
			cpool.connections[i] = nil
			cpool.gone++
			if cpool.gone == len(cpool.connections) {
				cpool.close()
			}
		}
	}
}

type pools struct {
	p        []*pool
	lastPool *pool
	deadlock.RWMutex
}

func (pls *pools) register(c *Client) {
	p := pls.lastPool
	if p == nil || p.full.Load() {
		p = &pool{
			connections: &[10]*Client{},
			sender:      make(chan *message, 10),
			inserter:    make(chan *Client, 2),
			remover:     make(chan int32, 2),
			full:        atomic.NewBool(false),
			next:        atomic.NewInt32(0),
		}
		pls.Lock()
		pls.p = append(pls.p, p)
		pls.Unlock()
		p.run()
	}
	err := p.insert(c)
	if err != nil && errors.Is(err, ErrPoolFull) {
		pls.register(c)
	}
}

func (pls *pools) publishBytes(b []byte) {
	m := &message{op: ws.OpBinary, msg: b}
	pls.RLock()
	for _, p := range pls.p {
		p.send(m)
	}
	pls.RUnlock()
}

func (pls *pools) publish(b []byte) {
	m := &message{op: ws.OpText, msg: b}
	pls.RLock()
	for _, p := range pls.p {
		p.send(m)
	}
	pls.RUnlock()
}

func (sub *Subscribers) PublishBytes(URI string, b []byte) {
	sub.RLock()
	ps, ok := sub.pools[URI]
	sub.RUnlock()
	if !ok {
		return
	}
	ps.publishBytes(b)
}

func (sub *Subscribers) Publish(URI string, b []byte) {
	sub.RLock()
	ps, ok := sub.pools[URI]
	sub.RUnlock()
	if !ok {
		return
	}
	ps.publish(b)
}

func C(w http.ResponseWriter, r *http.Request) *Client {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return nil
	}

	return &Client{
		conn:      conn,
		sendQueue: make(chan *message, 5),
		close:     make(chan bool),
	}
}

func (cl *Client) send(m *message) {
	for {
		select {
		case cl.sendQueue <- m:
		default:
			cl.close <- true
		}
	}
}

func (cl *Client) Send(op ws.OpCode, msg []byte) {
	cl.send(&message{op, msg})
}

func (cl *Client) MarshalSend(msg interface{}) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	cl.send(&message{ws.OpText, b})
	return nil
}

func (cl *Client) SendBinary(msg []byte) {
	cl.send(&message{ws.OpBinary, msg})
}

func (cl *Client) SendText(msg string) {
	cl.send(&message{ws.OpText, []byte(msg)})
}

func (cl *Client) Run() {
	if cl.sendQueue == nil {
		panic("client has not been initialised correctly")
	}

	defer cl.conn.Close()

	for {
		select {
		case m := <-cl.sendQueue:
			err := cl.conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
			if err != nil {
				return
			}
			err = wsutil.WriteServerMessage(cl.conn, m.op, m.msg)
			if err != nil {
				return
			}
		case <-cl.close:
			return
		}
	}
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		subscribers = &Subscribers{
			pools: make(map[string]*pools),
		}
	})
	client := C(w, r)
	Sub(r.URL.Path, client)

	go client.Run()
}
