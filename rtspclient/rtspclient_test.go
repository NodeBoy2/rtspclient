package rtspclient

import (
	"log"
	"strconv"
	"testing"
)

func CreateSession() *RtspClientSession {
	return &RtspClientSession{
		rtspContext: NewRtspClientContext(),
	}
}
func Test_parsingUrl_1(t *testing.T) {
	session := CreateSession()
	rtspURL := "rtsp://192.168.1.111"
	session.parsingURL(rtspURL)
	log.Print(session.address + ", " + strconv.Itoa(session.port) + ", " + session.username + ", " + session.password)
	if session.address == "192.168.1.111" && session.port == 0 && session.username == "" && session.password == "" {
		t.Log("_1 pass")
	} else {
		t.Error("_1 error")
	}
}

func Test_parsingUrl_2(t *testing.T) {
	session := CreateSession()
	rtspURL := "rtsp://www.baidu.com/test/1112.d"
	session.parsingURL(rtspURL)
	log.Print(session.address + ", " + strconv.Itoa(session.port) + ", " + session.username + ", " + session.password)
	if session.address == "www.baidu.com" && session.port == 0 && session.username == "" && session.password == "" {
		t.Log("_2 pass")
	} else {
		t.Error("_2 error")
	}
}
func Test_parsingUrl_3(t *testing.T) {
	session := CreateSession()
	rtspURL := "rtsp://192.168.1.111:10554/test"
	session.parsingURL(rtspURL)
	log.Print(session.address + ", " + strconv.Itoa(session.port) + ", " + session.username + ", " + session.password)
	if session.address == "192.168.1.111" && session.port == 10554 && session.username == "" && session.password == "" {
		t.Log("_3 pass")
	} else {
		t.Error("_3 error")
	}
}
func Test_parsingUrl_4(t *testing.T) {
	session := CreateSession()
	rtspURL := "rtsp://user@192.168.1.111:10554"
	session.parsingURL(rtspURL)
	log.Print(session.address + ", " + strconv.Itoa(session.port) + ", " + session.username + ", " + session.password)
	if session.address == "192.168.1.111" && session.port == 10554 && session.username == "user" && session.password == "" {
		t.Log("_4 pass")
	} else {
		t.Error("_4 error")
	}
}
func Test_parsingUrl_5(t *testing.T) {
	session := CreateSession()
	rtspURL := "rtsp://@192.168.1.111:10554/111"
	session.parsingURL(rtspURL)
	log.Print(session.address + ", " + strconv.Itoa(session.port) + ", " + session.username + ", " + session.password)
	if session.address == "192.168.1.111" && session.port == 10554 && session.username == "" && session.password == "" {
		t.Log("_5 pass")
	} else {
		t.Error("_5 error")
	}
}
func Test_parsingUrl_6(t *testing.T) {
	session := CreateSession()
	rtspURL := "rtsp://user:pass@192.168.1.111:10554/111"
	session.parsingURL(rtspURL)
	log.Print(session.address + ", " + strconv.Itoa(session.port) + ", " + session.username + ", " + session.password)
	if session.address == "192.168.1.111" && session.port == 10554 && session.username == "user" && session.password == "pass" {
		t.Log("_6 pass")
	} else {
		t.Error("_6 error")
	}
}
