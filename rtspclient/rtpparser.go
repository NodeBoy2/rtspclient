package rtspclient

import (
	"log"
)

const (
	MaxPayloadLength = 8 * 1024 * 1024
)

type RtpParser struct {
	payloadBuf       []byte
	payloadLen       int
	rtpSourceHandler func([]byte) (bool, int)
}

func newRtpParser(codecName string) *RtpParser {
	return &RtpParser{
		payloadBuf:       make([]byte, MaxPayloadLength),
		payloadLen:       0,
		rtpSourceHandler: getRTPSourceHandler(codecName),
	}
}

func (rtpParser *RtpParser) pushData(src []byte) []byte {
	isCompletersFrame, skipHeaderLen := rtpParser.rtpSourceHandler(src)
	payloadData := src[skipHeaderLen:]

	if rtpParser.payloadLen+len(payloadData) > MaxPayloadLength {
		log.Print("playload too long")
		data := make([]byte, rtpParser.payloadLen)
		copy(data, rtpParser.payloadBuf[:rtpParser.payloadLen])

		rtpParser.payloadLen = 0
		if len(payloadData) < MaxPayloadLength {
			copy(rtpParser.payloadBuf[rtpParser.payloadLen:rtpParser.payloadLen+len(payloadData)], payloadData)
			rtpParser.payloadLen += len(payloadData)
		}
		return data
	}

	copy(rtpParser.payloadBuf[rtpParser.payloadLen:rtpParser.payloadLen+len(payloadData)], payloadData)
	rtpParser.payloadLen += len(payloadData)

	if isCompletersFrame {
		data := make([]byte, rtpParser.payloadLen)
		copy(data, rtpParser.payloadBuf[:rtpParser.payloadLen])
		rtpParser.payloadLen = 0
		return data
	}
	return nil
}

func getRTPSourceHandler(codecName string) func([]byte) (bool, int) {
	switch codecName {
	case "H264":
		{
			return rtpSourceH264
		}
	case "H265", "HEVC":
		{
			return rtpSourceHevc
		}
	case "bbw":
		{
			return rtpSourceMark
		}
	default:
		{
			return rtpSourceSimple
		}
	}
}
