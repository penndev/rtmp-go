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

// 启动Tcp监听
// 处理golang net Listenconfig 参数
func (srv *Serve) listen() error {
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
			continue
		}
		var run = func() {
			conn := &Conn{
				serve:    srv,
				rwc:      &conn,
				IsPusher: false,
			}
			//阻塞函数-处理Rtmp协议内容
			conn.Connect()
		}
		go run()
	}
}

// 实例 Serve 结构体
// 配置Rtmp参数
func newServer() *Serve {
	serve := &Serve{
		Addr:    ":1935",
		Timeout: 10 * time.Second,
	}
	return serve
}

// 运行Rtmp协议。
// 阻塞函数
func NewRtmp() error {
	s := newServer()
	if err := s.listen(); err != nil {
		return err
	}
	return nil
}
