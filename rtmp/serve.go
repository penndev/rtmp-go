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
}

func (srv *Serve) handle(nc net.Conn) {
	defer nc.Close()

	log.Println(nc.RemoteAddr().String(), "-> nc connected")
	// 握手
	if err := ServeHandShake(nc); err != nil {
		panic(err)
	}

	chk := newChunk(nc)
	conn := newConn()
	// Connection
	if err := netConnectionCommand(chk, conn); err != nil {
		panic(err)
	}
	// 初始化流
	if err := netStreamCommand(chk, conn); err != nil {
		panic(err)
	}
	// 处理client消息。
	go handle(chk, conn)

	var client func(Pack)
	callback := func(pk Pack, client func(Pack)) {
		client(pk)
	}
	// 第 1 bit 设置 onMetaData
	// 第 2 bit 设置 Audit
	// 第 3 bit 设置 Video
	//0000 0111
	readyIng := 0

	//处理逻辑
	if conn.IsPublish {
		log.Println("publish")
		client = func(pk Pack) {
			if readyIng < 7 {
				if pk.MessageTypeID == 15 {
					readyIng |= 1 // onMetaData
					log.Println("pk.MessageTypeID=15")
					return
				}
				if pk.MessageTypeID == 18 {
					readyIng |= 1 // onMetaData
					log.Println(readyIng)
					return
				}
				if pk.MessageTypeID == 8 {

					readyIng |= 2 // onAuditInit
					log.Println(readyIng)
					return
				}
				if pk.MessageTypeID == 9 {
					readyIng |= 4 // onVideoInit
					log.Println(readyIng)
					return
				}
				log.Println(pk.ChunkMessageHeader)
			}
			conn.Closed = true
			log.Println(pk.MessageTypeID)
		}
	} else {
		log.Println("play")
		client = func(pk Pack) {
			log.Println(pk.MessageTypeID)
		}
	}

	for {
		if conn.Closed {
			break
		}
		select {
		case status := <-conn.CloseChan:
			conn.Closed = status
		case avpack := <-conn.AVPackChan:
			callback(avpack, client)
		}
	}
	// publish无效
	chk.setStreamEof(DefaultStreamID)
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
	}

	if err := s.listen(); err != nil {
		return err
	}
	return nil
}
