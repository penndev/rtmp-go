package rtmp

import (
	"bufio"
	"io"

	"github.com/pennfly/rtmp-go/util"
)

type ChunkBasicHeader struct {
	Format       byte
	ChunkSteamId uint32
}

type ChunkHeader struct {
	Timestamp       uint32 // 3 byte
	MessageLength   uint32 // 3 byte
	MessageTypeID   byte   // 1 byte
	MessageStreamID uint32 // 4 byte
	ExtendTimestamp uint32 // 4 byte
}

type ChunkMessage struct {
	basic   ChunkBasicHeader
	header  ChunkHeader
	message []byte
}

func readChunkBasicHeader(c *bufio.Reader) (ChunkBasicHeader, error) {
	var cbh ChunkBasicHeader
	basicHeader, err := c.ReadByte()
	checkErr(err)
	cbh.Format = basicHeader >> 6
	cbh.ChunkSteamId = uint32(basicHeader & 0x3f)
	if cbh.ChunkSteamId == 0 {
		csid, err := c.ReadByte()
		checkErr(err)
		cbh.ChunkSteamId = 64 + uint32(csid)
	} else if cbh.ChunkSteamId == 1 {
		csid := make([]byte, 2)
		_, err = io.ReadFull(c, csid)
		checkErr(err)
		cbh.ChunkSteamId = 64 + uint32(csid[0]) + 256*uint32(csid[1])
	}
	return cbh, nil
}

func readChunkHeader(c *bufio.Reader, b ChunkBasicHeader) (ChunkHeader, error) {
	var chunk ChunkHeader
	switch b.Format {
	case 0:
		buf := make([]byte, 3)
		_, err := io.ReadFull(c, buf)
		checkErr(err)
		chunk.Timestamp = util.BigEndian.Uint24(buf)

		_, err = io.ReadFull(c, buf)
		checkErr(err)
		chunk.MessageLength = util.BigEndian.Uint24(buf)

		// Message Type ID 1 bytes
		tp, err := c.ReadByte() // 读取Message Type ID
		checkErr(err)
		chunk.MessageTypeID = tp

		// Message Stream ID 4bytes
		buff := make([]byte, 4)
		_, err = io.ReadFull(c, buff)
		checkErr(err)
		chunk.MessageStreamID = util.LittleEndian.Uint32(buff)

		// ExtendTimestamp 4 bytes
		if chunk.Timestamp == 0xffffff {
			_, err = io.ReadFull(c, buff)
			chunk.ExtendTimestamp = util.BigEndian.Uint32(buff)
		}
	case 1:
		buf := make([]byte, 3)
		_, err := io.ReadFull(c, buf)
		checkErr(err)
		chunk.Timestamp = util.BigEndian.Uint24(buf)

		_, err = io.ReadFull(c, buf)
		checkErr(err)
		chunk.MessageLength = util.BigEndian.Uint24(buf)

		// Message Type ID 1 bytes
		tp, err := c.ReadByte() // 读取Message Type ID
		checkErr(err)
		chunk.MessageTypeID = tp

		// ExtendTimestamp 4 bytes
		if chunk.Timestamp == 0xffffff {
			buff := make([]byte, 4)
			_, err = io.ReadFull(c, buff)
			chunk.ExtendTimestamp = util.BigEndian.Uint32(buff)
		}
	case 2:
		buf := make([]byte, 3)
		_, err := io.ReadFull(c, buf)
		checkErr(err)
		chunk.Timestamp = util.BigEndian.Uint24(buf)

		// ExtendTimestamp 4 bytes
		if chunk.Timestamp == 0xffffff {
			buff := make([]byte, 4)
			_, err = io.ReadFull(c, buff)
			chunk.ExtendTimestamp = util.BigEndian.Uint32(buff)
		}
	case 3:
		// 什么都不用做？
	}
	return chunk, nil
}

//TmpRead 缓存临时消息
var TmpRead = make(map[uint32][]byte)

func ReadChunkMessage(conn *Conn) (ChunkMessage, error) {
	var msg ChunkMessage
	msg.basic, _ = readChunkBasicHeader(conn.br)
	msg.header, _ = readChunkHeader(conn.br, msg.basic)

	if TmpRead[msg.basic.ChunkSteamId] == nil {
		TmpRead[msg.basic.ChunkSteamId] = make([]byte, 0)
	}
	readed := uint32(len(TmpRead[msg.basic.ChunkSteamId]))
	needRead := uint32(RTMP_DEFAULT_CHUNK_SIZE)
	unRead := msg.header.MessageLength - readed
	if unRead < needRead {
		needRead = unRead
	}

	buf := make([]byte, needRead)
	_, err := io.ReadFull(conn.br, buf)
	checkErr(err)

	TmpRead[msg.basic.ChunkSteamId] = append(TmpRead[msg.basic.ChunkSteamId], buf...)
	if uint32(len(TmpRead[msg.basic.ChunkSteamId])) == msg.header.MessageLength {
		msg.message = TmpRead[msg.basic.ChunkSteamId]
		delete(TmpRead, msg.basic.ChunkSteamId)
		return msg, nil
	}
	return ReadChunkMessage(conn)
}
