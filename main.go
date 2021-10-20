package main

import (
	"fmt"

	"github.com/penndev/rtmp-go/rtmp"
)

func main() {

	err := rtmp.NewRtmp()
	if err != nil {
		fmt.Println(err)
	}
}
