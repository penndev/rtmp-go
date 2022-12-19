package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/penndev/rtmp-go/flv"
	"github.com/penndev/rtmp-go/handlers"
	"github.com/penndev/rtmp-go/hls"
	"github.com/penndev/rtmp-go/mpegts"
	"github.com/penndev/rtmp-go/rtmp"
)

func main() {
	rtmpSrv := rtmp.NewRtmp()
	rtmpSrv.AdapterRegister(flv.Adapterflv)   // 写入flv录播文件
	rtmpSrv.AdapterRegister(mpegts.Adapterts) // 生成mpegts文件
	fmt.Println("运行中")
	go func() {
		http.Handle("/runtime/", http.StripPrefix("/runtime/", http.FileServer(http.Dir("./runtime"))))
		http.Handle("/dirs", http.StripPrefix("/h5player/", http.FileServer(http.Dir("./h5player"))))
		http.HandleFunc("/play.m3u8", hls.Handlehls(rtmpSrv.SubscriptionTopic))
		http.HandleFunc("/play.flv", flv.Handleflv(rtmpSrv.SubscriptionTopic)) // http flv 播放
		err := http.ListenAndServe("127.0.0.1:80", nil)
		panic(err)
	}()
	go func() {
		router := gin.Default()
		router.LoadHTMLFiles("h5player\\flv\\index.html")
		router.Static("flvjs", "h5player\\flv\\flvjs")
		router.GET("/flvplay", handlers.Handleflv)
		router.Run("127.0.0.1:8080")
	}()
	err := rtmpSrv.Listen("127.0.0.1:1935")
	panic(err)
}
