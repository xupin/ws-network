package ws

import (
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// RFC 6455, section 11.8.
const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1
	// BinaryMessage denotes a binary data message.
	BinaryMessage = 2
	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8
	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9
	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10
)

const (
	// DefaultInChanSize 默认读队列大小
	DefaultInChanSize = 1024
	// DefaultOutChanSize 默认写队列大小
	DefaultOutChanSize = 1024
	// DefaultHeartbeatInterval 默认心跳检测间隔
	DefaultHeartbeatInterval = 300
)

type Message struct {
	MessageType int
	Data        []byte
}

type Connection struct {
	// id 标识id
	id string
	// conn 底层长连接
	conn *websocket.Conn
	// inChan 读队列
	inChan chan *Message
	// outChan 写队列
	outChan chan *Message
	// closeChan 关闭通知
	closeChan chan struct{}
	// heartbeatInterval 心跳检测间隔, 秒
	heartbeatInterval int
	// lastHeartbeatTime 最近一次心跳时间
	lastHeartbeatTime time.Time
	// mutex 保护 closeChan 只被执行一次
	mutex sync.Mutex
	// isClosed closeChan状态
	isClosed bool
}

type Options struct {
	InChanSize        int
	OutChanSize       int
	HeartbeatInterval int
}

func NewConnection(opts ...*Options) *Connection {
	inChanSize, outChanSize := DefaultInChanSize, DefaultOutChanSize
	heartbeatInterval := DefaultHeartbeatInterval
	if len(opts) > 0 {
		opt := opts[0]
		if opt.InChanSize > 0 {
			inChanSize = opt.InChanSize
		}
		if opt.OutChanSize > 0 {
			outChanSize = opt.OutChanSize
		}
		if opt.HeartbeatInterval > 0 {
			heartbeatInterval = opt.HeartbeatInterval
		}
	}
	return &Connection{
		id:                uuid.NewString(),
		conn:              nil,
		inChan:            make(chan *Message, inChanSize),
		outChan:           make(chan *Message, outChanSize),
		closeChan:         make(chan struct{}, 1),
		heartbeatInterval: heartbeatInterval,
		lastHeartbeatTime: time.Now(),
	}
}

func (c *Connection) Close() error {
	_ = c.conn.Close()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.isClosed {
		close(c.closeChan)
		c.isClosed = true
	}
	return nil
}

func (c *Connection) Open(w http.ResponseWriter, r *http.Request) error {
	upgrade := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	c.conn = conn
	go c.readLoop()
	go c.writeLoop()
	return nil
}

func (c *Connection) readLoop() {
	for {
		msgType, data, err := c.conn.ReadMessage()
		if err != nil {
			_ = c.Close()
			return
		}
		select {
		case c.inChan <- &Message{
			MessageType: msgType,
			Data:        data,
		}:
		case <-c.closeChan:
			return
		}
	}
}

func (c *Connection) writeLoop() {
	timer := time.NewTimer(time.Duration(c.heartbeatInterval) * time.Second)
	defer timer.Stop()
	for {
		select {
		case msg := <-c.outChan:
			_ = c.conn.WriteMessage(msg.MessageType, msg.Data)
		case <-c.closeChan:
			return
		}
	}
}

func (r *Connection) error() error {
	return errors.New("connection already closed")
}

// Receive 接收数据
func (c *Connection) Receive() (msg *Message, err error) {
	select {
	case msg = <-c.inChan:
	case <-c.closeChan:
		err = c.error()
	}
	return
}

func (c *Connection) Write(msg *Message) (err error) {
	select {
	case c.outChan <- msg:
	case <-c.closeChan:
		err = c.error()
	}
	return
}

func (c *Connection) GetConnID() string {
	return c.id
}

func (c *Connection) GetRemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}
