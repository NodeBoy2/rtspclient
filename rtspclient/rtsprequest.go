package rtspclient

import (
	"errors"
	"fmt"
	"log"
)

const (
	sUserAgent = "None"
	sControlID = "trackID"
)

type RtspClientContext struct {
	rtspURL      string
	userAgent    string
	controlID    string
	sessionID    string
	setupHeaders string
	cseq         int
	bandwidth    int //bps
}

func NewRtspClientContext() *RtspClientContext {
	return &RtspClientContext{
		userAgent: sUserAgent,
		controlID: sControlID,
	}
}

func (session *RtspClientSession) sendRequst(request []byte) {
	log.Println(string(request))
	session.rtspRequestInitial = false
	session.tcpConn.Send([]byte(request), false)
	session.rtspContext.cseq++
}

func (session *RtspClientSession) SendDescribe() error {
	if !session.rtspRequestInitial {
		return errors.New("waiting last request Reply")
	}
	request := fmt.Sprintf(("DESCRIBE %s RTSP/1.0\r\n" +
		"CSeq: %d\r\n" +
		"Accept: application/sdp\r\n" +
		"User-agent: %s\r\n"), session.rtspContext.rtspURL, session.rtspContext.cseq, session.rtspContext.userAgent)

	if 0 != session.rtspContext.bandwidth {
		request += fmt.Sprintf("Bandwidth: %d\r\n", session.rtspContext.bandwidth)
	}
	request += "\r\n"
	session.sendRequst([]byte(request))
	return nil
}

func (session *RtspClientSession) SendTcpSetup(inTrackURL string, inClientRTPid int, inClientRTCPid int) error {
	if !session.rtspRequestInitial {
		return errors.New("waiting last request Reply")
	}

	request := fmt.Sprintf(("SETUP %s RTSP/1.0\r\n" +
		"CSeq: %d\r\n" +
		"Session: %s\r\n" +
		"Transport: RTP/AVP/TCP;unicast;interleaved=%d-%d\r\n" +
		"%s" +
		"User-agent: %s\r\n"), inTrackURL, session.rtspContext.cseq, session.rtspContext.sessionID, inClientRTPid, inClientRTCPid, session.rtspContext.setupHeaders, session.rtspContext.userAgent)

	if 0 != session.rtspContext.bandwidth {
		request += fmt.Sprintf("Bandwidth: %d\r\n", session.rtspContext.bandwidth)
	}
	request += "\r\n"
	session.sendRequst([]byte(request))

	return nil
}

func (session *RtspClientSession) SendPlay(inStartTimeSec int, inSpeed int) error {
	if !session.rtspRequestInitial {
		return errors.New("waiting last request Reply")
	}

	var strSpeed string
	if inSpeed != 1 {
		strSpeed = fmt.Sprintf("Speed: %f5.2\r\n", float32(inSpeed))
	}

	request := fmt.Sprintf(("PLAY %s RTSP/1.0\r\n" +
		"CSeq: %d\r\n" +
		"Session: %s\r\n" +
		"Range: npt=%d.0-\r\n" +
		"%s" +
		"x-prebuffer: maxtime=3.0\r\n" +
		"User-agent: %s\r\n"), session.rtspContext.rtspURL, session.rtspContext.cseq, session.rtspContext.sessionID, inStartTimeSec, strSpeed, session.rtspContext.userAgent)

	if 0 != session.rtspContext.bandwidth {
		request += fmt.Sprintf("Bandwidth: %d\r\n", session.rtspContext.bandwidth)
	}
	request += "\r\n"
	session.sendRequst([]byte(request))

	return nil
}

func (session *RtspClientSession) SendTeardown() error {
	if !session.rtspRequestInitial {
		return errors.New("waiting last request Reply")
	}

	request := fmt.Sprintf(("TEARDOWN %s RTSP/1.0\r\n" +
		"CSeq: %d\r\n" +
		"User-agent: %s\r\n"), session.rtspContext.rtspURL, session.rtspContext.cseq, session.rtspContext.userAgent)

	request += "\r\n"
	session.sendRequst([]byte(request))

	return nil
}
