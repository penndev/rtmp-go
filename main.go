package main

import (
	"net/http"

	"github.com/penndev/rtmp-go/flag"
	"github.com/penndev/rtmp-go/flv"
	"github.com/penndev/rtmp-go/hls"
	"github.com/penndev/rtmp-go/mpegts"
	"github.com/penndev/rtmp-go/rtmp"
)

func main() {
	flag.Parse()

	rtmpSrv := rtmp.NewRtmp()
	rtmpSrv.AdapterRegister(flv.AdapterFlv) // 写入flv录播文件
	rtmpSrv.AdapterRegister(mpegts.Adapter) // 生成mpeg-ts文件
	go func() {
		http.Handle("/runtime/", http.StripPrefix("/runtime/", http.FileServer(http.Dir("./runtime"))))
		http.HandleFunc("/play.m3u8", hls.HandleHls(rtmpSrv.SubscriptionTopic))
		http.HandleFunc("/play.flv", flv.HandleFlv(rtmpSrv.SubscriptionTopic))
		print("Http Serve listening http:", flag.HttpAddr, "\n")
		err := http.ListenAndServe(flag.HttpAddr, nil)
		panic(err)
	}()
	print("Rtmp Serve listening rtmp://", flag.RtmpAddr, "\n")
	err := rtmpSrv.Listen(flag.RtmpAddr)
	panic(err)
}
