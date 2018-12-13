package main

import (
	"rtspclient/rtspclient"
)

type DefaultDataHandle struct {
}

func (dataHandler *DefaultDataHandle) SetMediaSubsession(media rtspclient.MediaSubsession) {
}

func (dataHandler *DefaultDataHandle) GetHeader() []byte {
	return make([]byte, 0)
}

func (dataHandler *DefaultDataHandle) ParsingData(src []byte) []byte {
	return src
}
