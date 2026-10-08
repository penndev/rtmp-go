package main

import (
	"fmt"

	"github.com/penndev/rtmp/pubsub"
	"github.com/penndev/rtmp/rtmp"
	"github.com/penndev/rtmp/rtmp/handler"
)

func main() {
	streams := pubsub.NewRtmp()
	rtmpSrv := rtmp.New(handler.NewDefaultHandler(), streams)

	addr := "127.0.0.1:1935"
	fmt.Printf("Listening on rtmp://%s\n", addr)
	err := rtmpSrv.Listen(addr)
	panic(err)
}
