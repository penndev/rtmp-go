package main

import (
	"log"
	"net"
)

const (
	// Type0 是fmt type = 0
	Type0 = 11
	// Type1 是fmt type = 1
	Type1 = 7
	// Type2 是fmt type = 2
	Type2 = 3
	// Type3 是fmt type = 3
	Type3 = 0
)

//  Chunk Header
type chunk struct {
	fmt           int
	csid          int
	timestamp     int
	messagelength int
	messagetypeid int
	msgstreamid   int
	messagebody   []byte
	debug         bool

	//以下数据为程序设计用。

}

func (chunk *chunk) initialization(conn net.Conn) {
	chunk.basicHeader(conn)
	switch chunk.fmt {
	case 0:

		bufType := make([]byte, Type0)
		len, err := conn.Read(bufType)
		if err != nil {
			log.Fatal("Rtmp Chunk Len and Err0：", len, err)
		}
		chunk.type0(bufType)
		if chunk.debug == true {
			log.Println("initialization 0 fmt:", bufType)
		}
	case 1:

		bufType := make([]byte, Type1)
		len, err := conn.Read(bufType)
		if err != nil {
			log.Fatal("Rtmp Chunk Len and Err1：", len, err)
		}
		chunk.type1(bufType)
		if chunk.debug == true {
			log.Println("initialization 1 fmt:", bufType)
		}
	case 2:

		bufType := make([]byte, Type2)
		len, err := conn.Read(bufType)
		if err != nil {
			log.Fatal("Rtmp Chunk Len and Err2：", len, err)
		}
		chunk.type2(bufType)
		if chunk.debug == true {
			log.Println("initialization 2 fmt:", bufType)
		}
	case 3:

		log.Println("Rtmp Chunk ChunkBasicHeader 3")

	default:
		// 读取错误为识别的数据。
		buf := make([]byte, 1024)
		len, err := conn.Read(buf)
		if err != nil {
			log.Fatal("Rtmp Chunk Len and Err：", len, err)
		}
		log.Println("Rtmp Chunk Error Body =", len)
	}

	chunk.messagebody = make([]byte, chunk.messagelength)
	len, err := conn.Read(chunk.messagebody)
	if err != nil {
		log.Fatal("Rtmp chunk get body err: ", err, len)
	}

}

// Chunk Basic Header
func (chunk *chunk) basicHeader(conn net.Conn) {
	bufFMT := make([]byte, 1)
	bufflen, err := conn.Read(bufFMT)
	if bufflen != 1 || err != nil {
		log.Fatal("net.conn error")
	}

	if chunk.debug == true {
		log.Println("basicHeader fmt:", bufFMT)
	}

	chunk.fmt = int(bufFMT[0]) >> 6
	chunk.csid = int(bufFMT[0]) & 0x3f

	if chunk.csid == 0 {
		bufCsid := make([]byte, 1)
		conn.Read(bufCsid)
		chunk.csid = int(bufCsid[0]) + 64
	} else if chunk.csid == 1 {
		bufCsid := make([]byte, 2)
		conn.Read(bufCsid)
		chunk.csid = int(bufCsid[1])>>8 + int(bufCsid[0]) + 64
	}
}

//Type 0
func (chunk *chunk) type0(header []byte) {
	chunk.messagelength = int(header[3])<<16 + int(header[4])<<8 + int(header[5])
	chunk.messagetypeid = int(header[6])

}

//Type 1
func (chunk *chunk) type1(header []byte) {
	chunk.messagelength = int(header[3])<<16 + int(header[4])<<8 + int(header[5])
	chunk.messagetypeid = int(header[6])
}

//Type 2
func (chunk *chunk) type2(b []byte) {

}
