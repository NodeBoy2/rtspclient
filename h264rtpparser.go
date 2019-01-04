package rtspclient

type H264RtpParser struct {
}

func (rtpParser *H264RtpParser) SplitHeader(src []byte) (isCompletesFrame bool, header []byte, payload []byte) {
	if len(src) < RtpHeaderLen+1 {
		return true, src[:RtpHeaderLen], src[RtpHeaderLen:]
	}
	rtpData := src[RtpHeaderLen:]
	packetNALUnitType := (rtpData[0] & 0x1f)
	if packetNALUnitType > 1 && packetNALUnitType <= 23 {
		packetNALUnitType = 1
	}

	var skipHeaderLen int
	switch packetNALUnitType {
	case 0, 1: // undefined, but pass them through
		{
			skipHeaderLen = 0
			isCompletesFrame = true
		}
		break
	case 24: // STAP-A
		{
			skipHeaderLen = 1
			isCompletesFrame = true
		}
		break
	case 25, 26, 27: // STAP-B MTAP16 MTAP24
		{
			skipHeaderLen = 3
			isCompletesFrame = true
		}
		break
	case 28, 29: // FU-A // FU-B
		{
			// For these NALUs, the first two bytes are the FU indicator and the FU header.
			// If the start bit is set, we reconstruct the original NAL header into byte 1:
			startBit := rtpData[1] & 0x80
			endBit := rtpData[1] & 0x40
			if 0 != startBit {
				// reset nal header
				rtpData[1] = (rtpData[0] & 0xe0) | (rtpData[1] & 0x1f)
				skipHeaderLen = 1
			} else {
				skipHeaderLen = 2
			}
			isCompletesFrame = (endBit != 0)
		}
		break
	default:
		{
			skipHeaderLen = 0
			isCompletesFrame = true
		}
	}

	return isCompletesFrame, src[:skipHeaderLen+RtpHeaderLen], src[skipHeaderLen+RtpHeaderLen:]
}
func (rtpParser *H264RtpParser) ParsingRtp(header []byte, payload []byte) (naluHeaderSize int, naluSize int) {
	if len(header) < RtpHeaderLen+2 {
		return 0, len(payload)
	}

	naluHeader := header[RtpHeaderLen:]
	packetNALUnitType := (naluHeader[0] & 0x7E) >> 1

	switch packetNALUnitType {
	case 24, 25: // STAP-A STAP-B
		{
			naluSize = int((payload[0] << 8) | naluHeader[1])
			naluHeaderSize = 2
		}
		break
	case 26: // MTAP16
		{
			naluSize = int((naluHeader[3] << 8) | naluHeader[4])
			naluHeaderSize = 5
		}
		break
	case 27: // MTAP24
		{
			naluSize = int((naluHeader[3] << 8) | naluHeader[4])
			naluHeaderSize = 6
		}
		break
	default:
		{
			naluSize = len(payload)
			naluHeaderSize = 0
		}
	}
	return naluHeaderSize, naluSize
}
