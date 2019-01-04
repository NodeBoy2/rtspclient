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
	rtpSourceHandler IRtpParseInterface
	isMarkFrame      bool
}

func newRtpParser(codecName string) *RtpParser {
	return &RtpParser{
		payloadBuf:       make([]byte, MaxPayloadLength),
		payloadLen:       0,
		rtpSourceHandler: getRTPSourceHandler(codecName),
	}
}

func (rtpParser *RtpParser) splitRtpPacket(src []byte) ([]byte, []byte) {
	var header, payload []byte
	rtpParser.isMarkFrame, header, payload = rtpParser.rtpSourceHandler.SplitHeader(src)
	return header, payload
}

func (rtpParser *RtpParser) pushData(header []byte, payload []byte) ([]byte, int) {
	naluHeaderSize, naluSize := rtpParser.rtpSourceHandler.ParsingRtp(header, payload)
	if naluSize > len(payload)-naluHeaderSize {
		naluSize = len(payload)
	}
	nalu := payload[naluHeaderSize:naluSize]

	if rtpParser.payloadLen+len(nalu) > MaxPayloadLength {
		log.Print("playload too long")
		data := make([]byte, rtpParser.payloadLen)
		copy(data, rtpParser.payloadBuf[:rtpParser.payloadLen])

		rtpParser.payloadLen = 0
		if len(nalu) < MaxPayloadLength {
			copy(rtpParser.payloadBuf[rtpParser.payloadLen:], nalu)
			rtpParser.payloadLen += len(nalu)
		}
		return data, naluSize
	}

	copy(rtpParser.payloadBuf[rtpParser.payloadLen:], nalu)
	rtpParser.payloadLen += len(nalu)

	if rtpParser.isMarkFrame {
		data := make([]byte, rtpParser.payloadLen)
		copy(data, rtpParser.payloadBuf[:rtpParser.payloadLen])
		rtpParser.payloadLen = 0
		return data, naluSize
	}
	return nil, naluSize
}

func getRTPSourceHandler(codecName string) IRtpParseInterface {
	switch codecName {
	case "H264":
		{
			return &H264RtpParser{}
		}
	case "H265", "HEVC":
		{
			return &HevcRtpParser{}
		}
	case "bbw":
		{
			return &MarkRtpParser{}
		}
	default:
		{
			return &SimpleRtpParser{}
		}
	}
}
