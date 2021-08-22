package rtmp

import (
	"net"
)

//依据rtmp对tcp进行封装
type Conn struct {
	// Serve 数据结构
	// 用来访问运行时数据
	serve *Serve

	// Tcp 网络IO
	// 进行数据通讯处理
	rwc *net.Conn

	App    string
	Stream string

	//是否有推送消息体的权限。
	//是否是主播
	IsPusher bool

	PackChan chan Pack
}

// 握手
func (c *Conn) handShake() error {
	err := ServeHandShake(*c.rwc)
	return err
}

// 关闭rtmp连接，做一些清理。
func (c *Conn) Close() {
	(*c.rwc).Close()
}

func (c *Conn) onPushMate(pk Pack) {
	c.serve.WorkPool.Publish(c.App, c.PackChan)
	c.PackChan <- pk
}

func (c *Conn) onPushAv(pk Pack) {
	c.PackChan <- pk
}

func (c *Conn) onSetPush(app string, stream string) {
	c.App = app
}

func (c *Conn) onPushStop() {
	c.serve.WorkPool.ClosePublish(c.App, c.PackChan)
}

func newConn(srv *Serve, nc *net.Conn) (*Conn, error) {
	c := &Conn{
		serve:    srv,
		rwc:      nc,
		IsPusher: false,
		PackChan: make(chan Pack),
	}

	err := c.handShake()
	return c, err
}
