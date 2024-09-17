package flag

import "flag"

var HttpAddr string
var RtmpAddr string

func Parse() {
	httpAddr := flag.String("http", "127.0.0.1:80", "HTTP server address (e.g., 127.0.0.1:80)")
	rtmpAddr := flag.String("rtmp", "127.0.0.1:1935", "RTMP server address (e.g., 127.0.0.1:1935)")
	flag.Parse() // 解析命令行参数

	HttpAddr = *httpAddr
	RtmpAddr = *rtmpAddr
}
