package rtmp

import (
	"bufio"
	"net"
)

// conn read size and write size
type rwByteSize struct {
	read  int
	write int
}

//依据rtmp对tcp进行封装
type Conn struct {
	// Serve 数据结构
	// 用来访问运行时数据
	serve *Serve
	// Tcp 网络IO
	// 进行数据通讯处理
	rwc *net.Conn
	r   *bufio.Reader
	w   *bufio.Writer
	//读写字节统计
	//rtmp协议中的需要
	ReadChunkSize  int
	WriteChunkSize int

	SteamID uint32
	// ChunkLists     map[uint32]Chunk
	// SendChunkLists map[uint32]Chunk
	App    string
	Stream string
	//是否有推送消息体的权限。
	//是否是主播
	IsPusher bool
}

// // 开始处理 流
// func (c *Conn) stream() error {
// 	chk := newChunk(c)
// 	for {
// 		//read chunk message
// 		if err := chk.ReadMsg(); err != nil {
// 			return err
// 		}
// 		//ctrl message
// 		if err := newMessage(chk); err != nil {
// 			return err
// 		}
// 		//exit the client
// 		if c.closed {
// 			return nil
// 		}
// 	}
// }

// 关闭rtmp连接，做一些清理。
func (c *Conn) Close() {
	(*c.rwc).Close()
}

//处理Rtmp消息协议
func (c *Conn) connect() {
	defer c.Close()
}
