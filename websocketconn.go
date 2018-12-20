package rtspclient

import (
	"net"
	"time"

	"github.com/gorilla/websocket"
)

type WebsocketConn struct {
	conn *websocket.Conn
}

func (conn *WebsocketConn) Read(b []byte) (int, error) {
	_, message, err := conn.conn.ReadMessage()
	if nil != err {
		return 0, err
	}
	copy(b[:len(message)], message)
	return len(message), nil
}

func (conn *WebsocketConn) Write(b []byte) (int, error) {
	err := conn.conn.WriteMessage(websocket.BinaryMessage, b)
	if nil != err {
		return 0, err
	}
	return len(b), nil
}

func (conn *WebsocketConn) Close() error {
	return conn.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

func (conn *WebsocketConn) LocalAddr() net.Addr {
	return conn.LocalAddr()
}

func (conn *WebsocketConn) RemoteAddr() net.Addr {
	return conn.RemoteAddr()
}

func (conn *WebsocketConn) SetDeadline(t time.Time) error {
	return conn.SetDeadline(t)
}

func (conn *WebsocketConn) SetReadDeadline(t time.Time) error {
	return conn.SetReadDeadline(t)
}

func (conn *WebsocketConn) SetWriteDeadline(t time.Time) error {
	return conn.SetWriteDeadline(t)
}
