package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"go-webSocket/server/impl"
	"net/http"
)

var upgrader = websocket.Upgrader{
	// 允许跨域
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		wsConn *websocket.Conn
		data   []byte
		conn   *impl.Connection
		err    error
	)

	// upgrade : websocket
	if wsConn, err = upgrader.Upgrade(w, r, nil); err != nil {
		fmt.Println(err)
		return
	}

	// upgrade.wsConn
	// Text, binary
	if conn, err = impl.InitConnection(wsConn); err != nil {
		return
	}

	//go func() {
	//	for {
	//		if err := conn.WriteMessage([]byte("heart beat")); err != nil {
	//			return
	//		}
	//		time.Sleep(5 * time.Second)
	//	}
	//}()

	for {
		if data, err = conn.ReadMessage(); err != nil {
			conn.Close()
		}

		if err = conn.WriteMessage(data); err != nil {
			conn.Close()
		}
	}

}

func main() {
	http.HandleFunc("/ws", wsHandler)
	http.ListenAndServe("0.0.0.0:7777", nil)
}
