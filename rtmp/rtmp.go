package rtmp

import (
	"bufio"
	"log"
	"net"
	"time"
)

//Connnect rtmp单个链接的struct
type Connnect struct {
	rw         *bufio.ReadWriter // Read and Write,用来握手
	conn       net.Conn
	rwByteSize map[string]uint32
	createTime string
	remoteAddr string
	url        string
	appName    string
}

// Close 关闭链接处理
func (c *Connnect) Close() {
	c.conn.Close()
	log.Println("Conn is close :", c.remoteAddr)
}

// HandShake 处理rtmp握手。
func (c *Connnect) HandShake() error {

	return nil
}

// NewConnnect 初始化一个新的链接。
func NewConnnect(conn net.Conn) Connnect {
	var c Connnect
	c.conn = conn
	c.rw = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	c.rwByteSize = make(map[string]uint32)
	c.rwByteSize["read"] = 0
	c.rwByteSize["write"] = 0

	c.remoteAddr = conn.RemoteAddr().String()
	c.createTime = time.Now().String()

	return c
}
