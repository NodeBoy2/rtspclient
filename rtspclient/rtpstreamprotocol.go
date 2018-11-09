package rtspclient

import (
	"bytes"
	"encoding/binary"
	"log"
)

type RTPStreamProtocol struct {
}

func (streamProtocol *RTPStreamProtocol) Init() {

}

func (streamProtocol *RTPStreamProtocol) GetHeaderLength() int {
	return 4
}

func (streamProtocol *RTPStreamProtocol) UnserializeHeader(header []byte) (int, []byte) {
	if header[0] == '$' {
		// rtp data
		lenBufBefore := header[2:4]
		reader := bytes.NewReader(lenBufBefore)
		var len uint16
		err := binary.Read(reader, binary.BigEndian, &len)
		if nil != err {
			return 0, nil
		}
		return int(len), nil
	} else if "rtsp" == string(header[0:4]) || "RTSP" == string(header[0:4]) {
		return 0, []byte("\r\n\r\n")
	} else {
		log.Print("header error")
		log.Print(header)
		return 0, nil
	}
}

func (streamProtocol *RTPStreamProtocol) SerializeHeader(data []byte) []byte {
	return nil
}

func (streamProtocol *RTPStreamProtocol) GetContentLength(body []byte) int {
	if "rtsp" != string(body[0:4]) && "RTSP" != string(body[0:4]) {
		return 0
	}
	rtspContext := &RtspResponseContext{}

	err := ParserRtspResponse(body, rtspContext)
	if nil != err {
		log.Print(err)
		return 0
	}
	return rtspContext.contentLength
}
