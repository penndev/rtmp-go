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
		SteamID:       2,
		Payload:       payloadByte,
		Conn:          m.Conn,
	}
	c.SendChunk()
}

func (m *Message) streamBegin() {
	payloadByte := make([]byte, 4)
	controType := []byte{0, 0}
	binary.BigEndian.PutUint32(payloadByte, 4)
	payloadByte = append(controType, payloadByte...)
	c := Chunk{
		MessageTypeID: 4,
		SteamID:       3,
		Payload:       payloadByte,
		Conn:          m.Conn,
	}
	c.SendChunk()
}

func (m *Message) respPublish(steam amf.Value) {

	m.Conn.Stream = steam.(string)
	pusher := make(chan byte)
	log.Println("live name := ", m.Conn.App+"/"+m.Conn.Stream)
	m.Server.WorkPool[m.Conn.App+"/"+m.Conn.Stream] = &WorkPool{
		Pusher:   pusher,
		Metadata: Chunk{},
		Player:   []chan Chunk{},
	}

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

func (m *Message) respPlay(steam amf.Value) {
	m.Conn.Stream = steam.(string)
	name := m.Conn.App + "/" + m.Conn.Stream
	if _, ok := m.Server.WorkPool[name]; !ok {
		m.Conn.Close()
		return
	}

	m.streamBegin()

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

	defer func() {
		metadata := m.Server.WorkPool[name].Metadata
		metadata.Conn = m.Conn
		metadata.SendChunk()

		player := make(chan Chunk)
		m.Server.WorkPool[name].Player = append(m.Server.WorkPool[name].Player, player)
		for {
			x := <-player
			//log.Println(x.Format, x.SteamID, x.Timestamp, x.MessageLength, x.MessageTypeID, x.MessageStreamID)
			x.Conn = m.Conn
			if err := x.SendChunk(); err != nil {
				log.Println(err)
				break
			}
		}
	}()

}

func (m *Message) respCreateSteam(nmb amf.Value) {
	t, ok := nmb.(float64)
	if !ok {
		t = 0
	}
	repByte := amf.Encode([]amf.Value{"_result", int(t), nil, int(t)})
	c := Chunk{
		MessageTypeID: 20,
		SteamID:       3,
		Payload:       repByte,
		Conn:          m.Conn,
	}
	c.SendChunk()
}

func (m *Message) respConnect(amfObj amf.Value) {
	m.setChunkSize()
	//set conn app
	app, ok := amfObj.(map[string]amf.Value)
	if !ok {
		m.Conn.Close()
	}
	m.Conn.App = app["app"].(string)
	//resp connect
	var arrSour []amf.Value
	repVer := make(map[string]amf.Value)
	repVer["fmsVer"] = "FMS/3,0,1,123"
	repVer["capabilities"] = 31
	repStatus := make(map[string]amf.Value)
	repStatus["level"] = "status"
	repStatus["code"] = "NetConnection.Connect.Success"
	repStatus["description"] = "Connection succeeded"
	repStatus["objectEncoding"] = 0
	// _error or _result
	repSour := append(arrSour, "_result", 1, repVer, repStatus)
	c := Chunk{
		MessageTypeID: 20,
		SteamID:       m.Conn.SteamID,
		Payload:       amf.Encode(repSour),
		Conn:          m.Conn,
	}
	c.SendChunk()
}

func (m *Message) createSteam() {
	item := amf.Decode(m.Chunk.Payload)
	switch item[0] {
	case "connect":
		m.respConnect(item[2])
	case "createStream":
		m.respCreateSteam(item[1])
	case "publish":
		m.respPublish(item[3])
	case "play":
		m.respPlay(item[3]) // onStatus-play-reset
	default:
		log.Println("Rtmp Message not resp:->", item[0])
	}
}

func (m *Message) sendAvPack() {
	//
	play := &m.Server.WorkPool[m.Conn.App+"/"+m.Conn.Stream].Player
	for _, v := range *play {
		v <- *m.Chunk
	}
}

//分发消息。
func (m *Message) assort() {
	switch m.Chunk.MessageTypeID {
	case 20, 17:
		m.createSteam()
	case 18, 15:
		m.Server.WorkPool[m.Conn.App+"/"+m.Conn.Stream].Metadata = *m.Chunk
	case 9, 8:
		m.sendAvPack()
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
