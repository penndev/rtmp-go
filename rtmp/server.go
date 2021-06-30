package rtmp

import (
	"net"
	"time"
)

//WorkPool rtmp 消息池
type WorkPool struct {
	Metadata     []byte
	VideoCode    []byte
	AudioCode    []byte
	videocodecid int
	Player       map[string]chan Chunk
}

// A Server defines parameters for running a RTMP server.
type Server struct {
	Addr      string
	ChunkSize int
	Timeout   time.Duration
	WorkPool  map[string]map[string]*WorkPool //['live'=>['room'=>'WorkPool']]
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

// Serve add WorkPool
func (srv *Server) addPool(app string, stream string) {
	room := make(map[string]*WorkPool)
	room[stream] = &WorkPool{
		Metadata: []byte{},
		Player:   map[string]chan Chunk{},
	}
	srv.WorkPool[app] = room
}

// Serve listens on the TCP network address addr and timeout
func Serve() error {
	// wp :=
	server := &Server{
		Addr:      ":1935",
		Timeout:   30,
		ChunkSize: 128,
		WorkPool:  map[string]map[string]*WorkPool{},
	}
	return server.Serve()
}
