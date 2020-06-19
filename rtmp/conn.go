package rtmp

import (
	"bufio"
	"fmt"
	"io"
	"net"
)

// conn read size and write size
type rwByteSize struct {
	read  int
	write int
}

// A Conn represents the server side of an HTTP connection.
type Conn struct {
	// server is the server on which the connection arrived.
	// Immutable; never nil.
	Server *Server
	rwc    net.Conn

	r          *bufio.Reader
	w          *bufio.Writer
	rwByteSize *rwByteSize
	remoteAddr string

	ReadChunkSize  int
	WriteChunkSize int
	SteamID        uint32
	ChunkLists     map[uint32]Chunk
	App            string
	Stream         string
}

func (c *Conn) handShake() error {
	err := ServeHandShake(c.rwc)
	return err
}

// Close this connection.
func (c *Conn) Close() {
	c.rwc.Close()
}

// ReadFull 读取net.Conn 数据，并且增加统计
func (c *Conn) ReadFull(length int) ([]byte, error) {
	buf := make([]byte, length)
	l, err := io.ReadFull(c.r, buf)
	c.rwByteSize.read += l
	return buf, err
}

//ReadByte 读取单个字节
func (c *Conn) ReadByte() (byte, error) {
	// log.Println(c.rwByteSize.read)
	c.rwByteSize.read++
	return c.r.ReadByte()
}

// Write 写数据。
func (c *Conn) Write(buf []byte) (int, error) {
	l, err := c.w.Write(buf)
	c.rwByteSize.write += l
	c.w.Flush()
	return l, err
}

func (c *Conn) stream() error {
	chk := newChunk(c)
	//msg := newMsg()
	for {
		if err := chk.ReadMsg(); err != nil {
			return err
		}
		newMessage(chk)
	}
}

// 开始处理Rtmp
func (c *Conn) serve() {
	c.remoteAddr = c.rwc.RemoteAddr().String()
	defer c.Close()
	//Handshake
	if err := c.handShake(); err != nil {
		fmt.Println("c.handShake err ->:", err)
	}

	if err := c.stream(); err != nil {
		fmt.Println("c.stream err ->:", err)
	}
}

// Return a Instantiated method
func newConn(srv *Server, rw net.Conn) *Conn {
	conn := &Conn{
		Server:         srv,
		ReadChunkSize:  srv.ChunkSize, //单个client独立一个chunksize。
		WriteChunkSize: 4096,

		rwc:        rw,
		r:          bufio.NewReader(rw),
		w:          bufio.NewWriter(rw),
		remoteAddr: rw.RemoteAddr().String(),
		rwByteSize: &rwByteSize{},

		SteamID:    3,
		ChunkLists: make(map[uint32]Chunk),
	}
	return conn
}
