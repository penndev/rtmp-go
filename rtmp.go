package main

import (
	"log"
	"net"
	// "bytes"
	"encoding/binary"
	// "github.com/pennfly/amf-go"
)

//处理主控。
func main() {
	listen, err := net.Listen("tcp", "127.0.0.1:1935")
	if err != nil {
		log.Fatal(err)
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		//  处理用户的新连接
		go handleServe(conn)
	}
}

//处理用户发送消息调度
func handleServe(c net.Conn) {
	defer c.Close()
	ip := c.RemoteAddr().String()
	log.Println("new Client:", ip)

	err := ConnHandshake(c)
	if err != nil {
		log.Println(err)
	}

	byt := 2048

	for {
		buf := make([]byte, byt)
		bl, err := c.Read(buf)
		if err != nil {
			log.Println(err)
			break
		}

		log.Println("原始传入数据  @", bl, buf[0:bl])

		rl := 0
		var m Rtmp
		m.c = c
		for rl < bl {
			rl += m.InitHeader(buf[rl:bl])
			log.Println("解析消息的type ID 是：", m.type_id)
			if m.type_id[0] == 20 {
				// CommandAMF0(m)
			}
		}

	}
}

// C012 S012
func ConnHandshake(c net.Conn) (err error) {
	buf := make([]byte, 1537)
	_, err = c.Read(buf)
	if err != nil {
		log.Println("rtmp hand shake fail")
	}
	// C0+C1->
	if buf[0] != 3 {
		log.Println("rtmp version is not right!")
	} else {
		// <-S0+S1+S2
		mg := buf[0:]
		for _, value := range buf[1:1537] {
			mg = append(mg, value)
		}
		c.Write(mg)
	}
	// throw C2->
	_, err = c.Read(buf)
	log.Println("rtmp hand shake finish!")
	return nil
}

//rtmp消息处理结构体
type Rtmp struct {
	c               net.Conn
	chunk_stream_id []byte //1 byte
	timestamp       []byte //3 byte
	body_size       []byte // 3 byte
	type_id         []byte //1 byte
	stream_id       []byte //4 byte
	body            []byte
}

func (m *Rtmp) InitHeader(meg []byte) int {
	m.chunk_stream_id = meg[0:1]
	m.timestamp = meg[1:4]
	m.type_id = meg[7:8]
	m.stream_id = meg[8:12]

	body_size := []byte{0, meg[4], meg[5], meg[6]}
	tmp_body_size := binary.BigEndian.Uint32(body_size)

	res := 12 + int(tmp_body_size)
	m.body = meg[12:res]

	//log.Println("分割出来的消息为 #", res, meg[0:res] )

	return res
}

func CommandAMF0(m Rtmp) {
	commandName := amf.Decode(m.body)

	log.Println(commandName)

	// switch commandName {
	// case "connect":

	// }

}
