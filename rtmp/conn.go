package rtmp

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/penndev/rtmp/amf"
)

type Conn struct {
	chk *Chunk
	nc  net.Conn

	App    string
	Stream string

	IsPublish bool
}

// 根据返回值处理连接是否继续
// return true 继续下一步
func (c *Conn) onConnect(app string) bool {
	c.App = app
	// fmt.Println("c.app ->", c.App)
	return true
}

func (c *Conn) onPublish(stream string) bool {
	//验证密钥。
	// fmt.Println("c.stream ->", stream)
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

func (c *Conn) HandleStream() error {
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

func (c *Conn) handlePlay(subscriberCh <-chan Pack) error {
	clientCh := make(chan error)
	go func() {
		err := c.handlePublishing(func(pk Pack) {})
		clientCh <- err
	}()
	for {
		select {
		case pk, ok := <-subscriberCh:
			if !ok {
				c.chk.setStreamEof(DefaultStreamID)
				return errors.New("subscriber chan handle close")
			}
			c.chk.sendPack(DefaultStreamID, pk)
		case clientCh := <-clientCh:
			return clientCh
		}
	}
}

// 5.2. Handshake . . . . . . . . . . . . . . . . . . . . . . . . 7
// 5.2.1. Handshake Sequence . . . . . . . . . . . . . . . . . . 7
// 5.2.2. C0 and S0 Format . . . . . . . . . . . . . . . . . . . 7
// 5.2.3. C1 and S1 Format . . . . . . . . . . . . . . . . . . . 8
// 5.2.4. C2 and S2 Format . . . . . . . . . . . . . . . . . . . 8
// 5.2.5. Handshake Diagram . . . . . . . . . . . . . . . . . . 10
func (c *Conn) Handshake() error {
	c.nc.SetReadDeadline(time.Now().Add(10 * time.Second))
	defer c.nc.SetReadDeadline(time.Time{})

	// C0: version(1)
	c0 := make([]byte, 1)
	_, err := io.ReadFull(c.nc, c0)
	if err != nil {
		return err
	}
	if c0[0] != VERSION {
		return errors.New("rtmp version is not support")
	}

	// S0 + S1
	// S0: version(1)
	// S1: time(4) + zero(4) + random(1528)
	s0s1 := make([]byte, 1537)
	s0s1[0] = VERSION
	s1rand := s0s1[9:1537]
	if _, err = rand.Read(s1rand); err != nil {
		return err
	}
	start := time.Now()
	n, err := c.nc.Write(s0s1)
	if err != nil {
		return err
	}
	if n != 1537 {
		return errors.New("write s1 failed")
	}

	// C1: time(4) + zero(4) + random(1528)
	c1 := make([]byte, 1536)
	_, err = io.ReadFull(c.nc, c1)
	if err != nil {
		return err
	}

	// // Flash Player 9 RTMP-e
	// if !bytes.Equal(c1[4:8], []byte{0, 0, 0, 0}) {
	// 	todo: handle flash player 9 rtmp-e
	// }

	// S2: time = C1.time, time2 = 从发出 S1 到此刻的毫秒, random = C1.random
	s2 := make([]byte, 1536)
	copy(s2[0:4], c1[0:4])
	s2Time2 := uint32(time.Since(start).Milliseconds())
	binary.BigEndian.PutUint32(s2[4:8], s2Time2)
	copy(s2[8:], c1[8:])
	n, err = c.nc.Write(s2)
	if err != nil {
		return err
	}
	if n != len(s2) {
		return errors.New("write s2 failed")
	}

	// C2: time(4) + time2(4) + random echo(1528)
	c2 := make([]byte, 1536)
	_, err = io.ReadFull(c.nc, c2)
	if err != nil {
		return err
	}
	// C2.time 应回显 S1.time。当前 S1.time 为 0。
	if !bytes.Equal(c2[0:4], s0s1[1:5]) {
		return errors.New("c2 time is not s1 time")
	}
	if !bytes.Equal(c2[8:], s1rand) {
		return errors.New("c2 random is not s1 random")
	}

	return nil
}

func NewConn(nc net.Conn) *Conn {
	return &Conn{
		chk: newChunk(nc),
		nc:  nc,
	}
}
