package main

import (
	"log"
	"net"
	"time"

	"gitlab.bmi/fyf/rtspclient.git/tcpnetwork"
)

type ClientEventQueue struct {
	eventQueue chan *tcpnetwork.ConnEvent
}

func (eventQueue *ClientEventQueue) Push(event *tcpnetwork.ConnEvent) {
	if nil == eventQueue.eventQueue {
		return
	}
	eventQueue.eventQueue <- event
}

func (eventQueue *ClientEventQueue) Pop() *tcpnetwork.ConnEvent {
	event, ok := <-eventQueue.eventQueue
	if !ok {
		// event queue already closed
		eventQueue.eventQueue = nil
		return nil
	}

	return event
}

func (eventQueue *ClientEventQueue) HandleConn() {
	defer func() {
		close(eventQueue.eventQueue)
	}()
	for {
		event := eventQueue.Pop()
		if nil == event {
			// channel closed, quit
			return
		}
		switch event.EventType {
		case tcpnetwork.ConnEventConnected:
			{
				log.Printf("conntion connected.")
			}
			break
		case tcpnetwork.ConnEventDisconnected:
			{
				log.Printf("conntion disconnected.")
				return
			}
		case tcpnetwork.ConnEventData:
			{
				if event.Data != nil {
					event.Conn.Send(event.Data, false)
					if string(event.Data) == "bye" {
						event.Conn.Close()
					}
				}
			}
			break
		case tcpnetwork.ConnEventClose:
			{
				log.Printf("connection closed")
				return
			}
		default:
			log.Printf("event type error.")
			break
		}
	}
}

type ClientStreamProtocol struct {
}

func (streamProtocol *ClientStreamProtocol) Init() {

}

func (streamProtocol *ClientStreamProtocol) GetHeaderLength() int {
	return 2
}

func (streamProtocol *ClientStreamProtocol) UnserializeHeader(header []byte) (int, []byte) {
	if header[0] == '0' {
		return int(header[1] - '0'), nil
	} else {
		return 0, []byte("eof")
	}

}

func (streamProtocol *ClientStreamProtocol) GetContentLength([]byte) int {
	return 3
}

func (streamProtocol *ClientStreamProtocol) SerializeHeader(data []byte) []byte {
	// header := make([]byte, 1)
	// header[0] = (byte)(len(data) + len(header))
	// header[0] += '0'
	// return header
	return nil
}

func main() {
	addr := "127.0.0.1:8080"
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if nil != err {
		log.Printf("connect " + addr + " error")
	}

	clientEvent := &ClientEventQueue{eventQueue: make(chan *tcpnetwork.ConnEvent)}
	connect := tcpnetwork.NewConnection(conn, 0x0fff, clientEvent.Push)

	clientProtocol := &ClientStreamProtocol{}
	connect.SetStreamProtocol(clientProtocol)
	connect.Run()
	clientEvent.HandleConn()
	log.Printf("connection over")
}
