package rtmp

import (
	"bufio"
	"net"
	"time"
)

// Conn 单个连接详情
type Conn struct {
	br  *bufio.Reader     // Read
	bw  *bufio.Writer     // Write
	brw *bufio.ReadWriter // Read and Write,用来握手

	conn       net.Conn
	remoteAddr string
	url        string
	appName    string
	createTime string
}

//NewConn 初始化新链接
func NewConn(conn net.Conn) Conn {
	var c Conn
	c.br = bufio.NewReader(conn)
	c.bw = bufio.NewWriter(conn)
	c.brw = bufio.NewReadWriter(c.br, c.bw)
	c.conn = conn

	c.remoteAddr = conn.RemoteAddr().String()
	c.createTime = time.Now().String()

	return c
}

// Close 关闭链接处理
func (c Conn) Close() {
	c.conn.Close()
}
