package rtspclient

import (
	"errors"
	"log"
	"net"
	"rtspclient/tcpnetwork"
	"strconv"
	"strings"
	"time"
)

const (
	// RtspEventNone none event
	RtspEventNone = iota
	// RtspEventRequestSuccess request server success
	RtspEventRequestSuccess
	// RtspEventRequestError request server success
	RtspEventRequestError
	// RtspEventDisconnected disconnected event
	RtspEventDisconnected
)

// RtspEvent rtsp session event
type RtspEvent struct {
	EventType int
	Session   *RtspClientSession
	Data      []byte // data
}

// RtspData rtp data
type RtspData struct {
	ChannelNum int
	Session    *RtspClientSession
	Data       []byte
}

func newRtspEvent(eventType int, session *RtspClientSession, data []byte) *RtspEvent {
	return &RtspEvent{
		EventType: eventType,
		Session:   session,
		Data:      data,
	}
}

func newRtspData(channelNum int, session *RtspClientSession, data []byte) *RtspData {
	return &RtspData{
		ChannelNum: channelNum,
		Session:    session,
		Data:       data,
	}
}

func (session *RtspClientSession) sendEvent(eventType int, data []byte) {
	if nil == session.eventHandle {
		log.Println("nil event queue")
		return
	}
	session.eventHandle(newRtspEvent(eventType, session, data))
}

type RtspClientSession struct {
	username           string
	password           string
	address            string
	port               int
	rtspRequestInitial bool
	timeoutSec         int
	rtspContext        *RtspClientContext
	dataHandle         func(*RtspData)
	eventHandle        func(*RtspEvent)
	rtpProtocol        *RTPStreamProtocol
	tcpConn            *tcpnetwork.Connection
	eventQueue         chan *tcpnetwork.ConnEvent
	rtspResponseQueue  chan *RtspResponseContext
	rtpChannelMap      map[int]*RtpParser
	sdpInfo            *SDPInfo
	RtpMediaMap        map[int]MediaSubsession
}

func NewRtspClientSession(rtpHandler func(*RtspData), eventHandler func(*RtspEvent)) *RtspClientSession {
	return &RtspClientSession{
		dataHandle:         rtpHandler,
		eventHandle:        eventHandler,
		eventQueue:         make(chan *tcpnetwork.ConnEvent),
		rtspResponseQueue:  make(chan *RtspResponseContext),
		rtpProtocol:        &RTPStreamProtocol{},
		rtspContext:        NewRtspClientContext(),
		rtspRequestInitial: true,
		timeoutSec:         2,
		rtpChannelMap:      make(map[int]*RtpParser),
		RtpMediaMap:        make(map[int]MediaSubsession),
	}
}

func (session *RtspClientSession) pushConnEvent(event *tcpnetwork.ConnEvent) {
	if nil == session.eventQueue {
		return
	}
	session.eventQueue <- event
}

