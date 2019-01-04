package rtspclient

type IRtpParseInterface interface {
	SplitHeader(src []byte) (isCompletesFrame bool, header []byte, payload []byte)
	ParsingRtp(header []byte, payload []byte) (naluHeaderSize int, naluSize int)
}

const (
	PacketHeaderLen = 4
	RtpHeaderLen    = 12
)
