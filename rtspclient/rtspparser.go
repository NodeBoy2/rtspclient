package rtspclient

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type RtspResponseContext struct {
	Status              int
	contentLength       int
	content             string
	sessionID           string
	basicAuthenticator  *Authenticator
	digestAuthenticator *Authenticator
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

	context.Status = 0
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
	context.Status, _ = strconv.Atoi(strStatus)

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
		} else if theKey == strings.ToUpper(sAuthenticateHeader) {
			authenticateType, realm, nonce, _ := parsingAuthenticate(theValue)
			if "Digest" == authenticateType {
				context.digestAuthenticator = NewAuthenticator()
				context.digestAuthenticator.realm = realm
				context.digestAuthenticator.nonce = nonce
				context.digestAuthenticator.authenicatorType = AuthenticatorTypeDigest
			} else if "Basic" == authenticateType {
				context.basicAuthenticator = NewAuthenticator()
				context.basicAuthenticator.realm = realm
				context.basicAuthenticator.authenicatorType = AuthenticatorTypeBasic
			}
		}
	}
	return nil
}

func parsingAuthenticate(authenticateValue string) (string, string, string, string) {
	authenticateValue = strings.Replace(authenticateValue, "\"", " ", -1)
	var realm string
	var nonce string
	var stale string
	var authenticateType string

	fmt.Sscanf(authenticateValue, " %s realm= %s , nonce= %s , stale= %s ", &authenticateType, &realm, &nonce, &stale)
	return authenticateType, realm, nonce, stale
}
