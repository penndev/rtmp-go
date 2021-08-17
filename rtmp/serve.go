package rtmp

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"
)

type Serve struct {
	Addr      string
	ChunkSize int
	Timeout   time.Duration

	WorkPool *WorkPool
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
	// 启动av转发工作池
	wp := newWorkPool()
	srv.WorkPool = wp
	go wp.run()
	// 进行监听新连接。
	for {
		nc, err := ln.Accept()
		if err != nil {
			fmt.Print(err)
			continue
		}

		go func() {
			log.Println(nc.RemoteAddr().String(), "-> nc connected")
			defer nc.Close()
			conn, err := newConn(srv, &nc)
			if err != nil {
				log.Println(err)
				return
			}
			chk := newChunk(&nc)
			err = chk.Handle(conn)
			if err != nil {
				log.Println(err)
			}
			log.Println(nc.RemoteAddr().String(), "-> nc closeing")
		}()
	}
}

// 运行Rtmp协议。
// 阻塞函数
func NewRtmp() error {
	s := &Serve{
		Addr:    ":1935",
		Timeout: 10 * time.Second,
	}

	if err := s.listen(); err != nil {
		return err
	}
	return nil
}
