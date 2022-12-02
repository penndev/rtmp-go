package main

import (
	"log"

	"github.com/penndev/rtmp-go/av"
	"github.com/penndev/rtmp-go/rtmp"
)

func main() {
	rtmpSrv := rtmp.NewRtmp()
	// 写入flv文件
	rtmpSrv.AdapterRegister(func(name string, ch <-chan rtmp.Pack) {
		var flv av.FLV
		flv.GenFlv(name)
		defer flv.Close()
		for pk := range ch {
			log.Println("pk ty id->", pk.MessageTypeID)
			flv.AddTag(pk.MessageTypeID, pk.Timestamp, pk.PayLoad)
		}
	})
	err := rtmpSrv.Listen()
	if err != nil {
		log.Println(err)
	}
}
