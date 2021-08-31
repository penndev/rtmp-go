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
	callClose := func(c func()) {
		c()
	}
	callConnect := func(pk Pack, c func(Pack)) {
		c(pk)
	}
	var client func(Pack)
	var close func()
	app := srv.App
	//处理逻辑
	if conn.IsPublish {
		stream := app.addPublish(conn.App, conn.Stream)
		readyIng := 0
		client = func(pk Pack) {
			if readyIng < 7 {
				stream.setMeta(pk, &readyIng)
				return
			}
			stream.setPack(pk)
		}
		close = func() {
			app.delPublish(conn.App, conn.Stream)
		}
	} else {
		// 初始化流不存在。
		if ok := app.addPlay(conn.App, conn.Stream, conn.AVPackChan); !ok {
			log.Println("Play stream not found:", conn.App, conn.Stream)
			conn.Closed = true
		}
		// =======================================================
		mt := app.getMeta(conn.App, conn.Stream)
		pk := Pack{
			PayLoad: mt.meta,
		}
		pk.MessageTypeID = 18
		chk.sendPack(DefaultStreamID, pk)
		//--
		pk = Pack{
			PayLoad: mt.video,
		}
		pk.MessageTypeID = 9
		chk.sendPack(DefaultStreamID, pk)
		//--
		pk = Pack{
			PayLoad: mt.audit,
		}
		pk.MessageTypeID = 8
		chk.sendPack(DefaultStreamID, pk)
		// =======================================================
		client = func(pk Pack) {
			chk.sendPack(DefaultStreamID, pk)
			// log.Println(pk.MessageTypeID)
		}
		close = func() {
			chk.setStreamEof(DefaultStreamID)
			app.delPlay(conn.App, conn.Stream, conn.AVPackChan)
		}
	}

	//处理message.
	if !conn.Closed {
		for avpack := range conn.AVPackChan {
			callConnect(avpack, client)
		}
	}

	callClose(close)
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
