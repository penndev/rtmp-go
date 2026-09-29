package main

import (
	"github.com/penndev/rtmp/flag"
	"github.com/penndev/rtmp/rtmp"
)

func main() {
	flag.Parse()

	rtmpSrv := rtmp.New()
	print("Rtmp Serve listening rtmp://", flag.RtmpAddr, "\n")
	err := rtmpSrv.Listen(flag.RtmpAddr)
	panic(err)
}
