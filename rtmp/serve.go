package rtmp

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/penndev/rtmp-go/flag"
)

var logTemplate = `[Notify] 
Rtmp new Publish [%s]
- - - - Play List URL - - - -
rtmp: rtmp://%s
flv: http://%s
hls: http://%s
`

type adapterListen func(string, <-chan Pack)

type Serve struct {
	mu    sync.RWMutex
	Addr  string
	Topic map[string]*PubSub

	// 全局订阅器，所有新的推流都会被加入到这个队列。
	Adapter []adapterListen
}

// 当有新的推送消息时。
func (srv *Serve) newPublisher(topic string) *PubSub {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	ps := &PubSub{
		buffer:     3,
		timeout:    3 * time.Second,
		subscriber: make(map[chan Pack]bool),
		mediaInfo:  metaInfo{},
	}
	// 处理全局adapter listen
	for _, adapterCallBack := range srv.Adapter {
		ch := make(chan Pack)
		go adapterCallBack(topic, ch)
		ps.subscriber[ch] = true
	}
	srv.Topic[topic] = ps
	return ps
}

// 播放客户端主动关闭
func (srv *Serve) closePublisher(topic string) {
	if ps, ok := srv.Topic[topic]; ok {
		srv.mu.Lock()
		defer srv.mu.Unlock()
		ps.Close()
		delete(srv.Topic, topic)
	}
}

// 播放客户端获取实例。
func (srv *Serve) getPublisher(topic string) (*PubSub, bool) {
	if pubsub, ok := srv.Topic[topic]; ok {
		return pubsub, true
	} else {
		return nil, false
	}
}

func (srv *Serve) handle(nc net.Conn) {
	defer func() {
		nc.Close()
		if err := recover(); err != nil {
			log.Printf("%s: %s", "recover: ", err)
		}
	}()
	// check rtmp handshake
	if err := ServeHandShake(nc); err != nil {
		log.Printf("%s ServeHandShake fail err[%s]", nc.RemoteAddr(), err.Error())
		return
	}
	// create new rtmp conn
	conn := NewConn(nc)
	if err := conn.handleConnect(); err != nil {
		log.Printf("%s handleConnect fail err[%s]", nc.RemoteAddr(), err.Error())
		return
	}
	if err := conn.handleStream(); err != nil {
		log.Printf("%s handleStream fail err[%s]", nc.RemoteAddr(), err.Error())
		return
	}
	if conn.IsPublish {
		topic := conn.App + conn.Stream
		log.Printf(
			logTemplate,                          // 模板
			topic,                                // 主题
			flag.RtmpAddr+"/"+topic,              // rtmp 播放拼接
			flag.HttpAddr+"/play.flv?top="+topic, // flv播放拼接
			flag.HttpAddr+"/play.m3u8?top="+topic, // hls播放拼接
		)
		pubsub := srv.newPublisher(topic)
		conn.handlePublishing(func(pk Pack) {
			pubsub.Publish(pk)
		})
		srv.closePublisher(topic)
	} else {
		topic := conn.App + conn.Stream
		if pubsub, ok := srv.getPublisher(topic); ok {
			sch := pubsub.Subscription()
			defer pubsub.SubscriptionClose(sch)
			if err := conn.handlePlay(sch); err != nil {
				log.Printf("%s: %s", "play fail", err)
			}
		} else {
			log.Printf("rtmp %s not found", topic)
			// 立即退出 defer nc.close
		}

	}
}

// 启动Tcp监听
// 处理golang net ListenConfig 参数 - 做优化
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

// 处理全局适配器，用来监听所有的推送流。
func (srv *Serve) AdapterRegister(al adapterListen) {
	srv.mu.Lock()
	srv.Adapter = append(srv.Adapter, al)
	srv.mu.Unlock()
}

func (srv *Serve) SubscriptionTopic(topic string) (*PubSub, bool) {
	if pubsub, ok := srv.Topic[topic]; ok {
		return pubsub, true
	} else {
		return nil, false
	}
}

// create new rtmp serve
func NewRtmp() *Serve {
	s := &Serve{
		Topic:   make(map[string]*PubSub),
		Adapter: []adapterListen{},
	}
	return s
}
