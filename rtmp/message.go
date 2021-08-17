package rtmp

import (
	"errors"
	"fmt"
	"log"
	"rtmp-go/amf"
	"rtmp-go/av"
)

// 处理消息
var VideoTime uint32

func (chk *Chunk) Handle(c *Conn) error {

	var flv av.FLV
	flv.GenFlv("room")
	defer flv.Close()

	for {
		payload, err := chk.readMsg()
		if err != nil {
			return err
		}
		header := chk.rChkList[chk.csid]
		switch header.MessageTypeID {
		case 20:
			item := amf.Decode(payload)
			if stop, err := chk.netCommands(item); err != nil || stop {
				return err
			}
		case 18, 15: // Metadata
			flv.AddTag(int(header.MessageTypeID), 0, payload[16:])
		case 8, 9: // Video data
			VideoTime += header.Timestamp
			flv.AddTag(int(header.MessageTypeID), VideoTime, payload)
		default:
			return errors.New("cant meet this MessageTypeID:" + string(header.MessageTypeID))
		}
	}
}

// 处理create stream
// bool 是否退出，通常当前推流结束
func (chk *Chunk) netCommands(item []amf.Value) (bool, error) {
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
			fmt.Println("处理messageStreamId=", int(tranId))
			if err := chk.sendMsg(20, 3, content); err != nil {
				return true, err
			}
		} else {
			return true, errors.New("err: connect->item[2].(float64);")
		}
		log.Println("createStream.finish.")

	// case "play":
	// 7.2.2.1. play . . . . . . . . . . . . . . . . . . . . . . . 38
	case "publish":
		res := make(map[string]amf.Value)
		res["level"] = "status"
		res["code"] = "NetStream.Publish.Start"
		res["description"] = "Start publishing"
		content := amf.Encode([]amf.Value{"onStatus", 0, nil, res})
		chk.sendMsg(20, 3, content)

	case "deleteStream":
		return true, nil

	// 协议不带，但obs发送了的
	case "releaseStream":
	case "FCPublish":
	case "FCUnpublish":
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
