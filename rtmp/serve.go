package rtmp

import (
	"context"
	"log"
	"net"
	"time"
)

type Serve struct {
	Addr    string
	Timeout time.Duration
	App     *App
}

func (srv *Serve) handle(nc net.Conn) {
	defer nc.Close()
	log.Println(nc.RemoteAddr().String(), "-> nc connected")

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
	log.Println(nc.RemoteAddr().String(), "-> nc closeID")
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

	if err := s.listen(); err != nil {
		return err
	}
	return nil
}
