package rtmp

import (
	"bufio"
	"log"
	"net"
)

type Conn struct {
	Chunk
	nc net.Conn
}

func (c *Conn) HandleStream() error {
	msg, err := c.Read()
	if err != nil {
		return err
	}
	log.Println("read message", msg.MessageType)
	return nil
}

func NewConn(nc net.Conn) *Conn {
	return &Conn{
		nc: nc,
		Chunk: Chunk{
			r:               bufio.NewReader(nc),
			w:               bufio.NewWriter(nc),
			readStreamList:  make(map[int]*Message),
			writeStreamList: make(map[int]*MessageHeader),
			readChunkSize:   DEFAULT_CHUNK_SIZE,
			writeChunkSize:  DEFAULT_CHUNK_SIZE,
		},
	}
}
