package rtspclient

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gitlab.bmi/fyf/rtspclient.git/tcpnetwork"

	"github.com/gorilla/websocket"
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
	lastVideoTimestamp uint32
	lastAudioTimestamp uint32
	curVideoPts        uint32
	curAudioPts        uint32
	audioCout          int32
	VideoCout          int32
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
	header := data[:4]
	rtpData := data[4:]

	timestampBufBefore := rtpData[4:8]
	reader := bytes.NewReader(timestampBufBefore)
	var timestamp uint32
	err := binary.Read(reader, binary.BigEndian, &timestamp)
	if nil != err {
		return
	}

	// rtp data
	channelNum := int(header[1])
	rtpParser, ok := session.rtpChannelMap[channelNum]
	if ok {
		payload := rtpParser.pushData(rtpData)
		if nil != payload {
			session.dataHandle(newRtspData(channelNum, session, payload))

			if 0 == channelNum {
				if session.lastVideoTimestamp != timestamp {
					session.VideoCout++
					if session.lastVideoTimestamp != 0 {
						session.curVideoPts += ((timestamp - session.lastVideoTimestamp) / 90)
						if timestamp < session.lastVideoTimestamp {
							log.Println("video cout: ", session.VideoCout, " video pts: ", session.curVideoPts)
							log.Println(timestamp, ",", session.lastVideoTimestamp)
						}
					}
					session.lastVideoTimestamp = timestamp
				}
			} else if 2 == channelNum {
				if session.lastAudioTimestamp != timestamp {
					session.audioCout++
					if session.lastAudioTimestamp != 0 {
						session.curAudioPts += ((timestamp - session.lastAudioTimestamp) / 16)

						if timestamp < session.lastAudioTimestamp {
							log.Println("audio cout: ", session.audioCout, " audio pts: ", session.curAudioPts)
							log.Println(timestamp, ",", session.lastAudioTimestamp)
						}
					}
					session.lastAudioTimestamp = timestamp
				}
			}
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
	session.rtspContext.authenicator = nil

	session.SendDescribe()

	response, errorInfo = session.WaitRtspResponse()
	if nil != errorInfo {
		return errorInfo
	}
	if 401 == response.Status && session.username != "" && session.password != "" {
		if nil != response.digestAuthenticator {
			session.rtspContext.authenicator = response.digestAuthenticator
		} else if nil != response.basicAuthenticator {
			session.rtspContext.authenicator = response.basicAuthenticator
		}
		if nil != session.rtspContext.authenicator {
			session.rtspContext.authenicator.username = session.username
			session.rtspContext.authenicator.password = session.password
		}

		session.SendDescribe()
		response, errorInfo = session.WaitRtspResponse()
		if nil != errorInfo {
			return errorInfo
		}
	}

	if 200 != response.Status {
		return errors.New("response error: " + strconv.Itoa(response.Status))
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
		if 200 != response.Status {
			return errors.New("response error: " + strconv.Itoa(response.Status))
		}
		session.rtspContext.sessionID = response.sessionID
	}

	session.SendPlay(0, 1)
	response, errorInfo = session.WaitRtspResponse()
	if nil != errorInfo {
		return errorInfo
	}
	if 200 != response.Status {
		return errors.New("response error: " + strconv.Itoa(response.Status))
	}

	sendRequestSuccess = true
	return nil
}

func (session *RtspClientSession) ParsingURL(rtspURL string) error {
	urlInfo, urlError := url.Parse(rtspURL)
	if nil != urlError {
		return errors.New("url parse error: " + rtspURL)
	}

	if "" != urlInfo.RawQuery {
		session.rtspContext.rtspURL += ("?" + urlInfo.RawQuery)
	}
	session.address = urlInfo.Host
	if !strings.Contains(session.address, ":") {
		session.address += ":554"
	}

	if nil != urlInfo.User {
		session.username = urlInfo.User.Username()
		session.password, _ = urlInfo.User.Password()
	}

	urlInfo.User = nil
	session.rtspContext.rtspURL = urlInfo.String()
	return nil
}

func (session *RtspClientSession) SetConnection(conn net.Conn) error {
	session.tcpConn = tcpnetwork.NewConnection(conn, 0x0fff, session.pushConnEvent)
	session.tcpConn.SetStreamProtocol(session.rtpProtocol)
	session.tcpConn.Run()
	go session.HandleConn()

	return session.sendRequest()
}

func (session *RtspClientSession) Close() {
	if session.tcpConn.GetStatus() == tcpnetwork.ConnEventConnected {
		session.SendTeardown()
		session.WaitRtspResponse()
	}
	session.tcpConn.Close()
}

func (session *RtspClientSession) PlayUseWebsocket(webURL string, rtspURL string) error {
	urlError := session.ParsingURL(rtspURL)
	if nil != urlError {
		return errors.New("url parse error: " + rtspURL)
	}

	websocketConn, _, err := websocket.DefaultDialer.Dial(webURL, nil)
	if nil != err {
		session.sendEvent(RtspEventDisconnected, nil)
		log.Println("connect error: ", webURL)
		return err
	}

	conn := &WebsocketConn{conn: websocketConn}

	session.SetConnection(conn)
	return nil
}

func (session *RtspClientSession) Play(rtspURL string) error {

	if nil != session.tcpConn && tcpnetwork.ConnStatusConnected == session.tcpConn.GetStatus() {
		return errors.New("session is connected")
	}

	urlError := session.ParsingURL(rtspURL)
	if nil != urlError {
		return errors.New("url parse error: " + rtspURL)
	}

	// create connection
	conn, err := net.DialTimeout("tcp", session.address, time.Duration(session.timeoutSec)*time.Second)
	if nil != err {
		session.sendEvent(RtspEventDisconnected, nil)
		println(err.Error())
		println(session.address)
		return errors.New("connect " + session.address + "")
	}

	return session.SetConnection(conn)
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

}
