package rtmp

import (
	"log"
	"net"
)

// Service 服务公用参数
type Service struct {
	Host     string //Server listen port
	RunDir   string //Keep Video dir
	Listener net.Listener
}

// Run 启动rtmp服务器
func (s *Service) Run() error {
	defer s.Close()
	if s.Host == "" {
		s.Host = ":1935"
	}

	if s.RunDir == "" {
		s.RunDir = "runtime"
	}

	if err := s.Listen(); err != nil {
		log.Println(err)
	}

}

//Accept 处理分发数据
func (s *Service) Accept(c Connnect) {

	for {
		netConn, err := s.Listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		conn := NewConn(netConn)
		conn.Run()

	}

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

//Listen return net.Listener
// net start listen s.
func (s Service) Listen() error {
	var err error
	s.Listener, err = net.Listen("tcp", s.Host)
	return err
}

// Close the serve
// defer the serve do someing
func (s *Service) Close() {
	s.Listener.Close()
}
