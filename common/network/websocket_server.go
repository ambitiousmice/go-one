package network

import (
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"time"
)

type WebsocketServerDelegate interface {
	ServeWebsocketConnection(w http.ResponseWriter, r *http.Request)
}

func ServeWebsocket(listenAddr string, delegate WebsocketServerDelegate) {
	log.Infof("Listening on websocket: %s ...", listenAddr)
	http.HandleFunc("/ws", delegate.ServeWebsocketConnection)
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		log.Panic("run websocket server error:" + err.Error())
	}
}

// WebSocketConn 是一个实现了 net.Conn 接口的结构体
type WebSocketConn struct {
	*websocket.Conn
}

// Read 实现了 net.Conn 接口的 Read 方法
func (wsc WebSocketConn) Read(b []byte) (int, error) {
	// 使用 WebSocket 连接的读方法
	_, reader, err := wsc.NextReader()
	if err != nil {
		return 0, err
	}
	return reader.Read(b)
}

// Write 实现了 net.Conn 接口的 Write 方法
func (wsc WebSocketConn) Write(b []byte) (int, error) {
	// 使用 WebSocket 连接的写方法
	writer, err := wsc.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, err
	}
	n, err := writer.Write(b)
	if err != nil {
		return n, err
	}
	err = writer.Close()
	return n, err
}

// Close 实现了 net.Conn 接口的 Close 方法
func (wsc WebSocketConn) Close() error {
	// 使用 WebSocket 连接的关闭方法
	return wsc.Conn.Close()
}

// LocalAddr 实现了 net.Conn 接口的 LocalAddr 方法
func (wsc WebSocketConn) LocalAddr() net.Addr {
	return wsc.Conn.LocalAddr()
}

// RemoteAddr 实现了 net.Conn 接口的 RemoteAddr 方法
func (wsc WebSocketConn) RemoteAddr() net.Addr {
	return wsc.Conn.RemoteAddr()
}

// SetDeadline 实现了 net.Conn 接口的 SetDeadline 方法
func (wsc WebSocketConn) SetDeadline(t time.Time) error {
	return wsc.Conn.SetReadDeadline(t)
}

// SetReadDeadline 实现了 net.Conn 接口的 SetReadDeadline 方法
func (wsc WebSocketConn) SetReadDeadline(t time.Time) error {
	return wsc.Conn.SetReadDeadline(t)
}

// SetWriteDeadline 实现了 net.Conn 接口的 SetWriteDeadline 方法
func (wsc WebSocketConn) SetWriteDeadline(t time.Time) error {
	return wsc.Conn.SetWriteDeadline(t)
}
