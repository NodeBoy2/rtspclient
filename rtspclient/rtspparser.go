package rtspclient

import (
	"errors"
	"strconv"
	"strings"
)

type RtspResponseContext struct {
	status        int
	contentLength int
	content       string
	sessionID     string
}

const (
	sSessionHeader      = "Session"
	sContentLenHeader   = "Content-length"
	sTransportHeader    = "Transport"
	sRTPInfoHeader      = "RTP-Info"
	sRTPMetaInfoHeader  = "x-RTP-Meta-Info"
	sAuthenticateHeader = "WWW-Authenticate"
	sSameAsLastHeader   = " ,"
	sPublic             = "Public"
)

func ParserRtspResponse(response []byte, context *RtspResponseContext) error {
	// rtp packet no context
	if string(response[:4]) != "RTSP" && string(response[:4]) != "rtsp" {
		return errors.New("not rtsp packet")
	}

	context.status = 0
	context.contentLength = 0
	context.content = ""

	rtspEndPos := strings.Index(string(response), "\r\n\r\n") // get rtsp response, remove content.
	if rtspEndPos == -1 {
		return errors.New("no eof flag")
	}
	rtspEndPos += len("\r\n\r\n") // add \r\n\r\n len

	rtspResponse := response[:rtspEndPos]

	// get status code
	strStatus := string(response[len("RTSP/1.0 ") : len("RTSP/1.0 ")+3]) // skip past RTSP/1.0
	context.status, _ = strconv.Atoi(strStatus)

	fields := strings.Split(string(rtspResponse), "\r\n")
	for _, field := range fields {
		keyValue := strings.Split(field, ":")
		if 2 != len(keyValue) {
			continue
		}
		theKey := strings.ToUpper(keyValue[0])
		theKey = strings.Replace(theKey, " ", "", -1)
		theValue := keyValue[1]
		if theKey == strings.ToUpper(sContentLenHeader) {
			theValue = strings.Replace(theValue, " ", "", -1)
			context.contentLength, _ = strconv.Atoi(theValue)
			if len(response)-rtspEndPos >= context.contentLength {
				context.content = string(response[rtspEndPos : rtspEndPos+context.contentLength])
			}
		} else if theKey == strings.ToUpper(sSessionHeader) {
			context.sessionID = theValue
		}
	}
	return nil
}
