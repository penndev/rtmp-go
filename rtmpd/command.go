package rtmp

import (
	"encoding/binary"
	"log"

	"github.com/pennfly/rtmp-go/amf"
)

func connect(rep string) {
	if rep == "_result" {

	} else if rep == "_error" {

	}

}
func call()         {}
func close()        {}
func createStream() {}

/*------------------------------------*/
func play()         {}
func play2()        {}
func deleteStream() {}
func closeStream()  {}
func receiveAudio() {}
func receiveVideo() {}
func publish()      {}
func seek()         {}
func pause()        {}

/*------------------------------------*/

// CommandMessage 处理
func CommandMessage(msg *Chunk) {
	amfItem := amf.Decode(msg.Payload)
	log.Println("amf command:", amfItem)
	switch amfItem[0] {
	case "connect":
		msg.c.WriteBuffer(append(ReplySetChunkSize(), ReplyConnect("_result")...))
	case "releaseStream", "FCPublish":

	case "createStream":
		msg.c.WriteBuffer(ReplyCreateStream())
	case "publish":
		msg.c.WriteBuffer(ReplyPublish())
	default:
		log.Println("Oop Cant find the amf command:", amfItem)
	}
}

// ReplyConnect 回复
func ReplyConnect(res string) []byte {
	repM := make(map[string]amf.Value)
	repM["fmsVer"] = "FMS/3,0,1,123"
	repM["capabilities"] = 31

	repO := make(map[string]amf.Value)
	repO["level"] = "status"
	repO["code"] = "NetConnection.Connect.Success"
	repO["description"] = "Connection succeeded"
	repO["objectEncoding"] = 0

	var arrSour []amf.Value
	repSour := append(arrSour, "_result", 1, repM, repO)
	repByte := amf.Encode(repSour)

	// 处理开头Chunkid,Time,bodyLen
	i := make([]byte, 4)
	binary.BigEndian.PutUint32(i, uint32(len(repByte)))
	r := append([]byte{3, 0, 0}, i...)
	r = append(r, 20)
	r = append(r, 0, 0, 0, 0)
	return append(r, repByte...)
}

// ReplySetChunkSize 4096
func ReplySetChunkSize() []byte {
	return []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00}
}

// ReplyCreateStream 返回。
func ReplyCreateStream() []byte {
	var arrSour []amf.Value
	arrSour = append(arrSour, "_result", 4, nil, 4)
	repByte := amf.Encode(arrSour)
	return InitMsgHead(repByte, 3, 20)
}

func ReplyPublish() []byte {
	var arrSour []amf.Value
	res := make(map[string]amf.Value)
	res["level"] = "status"
	res["code"] = "NetStream.Publish.Start"
	res["description"] = "Start publishing"
	arrSour = append(arrSour, "onStatus", 0, nil, res)
	repByte := amf.Encode(arrSour)
	return InitMsgHead(repByte, 4, 20)
}

func InitMsgHead(payload []byte, steamid byte, typeid byte) []byte {
	i := make([]byte, 4)
	binary.BigEndian.PutUint32(i, uint32(len(payload)))
	r := append([]byte{steamid, 0, 0}, i...)
	r = append(r, typeid)
	r = append(r, 0, 0, 0, 0)
	return append(r, payload...)
}
