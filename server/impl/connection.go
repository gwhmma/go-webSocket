package impl

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

type Connection struct {
	wsConn    *websocket.Conn
	inChan    chan []byte
	outChan   chan []byte

	sync.Mutex
	isClose bool
	closeChan chan byte
}

func InitConnection(wsConn *websocket.Conn) (conn *Connection, err error) {
	conn = &Connection{
		wsConn:    wsConn,
		inChan:    make(chan []byte, 1000),
		outChan:   make(chan []byte, 1000),
		closeChan: make(chan byte, 1),
	}

	// 启动读协程
	go conn.readLoop()
	// 启动写协程
	go conn.writeLoop()

	return
}

// API
func (c *Connection) ReadMessage() (data []byte, err error) {
	select {
	case data = <-c.inChan:
	case <-c.closeChan:
		err = errors.New("connection is closed")
	}
	return
}

func (c *Connection) WriteMessage(data []byte) (err error) {
	select {
	case c.outChan <- data:
	case <-c.closeChan:
		err = errors.New("connection is closed")
	}
	return
}

func (c *Connection) Close() {
	// 线程安全， 可重入的close
	c.wsConn.Close()
	// chan 只能关闭一次
	c.Mutex.Lock()
	if !c.isClose {
		close(c.closeChan)
		c.isClose = true
	}
	c.Mutex.Unlock()
}

// sync.Once
//func closeChan()  {
//	close()
//}

// 内部实现
func (c *Connection) readLoop() {
	var (
		data []byte
		err  error
	)

	for {
		if _, data, err = c.wsConn.ReadMessage(); err != nil {
			fmt.Println("readLoop : ", err)
			c.Close()
			return
		}

		// 阻塞在这里 等待inChan有位置
		select {
		case c.inChan <- data:
		case <-c.closeChan:
			// closeChan关闭的时候
			c.Close()
		}
	}
}

func (c *Connection) writeLoop() {
	var (
		data []byte
		err  error
	)

	for {
		select {
		case data = <-c.outChan:
		case <-c.closeChan:
			c.Close()
		}

		if err = c.wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
			fmt.Println("writeLoop : ", err)
			c.Close()
			return
		}
	}
}
