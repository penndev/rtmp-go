package rtmp

import (
	"log"
	"net"
	"sync"
	"time"
)

type adapterlisten func(string, <-chan Pack)

type Serve struct {
	mu      sync.RWMutex
	Addr    string
	Topic   map[string]*PubSub
	Adapter []adapterlisten
}

func (srv *Serve) handle(nc net.Conn) {
	ncaddr := nc.RemoteAddr().String()
	defer func() {
		nc.Close()
		log.Printf("[%s]-> nc closed", ncaddr)
	}()
	log.Printf("[%s]-> nc connected", ncaddr)
	// check rtmp handshake
	if err := ServeHandShake(nc); err != nil {
		log.Printf("[%s]-> conn handshake fail", ncaddr)
		return
	}
	// create new rtmp conn
	conn := NewConn(nc)
	if err := conn.handleConnect(); err != nil {
		log.Printf("[%s]-> rtmp connection fail(%s)", ncaddr, err)
		return
	}
	if err := conn.handleStream(); err != nil {
		log.Printf("[%s]-> rtmp stream fail(%s)", ncaddr, err)
		return
	}
	if conn.IsPublish {
		topic := conn.App + conn.Stream
		pubsub := srv.newPublisher(topic)
		conn.handlePublishing(func(pk Pack) {
			pubsub.Publish(pk)
		})
		srv.colsePublisher(topic)
	} else {
		// srv.mu.Lock()
		// 判断是否有发布者
		// srv.mu.Unlock()
	}
}

// 启动Tcp监听
// 处理golang net Listenconfig 参数
func (srv *Serve) Listen() error {
	ln, err := net.Listen("tcp", srv.Addr)
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

// 处理全局适配器，通常生成全局文件用
func (srv *Serve) AdapterRegister(al adapterlisten) {
	srv.mu.Lock()
	srv.Adapter = append(srv.Adapter, al)
	srv.mu.Unlock()
}

func (srv *Serve) newPublisher(topic string) *PubSub {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	ps := &PubSub{
		timeout:    3 * time.Second,
		subscriber: make(map[chan Pack]bool),
	}
	// 处理全局adapter listen
	for _, adcb := range srv.Adapter {
		ch := make(chan Pack)
		go adcb(topic, ch)
		ps.subscriber[ch] = true
	}
	srv.Topic[topic] = ps
	return ps
}

func (srv *Serve) colsePublisher(topic string) {
	if ps, ok := srv.Topic[topic]; ok {
		srv.mu.Lock()
		defer srv.mu.Unlock()
		ps.Close()
		delete(srv.Topic, topic)
	}
}

// create new rtmp serve
func NewRtmp() *Serve {
	s := &Serve{
		Addr:    "127.0.0.1:1935",
		Topic:   make(map[string]*PubSub),
		Adapter: []adapterlisten{},
	}

	return s
}
