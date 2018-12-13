package tcpnetwork

import (
	"bufio"
	"bytes"
	"errors"
	"log"
	"net"
	"time"
)

const (
	// ConnStatusNone none status
	ConnStatusNone = iota
	// ConnStatusConnected connected status
	ConnStatusConnected
	// ConnStatusDisconnected disconnected status
	ConnStatusDisconnected
)

const (
	// ConnEventNone none event
	ConnEventNone = iota
	// ConnEventConnected connected event
	ConnEventConnected
	// ConnEventDisconnected disconnected event
	ConnEventDisconnected
	// ConnEventData recive data event
	ConnEventData
	// ConnEventClose close event
	ConnEventClose
)

const (
	connConfDefaultSendTimeoutSec = 5
	connConfMaxReadBufferLength   = 0xffff
)

// Connection TCP connection context
type Connection struct {
	conn                net.Conn
	connRW              *bufio.ReadWriter
	status              int
	sendMsgQueue        chan []byte
	sendBufferSize      int
	sendTimeoutSec      int
	readTimeoutSec      int
	eventHandler        func(*ConnEvent)
	streamProtocol      IStreamProtocol
	maxReadBufferLength int
}

// ConnEvent TCP connnection event
type ConnEvent struct {
	EventType int
	Conn      *Connection
	Data      []byte // data
}

func newConnEvent(eventType int, connect *Connection, data []byte) *ConnEvent {
	return &ConnEvent{
		EventType: eventType,
		Conn:      connect,
		Data:      data,
	}
}

// NewConnection create a Connection
func NewConnection(c net.Conn, sendBufferSize int, eventHandler func(*ConnEvent)) *Connection {
	return &Connection{
		conn:                c,
		connRW:              bufio.NewReadWriter(bufio.NewReaderSize(c, connConfMaxReadBufferLength), bufio.NewWriterSize(c, connConfMaxReadBufferLength)),
		status:              ConnStatusNone,
		sendMsgQueue:        make(chan []byte, sendBufferSize),
		sendBufferSize:      sendBufferSize,
		sendTimeoutSec:      connConfDefaultSendTimeoutSec,
		maxReadBufferLength: connConfMaxReadBufferLength,
		eventHandler:        eventHandler,
	}
}

// directly close, packages in queue will not be sent
func (connection *Connection) close() {
	if ConnStatusConnected != connection.status {
		return
	}

	connection.conn.Close()
	connection.status = ConnStatusDisconnected
}

// Close close tcp connection
func (connection *Connection) Close() {
	if connection.status != ConnStatusConnected {
		return
	}

	select {
	case connection.sendMsgQueue <- nil:
		{
			// nothing
		}
	case <-time.After(time.Duration(connection.sendTimeoutSec) * time.Second):
		{
			// timeout, close the connection
			connection.close()
			log.Printf("Conn send message timeout, close it")
		}
	}
}

func (connection *Connection) pushEvent(eventType int, data []byte) {
	if nil == connection.eventHandler {
		log.Println("nil event queue")
		return
	}
	connection.eventHandler(newConnEvent(eventType, connection, data))
}

// GetStatus get connection status
func (connection *Connection) GetStatus() int {
	return connection.status
}

func (connection *Connection) setStatus(status int) {
	connection.status = status
}

// SetReadTimeoutSec set read time out
func (connection *Connection) SetReadTimeoutSec(sec int) {
	connection.readTimeoutSec = sec
}

// GetReadTimeoutSec get read time out
func (connection *Connection) GetReadTimeoutSec() int {
	return connection.readTimeoutSec
}

// SetStreamProtocol set stream protocol interface
func (connection *Connection) SetStreamProtocol(streamProtocol IStreamProtocol) {
	connection.streamProtocol = streamProtocol
}

func (connection *Connection) sendRaw(msg []byte) {
	if connection.status != ConnStatusConnected {
		return
	}

	select {
	case connection.sendMsgQueue <- msg:
		{
			// nothing
		}
	case <-time.After(time.Duration(connection.sendTimeoutSec) * time.Second):
		{
			// timeout, close the connection
			connection.close()
			log.Printf("Conn send message timeout, close it")
		}
	}
}

// Send send data
func (connection *Connection) Send(msg []byte, needCopy bool) {
	if connection.status != ConnStatusConnected {
		return
	}

	buf := msg
	if needCopy {
		msgCopy := make([]byte, len(msg))
		copy(msgCopy, msg)
		buf = msgCopy
	}

	connection.sendRaw(buf)
}

// Run a routine to process connection connection
func (connection *Connection) Run() {
	go connection.routineMain()
}

