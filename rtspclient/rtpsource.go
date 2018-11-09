package rtspclient

const (
	PacketHeaderLen = 4
	RtpHeaderLen    = 12
)

func rtpSourceSimple([]byte) (bool, int) {
	return true, RtpHeaderLen
}

func rtpSourceMark(src []byte) (bool, int) {
	if len(src) < 2 {
		return true, RtpHeaderLen
	}
	mark := ((src[1] & 0x80) >> 7)
	isCompletesFrame := (mark == 1)

	return isCompletesFrame, RtpHeaderLen
}

func rtpSourceH264(src []byte) (bool, int) {
	if len(src) < RtpHeaderLen+1 {
		return true, RtpHeaderLen
	}

	payloadData := src[RtpHeaderLen:]
	packetNALUnitType := (payloadData[0] & 0x1f)
	if packetNALUnitType > 1 && packetNALUnitType <= 23 {
		packetNALUnitType = 1
	}

	var skipHeaderLen int
	var isCompletesFrame bool
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
			startBit := payloadData[1] & 0x80
			endBit := payloadData[1] & 0x40
			if 0 != startBit {
				// reset nal header
				payloadData[1] = (payloadData[0] & 0xe0) | (payloadData[1] & 0x1f)
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

	return isCompletesFrame, skipHeaderLen + RtpHeaderLen
}

func rtpSourceHevc(src []byte) (bool, int) {
	if len(src) < RtpHeaderLen+2 {
		return true, RtpHeaderLen
	}

	payloadData := src[RtpHeaderLen:]
	packetNALUnitType := (payloadData[0] & 0x7E) >> 1

	var skipHeaderLen int
	var isCompletesFrame bool
	switch packetNALUnitType {
	case 48: // Aggregation Packet (AP)
		{
			// We skip over the 2-byte Payload Header, and the DONL header (if any).
			skipHeaderLen = 4
			isCompletesFrame = true
		}
		break
	case 49: // Fragmentation Unit (FU)
		{
			// This NALU begins with the 2-byte Payload Header, the 1-byte FU header, and (optionally)
			// the 2-byte DONL header.
			// If the start bit is set, we reconstruct the original NAL header at the end of these
			// 3 (or 5) bytes, and skip over the first 1 (or 3) bytes.
			startBit := payloadData[2] & 0x80
			endBit := payloadData[2] & 0x40
			if 0 != startBit {
				nalUnitType := payloadData[2] & 0x3F
				newNalHeader := make([]byte, 2)
				newNalHeader[0] = (payloadData[0] & 0x81) | (nalUnitType << 1)
				newNalHeader[1] = payloadData[1]

				payloadData[1] = newNalHeader[0]
				payloadData[2] = newNalHeader[1]

				skipHeaderLen = 1
			} else {
				skipHeaderLen = 3
			}
			isCompletesFrame = (endBit != 0)
		}
		break
	default:
		{
			isCompletesFrame = true
			skipHeaderLen = 0
		}
	}
	return isCompletesFrame, skipHeaderLen + RtpHeaderLen
}
