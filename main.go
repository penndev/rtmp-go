package main

import (
	"net/http"

	"github.com/penndev/rtmp-go/flv"
	"github.com/penndev/rtmp-go/rtmp"
)

func main() {
	rtmpSrv := rtmp.NewRtmp()
	rtmpSrv.AdapterRegister(flv.Adapterflv) // 写入flv文件
	go func() {
		http.HandleFunc("/play.flv", flv.Handleflv(rtmpSrv.SubscriptionTopic)) // http flv 播放
		err := http.ListenAndServe("127.0.0.1:80", nil)
		panic(err)
	}()
	err := rtmpSrv.Listen("127.0.0.1:1935")
	panic(err)
}
