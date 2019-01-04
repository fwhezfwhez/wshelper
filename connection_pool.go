package wshelper

import (
	"context"
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"golang.org/x/net/websocket"
	"io"
	"sync"
	"time"
)

type ConnectionPool struct {
	Full bool
	Pool map[string]*websocket.Conn
	M    *sync.RWMutex
}

// new a concurrently safe pool to restore connections
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		Pool: make(map[string]*websocket.Conn),
		M:    &sync.RWMutex{},
	}
}

// get length
func (cp *ConnectionPool) Length() int {
	cp.M.RLock()
	defer cp.M.RUnlock()
	return len(cp.Pool)
}

// add
func (cp *ConnectionPool) Add(key string, conn *websocket.Conn) {
	if !cp.IfExist(key) {
		cp.M.Lock()
		defer cp.M.Unlock()
		cp.Pool[key] = conn
	}
}

func (cp *ConnectionPool) SetFull(state bool) {
	cp.M.Lock()
	defer cp.M.Unlock()
	cp.Full = state
}

// delete
func (cp *ConnectionPool) Remove(key string) {
	if cp.IfExist(key) {
		cp.M.Lock()
		defer cp.M.Unlock()
		delete(cp.Pool, key)
	}
}

// get
func (cp *ConnectionPool) Get(key string) (*websocket.Conn, bool) {
	cp.M.RLock()
	defer cp.M.RUnlock()
	con, ok := cp.Pool[key]
	return con, ok
}

// whether a key exists
func (cp *ConnectionPool) IfExist(key string) bool {
	cp.M.RLock()
	defer cp.M.RUnlock()
	_, ok := cp.Pool[key]
	return ok
}

// a supervisor to keep pool stable
func (cp *ConnectionPool) Supervisor() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context) {
		fmt.Println("supervisor for connection pool has been auto-started")
		for {
			select {
			case <-ctx.Done():
				fmt.Println("connection pool supervisor successfully canceled")
			default:
				if cp.Length() > v.GetInt("maxOnlineConnPerPool") {
					// the max num of conn is weakly consistent, it's ok to overweight not far,so no need to add lock here
					cp.SetFull(true)
				} else {
					cp.SetFull(false)
				}
			}
			time.Sleep(10 * time.Minute)
		}
	}(ctx)
	return cancel
}

// send msg to a user
func (cp *ConnectionPool) SendOne(data []byte, to string) error {
	cp.M.RLock()
	con, ok := cp.Get(to)
	cp.M.RUnlock()
	if !ok {
		return nil
	}
	return websocket.Message.Send(con, data)
}

// eof and user offline is not regarded as error, since record will be saved to database,
// when user off-line or connection closed by client, this real-time chat  does nothing
func (cp *ConnectionPool) SendMany(data []byte, tos ... string) error {
	var er = make(chan error, len(tos))
	var wg = sync.WaitGroup{}
	wg.Add(len(tos))

	for _, to := range tos {
		go func(to string, wg *sync.WaitGroup) {
			defer wg.Done()
			cp.M.RLock()
			con, ok := cp.Get(to)
			cp.M.RUnlock()
			if !ok {
				return
			}

			e := websocket.Message.Send(con, data)
			if e != nil {
				if e != io.EOF {
					er <- errorx.New(e)
				}
			}
		}(to, &wg)
	}
	wg.Wait()
	var errors = make([]error, 0, len(tos))
L:
	for {
		e, ok := <-er
		if !ok {
			break L
		}
		errors = append(errors,e)
	}
	return errorx.GroupErrors(errors...)
}
