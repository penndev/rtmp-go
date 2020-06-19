package rtmp

import (
	"net"
	"time"
)

//WorkPool rtmp 消息池
type WorkPool struct {
	Metadata Chunk
	Pusher   chan byte
	Player   []chan Chunk
}

// A Server defines parameters for running a RTMP server.
type Server struct {
	Addr      string
	ChunkSize int
	Timeout   time.Duration
	WorkPool  map[string]*WorkPool
}

// Serve start net tcp serve
func (srv *Server) Serve() error {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		rw, err := ln.Accept()
		if err != nil {
			return err
		}
		c := newConn(srv, rw)
		go c.serve()
	}

}

// Serve listens on the TCP network address addr and timeout
func Serve() error {
	server := &Server{
		Addr:      ":1935",
		Timeout:   30,
		ChunkSize: 128,
		WorkPool:  map[string]*WorkPool{},
	}
	return server.Serve()
}
