package rtmp

import (
	"log"
	"net"
)

// Service 服务公用参数
type Service struct {
	Listen string
}

// Server 启动rtmp服务器
func (s Service) Server() error {
	listen, err := net.Listen("tcp", s.Listen)
	checkErr(err)
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		checkErr(err)
		connnect := NewConnnect(conn)
		s.handle(connnect)
	}
	return nil
}

// handle 处理分发数据
func (s Service) handle(c Connnect) error {
	defer c.Close()

	// 首先client to service handshake
	err := c.HandShake() // 握手
	c.ReadRtmpMsg()
	checkErr(err)

	return nil
}

// checkErr 统一处理错误
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
