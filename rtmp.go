package main

import (
	// "bytes"
	"encoding/binary"
	"log"
	"net"

	"github.com/pennfly/amf-go"
)

//处理主控。
func main() {

	listen, err := net.Listen("tcp", "127.0.0.1:1935")
	if err != nil {
		log.Fatal(err)
	}
	defer listen.Close()

	for {
		c, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		//  处理用户的新连接
		// go handleServe(conn)

		defer c.Close()
		ip := c.RemoteAddr().String()
		log.Println("new Client:", ip)

		err = ConnHandshake(c)
		if err != nil {
			log.Println(err)
		}

		byt := 4096

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
				log.Println("解析消息的type ID 是：", m.type_id, ";长度是 rl = ", rl)
				if m.type_id[0] == 20 {
					CommandAMF0(m)
				}
			}

		}

	}
}

//处理用户发送消息调度
// func handleServe(c net.Conn) {

// }

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

	// log.Println("分割出来的消息为 #", res, tmp_body_size)

	return res
}

func (m *Rtmp) CreateHeader(meg []byte) {
	m.body = meg

}

func CommandAMF0(m Rtmp) {
	commandName := amf.Decode(m.body)
	log.Println(commandName)
	switch commandName[0] {
	case "connect":
		bs := []byte{2, 0, 0, 0, 0, 0, 4, 1, 0, 0, 0, 0, 0, 0, 4, 0 /**/, 2, 0, 0, 0, 0, 0, 4, 5, 0, 0, 0, 0, 0, 38, 37, 160 /**/, 3, 0, 0, 0, 0, 1, 5, 20, 0, 0, 0, 0, 2, 0, 7, 95, 114, 101, 115, 117, 108, 116, 0, 63, 240, 0, 0, 0, 0, 0, 0, 3, 0, 6, 102, 109, 115, 86, 101, 114, 2, 0, 14, 70, 77, 83, 47, 51, 44, 53, 44, 53, 44, 50, 48, 48, 52, 0, 12, 99, 97, 112, 97, 98, 105, 108, 105, 116, 105, 101, 115, 0, 64, 63, 0, 0, 0, 0, 0, 0, 0, 4, 109, 111, 100, 101, 0, 63, 240, 0, 0, 0, 0, 0, 0, 0, 0, 9, 3, 0, 5, 108, 101, 118, 101, 108, 2, 0, 6, 115, 116, 97, 116, 117, 115, 0, 4, 99, 111, 100, 101, 2, 0, 29, 78, 101, 116, 67, 111, 110, 110, 101, 99, 116, 105, 111, 110, 46, 67, 111, 110, 110, 101, 99, 116, 46, 83, 117, 99, 99, 101, 115, 115, 0, 11, 100, 101, 115, 99, 114, 105, 112, 116, 105, 111, 110, 2, 0, 21, 67, 111, 110, 110, 101, 99, 116, 105, 111, 110, 32, 115, 117, 99, 99, 101, 101, 100, 101, 100, 46, 0, 4, 100, 97, 116, 97, 8, 0, 0, 0, 1, 0, 7, 118, 101, 114, 115, 105, 111, 110, 2, 0, 10, 51, 44, 53, 44, 53, 44, 50, 48, 48, 52, 0, 0, 9, 0, 8, 99, 108, 105, 101, 110, 116, 73, 100, 0, 65, 215, 155, 120, 124, 192, 0, 0, 0, 14, 111, 98, 106, 101, 99, 116, 69, 110, 99, 111, 100, 105, 110, 103, 0, 64, 8, 0, 0, 0, 0, 0, 0, 0, 0, 9}
		m.c.Write(bs)
		log.Println("CommandAMF0 -connect")
	}

	log.Println("CommandAMF0 FINISH")

}
