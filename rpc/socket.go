package rpc

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type WsServer struct {
	listener net.Listener
	addr     string
	upgrade  *websocket.Upgrader
	entity   *ServerStarter
}

func NewWsServer(addr string, server *ServerStarter) *WsServer {
	ws := new(WsServer)
	ws.addr = addr
	ws.entity = server
	ws.upgrade = &websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			if r.Method != "GET" {
				fmt.Println("method is not GET")
				return false
			}
			if r.URL.Path != "/" {
				fmt.Println("path error")
				return false
			}
			return true
		},
	}
	return ws
}

func (self *WsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		httpCode := http.StatusInternalServerError
		reasePhrase := http.StatusText(httpCode)
		fmt.Println("path error ", reasePhrase)
		http.Error(w, reasePhrase, httpCode)
		return
	}
	conn, err := self.upgrade.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("websocket error:", err)
		return
	}
	// fmt.Println("client connect :", conn.RemoteAddr())
	go self.connHandle(conn)

}
func (self *WsServer) connHandle(conn *websocket.Conn) {
	defer func() {
		conn.Close()
	}()
	// stopCh := make(chan int)
	rep := make(map[string]interface{})
	req := make(map[string]interface{})
	// go self.send(conn, stopCh)
	// res := make(map[string]interface{})
	for {
		conn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(5000)))
		_, msg, err := conn.ReadMessage()
		if err != nil {
			// close(stopCh)
			if netErr, ok := err.(net.Error); ok {
				if netErr.Timeout() {
					// fmt.Printf("ReadMessage timeout remote: %v\n", conn.RemoteAddr())
					return
				}
			}
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				fmt.Printf("ReadMessage other remote:%v error: %v \n", conn.RemoteAddr(), err)
			}
			return
		}
		json.Unmarshal(msg, &req)

		id, datas, err := self.entity.HandleJsonRPCRequest(req)

		if err != nil {
			rep["jsonrpc"] = "2.0"
			rep["error"] = err
			rep["id"] = id
		} else {
			rep["jsonrpc"] = "2.0"
			rep["id"] = id
			rep["result"] = datas
		}
		jsonInfo, err := json.Marshal(rep)

		self.send(conn, jsonInfo)
		// id, datas, err := entity.common.HandleJsonRPCRequest(data)

	}
}

func (self *WsServer) send(conn *websocket.Conn, info []byte) {

	// for {
	// select {
	// case <-stopCh:
	// 	fmt.Println("connect closed")
	// 	return
	// case <-time.After(time.Second * 1):
	// 	data := fmt.Sprintf("hello websocket test from server %v", time.Now().UnixNano())
	err := conn.WriteMessage(1, []byte(info))
	// 	fmt.Println("sending....")
	if err != nil {
		fmt.Println("send msg faild ", err)
		return
	}
	// }
	// }
}

func (w *WsServer) Start() (err error) {
	w.listener, err = net.Listen("tcp", w.addr)
	if err != nil {
		fmt.Println("net listen error:", err)
		return
	}
	err = http.Serve(w.listener, w)
	if err != nil {
		fmt.Println("http serve error:", err)
		return
	}
	return nil
}