func (connection *Connection) routineMain() {
	defer func() {
		// routine end
		log.Printf("Rotine of connection quit")
		e := recover()
		if e != nil {
			log.Println("Panic: ", e)
		}

		// close the connection
		connection.close()

		// free channel
		close(connection.sendMsgQueue)
		connection.sendMsgQueue = make(chan []byte, connection.sendBufferSize)

		// post event
		connection.pushEvent(ConnEventDisconnected, nil)
	}()

	if nil == connection.streamProtocol {
		log.Println("Nil stream protocol")
		return
	}
	connection.streamProtocol.Init()

	// connected
	connection.pushEvent(ConnEventConnected, nil)
	connection.status = ConnEventConnected

	go connection.routineSend()
	connection.routineRead()
}

func (connection *Connection) routineSend() error {
	defer func() {
		log.Println("Connection send log return")
	}()

	for {
		select {
		case sendMsg, ok := <-connection.sendMsgQueue:
			{
				if !ok {
					// channel closed, quit
					return nil
				}

				if nil == sendMsg {
					log.Println("User disconnect")
					connection.close()
					return nil
				}

				var err error
				headerBytes := connection.streamProtocol.SerializeHeader(sendMsg)
				if nil != headerBytes {
					// write header first
					_, err = connection.connRW.Write(headerBytes)
					if err != nil {
						log.Println("Conn write error:", err)
						return err
					}
					connection.connRW.Flush()
				}
				_, err = connection.connRW.Write(sendMsg)
				if err != nil {
					log.Println("Conn write error:", err)
					return err
				}
				connection.connRW.Flush()
			}
		}
	}
}

func (connection *Connection) routineRead() error {
	// default buffer
	buf := make([]byte, connection.maxReadBufferLength)

	for {
		data, err := connection.unpack(buf)
		if err != nil {
			log.Println("Conn read error: ", err)
			return err
		}

		connection.pushEvent(ConnEventData, data)
	}
}

func (connection *Connection) read(buf []byte) error {
	readLen := 0
	for readLen < len(buf) {
		len, err := connection.connRW.Read(buf[readLen:])
		if nil != err {
			return err
		}
		readLen += len
	}
	return nil
}

func (connection *Connection) unpack(buf []byte) ([]byte, error) {
	if 0 != connection.readTimeoutSec {
		connection.conn.SetReadDeadline(time.Now().Add(time.Duration(connection.readTimeoutSec) * time.Second))
	}

	bufStartPos := 0
	headerLength := connection.streamProtocol.GetHeaderLength()
	bodyLength := 0
	contextLength := 0
	headBuf := buf[bufStartPos:headerLength]
	err := connection.read(headBuf)
	if err != nil {
		return nil, err
	}
	bufStartPos += headerLength
	// check length
	var msg []byte
	packetLength, strEOF := connection.streamProtocol.UnserializeHeader(headBuf)
	if nil != strEOF {
		curBodyPos := 0
		// read body
		if 0 != connection.readTimeoutSec {
			connection.conn.SetReadDeadline(time.Now().Add(time.Duration(connection.readTimeoutSec) * time.Second))
		}
		for {
			if bufStartPos+curBodyPos+1 > connection.maxReadBufferLength {
				return nil, errors.New("The stream data is too long")
			}

			err = connection.read(buf[bufStartPos+curBodyPos : bufStartPos+curBodyPos+1])
			if err != nil {
				return nil, err
			}
			curBodyPos++
			if curBodyPos < len(strEOF) {
				continue
			}
			if bytes.Compare(strEOF, buf[bufStartPos+curBodyPos-len(strEOF):bufStartPos+curBodyPos]) == 0 {
				break
			}
		}
		bodyLength = curBodyPos
	} else if 0 != packetLength {
		if packetLength > connection.maxReadBufferLength {
			return nil, errors.New("The stream data is too long")
		}

		// read body
		if 0 != connection.readTimeoutSec {
			connection.conn.SetReadDeadline(time.Now().Add(time.Duration(connection.readTimeoutSec) * time.Second))
		}
		bodyLength = packetLength
		err = connection.read(buf[bufStartPos : bufStartPos+bodyLength])
		if err != nil {
			return nil, err
		}
	} else {
		allDataLen, _ := connection.connRW.Read(buf[bufStartPos:])
		bufStartPos += allDataLen
		msg = make([]byte, bufStartPos)
		copy(msg, buf[:bufStartPos])
		return msg, nil
	}

	bufStartPos += bodyLength

	contextLength = connection.streamProtocol.GetContentLength(buf[:bufStartPos])
	if 0 != contextLength {
		// read context
		if 0 != connection.readTimeoutSec {
			connection.conn.SetReadDeadline(time.Now().Add(time.Duration(connection.readTimeoutSec) * time.Second))
		}

		err = connection.read(buf[bufStartPos : bufStartPos+contextLength])
		if err != nil {
			return nil, err
		}
	}
	bufStartPos += contextLength

	msg = make([]byte, bufStartPos)
	copy(msg, buf[:bufStartPos])
	if 0 != connection.readTimeoutSec {
		connection.conn.SetReadDeadline(time.Time{})
	}

	return msg, nil
}
