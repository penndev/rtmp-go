package rtmp

import (
	"bufio"
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
func (srv *Serve) Server() error {
	// 		c := newConn(srv, rw)
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
		//
		client := &Conn{
			serve: srv,
			rwc:   &conn,
			r:     bufio.NewReader(conn),
			w:     bufio.NewWriter(conn),

			ReadChunkSize:  srv.ChunkSize,
			WriteChunkSize: 4096,

			// 	SteamID:        4,
			// 	ChunkLists:     make(map[uint32]Chunk),
			// 	SendChunkLists: make(map[uint32]Chunk),
			IsPusher: false,
		}
		go client.connect()
	}
}

//使用默认参数 配置Rtmp
func Server() error {
	serve := &Serve{
		Addr:      ":1935",
		Timeout:   10 * time.Second,
		ChunkSize: 128,
		// WorkPool:  map[string]map[string]*WorkPool{},
	}
	return serve.Server()
}
