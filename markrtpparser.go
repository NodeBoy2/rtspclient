package rtspclient

type MarkRtpParser struct {
}

func (rtpParser *MarkRtpParser) SplitHeader(src []byte) (isCompletesFrame bool, header []byte, payload []byte) {
	if len(src) < 2 {
		return true, src[:RtpHeaderLen], src[RtpHeaderLen:]
	}

	mark := ((header[1] & 0x80) >> 7)
	isCompletesFrame = (mark == 1)

	return isCompletesFrame, src[:RtpHeaderLen], src[RtpHeaderLen:]
}

func (rtpParser *MarkRtpParser) ParsingRtp(header []byte, payload []byte) (naluHeaderSize int, naluSize int) {
	return 0, len(payload)
}
