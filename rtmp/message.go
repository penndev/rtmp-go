package rtmp

import (
	"encoding/binary"
	"errors"
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
	play := &m.Server.WorkPool[m.Conn.App][m.Conn.Stream].Player
	for _, v := range *play {
		v <- *m.Chunk
	}
}

//分发消息。
func (m *Message) main() error {
	switch m.Chunk.MessageTypeID {
	case 20, 17:
		m.createSteam()
	case 18, 15:
		m.setMetadata()
	case 9, 8:
		m.sendAvPack()
	default:
		return errors.New("rtmp message had err typeid " + string(m.Chunk.MessageTypeID))
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
