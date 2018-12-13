package main

import (
	"encoding/base64"
	"rtspclient/rtspclient"
)

type H265DataHandle struct {
	byteVps []byte
	byteSps []byte
	bytePps []byte
	gotIDR  bool
}

func (dataHandler *H265DataHandle) SetMediaSubsession(media rtspclient.MediaSubsession) {
	vpsInfo := media.Fmtp["sprop-vps"]
	spsInfo := media.Fmtp["sprop-sps"]
	ppsInfo := media.Fmtp["sprop-pps"]

	dataHandler.byteVps, _ = base64.StdEncoding.DecodeString(vpsInfo)
	dataHandler.byteSps, _ = base64.StdEncoding.DecodeString(spsInfo)
	dataHandler.bytePps, _ = base64.StdEncoding.DecodeString(ppsInfo)
	dataHandler.gotIDR = false
}

func (dataHandler *H265DataHandle) GetHeader() []byte {
	headerByte := []byte{0x00, 0x00, 0x00, 0x01}
	header := make([]byte, 0)
	header = append(header, headerByte...)
	header = append(header, dataHandler.byteVps...)
	header = append(header, headerByte...)
	header = append(header, dataHandler.byteSps...)
	header = append(header, headerByte...)
	header = append(header, dataHandler.bytePps...)
	return header
}

func (dataHandler *H265DataHandle) addHeader(data []byte) bool {
	var code int32
	code = -1
	code = (code << 8) + int32(data[0])
	naluType := (code & 0x7e) >> 1

	isIDR := false
	switch naluType {
	case 16, 17, 18, 19, 20, 21, 22, 23:
		isIDR = true
		break
	default:
		dataHandler.gotIDR = false
		break
	}
	addHeader := isIDR && !dataHandler.gotIDR
	dataHandler.gotIDR = isIDR || dataHandler.gotIDR
	return addHeader
}

func (dataHandler *H265DataHandle) ParsingData(src []byte) []byte {
	headerByte := []byte{0x00, 0x00, 0x00, 0x01}
	dst := make([]byte, 0)
	if dataHandler.addHeader(src) {
		dst = append(dst, dataHandler.GetHeader()...)
	}
	dst = append(dst, headerByte...)
	dst = append(dst, src...)
	return dst
}
