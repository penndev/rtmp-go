package rtmp

import (
	"log"
	"net"
	"sync"
)

type Serve struct {
	mu   sync.RWMutex
	Addr string
}

func (srv *Serve) handle(nc net.Conn) {
	log.Printf("rtmp handle %s", nc.RemoteAddr())
	defer func() {
		nc.Close()
		if err := recover(); err != nil {
			log.Printf("%s: %s", "recover: ", err)
		}
	}()

	conn := NewConn(nc)
	// check rtmp handshake
	if err := conn.Handshake(); err != nil {
		log.Printf("%s ServeHandShake fail err[%s]", nc.RemoteAddr(), err.Error())
		return
	}
	if err := conn.HandleStream(); err != nil {
		log.Printf("%s handleStream fail err[%s]", nc.RemoteAddr(), err.Error())
		return
	}
}

// rtmp server listen
func (srv *Serve) Listen(address string) error {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer ln.Close()
	for {
		nc, err := ln.Accept()
		if err != nil {
			return err
		}
		go srv.handle(nc)
	}
}

// create new rtmp serve
func NewRtmp() *Serve {
	s := &Serve{}
	return s
}
