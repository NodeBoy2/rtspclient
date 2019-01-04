package main

import (
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/NodeBoy2/rtspclient"
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
	if 20 > len(data.Data) {
		log.Printf("%x", data.Data)
	} else {
		log.Printf("%x", data.Data[:20])
	}

	if len(handler.mediaHandler) < 0 {
		log.Println("media handler not init")
		return
	}

	file, _ := os.OpenFile("D://test//"+strconv.Itoa(data.ChannelNum), os.O_WRONLY|os.O_APPEND, os.ModeAppend)
	defer file.Close()

	file.Write(handler.mediaHandler[data.ChannelNum].ParsingData(data.Data))
}

func main() {
	rtspHandler := &RtspHandler{}
	rtspSession := rtspclient.NewRtspClientSession(rtspHandler.RtpEventHandler, rtspHandler.RtspEventHandler)
	// err := rtspSession.Play("rtsp://192.168.10.50:10556/playback?serial=6072b43ec06d49f79c49febac8c64676&channel=67&starttime=20181212000000&endtime=20181212000400&isreduce=0")
	// err := rtspSession.Play("rtsp://192.168.1.247:10554/55c6516500514c8684c323ea60f59068?channel=7")
	// err := rtspSession.PlayUseWebsocket("ws://192.168.1.76:8080/websocket", "rtsp://admin:hk234567@192.168.10.103:554/Streaming/Channels/101?transportmode=unicast&profile=Profil_1")
	err := rtspSession.Play("rtsp://192.168.1.103:10556/planback?channel=1&starttime=20181228134420&endtime=20181228151000&isreduce=0")
	// err := rtspSession.Play("rtsp://192.168.1.186:10554/dahua-55c6516500514c8684c323ea60f59068-2?channel=1")
	// err := rtspSession.Play("rtsp://192.168.1.233:10554/fbc52b8f4d5b4114bf2289ed6e334b85?channel=2")
	// err := rtspSession.Play("rtsp://192.168.1.180:10554/8b8351e227084952b7ebc357e4d72cdd?channel=2")
	if nil != err {
		log.Print(err)
		return
	}

	count := 0
	for count < 5 {
		select {
		case <-time.After(5 * time.Second):
			{
				log.Print("time over")
			}
		}

		err = rtspSession.SendPause()
		if nil != err {
			log.Print(err)
			return
		}

		response, errorInfo := rtspSession.WaitRtspResponse()
		if nil != errorInfo {
			log.Print(errorInfo)
		}
		if 200 != response.Status {
			log.Print("Send Pause error")
			return
		}

		select {
		case <-time.After(3 * time.Second):
			{
				log.Print("time over")
			}
		}

		err = rtspSession.SendPlay(0, 4)
		if nil != err {
			log.Print(err)
			return
		}

		response, errorInfo = rtspSession.WaitRtspResponse()
		if nil != errorInfo {
			log.Print(errorInfo)
		}
		if 200 != response.Status {
			log.Print("Send Play error")
			return
		}
		count++
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
