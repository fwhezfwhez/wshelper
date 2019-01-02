package wshelper

import (
	"context"
	"fmt"
	"golang.org/x/net/websocket"
	"sync"
	"time"
)

type ConnectionPool struct{
	Full bool
	Pool map[string]*websocket.Conn
	M *sync.RWMutex
}

// new a concurrently safe pool to restore connections
func NewConnectionPool() *ConnectionPool{
	return &ConnectionPool{
		Pool: make(map[string]*websocket.Conn),
		M: &sync.RWMutex{},
	}
}

// get length
func (cp *ConnectionPool) Length() int{
	cp.M.RLock()
	defer cp.M.RUnlock()
	return len(cp.Pool)
}

// add
func (cp *ConnectionPool) Add(key string, conn *websocket.Conn){
	if !cp.IfExist(key){
		cp.M.Lock()
		defer cp.M.Unlock()
		cp.Pool[key] = conn
	}
}

func (cp *ConnectionPool) SetFull(state bool){
	cp.M.Lock()
	defer cp.M.Unlock()
	cp.Full = state
}

// delete
func (cp *ConnectionPool) Remove(key string){
	if cp.IfExist(key) {
		cp.M.Lock()
		defer cp.M.Unlock()
		delete(cp.Pool, key)
	}
}

// get
func (cp *ConnectionPool) Get(key string) *websocket.Conn{
	cp.M.RLock()
	defer cp.M.RUnlock()
	return cp.Pool[key]
}

// whether a key exists
func (cp *ConnectionPool) IfExist(key string) bool{
	cp.M.RLock()
	defer cp.M.RUnlock()
	_,ok:= cp.Pool[key]
	return ok
}

// a supervisor to keep pool stable
func (cp *ConnectionPool) Supervisor() context.CancelFunc{
	ctx,cancel := context.WithCancel(context.Background())

	go func(ctx context.Context){
		fmt.Println("supervisor for connection pool has been auto-started")
		for{
			select{
				case <-ctx.Done():
					fmt.Println("connection pool supervisor successfully canceled")
				default:
					if cp.Length() > v.GetInt("MaxOnlineConnPerPool") {
						// the max num of conn is weakly consistent, it's ok to overweight not far,so no need to add lock here
						cp.SetFull(true)
					} else{
						cp.SetFull(false)
					}
			}
			time.Sleep(10 * time.Minute)
		}
	}(ctx)
	return cancel
}
