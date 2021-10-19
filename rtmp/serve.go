package rtmp

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"rtmp-go/httpflv"
	"strings"
	"time"
)

type Serve struct {
	Addr    string
	Timeout time.Duration
	App     *App
}

func (srv *Serve) handle(nc net.Conn) {
	defer nc.Close()

	// 处理握手相关
	if err := ServeHandShake(nc); err != nil {
		panic(err)
	}

	chk := newChunk(nc)
	conn := newConn()

	// 处理连接流
	if err := netConnectionCommand(chk, conn); err != nil {
		panic(err)
	}
	// 处理初始化流
	if err := netStreamCommand(chk, conn); err != nil {
		panic(err)
	}
	// 主流程
	if err := netHandleCommand(chk, conn, srv.App); err != nil {
		panic(err)
	}
}

// 启动Tcp监听
// 处理golang net Listenconfig 参数
func (srv *Serve) listen() error {
	var lc = net.ListenConfig{
		KeepAlive: srv.Timeout,
	}
	// 启动TCP监听
	ln, err := lc.Listen(context.Background(), "tcp", srv.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		nc, err := ln.Accept()
		if err != nil {
			return err
		}
		// 开始业务流程
		go srv.handle(nc)
	}
}

// 运行Rtmp协议。
// 阻塞函数
func NewRtmp() error {
	s := &Serve{
		Addr:    ":1935",
		Timeout: 10 * time.Second,
		App:     newApp(),
	}

	fmt.Print(`
             _                                                                     
        .___| |_. _ ___ _  _ __ ______ __ _  ___  
        | __| __|/ _   _ \| '_ \______/ _. |/ _ \ 
        | | | |_| | | | | | |_) |    | (_| | (_) |
        |_|  \__|_| |_| |_| .__/      \__, |\___/ 
                          | |          __/ |      
                          |_|         |___/                 
	`)
	name, _ := os.Hostname()
	addrs, _ := net.LookupHost(name)
	fmt.Print("\n     RTMP推流地址(demo): rtmp://" + addrs[0] + ":1935/live/room \n\n")

	fmt.Println(httpflv.Serve(func(w http.ResponseWriter, req *http.Request) {

		flvPath := strings.Split(req.URL.Path, ".")
		if len(flvPath) != 2 || flvPath[1] != "flv" {
			http.NotFound(w, req)
			return
		}

		appPath := strings.Split(flvPath[0], "/")[1:]
		if len(appPath) != 2 {
			http.NotFound(w, req)
			return
		}

		f, err := os.Open("./runtime/live_boot1634624720.flv")
		if err != nil {
			panic(err)
		}
		io.Copy(w, f)

	}))

	if err := s.listen(); err != nil {
		return err
	}
	return nil
}
