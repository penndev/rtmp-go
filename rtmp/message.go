package rtmp

import (
	"encoding/binary"
	"errors"
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

func (m *Message) sendAvPack() {
	meta := m.Server.WorkPool[m.Conn.App][m.Conn.Stream]
	if meta.videocodecid == 7 {
		if m.Chunk.MessageTypeID == 9 && len(meta.VideoCode) == 0 {
			meta.VideoCode = m.Chunk.Payload
		} else if m.Chunk.MessageTypeID == 8 && len(meta.AudioCode) == 0 {
			meta.AudioCode = m.Chunk.Payload
		}
	}

	for _, v := range meta.Player {
		v <- *m.Chunk
	}
}

func (m *Message) createSteam() error {
	var item []amf.Value
	if m.Chunk.MessageTypeID == 17 {
		item = amf.Decode3(m.Chunk.Payload)
	} else {
		item = amf.Decode(m.Chunk.Payload)
	}

	switch item[0] {
	case "connect":
		m.respConnect(item[2])

	case "createStream":
		m.respCreateSteam(item[1])
	case "publish":
		m.respPublish(item[3])
	case "play":
		m.respPlay(item[3]) // onStatus-play-reset
	case "deleteStream":
		m.deleteStream()
	case "FCPublish":
	case "getStreamLength":
	case "releaseStream":
	default:
		log.Println("message createsteam unprocessed: ", item[0].(string))
	}
	return nil
}

//分发消息。
func (m *Message) main() error {
	switch m.Chunk.MessageTypeID {
	case 20, 17:
		return m.createSteam()
	case 18, 15:
		m.setMetadata()
	case 9, 8:
		m.sendAvPack()
	default:
		return errors.New("message typeid error: " + string(m.Chunk.MessageTypeID))
	}
	return nil
}

func newMessage(chk *Chunk) error {
	msg := Message{
		Chunk:  chk,
		Conn:   chk.Conn,
		Server: chk.Conn.Server,
	}
	err := msg.main()
	return err
}
