package rtmp

import (
	"encoding/binary"
	"log"

	"github.com/pennfly/rtmp-go/amf"
)

// Message 详细消息处理。
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

func (m *Message) streamBegin() {
	payloadByte := make([]byte, 4)
	controType := []byte{0, 0}
	binary.BigEndian.PutUint32(payloadByte, 0)
	payloadByte = append(controType, payloadByte...)
	c := Chunk{
		MessageTypeID: 4,
		SteamID:       m.Conn.SteamID,
		Payload:       payloadByte,
		Conn:          m.Conn,
	}
	c.SendChunk()
}

func (m *Message) respConnect() {

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

func (m *Message) respCreateSteam(n int) {
	repByte := amf.Encode([]amf.Value{"_result", n, nil, n})
	c := Chunk{
		MessageTypeID: 20,
		SteamID:       3,
		Payload:       repByte,
		Conn:          m.Conn,
	}
	c.SendChunk()
}

func (m *Message) respPublish() {
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

func (m *Message) respPlay() {
	res := make(map[string]amf.Value)
	res["level"] = "status"
	res["code"] = "NetStream.Play.Start"
	res["description"] = "Start playing"
	resp := []amf.Value{"onStatus", 0, nil, res}

	c := Chunk{
		MessageTypeID: 20,
		SteamID:       4,
		Payload:       amf.Encode(resp),
		Conn:          m.Conn,
	}
	c.SendChunk()
}

func (m *Message) createSteam() {
	item := amf.Decode(m.Chunk.Payload)

	switch item[0] {
	case "connect":
		m.setChunkSize()
		m.respConnect()
	case "createStream":
		tID := item[1]
		t, ok := tID.(float64)
		if !ok {
			t = 0
		}
		m.respCreateSteam(int(t))
	case "publish":
		m.respPublish()
	case "play":
		m.streamBegin() // m.StreamIsRecorded
		m.respPlay()    // onStatus-play-reset
	default:
		log.Println("Rtmp Message not resp:->", item[0])
	}
}

//分发消息。
func (m *Message) assort() {
	switch m.Chunk.MessageTypeID {
	case 20, 17:
		m.createSteam()
	case 18, 15, 8, 9:
		//发送video Data
	default:
		log.Println("Rtmp Message had err typeid ->:", m.Chunk.MessageTypeID)
	}
}

func newMessage(chk *Chunk) error {
	msg := Message{
		Chunk:  chk,
		Conn:   chk.Conn,
		Server: chk.Conn.Server,
	}
	msg.assort()
	return nil
}
