package rtmp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"rtmp-go/amf"
)

// 处理消息
var VideoTime uint32

func (chk *Chunk) Handle(c *Conn) error {

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
				c.onPushStop()
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
			VideoTime += header.Timestamp
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
			return errors.New("cant meet this MessageTypeID:" + fmt.Sprint(header.MessageTypeID))
		}
	}
}

// 处理create stream
// bool 是否退出，通常当前推流结束
func (chk *Chunk) netCommands(item []amf.Value, c *Conn) (bool, error) {
	switch item[0] {
	case "connect":
		_, ok := item[2].(map[string]amf.Value)
		if !ok {
			return true, errors.New("err: connect->item[2].(map[string]amf.Value)")
		}

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
	// case "Call":
	// 7.2.1.2. Call . . . . . . . . . . . . . . . . . . . . . . . 35
	case "createStream":
		if tranId, ok := item[1].(float64); ok {
			content := amf.Encode([]amf.Value{"_result", int(tranId), nil, int(tranId)})
			if err := chk.sendMsg(20, 3, content); err != nil {
				return true, err
			}
		} else {
			return true, errors.New("err: connect->item[2].(float64);")
		}
		log.Println("createStream.finish.")

	case "play":

		streamContent := make([]byte, 6)
		binary.BigEndian.PutUint32(streamContent[2:], 4)
		chk.sendMsg(4, 3, streamContent)

		res := make(map[string]amf.Value)
		res["level"] = "status"
		res["code"] = "NetStream.Play.Start"
		res["description"] = "Start playing"
		resp := amf.Encode([]amf.Value{"onStatus", 0, nil, res})
		chk.sendMsg(20, 3, resp)

		c.onPlay()
		pack := c.serve.WorkPool.MateList["live"]
		chk.sendMsg(20, 3, pack.Content)
		go func() {
			//首先初始化关键帧
			for pck := range c.PackChan {
				if pck.Type == 9 {
					k := pck.Content[0]
					if k>>4 == 1 {
						log.Println(pck.Type, pck.Time, len(pck.Content))
						chk.sendAv(pck.Type, 4, pck.Time, pck.Content)
						break
					}
				}
			}
			// panic("debug.")
			for pck := range c.PackChan {
				chk.sendAv(pck.Type, 4, pck.Time, pck.Content)
			}
		}()

	// 7.2.2.1. play . . . . . . . . . . . . . . . . . . . . . . . 38
	case "publish":
		res := make(map[string]amf.Value)
		res["level"] = "status"
		res["code"] = "NetStream.Publish.Start"
		res["description"] = "Start publishing"
		content := amf.Encode([]amf.Value{"onStatus", 0, nil, res})
		chk.sendMsg(20, 3, content)

		var app, stream string
		var ok bool
		if app, ok = item[4].(string); !ok {
			return true, errors.New("cant find app name")
		}
		if stream, ok = item[3].(string); !ok {
			return true, errors.New("cant find app stream")
		}
		c.onSetPush(app, stream)
	case "deleteStream":
		return true, nil
	// 协议不带，但obs发送了的
	case "releaseStream":
	case "FCPublish":
	case "FCUnpublish":
	case "getStreamLength":
	default:
		// 7.2.2.2. play2 . . . . . . . . . . . . . . . . . . . . . . 42
		// 7.2.2.4. receiveAudio . . . . . . . . . . . . . . . . . . . 44
		// 7.2.2.5. receiveVideo . . . . . . . . . . . . . . . . . . . 45
		// 7.2.2.7. seek . . . . . . . . . . . . . . . . . . . . . . . 46
		// 7.2.2.8. pause . . . . . . . . . . . . . . . . . . . . . . 47
		return true, errors.New("netCommands handle error:" + item[0].(string))
	}
	return false, nil
}
