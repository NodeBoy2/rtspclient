package rtspclient

type HevcRtpParser struct {
}

func (rtpParser *HevcRtpParser) SplitHeader(src []byte) (isCompletesFrame bool, header []byte, payload []byte) {
	if len(src) < RtpHeaderLen+2 {
		return false, src[:RtpHeaderLen], src[RtpHeaderLen:]
	}

	rtpData := src[RtpHeaderLen:]
	packetNALUnitType := (rtpData[0] & 0x7E) >> 1

	var skipHeaderLen int
	switch packetNALUnitType {
	case 48: // Aggregation Packet (AP)
		{
			// We skip over the 2-byte Payload Header, and the DONL header (if any).
			skipHeaderLen = 2
			isCompletesFrame = true
		}
		break
	case 49: // Fragmentation Unit (FU)
		{
			// This NALU begins with the 2-byte Payload Header, the 1-byte FU header, and (optionally)
			// the 2-byte DONL header.
			// If the start bit is set, we reconstruct the original NAL header at the end of these
			// 3 (or 5) bytes, and skip over the first 1 (or 3) bytes.
			startBit := rtpData[2] & 0x80
			endBit := rtpData[2] & 0x40
			if 0 != startBit {
				nalUnitType := rtpData[2] & 0x3F
				newNalHeader := make([]byte, 2)
				newNalHeader[0] = (rtpData[0] & 0x81) | (nalUnitType << 1)
				newNalHeader[1] = rtpData[1]

				rtpData[1] = newNalHeader[0]
				rtpData[2] = newNalHeader[1]

				skipHeaderLen = 1
			} else {
				skipHeaderLen = 3
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

func (rtpParser *HevcRtpParser) ParsingRtp(header []byte, payload []byte) (naluHeaderSize int, naluSize int) {
	if len(header) < RtpHeaderLen+2 {
		return 0, len(payload)
	}

	naluHeader := header[RtpHeaderLen:]
	packetNALUnitType := (naluHeader[0] & 0x7E) >> 1

	switch packetNALUnitType {
	case 48: // Aggregation Packet (AP)
		{
			// We skip over the 2-byte Payload Header, and the DONL header (if any).
			naluSize = int((payload[0] << 8) | payload[1])
			naluHeaderSize = 2
		}
		break
	default:
		{
			naluSize = len(payload)
			naluHeaderSize = 0
		}
	}
	return naluHeaderSize, naluSize + naluHeaderSize
}
