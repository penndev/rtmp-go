package main

import (
	"fmt"
	"rtmp-go/rtmp"
)

func main() {
	err := rtmp.Server()
	if err != nil {
		fmt.Println(err)
	}
}
