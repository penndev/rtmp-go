package main

import (
	"net/http"

	"github.com/penndev/rtmp-go/flv"
	"github.com/penndev/rtmp-go/hls"
	"github.com/penndev/rtmp-go/mpegts"
	"github.com/penndev/rtmp-go/rtmp"
)

func main() {
	rtmpSrv := rtmp.NewRtmp()
	rtmpSrv.AdapterRegister(flv.AdapterFlv) // 写入flv录播文件
	rtmpSrv.AdapterRegister(mpegts.Adapter) // 生成mpeg-ts文件
	go func() {
		http.Handle("/runtime/", http.StripPrefix("/runtime/", http.FileServer(http.Dir("./runtime"))))
		http.HandleFunc("/play.m3u8", hls.HandleHls(rtmpSrv.SubscriptionTopic))
		http.HandleFunc("/play.flv", flv.HandleFlv(rtmpSrv.SubscriptionTopic)) // http flv 播放
		err := http.ListenAndServe("127.0.0.1:80", nil)
		panic(err)
	}()
	print("Serve listening rtmp://127.0.0.1:1935\n")
	err := rtmpSrv.Listen("127.0.0.1:1935")
	panic(err)
}
