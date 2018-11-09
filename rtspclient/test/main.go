package main

import (
	"encoding/base64"
	"log"
	"os"
	"rtspclient/rtspclient"
	"strconv"
	"strings"
	"time"
)

func writeH264Header(file *os.File, media rtspclient.MediaSubsession) {
	spsInfo := media.Fmtp["sprop-parameter-sets"]
	if "" == spsInfo {
		return
	}
	spsPps := strings.Split(spsInfo, ",")
	var sps, pps string
	if len(spsPps) == 1 {
		sps = spsPps[0]
	} else if len(spsPps) == 2 {
		sps = spsPps[0]
		pps = spsPps[1]
	}

	byteSps, _ := base64.StdEncoding.DecodeString(sps)
	bytePps, _ := base64.StdEncoding.DecodeString(pps)
	headerByte := []byte{0x00, 0x00, 0x00, 0x01}
	file.Write(headerByte)
	file.Write(byteSps)
	file.Write(headerByte)
	file.Write(bytePps)
}

func writeH265Header(file *os.File, media rtspclient.MediaSubsession) {
	vpsInfo := media.Fmtp["sprop-vps"]
	spsInfo := media.Fmtp["sprop-sps"]
	ppsInfo := media.Fmtp["sprop-pps"]

	byteVps, _ := base64.StdEncoding.DecodeString(vpsInfo)
	byteSps, _ := base64.StdEncoding.DecodeString(spsInfo)
	bytePps, _ := base64.StdEncoding.DecodeString(ppsInfo)
	headerByte := []byte{0x00, 0x00, 0x00, 0x01}
	file.Write(headerByte)
	file.Write(byteVps)
	file.Write(headerByte)
	file.Write(byteSps)
	file.Write(headerByte)
	file.Write(bytePps)
}

func WriteMediaHeader(session *rtspclient.RtspClientSession) {
	for index, media := range session.RtpMediaMap {
		file, err := os.OpenFile("D://test//"+strconv.Itoa(index), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModeExclusive)
		defer file.Close()
		if nil != err {
			log.Print(err)
			continue
		}

		if "H264" == media.CodecName {
			writeH264Header(file, media)
		} else if "H265" == media.CodecName {
			writeH265Header(file, media)
		}
	}
}

func RtspEventHandler(event *rtspclient.RtspEvent) {
	log.Print("event type: ", event.EventType)
	switch event.EventType {
	case rtspclient.RtspEventRequestSuccess:
		{
			WriteMediaHeader(event.Session)
			break
		}
	}
}

func RtpEventHandler(data *rtspclient.RtspData) {
	if 20 > len(data.Data) {
		log.Printf("%x", data.Data)
	} else {
		log.Printf("%x", data.Data[:20])
	}

	file, _ := os.OpenFile("D://test//"+strconv.Itoa(data.ChannelNum), os.O_WRONLY|os.O_APPEND, os.ModeAppend)
	defer file.Close()

	if "H264" == data.Session.RtpMediaMap[data.ChannelNum].CodecName {
		headerByte := []byte{0x00, 0x00, 0x00, 0x01}
		file.Write(headerByte)
	} else if "H265" == data.Session.RtpMediaMap[data.ChannelNum].CodecName {
		if 0x26 == data.Data[0] {
			writeH265Header(file, data.Session.RtpMediaMap[data.ChannelNum])
		}
		headerByte := []byte{0x00, 0x00, 0x00, 0x01}
		file.Write(headerByte)
	}
	file.Write(data.Data)
}

func main() {
	rtspSession := rtspclient.NewRtspClientSession(RtpEventHandler, RtspEventHandler)
	err := rtspSession.Play("rtsp://103.60.165.57:10554/A8BE16020167?channel=0")
	// err := rtspSession.Play("rtsp://fengyf:fengyf@192.168.10.113:554/cam/realmonitor?channel=1&subtype=0&unicast=true&proto=Onvif")
	// err := rtspSession.Play("rtsp://admin:hk234567@192.168.10.103:554/Streaming/Channels/101?transportmode=unicast&profile=Profil_1")
	if nil != err {
		log.Print(err)
	}

	select {
	case <-time.After(30 * time.Second):
		{
			log.Print("time over")
		}
	}

	rtspSession.Close()

	select {
	case <-time.After(1 * time.Second):
		{
			log.Print("time over")
		}
	}
	log.Print("main() disconnect.")
}
