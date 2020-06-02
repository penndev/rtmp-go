package rtmp

import "net"

type Conn struct {
	conn net.Conn
}

func NewConn(netConn net.Conn) (Conn, error) {
	var c Conn
	c.conn = netConn
	return c, nil
}
