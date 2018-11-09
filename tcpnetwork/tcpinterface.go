package tcpnetwork

// IStreamProtocol tcp data unpack callback
type IStreamProtocol interface {
	// Init
	Init()
	// get the header length of the stream
	GetHeaderLength() int
	// read the header length of the stream
	UnserializeHeader([]byte) (int, []byte)
	// format header
	SerializeHeader([]byte) []byte

	// get content length
	GetContentLength([]byte) int
}
