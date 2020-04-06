package rtmp

import (
	"bufio"
	"net"
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
	connected  bool   // 连接是否完成
	streamID   uint32 // 流ID
}

//NewConn 初始化新链接
func NewConn(conn net.Conn) Conn {
	var c Conn
	c.conn = conn
	return c
}
