package rtmp

import (
	"encoding/binary"
	"log"

	"github.com/pennfly/rtmp-go/amf"
)

type Message struct {
	Conn   *Conn
	Server *Server
	Chunk  *Chunk
}

func (m *Message) setChunkSize() {
	payloadByte := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadByte, uint32(m.Conn.WriteChunkSize))
	c := Chunk{
		MessageTypeID: 1,
		SteamID:       m.Conn.SteamID,
		Payload:       payloadByte,
		Conn:          m.Conn,
	}
	c.SendChunk()
}

func (m *Message) connectResult() {

	var arrSour []amf.Value

	repVer := make(map[string]amf.Value)
	repVer["fmsVer"] = "FMS/3,0,1,123"
	repVer["capabilities"] = 31

	repStatus := make(map[string]amf.Value)
	repStatus["level"] = "status"
	repStatus["code"] = "NetConnection.Connect.Success"
	repStatus["description"] = "Connection succeeded"
	repStatus["objectEncoding"] = 0

	repSour := append(arrSour, "_result", 1, repVer, repStatus)

	c := Chunk{
		MessageTypeID: 20,
		SteamID:       m.Conn.SteamID,
		Payload:       amf.Encode(repSour),
		Conn:          m.Conn,
	}
	c.SendChunk()
}

func (m *Message) createSteamResult() {
	repByte := amf.Encode([]amf.Value{"_result", 4, nil, 4})
	c := Chunk{
		MessageTypeID: 20,
		SteamID:       3,
		Payload:       repByte,
		Conn:          m.Conn,
	}
	c.SendChunk()
}

func (m *Message) publishResult() {

	res := make(map[string]amf.Value)
	res["level"] = "status"
	res["code"] = "NetStream.Publish.Start"
	res["description"] = "Start publishing"
	resp := []amf.Value{"onStatus", 0, nil, res}

	c := Chunk{
		MessageTypeID: 20,
		SteamID:       4,
		Payload:       amf.Encode(resp),
		Conn:          m.Conn,
	}
	c.SendChunk()

}

func (m *Message) createSteam(item []amf.Value) {

	switch item[0] {
	case "connect":
		m.setChunkSize()
		m.connectResult()
	case "createStream":
		m.createSteamResult()
	case "publish":
		m.publishResult()
	default:

	}
}

func (m *Message) doAction() {
	switch m.Chunk.MessageTypeID {
	case 20, 17:
		item := amf.Decode(m.Chunk.Payload)
		if len(item) >= 1 {
			m.createSteam(item)
		}
	case 18, 15, 8, 9:
		//发送video Data
	default:
		log.Println("rtmp create steam had err typeid ->:", m.Chunk.MessageTypeID)
	}
}

func newMessage(chk *Chunk) error {
	msg := Message{
		Chunk:  chk,
		Conn:   chk.Conn,
		Server: chk.Conn.Server,
	}
	msg.doAction()

	return nil
}
