package rtmp

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"github.com/penndev/rtmp/amf"
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

	switch msg.MessageType {
	case AMF0CommandMessage:
		values, err := amf.Decode0(msg.PayLoad)
		if err != nil {
			return err
		}
		log.Println("read message amf0 command message", values)
	case AMF3CommandMessage:
		values, err := amf.Decode3(msg.PayLoad)
		if err != nil {
			return err
		}
		log.Println("read message amf3 command message", values)
	default:
		return fmt.Errorf("unknown message type: %d", msg.MessageType)
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
