package rtmp

import (
	"log"
	"net"
	"net/url"
	"os"
)

type Conn struct {
	App    string
	Stream string

	// 主播端 true
	// 播放端 false
	IsPublish bool
	IsStoped  bool

	AVPackChan chan Pack
	CloseChan  chan bool
}

// 根据返回值处理连接是否继续
// return true 继续下一步
func (c *Conn) onConnect(app string) bool {
	c.App = app
	return true
}

func (c *Conn) onPublish(stream string) bool {
	//验证密钥。
	str, _ := url.Parse(stream)
	c.Stream = str.Path
	c.IsPublish = true

	// 如果是push端口则printf
	name, _ := os.Hostname()
	addrs, _ := net.LookupHost(name)
	log.Print("传输了信号: ", addrs[0], ":1935/", c.App, "/", c.Stream)
	log.Print("RTMP播放地址: rtmp://", addrs[0], ":1935/", c.App, "/", c.Stream)
	log.Print("http-flv播放地址: http://", addrs[0], ":8080/", c.App, "/", c.Stream, ".flv")
	return true
}

func (c *Conn) onPlay(stream string) bool {
	c.Stream = stream
	return true
}

func (c *Conn) onClose() {
}

func newConn() *Conn {
	return &Conn{
		AVPackChan: make(chan Pack),
		CloseChan:  make(chan bool),
	}
}
