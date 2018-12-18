package main

import (
	"log"
	"os"
	"rtspclient/rtspclient"
	"strconv"
	"sync"
	"time"
)

type MediaDataHandler interface {
	SetMediaSubsession(media rtspclient.MediaSubsession)
	GetHeader() []byte
	ParsingData([]byte) []byte
}

type RtspHandler struct {
	mediaHandler map[int]MediaDataHandler
	lock         sync.Mutex
}

func (handler *RtspHandler) WriteMediaHeader(session *rtspclient.RtspClientSession) {
	handler.lock.Lock()
	defer handler.lock.Unlock()
	handler.mediaHandler = make(map[int]MediaDataHandler)
	for index, media := range session.RtpMediaMap {
		file, err := os.OpenFile("D://test//"+strconv.Itoa(index), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModeExclusive)
		defer file.Close()
		if nil != err {
			log.Print(err)
			continue
		}

		if "H264" == media.CodecName {
			handler.mediaHandler[index] = &H264DataHandle{}
		} else if "H265" == media.CodecName {
			handler.mediaHandler[index] = &H265DataHandle{}
		} else {
			handler.mediaHandler[index] = &DefaultDataHandle{}
		}
		handler.mediaHandler[index].SetMediaSubsession(media)
		file.Write(handler.mediaHandler[index].GetHeader())
	}
}

func (handler *RtspHandler) RtspEventHandler(event *rtspclient.RtspEvent) {
	log.Print("event type: ", event.EventType)
	switch event.EventType {
	case rtspclient.RtspEventRequestSuccess:
		{
			handler.WriteMediaHeader(event.Session)
			break
		}
	}
}

func (handler *RtspHandler) RtpEventHandler(data *rtspclient.RtspData) {
	handler.lock.Lock()
	defer handler.lock.Unlock()
	// if 20 > len(data.Data) {
	// 	log.Printf("%x", data.Data)
	// } else {
	// 	log.Printf("%x", data.Data[:20])
	// }

	if len(handler.mediaHandler) < 0 {
		log.Println("media handler not init")
		return
	}

	file, _ := os.OpenFile("D://test//"+strconv.Itoa(data.ChannelNum), os.O_WRONLY|os.O_APPEND, os.ModeAppend)
	defer file.Close()

	// file.Write(handler.mediaHandler[data.ChannelNum].ParsingData(data.Data))
}

func CalculateFramerate(session *rtspclient.RtspClientSession) {
	lastVideoCout := int32(0)
	for {
		select {
		case <-time.NewTicker(1 * time.Second).C:
			log.Println("Current Frame Rate: ", session.VideoCout-lastVideoCout)
			lastVideoCout = session.VideoCout
		}
	}
}
func main() {
	rtspHandler := &RtspHandler{}
	rtspSession := rtspclient.NewRtspClientSession(rtspHandler.RtpEventHandler, rtspHandler.RtspEventHandler)
	err := rtspSession.Play("rtsp://192.168.10.50:10556/playback?serial=6072b43ec06d49f79c49febac8c64676&channel=67&starttime=20181212000000&endtime=20181212000400&isreduce=0")
	// err := rtspSession.Play("rtsp://192.168.1.247:10554/55c6516500514c8684c323ea60f59068?channel=7")
	// err := rtspSession.PlayUseWebsocket("ws://192.168.1.76:8080/websocket", "rtsp://admin:hk234567@192.168.10.103:554/Streaming/Channels/101?transportmode=unicast&profile=Profil_1")
	// err := rtspSession.Play("rtsp://fengyf:fengyf@192.168.10.113/cam/realmonitor?channel=1&subtype=0&unicast=true&proto=Onvif")
	// err := rtspSession.Play("rtsp://admin:hk234567@192.168.10.103:554/Streaming/Channels/101?transportmode=unicast&profile=Profil_1")
	// err := rtspSession.Play("rtsp://admin:Hk123456@192.168.10.107:554/Streaming/Channels/101?transportmode=unicast&profile=Profil_1")
	if nil != err {
		log.Print(err)
		return
	}

	err = rtspSession.SendPlay(0, 4)
	if nil != err {
		log.Print(err)
		return
	}
	go CalculateFramerate(rtspSession)

	response, errorInfo := rtspSession.WaitRtspResponse()
	if nil != errorInfo {
		log.Print(errorInfo)
	}
	if 200 != response.Status {
		log.Print("Send Play error")
		return
	}

	select {
	case <-time.After(50 * time.Second):
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
