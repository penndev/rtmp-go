package rtmp

import (
	"encoding/binary"
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

// 制作基础头。
// Chunk Basic  Header field may be 1, 2, or 3 bytes, depending on the chunk stream ID.
// The protocol supports up to 65597 streams with IDs 3-65599.
//  2-63  leng = 1
// Value 0 indicates the 2 byte ；  64-319 (the second byte + 64)
// Value 1 indicates  the 3 byte；  64-65599 ((the third byte)*256 + the second byte + 64)
func (chk *Chunk) genBasicHeader() []byte {
	BasicHead := make([]byte, 1)
	Format := int(chk.Format) << 6
	FirstSteam := int(chk.SteamID)

	if FirstSteam > 1 && FirstSteam < 64 {
		BasicHead[0] = byte(Format + FirstSteam)
		return BasicHead
	}
	// value 0
	if FirstSteam > 63 && FirstSteam < 320 {
		BasicHead[0] = byte(Format)
		return append(BasicHead, byte(FirstSteam-64))
	}
	// valuee 1
	if FirstSteam > 63 && FirstSteam < 65510 {
		BasicHead[0] = byte(Format + 1)
		Second := FirstSteam - 64
		if Second > 255 {
			BasicHead = append(BasicHead, byte(Second%256))
			BasicHead = append(BasicHead, byte(Second/256))
		} else {
			BasicHead = append(BasicHead, byte(Second), 0)
		}
	}

	return BasicHead
}

// 读取消息头。
func (chk *Chunk) readMessageHeader() error {
	// type 3
	if chk.Format > 2 {
		if chunk, ok := chk.Conn.ChunkLists[chk.SteamID]; ok {
			//防止隔流串联type = 3
			chk.MessageLength = chunk.MessageLength
			chk.MessageTypeID = chunk.MessageTypeID
			chk.MessageStreamID = chunk.MessageStreamID

		}
		return nil
	}
	// Type 0 - 1 - 2 had Timestamp
	if chk.Format < 3 {
		if timestamp, err := chk.Conn.ReadFull(3); err == nil {
			timestamp = append([]byte{0}, timestamp...)
			chk.Timestamp = binary.BigEndian.Uint32(timestamp)
		} else {
			return err
		}
		if chunk, ok := chk.Conn.ChunkLists[chk.SteamID]; ok {
			chk.MessageLength = chunk.MessageLength
			chk.MessageTypeID = chunk.MessageTypeID
			chk.MessageStreamID = chunk.MessageStreamID
		}
	}
	// type 0 - 1 had MessageLength and MessageType
	if chk.Format < 2 {
		messagelength, err := chk.Conn.ReadFull(3)
		messagelength = append([]byte{0}, messagelength...)
		chk.MessageLength = binary.BigEndian.Uint32(messagelength)
		chk.MessageTypeID, err = chk.Conn.ReadByte()
		if err != nil {
			return err
		}
	}
	// type 0
	if chk.Format < 1 {
		if steamid, err := chk.Conn.ReadFull(4); err == nil {
			chk.MessageStreamID = binary.LittleEndian.Uint32(steamid)
		} else {
			return err
		}
	}
	//判断时间拓展字段是否存在-
	if chk.Timestamp == 0xFFFFFF {
		extendTimestamp, err := chk.Conn.ReadFull(4)
		if err != nil {
			return err
		}
		chk.ExtendTimestamp = binary.BigEndian.Uint32(extendTimestamp)
	}
	chk.Payload = []byte{}
	chk.Conn.ChunkLists[chk.SteamID] = *chk
	return nil
}

// 读取源消息。
func (chk *Chunk) originMessage() error {
	var readed int
	for {
		//读基础头
		if err := chk.readBasicHeader(); err != nil {
			return err
		}

		//读消息头
		if err := chk.readMessageHeader(); err != nil {
			return err
		}

		//判断本次读取的消息长度。
		length := int(chk.MessageLength) - readed
		if length > chk.Conn.ReadChunkSize {
			length = chk.Conn.ReadChunkSize
		}

		//读取消息。
		if Payload, err := chk.Conn.ReadFull(length); err == nil {
			chk.Payload = append(chk.Payload, Payload...)
		} else {
			return err
		}

		//读取结束，退出循环。
		readed += length
		if uint32(readed) >= chk.MessageLength {
			break
		}
	}
	return nil
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

// 制作消息头
func (chk *Chunk) genMessageHeader() []byte {
	if chk.Format > 2 || chk.Format < 0 {
		return nil
	}
	var headArr []byte
	// Type 0 - 1 - 2 had Timestamp
	if chk.Format < 3 {
		readTime := make([]byte, 4)
		binary.BigEndian.PutUint32(readTime, chk.Timestamp)
		headArr = readTime[1:]
	}
	// type 0 - 1 had MessageLength and MessageType
	if chk.Format < 2 {
		readLen := make([]byte, 4)
		binary.BigEndian.PutUint32(readLen, chk.MessageLength)
		headArr = append(headArr, readLen[1:]...)
		headArr = append(headArr, chk.MessageTypeID)
	}
	// type 0 had steam id
	if chk.Format < 1 {
		readSteamid := make([]byte, 4)
		binary.BigEndian.PutUint32(readSteamid, chk.MessageStreamID)
		headArr = append(headArr, readSteamid...)
	}

	if chk.Timestamp == 0xFFFFFF {
		readExtedtime := make([]byte, 4)
		binary.BigEndian.PutUint32(readExtedtime, chk.ExtendTimestamp)
		headArr = append(headArr, readExtedtime...)
	}
	return headArr
}

//SendChunk 回复消息。
func (chk *Chunk) SendChunk() error {
	if chk.MessageLength == 0 {
		chk.MessageLength = uint32(len(chk.Payload))
	}

	if _, ok := chk.Conn.SendChunkLists[chk.SteamID]; ok {
		chk.Format = 1
	} else {
		chk.Format = 0
		chk.Conn.SendChunkLists[chk.SteamID] = *chk
	}

	var mArr []byte
	mArr = chk.genBasicHeader()
	mArr = append(mArr, chk.genMessageHeader()...)
	writeSize := 0
	payloadSize := int(chk.MessageLength)
	for {
		slicSize := writeSize + chk.Conn.WriteChunkSize
		if slicSize > payloadSize {
			slicSize = payloadSize //最后一次剩余的切片。
		}
		payload := chk.Payload[writeSize:slicSize]
		// check fill message header
		if writeSize == 0 {
			mArr = append(mArr, payload...)
		} else {
			chk.Format = 3
			mArr = chk.genBasicHeader()
			mArr = append(mArr, payload...)
		}
		//
		_, err := chk.Conn.Write(mArr)
		if err != nil {
			log.Println("message error->", err)
			return err
		}
		//发送完成
		mArr = []byte{}
		if slicSize == payloadSize {
			break
		}
		writeSize = slicSize
	}
	return nil
}

func newChunk(conn *Conn) *Chunk {
	chk := &Chunk{
		Conn: conn,
	}
	return chk
}
