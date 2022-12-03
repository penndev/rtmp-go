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
		for pk := range ch {
			flv.AddTag(pk.MessageTypeID, pk.Timestamp, pk.PayLoad)
		}
		flv.Close()
	})
	err := rtmpSrv.Listen()
	if err != nil {
		log.Println(err)
	}
}
