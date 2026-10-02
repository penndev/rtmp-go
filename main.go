package main

import (
	"fmt"

	"github.com/penndev/rtmp/rtmp"
)

func main() {
	addr := "127.0.0.1:1935"
	rtmpSrv := rtmp.New()
	fmt.Printf("Rtmp Serve listening rtmp://%s", addr)
	err := rtmpSrv.Listen(addr)
	panic(err)
}
