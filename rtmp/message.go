package rtmp

import (
	"errors"
	"fmt"
	"log"
	"rtmp-go/amf"
)

type Pack struct {
	ChunkMessageHeader
	PayLoad []byte
}

func netConnectionCommand(chk *Chunk, conn *Conn) error {
	read := 0
	for {
		pk, err := chk.handlesMsg()
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
			stu := conn.onConnect(app)
			chk.setChunkSize(SetChunkSize)
			chk.sendMsg(20, 3, respConnect(stu))
			if !stu {
				return errors.New("netConnectionCommand connect err:) cat conntect app " + app)
			}
			chk.setWindowAcknowledgementSize(2500000)
		case "createStream":
			tranId, ok := item[1].(float64)
			if !ok {
				return errors.New("netConnectionCommand createStream err:) cant find tranid")
			}
			chk.sendMsg(20, 3, respCreateStream(true, int(tranId), DefaultStreamID))
			if read == 1 {
				read = 2
			} else {
				return errors.New("netConnectionCommand err:) not do connect action")
			}
		case "releaseStream":
		case "FCPublish":
		default:
			log.Println("netConnectionCommand err: cant handle command->", item[0])
		}
		if read == 2 {
			break
		}
	}
	return nil
}

func netStreamCommand(chk *Chunk, conn *Conn) error {
	for {
		pk, err := chk.handlesMsg()
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
			status := conn.onPublish(streamName)
			chk.sendMsg(20, 3, respPublish(status))
			if !status {
				return errors.New("netStreamCommand err: streamname checkout fail")
			}
			chk.setStreamBegin(uint32(streamId))
			return nil
		case "play":
			streamName, ok := item[3].(string)
			if !ok {
				return errors.New("netStreamCommand play err: streamName error")
			}
			status := conn.onPlay(streamName)
			chk.sendMsg(20, 3, respPlay(status))
			if !status {
				return errors.New("netStreamCommand play err: streamname checkout fail")
			}
			return nil
		}

	}
}

func handle(chk *Chunk, conn *Conn) {
	log.Println("handle - - - - - -> 创建完成")
	defer log.Println("handle - - - - - -> 回收完成")
	for {
		pk, err := chk.handlesMsg()
		if err != nil {
			//true  我方关闭了tcp
			//false 客户端关闭了tcp
			if !conn.Closed {
				conn.onClose()
			}
			return
		}
		switch pk.MessageTypeID {
		case 8, 9, 15, 18:
			if conn.Closed {
				return
			}
			conn.AVPackChan <- pk
		case 20:
			item := amf.Decode(pk.PayLoad)
			switch item[0] {
			case "FCUnpublish":
			case "deleteStream":
				// 主动关闭
				conn.Closed = true
				close(conn.AVPackChan)
				conn.onClose()
				return
			default:
				log.Println("未遇到的消息(type 20)->", item[0])
			}
		default:
			log.Println("未遇到的type->", pk.MessageTypeID)
		}
	}
}

func respConnect(b bool) []byte {
	if !b {
		return amf.Encode([]amf.Value{"_error", 1, nil, nil})
	}
	repVer := make(map[string]amf.Value)
	repVer["fmsVer"] = "FMS/3,0,1,123"
	repVer["capabilities"] = 31
	repStatus := make(map[string]amf.Value)
	repStatus["level"] = "status"
	repStatus["code"] = "NetConnection.Connect.Success"
	repStatus["description"] = "Connection succeeded."
	repStatus["objectEncoding"] = 3
	return amf.Encode([]amf.Value{"_result", 1, repVer, repStatus})
}

func respCreateStream(b bool, transaId int, streamId int) []byte {
	return amf.Encode([]amf.Value{"_result", transaId, nil, streamId})
}

func respPublish(b bool) []byte {
	res := make(map[string]amf.Value)
	res["level"] = "status"
	if b {
		res["code"] = "NetStream.Publish.Start"
	} else {
		res["code"] = "NetStream.Publish.BadName"
	}
	res["description"] = "Start publishing"
	return amf.Encode([]amf.Value{"onStatus", 0, nil, res})
}

func respPlay(b bool) []byte {
	res := make(map[string]amf.Value)
	res["level"] = "status"
	if b {
		res["code"] = "NetStream.Play.Start"
	} else {
		res["code"] = "NetStream.Play.Failed"
	}
	res["description"] = "Start playing"
	return amf.Encode([]amf.Value{"onStatus", 0, nil, res})
}
