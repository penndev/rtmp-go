package rtmp

import (
	"encoding/binary"
	"errors"
	"log"
)

// Chunk 处理rtmp中的流数据
type Chunk struct {
	c               *Connnect
	Format          byte   //2 bit
	SteamID         uint32 //6 bit
	Timestamp       uint32 // 3 byte
	MessageLength   uint32 // 3 byte
	MessageTypeID   byte   // 1 byte
	MessageStreamID uint32 // 4 byte
	ExtendTimestamp uint32 // 4 byte
	Payload         []byte
}

// 读取基础标头。
func (chunk *Chunk) readBasicHeader() error {
	basicHeader, err := chunk.c.ReadByte()
	chunk.Format = basicHeader >> 6
	chunk.SteamID = uint32(basicHeader & 0x3f)

	if chunk.SteamID == 0 {
		csid, err := chunk.c.ReadByte()
		if err != nil {
			return err
		}
		chunk.SteamID = 64 + uint32(csid)
	} else if chunk.SteamID == 1 {
		csid, err := chunk.c.ReadBuffer(2)
		if err != nil {
			return err
		}
		chunk.SteamID = 64 + uint32(csid[0]) + 256*uint32(csid[1])
	}

	return err
}

//
func (chunk *Chunk) readMessageHeader() error {
	if chunk.Format > 3 {
		errors.New("rtmp chunk steam id > 3")
	}

	// Type 0 - 1 - 2 had Timestamp
	if chunk.Format <= 2 {
		timestamp, err := chunk.c.ReadBuffer(3)
		if err != nil {
			return nil
		}
		timestamp = append(timestamp, 0)
		chunk.Timestamp = binary.BigEndian.Uint32(timestamp)
	}

	// type 0 - 1 had MessageLength and MessageType
	if chunk.Format <= 1 {

		messagelength, err := chunk.c.ReadBuffer(3)
		messagelength = append([]byte{0}, messagelength...)
		chunk.MessageLength = binary.BigEndian.Uint32(messagelength)

		chunk.MessageTypeID, err = chunk.c.ReadByte()

		if err != nil {
			return err
		}
	}

	if chunk.Format == 0 {
		steamid, err := chunk.c.ReadBuffer(4)
		chunk.MessageStreamID = binary.LittleEndian.Uint32(steamid)
		if err != nil {
			return err
		}
	}
	//判断时间戳是否溢出
	if 2 == 1 {

	}

	return nil
}

// ReadChunkMsg 读取一个完整的消息块
func ReadChunkMsg(c *Connnect) (Chunk, error) {
	var err error

	var readed uint32

	var chunk Chunk
	chunk.c = c
	for {
		chunk.readBasicHeader()
		chunk.readMessageHeader()
		length := chunk.MessageLength - readed
		if length <= chunk.c.chunkSize {
			chunk.Payload, err = chunk.c.ReadBuffer(int(length))
			break
		}
		payload := make([]byte, chunk.c.chunkSize)
		payload, err = chunk.c.ReadBuffer(int(chunk.c.chunkSize))
		chunk.Payload = append(chunk.Payload, payload...)
		readed += chunk.c.chunkSize //标记已经读取的位数

		if err != nil {
			return chunk, nil
		}
	}
	log.Println(chunk.Format, chunk.SteamID, chunk.Timestamp, chunk.MessageLength, chunk.MessageTypeID, chunk.MessageStreamID, chunk.Payload)
	return chunk, err
}
