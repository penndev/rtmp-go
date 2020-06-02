package rtmp

import (
	"encoding/binary"
	"errors"
	"log"
)

// Chunk 处理rtmp中的流数据
type Chunk struct {
	Format          byte   //2 bit
	SteamID         uint32 //6 bit
	Timestamp       uint32 // 3 byte
	MessageLength   uint32 // 3 byte
	MessageTypeID   byte   // 1 byte
	MessageStreamID uint32 // 4 byte
	ExtendTimestamp uint32 // 4 byte
	Payload         []byte

	Conn *Conn
}

func (chk *Chunk) SendChunk() error {
	cache := make([]byte, 4)
	write := []byte{}

	if chk.SteamID < 64 {
		write = append(write, byte(chk.SteamID))
	}

	if chk.Timestamp == 0 {
		write = append(write, 0, 0, 0)
	}

	msgLen := uint32(len(chk.Payload))
	binary.BigEndian.PutUint32(cache, msgLen)
	write = append(write, cache[1:]...)

	write = append(write, chk.MessageTypeID)

	binary.BigEndian.PutUint32(cache, chk.MessageStreamID)
	write = append(write, cache...)

	write = append(write, chk.Payload...)
	//log.Println(write)
	_, err := chk.Conn.Write(write)
	return err
}

// 读取基础头。
func (chk *Chunk) readBasicHeader() error {
	basicHeader, err := chk.Conn.ReadByte()
	chk.Format = basicHeader >> 6
	chk.SteamID = uint32(basicHeader & 0x3f)

	if chk.SteamID == 0 {
		csid, err := chk.Conn.ReadByte()
		if err != nil {
			return err
		}
		chk.SteamID = 64 + uint32(csid)
	} else if chk.SteamID == 1 {
		csid, err := chk.Conn.ReadFull(2)
		if err != nil {
			return err
		}
		chk.SteamID = 64 + uint32(csid[0]) + 256*uint32(csid[1])
	}

	return err
}

// 读取消息头。
func (chk *Chunk) readMessageHeader() error {

	if chk.Format > 3 && chk.Format < 0 {
		return errors.New("rtmp chk steam id > 3")
	}

	// Type 0 - 1 - 2 had Timestamp
	if chk.Format <= 2 {
		timestamp, err := chk.Conn.ReadFull(3)
		if err != nil {
			return err
		}
		adr := []byte{0}
		timestamp = append(adr, timestamp...)
		chk.Timestamp = binary.BigEndian.Uint32(timestamp)
		// If typo = 2,3, had same last value
		// chk.MessageLength = chk.MessageLength
		// chk.MessageTypeID = chk.MessageTypeID
	}

	// type 0 - 1 had MessageLength and MessageType
	if chk.Format <= 1 {

		messagelength, err := chk.Conn.ReadFull(3)
		messagelength = append([]byte{0}, messagelength...)
		chk.MessageLength = binary.BigEndian.Uint32(messagelength)

		chk.MessageTypeID, err = chk.Conn.ReadByte()

		if err != nil {
			return err
		}
	}

	if chk.Format == 0 {
		steamid, err := chk.Conn.ReadFull(4)
		chk.MessageStreamID = binary.LittleEndian.Uint32(steamid)
		if err != nil {
			return err
		}
	}

	//判断时间拓展字段是否存在-
	if chk.Timestamp > 0xFFFFFF {
		extendTimestamp, err := chk.Conn.ReadFull(4)
		if err != nil {
			return err
		}
		chk.ExtendTimestamp = binary.BigEndian.Uint32(extendTimestamp)
	}

	return nil
}

// 读取源消息。
func (chk *Chunk) originMessage() error {
	var readed int
	chk.Payload = []byte{}
	for {
		if err := chk.readBasicHeader(); err != nil {
			return err
		}

		if err := chk.readMessageHeader(); err != nil {
			return err
		}
		//是否可以结束读取。。
		if length := int(chk.MessageLength) - readed; length <= chk.Conn.ReadChunkSize {

			Payload, err := chk.Conn.ReadFull(length)
			if err != nil {
				return err
			}
			chk.Payload = append(chk.Payload, Payload...)
			return nil
		}
		Payload, err := chk.Conn.ReadFull(chk.Conn.ReadChunkSize)
		if err != nil {
			return err
		}
		chk.Payload = append(chk.Payload, Payload...)

		readed += chk.Conn.ReadChunkSize //标记已经读取的位数
	}
}

//ReadMsg 读取消息。
func (chk *Chunk) ReadMsg() error {
	//ReadMsg 读取一个需要处理的消息。循环处理，如果是协议控制消息则继续读取。
	err := chk.originMessage()
	if err != nil {
		return err
	}

	// 处理与提取Msg无关的控制协议。
	switch int(chk.MessageTypeID) {
	case 1:
		chk.Conn.ReadChunkSize = int(binary.BigEndian.Uint32(chk.Payload))
		log.Println("Rtmp SetChunkSize", chk.Conn.ReadChunkSize)
		err = chk.ReadMsg()
	case 2:
		log.Println("Rtmp AbortMessage")
		err = chk.ReadMsg()
	case 3:
		log.Println("Rtmp Acknowledgement")
		err = chk.ReadMsg()
	case 4:
		log.Println("Rtmp SetBufferLength")
		err = chk.ReadMsg()
	case 5:
		log.Println("Rtmp WindowAcknowledgementSize")
		err = chk.ReadMsg()
	case 6:
		log.Println("Rtmp SetPeerBandwidth")
		err = chk.ReadMsg()
	}

	return err
}

func newChunk(conn *Conn) *Chunk {
	chk := &Chunk{
		Conn: conn,
	}
	return chk
}
