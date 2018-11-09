package main

import (
	"rtspclient/jsrtspclient"
)

func main() {
	manager := jsrtspclient.NewRtspManager()
	manager.Start()
}