func (session *RtspClientSession) HandleConn() {
	defer func() {
		close(session.eventQueue)
		session.eventQueue = make(chan *tcpnetwork.ConnEvent)
		close(session.rtspResponseQueue)
		session.rtspResponseQueue = make(chan *RtspResponseContext)
	}()
	for {
		event, ok := <-session.eventQueue
		if !ok {
			// channel closed, quit
			return
		}
		if nil == event {
			// channel closed, quit
			event.Conn.Close()
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
				session.sendEvent(RtspEventDisconnected, nil)
				return
			}
		case tcpnetwork.ConnEventData:
			{
				if event.Data != nil {
					session.parsingData(event.Data)
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

func (session *RtspClientSession) parsingData(data []byte) {
	headerData := data[:4]
	if headerData[0] == '$' {
		session.parsingRtp(data)
	} else if string(headerData[:4]) == "RTSP" || string(headerData[:4]) == "rtsp" {
		session.parsingRtsp(data)
	}
}

func (session *RtspClientSession) parsingRtp(data []byte) {
	// rtp data
	channelNum := int(data[1])
	rtpParser, ok := session.rtpChannelMap[channelNum]
	if ok {
		payload := rtpParser.pushData(data[4:])
		if nil != payload {
			session.dataHandle(newRtspData(channelNum, session, payload))
		}
	}
}

func (session *RtspClientSession) parsingRtsp(data []byte) {
	log.Print(string(data))
	rtspResponseContext := &RtspResponseContext{}
	err := ParserRtspResponse(data, rtspResponseContext)
	if nil != err {
		session.rtspResponseQueue <- nil
		return
	}
	session.rtspResponseQueue <- rtspResponseContext
}

func (session *RtspClientSession) parsingURL(url string) {
	// Parse the URL as "rtsp://[<username>[:<password>]@]<server-address-or-name>[:<port>][/<path>]"
	headerLen := len("rtsp://")
	if headerLen >= len(url) {
		return
	}

	session.rtspContext.rtspURL = url
	session.username, session.password, session.address = "", "", ""
	session.port = 0

	containUser := strings.Contains(url, "@")
	addrStartPos := headerLen
	if containUser {
		// parse username and password
		userInfoEndPos := strings.IndexRune(url, '@')
		userInfo := string(url[headerLen:userInfoEndPos])
		usernameEndPos := strings.IndexRune(userInfo, ':')
		if -1 != usernameEndPos {
			session.password = string(userInfo[usernameEndPos+1:])
			session.username = string(userInfo[:usernameEndPos])
		} else {
			session.username = userInfo
		}
		addrStartPos = userInfoEndPos + 1
	}
	noUserURL := string(url[addrStartPos:])
	addrEndPos := strings.IndexRune(noUserURL, '/')
	var addrInfo string
	if -1 != addrEndPos {
		addrInfo = string(noUserURL[:addrEndPos])
	} else {
		addrInfo = noUserURL
	}
	ipEndPos := strings.IndexRune(addrInfo, ':')
	if -1 != ipEndPos {
		session.address = string(addrInfo[:ipEndPos])
		strPort := string(addrInfo[ipEndPos+1:])
		port, err := strconv.Atoi(strPort)
		if nil != err {
			log.Println("prot error.")
		} else {
			session.port = port
		}
	} else {
		session.address = addrInfo
	}
}

func (session *RtspClientSession) sendRequest() error {
	var response *RtspResponseContext
	var errorInfo error

	sendRequestSuccess := false

	defer func() {
		if sendRequestSuccess == false {
			session.sendEvent(RtspEventRequestError, nil)
		} else {
			session.sendEvent(RtspEventRequestSuccess, nil)
		}
	}()

	session.sdpInfo = nil
	session.SendDescribe()

	response, errorInfo = session.WaitRtspResponse()
	if nil != errorInfo {
		return errorInfo
	}
	if 200 != response.status {
		return errors.New("response error: " + strconv.Itoa(response.status))
	}

	session.sdpInfo = parsingSDP(response.content)
	if nil == session.sdpInfo {
		return errors.New("parse sdp error")
	}

	session.rtspContext.sessionID = ""
	for index, media := range session.sdpInfo.Medias {
		strTrackURL := media.TrackURL
		if len(strTrackURL) < 4 || ("rtsp" != string(strTrackURL[:4]) && "RTSP" != string(strTrackURL[:4])) {
			// if strings.Contains(session.rtspContext.rtspURL, "?") {
			// 	paramStartIndex := strings.Index(session.rtspContext.rtspURL, "?")
			// 	strTrackURL = session.rtspContext.rtspURL[:paramStartIndex] + "/" + strTrackURL + session.rtspContext.rtspURL[paramStartIndex:]
			// } else {
			// 	strTrackURL = session.rtspContext.rtspURL + "/" + strTrackURL
			// }
			strTrackURL = session.rtspContext.rtspURL + "/" + strTrackURL
		}
		rtpIndex := index * 2
		rtcpIndex := index*2 + 1
		session.rtpChannelMap[rtpIndex] = newRtpParser(media.CodecName)
		session.RtpMediaMap[rtpIndex] = media
		session.SendTcpSetup(strTrackURL, rtpIndex, rtcpIndex)

		response, errorInfo = session.WaitRtspResponse()
		if nil != errorInfo {
			return errorInfo
		}
		if 200 != response.status {
			return errors.New("response error: " + strconv.Itoa(response.status))
		}
		session.rtspContext.sessionID = response.sessionID
	}

	session.SendPlay(0, 1)
	response, errorInfo = session.WaitRtspResponse()
	if nil != errorInfo {
		return errorInfo
	}
	if 200 != response.status {
		return errors.New("response error: " + strconv.Itoa(response.status))
	}

	sendRequestSuccess = true
	return nil
}

func (session *RtspClientSession) Close() {
	if session.tcpConn.GetStatus() == tcpnetwork.ConnEventConnected {
		session.SendTeardown()
		session.WaitRtspResponse()
	}
	session.tcpConn.Close()
}

func (session *RtspClientSession) Play(url string) error {

	if nil != session.tcpConn && tcpnetwork.ConnStatusConnected == session.tcpConn.GetStatus() {
		return errors.New("session is connected")
	}

	session.parsingURL(url)
	// create connection
	tcpAddr := session.address
	if 0 != session.port {
		tcpAddr += ":"
		tcpAddr += strconv.Itoa(session.port)
	}

	conn, err := net.DialTimeout("tcp", tcpAddr, time.Duration(session.timeoutSec)*time.Second)
	if nil != err {
		session.sendEvent(RtspEventDisconnected, nil)
		println(err.Error())
		println(tcpAddr)
		return errors.New("connect " + tcpAddr + "")
	}

	session.tcpConn = tcpnetwork.NewConnection(conn, 0x0fff, session.pushConnEvent)
	session.tcpConn.SetStreamProtocol(session.rtpProtocol)
	session.tcpConn.Run()
	go session.HandleConn()

	return session.sendRequest()
}

func (session *RtspClientSession) WaitRtspResponse() (*RtspResponseContext, error) {
	defer func() {
		session.rtspRequestInitial = true
	}()

	select {
	case response, ok := <-session.rtspResponseQueue:
		{
			if !ok {
				return nil, errors.New("rtsp connection disconnect")
			}
			if 200 != response.status {
				return nil, errors.New("rtsp response error: " + strconv.Itoa(response.status))
			}
			if nil == response {
				return nil, errors.New("parsing rtsp response error")
			}
			return response, nil
		}
	case <-time.After(time.Duration(session.timeoutSec) * time.Second):
		{
			session.Close()
			return nil, errors.New("recv response time out")
		}
	}

	return nil, nil
}
