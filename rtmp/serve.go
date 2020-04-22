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
	if err != nil {
		return err
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		connnect := NewConnnect(conn)
		go s.handle(connnect)
	}

}

// handle 处理分发数据
func (s Service) handle(c Connnect) {
	defer c.Close()

	// 处理rtmp握手消息
	if err := c.HandShake(); err != nil {
		log.Println("握手失败。")
		return
	}

	//	NetConnection
	if err := c.Steam(); err != nil {
		log.Println("Handle err:", err)
	}

}
