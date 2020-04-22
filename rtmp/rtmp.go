package rtmp

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"net"
	"time"
)

//Connnect rtmp单个链接的struct
type Connnect struct {
	r          *bufio.Reader
	w          *bufio.Writer
	rw         *bufio.ReadWriter // Read and Write,用来握手
	conn       net.Conn
	rwByteSize map[string]uint32
	createTime string
	remoteAddr string
	url        string
	appName    string
	chunkSize  uint32
}

// Close 关闭链接处理
func (c *Connnect) Close() {
	c.conn.Close()
	log.Println("Conn is close :", c.remoteAddr)
}

// Read 读取net.Conn 数据，并且增加统计
func (c *Connnect) Read(length int) ([]byte, error) {
	buf := make([]byte, length)
	l, err := io.ReadFull(c.r, buf)
	c.rwByteSize["read"] += uint32(l)
	return buf, err
}

//ReadByte 读取单个字节
func (c *Connnect) ReadByte() (byte, error) {
	c.rwByteSize["read"]++
	return c.r.ReadByte()
}

// WriteBuffer 写数据。
func (c *Connnect) WriteBuffer(buf []byte) error {
	l, err := c.w.Write(buf)
	c.rwByteSize["write"] += uint32(l)
	c.w.Flush()
	return err
}

// HandShake 处理rtmp握手。
func (c *Connnect) HandShake() error {
	err := Handshake(c)
	return err
}

//ReadMsg 读取一个需要处理的消息。循环处理，如果是协议控制消息则继续读取。
func (c *Connnect) ReadMsg() (Chunk, error) {
	msg, err := ReadChunkMsg(c)
	if err != nil {
		return msg, err
	}
	// 处理与提取Msg无关的控制协议。
	switch int(msg.MessageTypeID) {
	case SetChunkSize:
		c.chunkSize = binary.BigEndian.Uint32(msg.Payload)
		log.Println("Rtmp SetChunkSize", c.chunkSize)
		msg, err = c.ReadMsg()
	case AbortMessage:
		log.Println("Rtmp AbortMessage")
		msg, err = c.ReadMsg()
	case Acknowledgement:
		log.Println("Rtmp Acknowledgement")
		msg, err = c.ReadMsg()
	case WindowAcknowledgementSize:
		log.Println("Rtmp WindowAcknowledgementSize")
		msg, err = c.ReadMsg()
	case SetPeerBandwidth:
		log.Println("Rtmp SetPeerBandwidth")
		msg, err = c.ReadMsg()
	}

	return msg, err
}

//Steam 处理流数据
func (c *Connnect) Steam() error {
	// 处理消息
	for {

		msg, err := c.ReadMsg()
		if err != nil {
			return err
		}

		switch int(msg.MessageTypeID) {
		case CommandMessageAMF0:
			CommandMessage(&msg)
		default:
			log.Println("new msg:Typeid,MsgLength,Steamid", msg.MessageTypeID, msg.MessageLength, msg.MessageStreamID)
		}
	}
}

// NewConnnect 初始化一个新的链接。
func NewConnnect(conn net.Conn) Connnect {
	var c Connnect
	c.conn = conn
	c.r = bufio.NewReader(conn)
	c.w = bufio.NewWriter(conn)
	c.rw = bufio.NewReadWriter(c.r, c.w)

	c.rwByteSize = make(map[string]uint32)
	c.rwByteSize["read"] = 0
	c.rwByteSize["write"] = 0

	c.remoteAddr = conn.RemoteAddr().String()
	c.createTime = time.Now().String()
	c.chunkSize = ChunkDefaultSize
	return c
}
