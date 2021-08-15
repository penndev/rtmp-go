package rtmp

import (
	"errors"
	"fmt"
	"rtmp-go/amf"
)

type MsgHeader struct {
	Timestamp       uint32 // 3 byte
	MessageLength   uint32 // 3 byte
	MessageTypeID   byte   // 1 byte
	MessageStreamID uint32 // 4 byte
	ExtendTimestamp uint32 // 4 byte
}

// 处理消息
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
			if run, err := chk.netCommands(item); err != nil || run == false {
				return err
			}
		default:
			fmt.Println("cant meet this MessageTypeID:", header.MessageTypeID)
		}
	}
}

// 处理create stream
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
	case "Call":
		// 7.2.1.2. Call . . . . . . . . . . . . . . . . . . . . . . . 35
	case "createStream":
		if tranId, ok := item[1].(float64); ok {
			content := amf.Encode([]amf.Value{"_result", int(tranId), nil, int(tranId)})
			fmt.Println("处理messageStreamId=", int(tranId))
			fmt.Println(content)
			if err := chk.sendMsg(20, 3, content); err != nil {
				return true, err
			}
		} else {
			return true, errors.New("err: connect->item[2].(float64);")
		}
		fmt.Println("createStream.finish.")

	case "play":
		// 7.2.2.1. play . . . . . . . . . . . . . . . . . . . . . . . 38
	case "publish":
		res := make(map[string]amf.Value)
		res["level"] = "status"
		res["code"] = "NetStream.Publish.Start"
		res["description"] = "Start publishing"
		content := amf.Encode([]amf.Value{"onStatus", 0, nil, res})
		chk.sendMsg(20, 3, content)

	case "deleteStream":
		return false, nil
	// 协议不带，但obs发送了的
	case "releaseStream":
	// 协议不带，但obs发送了的
	case "FCPublish":
	case "FCUnpublish":
	default:
		// 7.2.2.2. play2 . . . . . . . . . . . . . . . . . . . . . . 42
		// 7.2.2.4. receiveAudio . . . . . . . . . . . . . . . . . . . 44
		// 7.2.2.5. receiveVideo . . . . . . . . . . . . . . . . . . . 45
		// 7.2.2.7. seek . . . . . . . . . . . . . . . . . . . . . . . 46
		// 7.2.2.8. pause . . . . . . . . . . . . . . . . . . . . . . 47
		panic("netCommands handle error:" + item[0].(string))
	}
	return true, nil
}
