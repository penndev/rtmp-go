package main

import (
	"fmt"
	"rtmp-go/rtmp"
)

func main() {
	err := rtmp.NewRtmp()
	if err != nil {
		fmt.Println(err)
	}
}
