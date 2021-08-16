package rtmp

import (
	"bufio"
	"fmt"
	"net"
)

//依据rtmp对tcp进行封装
type Conn struct {
	// Serve 数据结构
	// 用来访问运行时数据
	serve *Serve

	// Tcp 网络IO
	// 进行数据通讯处理
	rwc *net.Conn

	App    string
	Stream string

	//是否有推送消息体的权限。
	//是否是主播
	IsPusher bool
}

// 握手
func (c *Conn) handShake() error {
	err := ServeHandShake(*c.rwc)
	return err
}

// 关闭rtmp连接，做一些清理。
func (c *Conn) Close() {
	(*c.rwc).Close()
}

func newConn() *Conn {
	return nil
}

//处理Rtmp消息协议
func (c *Conn) Connect() {
	defer c.Close()

	// 握手
	if err := c.handShake(); err != nil {
		fmt.Println(err)
	}

	// 创建 Chunk Stream
	chk := Chunk{
		r:        bufio.NewReader(*c.rwc),
		w:        bufio.NewWriter(*c.rwc),
		rChkSize: uint32(DefaultChunkSize),
		wChkSize: uint32(DefaultChunkSize),
		rChkList: make(map[uint32]*MsgHeader),
		wChkList: make(map[uint32]*MsgHeader),
	}

	//阻塞处理
	if err := chk.Handle(c); err != nil {
		fmt.Println(err)
	}
}
