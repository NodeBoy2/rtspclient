package rtspclient

type SimpleRtpParser struct {
}

func (rtpParser *SimpleRtpParser) SplitHeader(src []byte) (isCompletesFrame bool, header []byte, payload []byte) {
	if len(src) < 2 {
		return true, src[:RtpHeaderLen], src[RtpHeaderLen:]
	}

	return true, src[:RtpHeaderLen], src[RtpHeaderLen:]
}

func (rtpParser *SimpleRtpParser) ParsingRtp(header []byte, payload []byte) (naluHeaderSize int, naluSize int) {
	return 0, len(payload)
}
