package main

import (
	"github.com/pennfly/rtmp-go/rtmp"
)

func main() {
	var rtmp rtmp.Service
	rtmp.Listen = ":1935"
	rtmp.Server()
}
