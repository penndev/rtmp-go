package main

import (
	"fmt"

	"github.com/penndev/rtmp/rtmp"
)

func main() {
	addr := "127.0.0.1:1935"
	rtmpSrv := rtmp.New()
	fmt.Printf("Listening on rtmp://%s\n", addr)
	err := rtmpSrv.Listen(addr)
	panic(err)
}
