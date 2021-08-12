package rtmp

import (
	"context"
	"fmt"
	"net"
	"time"
)

type Serve struct {
	Addr      string
	ChunkSize int
	Timeout   time.Duration
}

//启动Tcp监听
func (srv *Serve) Listen() error {
	var lc = net.ListenConfig{
		KeepAlive: srv.Timeout,
	}
	ln, err := lc.Listen(context.Background(), "tcp", srv.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Print(err)
			conn.Close()
			continue
		}
		var run = func() {
			conn := &Conn{
				serve:    srv,
				rwc:      &conn,
				IsPusher: false,
			}
			//阻塞函数
			conn.Connect()
		}
		go run()
		//
		// 启动媒体处理进程。
	}
}

//使用默认参数 配置Rtmp
func Server() error {
	serve := &Serve{
		Addr:    ":1935",
		Timeout: 10 * time.Second,
	}
	return serve.Listen()
}
