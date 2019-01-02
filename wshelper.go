package wshelper

import (
	"encoding/json"
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"golang.org/x/net/websocket"
	"io"
	"strconv"
	"sync"
)

const (
	B = 1 << (iota * 10)
	KB
	MB
	GB
)

type WebSocketHelper struct {
	M           *sync.RWMutex
	Commands    []int
	commandHash map[string]int
	Serializer  Marshaller
	pool        *ConnectionPool
}

type Marshaller interface {
	Marshal(obj interface{}) ([]byte, error)
	Unmarshal(src []byte, dest interface{}) error
}

type Jsoner struct {
}

func (j Jsoner) Marshal(dest interface{}) ([]byte, error) {
	return json.Marshal(dest)
}
func (j Jsoner) Unmarshal(data []byte, dest interface{}) error {
	return json.Unmarshal(data, dest)
}

// init a ws helper instance.
// a serializer implementing wshelper.Marshaller should be correctly set whatever protobuf/xml/json.
// if 'dest' is set nil, the default jsoner will be used
func NewWsHelper(dest Marshaller) *WebSocketHelper {
	wsh := &WebSocketHelper{
		pool:        NewConnectionPool(),
		M:           &sync.RWMutex{},
		Commands:    make([]int, 0, 10),
		commandHash: make(map[string]int, 0),
	}
	if dest == nil {
		dest = Jsoner{}
	}
	wsh.Serializer = dest
	return wsh
}

func (wsh *WebSocketHelper) Marshal(obj interface{}) ([]byte, error) {
	return wsh.Serializer.Marshal(obj)
}

func (wsh *WebSocketHelper) Unmarshal(src []byte, dest interface{}) error {
	return wsh.Serializer.Unmarshal(src, dest)
}

func (wsh *WebSocketHelper) SetMarshaller(dest Marshaller) {
	wsh.Serializer = dest
}

// generate a md5 key for a command
func (wsh *WebSocketHelper) genCommandHash(command int) string {
	hash1 := MD5(strconv.Itoa(command))
	return MD5(hash1 + strconv.Itoa(command))
}

// Set commands and commandHash , when exists,replace the old
// the hash of each command is 32 bit md5 salted by command itself in the depth of 1, details refers 'genCommandHash(int)string'
func (wsh *WebSocketHelper) SetCommands(commands ...int) {
	wsh.M.Lock()
	defer wsh.M.Unlock()
	wsh.Commands = commands
	for _, v := range wsh.Commands {
		wsh.commandHash[wsh.genCommandHash(v)] = v
	}
}

// get a msg command from its hash
func (wsh *WebSocketHelper) GetCommand(hash string) int {
	wsh.M.RLock()
	defer wsh.M.RUnlock()
	return wsh.commandHash[hash]
}

// get commandHashes
func (wsh *WebSocketHelper) ListCommandsHash() map[string]int {
	wsh.M.RLock()
	defer wsh.M.RUnlock()
	return wsh.commandHash
}

func (wsh *WebSocketHelper) Online(key string, conn *websocket.Conn) {
	wsh.pool.Add(key, conn)
}

func (wsh *WebSocketHelper) Offline(key string, conn *websocket.Conn) {
	wsh.pool.Remove(key)
}

// bind a message []byte to a struct 'dest' and the command value to 'command'
func (wsh *WebSocketHelper) Bind(conn *websocket.Conn, command *int, dest interface{}) error {
	var receive []byte

	er := websocket.Message.Receive(conn, &receive)
	if er != nil {
		if er == io.EOF {
			return er
		}
		return errorx.New(er)
	}

	*command = wsh.GetCommand(string(receive[:32]))
	return wsh.Unmarshal(receive[32:], dest)
}

// get a raw bytes of a request
func (wsh *WebSocketHelper) RawBytesOf(conn *websocket.Conn) ([]byte, error) {
	var buf []byte
	er := websocket.Message.Receive(conn, &buf)
	if er != nil {
		if er == io.EOF {
			return nil, er
		}
		return nil, errorx.New(er)
	}
	return buf, nil
}

// get raw bytes of a request,
// rs is the result per command,
// buf is a buffer per read,
// startFrom is the byte number calculated,
// maxSize is the limit max size of body
func (wsh *WebSocketHelper) SafeRawBytesOf(rs []byte, buf []byte, startFrom *int, conn *websocket.Conn, maxSize int) ([]byte, error) {
	if rs ==nil {
		rs = make([]byte, 0, 512)
	}

	if buf == nil {
		buf = make([]byte, 512)
	}
	n, e := conn.Read(buf)
	if e != nil {
		if e == io.EOF {
			return nil, e
		}
		return nil, errorx.New(e)
	}
	*startFrom += n
	if *startFrom > maxSize {
		return nil, errorx.NewFromString(fmt.Sprintf("receive size '%d' bigger than the max '%d'", *startFrom, maxSize))
	}
	rs = append(rs, buf[:n]...)
	return rs, nil
}

// get the command from the raw bytes
func (wsh *WebSocketHelper) CommandOf(buf []byte) int {
	if len(buf) < 32 {
		panic(errorx.NewFromStringf("want buf more than 32 bit but got '%d'", len(buf)))
	}
	return wsh.GetCommand(string(buf[:32]))
}

// get the core struct from the raw bytes
func (wsh *WebSocketHelper) CoreOf(buf []byte, dest interface{}) error {
	return wsh.Unmarshal(buf[32:], dest)
}
