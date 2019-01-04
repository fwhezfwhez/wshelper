package wshelper

import (
	"encoding/json"
	"eyas/wshelper/util"
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"golang.org/x/net/websocket"
	"io"
	"strconv"
	"sync"
	"time"
)

const (
	B = 1 << (iota * 10)
	KB
	MB
	GB
)

type WebSocketHelper struct {
	M *sync.RWMutex
	// box all commands supported
	Commands []int
	// hash each command to an unique and same-length header
	commandHash         map[string]int
	// command mapper to handle connection by command value
	commandHandleMapper map[int]func(pool *ConnectionPool, rawBytes []byte, cache map[string]interface{}) error

	// a common function to handle error
	handleE func(error)

	// marshal and unmarshal message
	Serializer Marshaller
	// save all connections online
	pool *ConnectionPool
}

type Marshaller interface {
	Marshal(obj interface{}) ([]byte, error)
	Unmarshal(src []byte, dest interface{}) error
	TypeName() string
}

// an realization of Marshaller.
// if NewWsHelper(nil), the Jsoner will be put to use.
type Jsoner struct {
}

// marshal
func (j Jsoner) Marshal(dest interface{}) ([]byte, error) {
	return json.Marshal(dest)
}
// unmarshal
func (j Jsoner) Unmarshal(data []byte, dest interface{}) error {
	return json.Unmarshal(data, dest)
}

// type name
func (j Jsoner) TypeName() string{
	return "json"
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
		handleE:     Panic,
	}
	if dest == nil {
		dest = Jsoner{}
	}
	wsh.Serializer = dest
	return wsh
}

// marshal
func (wsh *WebSocketHelper) Marshal(obj interface{}) ([]byte, error) {
	return wsh.Serializer.Marshal(obj)
}

// unmarshal
func (wsh *WebSocketHelper) Unmarshal(src []byte, dest interface{}) error {
	return wsh.Serializer.Unmarshal(src, dest)
}

// set marshaller
// I assume you're aware of the danger of changing marshaller while server is on .
// It's best set it right before the ws server has listened on
func (wsh *WebSocketHelper) SetMarshaller(dest Marshaller) {
	wsh.M.Lock()
	defer wsh.M.Unlock()
	wsh.Serializer = dest
}

// set error handler
// if not set , panic(error) is the officials
func (wsh *WebSocketHelper) SetHandleE(f func(error)){
	wsh.M.Lock()
	defer wsh.M.Unlock()
	wsh.handleE = f
}

// generate a md5 key for a command
func (wsh *WebSocketHelper) genCommandHash(command int) string {
	hash1 := util.MD5(strconv.Itoa(command))
	return util.MD5(hash1 + strconv.Itoa(command))
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

// online a user
func (wsh *WebSocketHelper) Online(key string, conn *websocket.Conn) {
	wsh.pool.Add(key, conn)
}

// offline a user
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
	if len(receive) < 32 {
		return errorx.NewFromStringf("required message  more than 32 bit but got '%s' length '%d'", string(receive), len(receive))
	}

	*command = wsh.GetCommand(string(receive[:32]))

	if *command == 0 {
		return errorx.NewFromStringf("unknown command '%d'", *command)
	}
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

	if len(buf) < 32 {
		return nil,errorx.NewFromStringf("required message  more than 32 bit but got '%s' length '%d'", string(buf), len(buf))
	}
	command := wsh.GetCommand(string(buf[:32]))

	if command == 0 {
		return nil, errorx.NewFromStringf("unknown command '%d'", command)
	}
	return buf, nil
}

// get raw bytes of a request,
// rs is the result per command,
// buf is a buffer per read,
// startFrom is the byte number calculated,
// maxSize is the limit max size of body
func (wsh *WebSocketHelper) SafeRawBytesOf(rs []byte, buf []byte, startFrom *int, conn *websocket.Conn, maxSize int) ([]byte, error) {
	if rs == nil {
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

	if len(buf[:n]) < 32 {
		return nil,errorx.NewFromStringf("required message  more than 32 bit but got '%s' length '%d'", string(buf), len(buf))
	}
	command := wsh.GetCommand(string(buf[:32]))

	if command == 0 {
		return nil, errorx.NewFromStringf("unknown command '%d'", command)
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

// a dispatcher model,how to use?
// assume you raise a ws server:
// func main() {
//     wsh := NewWsHelper(nil)
//     f := func(e error){
//         fmt.Println(e.Error())
//     }
//     http.Handle("/test-ws/", websocket.Handler(wsh.Dispatcher(f)))
//     err := http.ListenAndServe("127.0.0.1:8787", nil)
//     if err != nil {
//         panic(err)
//     }
// }
//
// a wsHelper instance serves only one url.
// it does not support dispatcher two or more url like:
//	http.Handle("/test-ws1/", websocket.Handler(wsh.Dispatcher(f)))
//	http.Handle("/test-ws2/", websocket.Handler(wsh.Dispatcher(f)))
// why this?
// we limit that each guest has only one single ws connection to the server, and different function is dispatcher by command.
// if design more than one url, each url would be a seperate connection, however the total tcp connections are number limited in a computer.
func (wsh *WebSocketHelper) Dispatcher(handleE func(error)) func(conn *websocket.Conn) {
	return func(conn *websocket.Conn) {
		defer func(){
			if e:= recover(); e!=nil {
				logger.Println(fmt.Sprintf("recover from '%v'", e))
			}
		}()
		defer conn.Close()
		// conn.SetDeadline(v.GetInt("deadline") *time.Second))
		// conn.MaxPayloadBytes = v.GetInt(maxPayloadBytes)

		conn.SetDeadline(time.Now().Add(10 * 60 * 60 * time.Second))
		conn.MaxPayloadBytes = 1 * GB
		var er error
		var raw []byte
		var cache map[string]interface{}
		for {
			raw, er = wsh.RawBytesOf(conn)
			if er != nil {
				if er == io.EOF {
					return
				}
				handleE(er)
				return
			}
			command := wsh.CommandOf(raw)
			// the command mapper should be set right on the init stage,thus,no need to add lock
			handler, ok :=wsh.commandHandleMapper[command]
			if !ok {
				continue
			}
			er = handler(wsh.pool, raw, cache)
			if er != nil {
				if er == io.EOF {
					return
				}
				handleE(er)
				return
			}
		}
	}
}

// fmt an error
func Panic(e error) {
	panic(e)
}

// log to file
func LogToFile(e error){
    logger.Println(e.Error())
}
