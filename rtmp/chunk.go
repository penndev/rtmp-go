package rtmp

import (
	"bufio"
	"encoding/binary"
	"io"
)

type FmtHeader struct {
	Timestamp       uint32 // 3 byte
	MessageLength   uint32 // 3 byte
	MessageTypeID   byte   // 1 byte
	MessageStreamID uint32 // 4 byte
	ExtendTimestamp uint32 // 4 byte
}

// Chunk 处理rtmp中的流数据
type Chunk struct {
	//There are four different formats for the chunk message header,
	//selected by the "fmt" field in the chunk basic header.
	// Value 0
	// Value 1
	// Value 2
	// Value 3 消息段，主要用来传输chunk段
	fmt byte //2 bit
	// The protocol supports up to 65597 streams with IDs 3-65599.
	// Chunk Basic  Header field may be 1, 2, or 3 bytes, depending on the chunk stream ID.
	// Value 3-63 no read more
	// Value 0 indicates the 2 byte ；  64-319 (the second byte + 64)
	// Value 1 indicates  the 3 byte；  64-65599 ((the third byte)*256 + the second byte + 64)
	// Value 2 保留
	csid uint32 //6 bit
	//
	Payload []byte
	//
	r *bufio.Reader
	w *bufio.Writer
	// read size 从客户端共读取了多少字节的数据（不包括握手）
	rSize uint
	//rtmp 协议 chunk 限制大小
	rChkSize uint32
	// 消息头暂存信息-多路复用使用
	rChkList map[uint32]*FmtHeader
}

//从bufio读取数据并进行字节统计
func (chk *Chunk) Read(l int) ([]byte, error) {
	buf := make([]byte, l)
	l, err := io.ReadFull(chk.r, buf)
	chk.rSize += uint(l)
	return buf, err
}

//读取Rtmp基础消息头
func (chk *Chunk) reqBasicHeader() error {
	bs, err := chk.Read(1)
	if err != nil {
		return err
	}
	chk.fmt = bs[0] >> 6
	chk.csid = uint32(bs[0] & 0x3f)
	if chk.csid == 0 {
		csid, err := chk.Read(1)
		if err != nil {
			return err
		}
		chk.csid = 64 + uint32(csid[0])
	} else if chk.csid == 1 {
		csid, err := chk.Read(2)
		if err != nil {
			return err
		}
		chk.csid = uint32(64 + csid[0] + csid[1]>>8)
	}
	return nil
}

// 根据 fmt 来处理获取message Header
func (chk *Chunk) reqMsgHeader() error {
	//fmt type = 3
	if chk.fmt == 3 {
		return nil
	}

	if _, ok := chk.rChkList[chk.csid]; ok == false {
		chk.rChkList[chk.csid] = &FmtHeader{}
	}

	//fmt type=[0 1 2]  have Timestamp
	if chk.fmt < 3 {
		timeByte, err := chk.Read(3)
		if err != nil {
			return err
		}
		timestamp := make([]byte, 4)
		copy(timestamp[1:], timeByte)
		chk.rChkList[chk.csid].Timestamp = binary.BigEndian.Uint32(timestamp)
	}

	//fmt type [0 1] MessageLength MessageType
	if chk.fmt < 2 {
		msgLenType, err := chk.Read(3)
		if err != nil {
			return nil
		}
		messagelength := make([]byte, 4)
		copy(messagelength[1:], msgLenType)
		chk.rChkList[chk.csid].MessageLength = binary.BigEndian.Uint32(messagelength)

		messageTypeID, err := chk.Read(1)
		if err != nil {
			return err
		}
		chk.rChkList[chk.csid].MessageTypeID = messageTypeID[0]
	}

	//fmt type 0 MessageStreamID
	if chk.fmt < 1 {
		steamid, err := chk.Read(4)
		if err != nil {
			return err
		}
		chk.rChkList[chk.csid].MessageStreamID = binary.LittleEndian.Uint32(steamid)
	}

	//判断时间拓展字段是否存在
	if chk.rChkList[chk.csid].Timestamp == 0xFFFFFF {
		extendTimestamp, err := chk.Read(4)
		if err != nil {
			return err
		}
		chk.rChkList[chk.csid].ExtendTimestamp = binary.BigEndian.Uint32(extendTimestamp)
	}

	return nil
}

// 读取一条消息。并返回
func (chk *Chunk) reqMsg() error {
	readedLen := uint32(0)
	for {
		if err := chk.reqBasicHeader(); err != nil {
			return err
		}
		if err := chk.reqMsgHeader(); err != nil {
			return err
		}
		//处理剩余未读字节数
		remaining := chk.rChkList[chk.csid].MessageLength - readedLen
		if remaining > chk.rChkSize {
			remaining = chk.rChkSize
		}
		//\本次读取多少数据。
		Payload, err := chk.Read(int(remaining))
		if err != nil {
			return err
		}
		//叠加内容体
		chk.Payload = append(chk.Payload, Payload...)
		readedLen += remaining
		//读取数据够数了。break =，panic >
		if readedLen >= chk.rChkSize {
			break
		}
	}
	return nil
}

// case 1:
// 	chk.Conn.ReadChunkSize = int(binary.BigEndian.Uint32(chk.Payload))
// 	log.Println("Rtmp SetChunkSize", chk.Conn.ReadChunkSize)
// 	err = chk.ReadMsg()
// case 2:
// 	log.Println("Rtmp AbortMessage")
// 	err = chk.ReadMsg()
// case 3:
// 	log.Println("Rtmp Acknowledgement")
// 	err = chk.ReadMsg()
// case 4:
// 	log.Println("Rtmp SetBufferLength")
// 	err = chk.ReadMsg()
// case 5:
// 	log.Println("Rtmp WindowAcknowledgementSize")
// 	err = chk.ReadMsg()
// case 6:
// 	log.Println("Rtmp SetPeerBandwidth")
// 	err = chk.ReadMsg()
// }
