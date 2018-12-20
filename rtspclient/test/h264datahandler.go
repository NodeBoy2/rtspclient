package main

import (
	"encoding/base64"
	"strings"

	"gitlab.bmi/fyf/rtspclient.git/rtspclient"
)

type H264DataHandle struct {
	byteSps []byte
	bytePps []byte
}

func (dataHandler *H264DataHandle) SetMediaSubsession(media rtspclient.MediaSubsession) {
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

	dataHandler.byteSps, _ = base64.StdEncoding.DecodeString(sps)
	dataHandler.bytePps, _ = base64.StdEncoding.DecodeString(pps)
}

func (dataHandler *H264DataHandle) GetHeader() []byte {
	headerByte := []byte{0x00, 0x00, 0x00, 0x01}
	header := make([]byte, 0)
	header = append(header, headerByte...)
	header = append(header, dataHandler.byteSps...)
	header = append(header, headerByte...)
	header = append(header, dataHandler.bytePps...)
	return header
}

func (dataHandler *H264DataHandle) ParsingData(src []byte) []byte {
	headerByte := []byte{0x00, 0x00, 0x00, 0x01}
	dst := append(headerByte, src...)
	return dst
}
