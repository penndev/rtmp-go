package rtmp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"rtmp-go/amf"
)

// 处理消息
func (chk *Chunk) Handle(c *Conn) error {
	defer func() {
		if c.IsPusher {
			c.serve.WorkPool.Close(c.App, c.PackChan)
		} else {
			delete(c.serve.WorkPool.PlayList[c.App], c.PackChan)
		}
		// close(c.PackChan)
		//关闭了两次chan 引起 panic
	}()
	for {
		payload, err := chk.readMsg()
		if err != nil {
			return err
		}
		header := chk.rChkList[chk.csid]
		switch header.MessageTypeID {
		case 20:
			item := amf.Decode(payload)
			if stop, err := chk.netCommands(item, c); err != nil || stop {
				return err
			}
		case 18, 15: // Metadata
			pk := Pack{
				Type:    header.MessageTypeID,
				Time:    0,
				Content: payload,
			}
			c.onPushMate(pk)
		case 8, 9: // Video data
			pk := Pack{
				Type:    header.MessageTypeID,
				Time:    header.Timestamp,
				Content: payload,
			}
			c.onPushAv(pk)
		case 4:
			item := amf.Decode(payload)
			fmt.Println("message 4:", item)
		default:
			return errors.New("Cont handle todo MessageTypeID:" + fmt.Sprint(header.MessageTypeID))
		}
	}
}

func (chk *Chunk) netCommands(item []amf.Value, c *Conn) (bool, error) {
	switch item[0] {
	case "connect":
		media, ok := item[2].(map[string]amf.Value)
		if !ok {
			return true, errors.New("err: connect->item[2].(map[string]amf.Value)")
		}
		app := media["app"].(string)
		c.onConnect(app)

		repVer := make(map[string]amf.Value)
		repVer["fmsVer"] = "FMS/3,0,1,123"
		repVer["capabilities"] = 31
		repStatus := make(map[string]amf.Value)
		repStatus["level"] = "status"
		repStatus["code"] = "NetConnection.Connect.Success"
		repStatus["description"] = "Connection succeeded."
		repStatus["objectEncoding"] = 3
		content := amf.Encode([]amf.Value{"_result", 1, repVer, repStatus})
		chk.sendMsg(20, 3, content)
		chk.setWindowAcknowledgementSize(2500000)
	case "createStream":
		if tranId, ok := item[1].(float64); ok {
			content := amf.Encode([]amf.Value{"_result", int(tranId), nil, int(tranId)})
			if err := chk.sendMsg(20, 3, content); err != nil {
				return true, err
			}
		} else {
			return true, errors.New("err: connect->item[2].(float64);")
		}
	case "play":
		streamContent := make([]byte, 6)
		binary.BigEndian.PutUint32(streamContent[2:], 4)
		chk.sendMsg(4, 2, streamContent)

		res := make(map[string]amf.Value)
		res["level"] = "status"
		res["code"] = "NetStream.Play.Start"
		res["description"] = "Start playing"
		resp := amf.Encode([]amf.Value{"onStatus", 0, nil, res})
		chk.sendMsg(20, 3, resp)

		c.onPlay()
		pack := c.serve.WorkPool.MateList[c.App]
		chk.sendMsg(20, 3, pack.Content)

		packV := c.serve.WorkPool.VideoList[c.App]
		chk.sendMsg(9, 4, packV.Content)

		packA := c.serve.WorkPool.AudioList[c.App]
		chk.sendMsg(8, 4, packA.Content)

		go func() {
			//首先初始化关键帧
			for pck := range c.PackChan {
				if pck.Type == 9 {
					k := pck.Content[0]
					if k>>4 == 1 {
						chk.sendAv(pck.Type, 4, pck.Time, pck.Content)
						break
					}
				}
			}
			for pck := range c.PackChan {
				chk.sendAv(pck.Type, 4, pck.Time, pck.Content)
			}
			streamContent := make([]byte, 6)
			binary.BigEndian.PutUint32(streamContent[2:], 4)
			streamContent[1] = 1
			chk.sendMsg(4, 2, streamContent)
			// 关闭。
			c.Close()
		}()
	case "publish":
		res := make(map[string]amf.Value)
		res["level"] = "status"
		res["code"] = "NetStream.Publish.Start"
		res["description"] = "Start publishing"
		content := amf.Encode([]amf.Value{"onStatus", 0, nil, res})
		chk.sendMsg(20, 3, content)

		// var app, stream string
		// var ok bool
		// if app, ok = item[4].(string); !ok {
		// 	return true, errors.New("cant find app name")
		// }
		// if stream, ok = item[3].(string); !ok {
		// 	return true, errors.New("cant find app stream")
		// }
		// c.onSetPush(app, stream)
	case "deleteStream":
		return true, nil
	// 处理兼容性
	// 协议不带，但(obs vlc ffmpeg)发送了的
	case "releaseStream":
	case "FCPublish":
	case "FCUnpublish":
	case "getStreamLength":
	//  处理兼容性匹配End
	default:
		return true, errors.New("netCommands handle error:" + item[0].(string))
	}
	return false, nil
}
