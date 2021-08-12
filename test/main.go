package main

import (
	"fmt"
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
			fmt.Print(err)
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
	fmt.Print("event type: ", event.EventType)
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
		fmt.Printf("%x", data.Data)
	} else {
		fmt.Printf("%x", data.Data[:20])
	}

	if len(handler.mediaHandler) < 0 {
		fmt.Println("media handler not init")
		return
	}

	file, _ := os.OpenFile("D://test//"+strconv.Itoa(data.ChannelNum), os.O_WRONLY|os.O_APPEND, os.ModeAppend)
	defer file.Close()

	file.Write(handler.mediaHandler[data.ChannelNum].ParsingData(data.Data))
}

func main() {
	rtspHandler := &RtspHandler{}
	rtspSession := rtspclient.NewRtspClientSession(rtspHandler.RtpEventHandler, rtspHandler.RtspEventHandler)
	err := rtspSession.Play("rtsp://192.168.1.103:10556/planback?channel=1&starttime=20181228134420&endtime=20181228151000&isreduce=0")
	if nil != err {
		fmt.Print(err)
		return
	}
	err = rtspSession.SendPlay(0, 2)
	if err != nil {
		fmt.Print(err)
	}
	rtspSession.WaitRtspResponse()

	select {
	case <-time.After(50 * time.Second):
		{
			fmt.Print("time over")
		}
	}

	rtspSession.Close()

	select {
	case <-time.After(1 * time.Second):
		{
			fmt.Print("time over")
		}
	}
	fmt.Print("main() disconnect.")
}
