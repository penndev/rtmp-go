package rtmp

import (
	"errors"
	"fmt"
	"net"

	"github.com/penndev/rtmp-go/amf"
)

type Conn struct {
	chk *Chunk

	App    string
	Stream string

	IsPublish bool
}

// 根据返回值处理连接是否继续
// return true 继续下一步
func (c *Conn) onConnect(app string) bool {
	c.App = app
	fmt.Println("c.app ->", c.App)
	return true
}

func (c *Conn) onPublish(stream string) bool {
	//验证密钥。
	fmt.Println("c.stream ->", stream)
	c.Stream = stream
	c.IsPublish = true

	// 验证必须可以才允许连接。

	// addrs, _ := net.LookupHost(name)
	// log.Print("传输了信号: ", addrs[0], ":1935/", c.App, "/", c.Stream)
	// log.Print("RTMP播放地址: rtmp://", addrs[0], ":1935/", c.App, "/", c.Stream)
	// log.Print("http-flv播放地址: http://", addrs[0], ":8080/", c.App, "/", c.Stream, ".flv")
	return true
}

func (c *Conn) onPlay(stream string) bool {
	c.Stream = stream
	c.IsPublish = false
	return true
}

func (c *Conn) handleConnect() error {
	read := 0
	for {
		pk, err := c.chk.handlesMsg()
		if err != nil {
			return err
		}
		if pk.MessageTypeID != 20 {
			return errors.New("netConnectionCommand err: cant handle type id" + fmt.Sprint(pk.MessageTypeID))
		}
		item := amf.Decode(pk.PayLoad)
		switch item[0] {
		case "connect":
			read = 1
			media, ok := item[2].(map[string]amf.Value)
			if !ok {
				return errors.New("netConnectionCommand connect err:) catn find media")
			}
			app, ok := media["app"].(string)
			if !ok {
				return errors.New("netConnectionCommand connect err:) cant find app")
			}
			stu := c.onConnect(app)
			c.chk.setChunkSize(SetChunkSize)
			c.chk.sendMsg(20, 3, respConnect(stu))
			if !stu {
				return errors.New("netConnectionCommand connect err:) cat conntect app " + app)
			}
			c.chk.setWindowAcknowledgementSize(2500000)
		case "createStream":
			tranId, ok := item[1].(float64)
			if !ok {
				return errors.New("netConnectionCommand createStream err:) cant find tranid")
			}
			c.chk.sendMsg(20, 3, respCreateStream(true, int(tranId), DefaultStreamID))
			if read == 1 {
				read = 2
			} else {
				return errors.New("netConnectionCommand err:) not do connect action")
			}
		case "releaseStream":
		case "FCPublish":
		default:
			return errors.New("netConnectionCommand err: cant handle command->" + fmt.Sprint(item[0]))
		}
		if read == 2 {
			break
		}
	}
	return nil
}

func (c *Conn) handleStream() error {
	for {
		pk, err := c.chk.handlesMsg()
		if err != nil {
			return err
		}
		if pk.MessageTypeID != 20 {
			return errors.New("netStreamCommand err: cant handle type id" + fmt.Sprint(pk.MessageTypeID))
		}
		item := amf.Decode(pk.PayLoad)
		switch item[0] {
		case "publish":
			streamId, ok := item[1].(float64)
			if !ok {
				return errors.New("netStreamCommand err: streamId error")
			}
			streamType, ok := item[4].(string)
			if !ok || streamType != "live" {
				return errors.New("netStreamCommand err: streamType error")
			}
			streamName, ok := item[3].(string)
			if !ok {
				return errors.New("netStreamCommand err: streamName error")
			}
			status := c.onPublish(streamName)
			c.chk.sendMsg(20, 3, respPublish(status))
			if !status {
				return errors.New("netStreamCommand err: streamname checkout fail")
			}
			c.chk.setStreamBegin(uint32(streamId))
			return nil
		case "play":
			streamName, ok := item[3].(string)
			if !ok {
				return errors.New("netStreamCommand play err: streamName error")
			}
			status := c.onPlay(streamName)
			c.chk.sendMsg(20, 3, respPlay(status))
			if !status {
				return errors.New("netStreamCommand play err: streamname checkout fail")
			}
			return nil
		}
	}
}

func (c *Conn) handlePublishing(cb func(Pack)) error {
	for {
		pk, err := c.chk.handlesMsg()
		if err != nil {
			return err
		}
		switch pk.MessageTypeID {
		case 8, 9, 15, 18:
			//不允许向已关闭的chan传输数据。
			// fmt.Println("收到消息->", pk.MessageTypeID)
			cb(pk)
		case 20:
			item := amf.Decode(pk.PayLoad)
			switch item[0] {
			case "FCUnpublish":
			case "deleteStream":
				return errors.New("handle deleteStream rtmp message")
			default:
				if ms, ok := item[0].(string); ok {
					return errors.New("handle undefined rtmp message:" + ms)
				} else {
					return errors.New("handle undefined rtmp message")
				}
			}
		default:
			return errors.New("handle undefined rtmp message type:" + fmt.Sprint(pk.MessageTypeID))
		}
	}
}

func (c *Conn) handlePlay() error {
	// 多路复用
	select {
	// 监听用户关闭消息
	// 监听play播放的流
	}
}

func NewConn(nc net.Conn) *Conn {
	return &Conn{
		chk: newChunk(nc),
	}
}
