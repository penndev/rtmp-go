package rtmp

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"log"
)

type MessageHeader struct {
	Timestamp       uint32 // 3 byte
	MessageLength   uint32 // 3 byte
	MessageType     MessageType // 1 byte
	MessageStreamID uint32 // 4 byte
	ExtendTimestamp uint32 // 4 byte
}

type Message struct {
	MessageHeader
	PayLoad []byte
}

type Chunk struct {
	r *bufio.Reader
	w *bufio.Writer

	readStreamList  map[int]*Message
	writeStreamList map[int]*MessageHeader

	readChunkSize  uint32
	writeChunkSize uint32
}

// 从net.conn 读取数据，阻塞型函数
func (chk *Chunk) read(l int) ([]byte, error) {
	buf := make([]byte, l)
	_, err := io.ReadFull(chk.r, buf)
	return buf, err
}

// 写入到net.conn数据，并不一定会发送
func (chk *Chunk) write(buf []byte) error {
	_, err := chk.w.Write(buf)
	return err
}

// 5.3.1.2. Chunk Message Header
func (chk *Chunk) readMsgHeader(fmt int, csid int) error {
	if _, ok := chk.readStreamList[csid]; !ok {
		chk.readStreamList[csid] = &Message{}
	}
	//fmt type=[0 1 2]  have Timestamp
	if fmt < 3 {
		buf, err := chk.read(3)
		if err != nil {
			return err
		}
		chk.readStreamList[csid].Timestamp = uint32(buf[0])<<16 | uint32(buf[1])<<8 | uint32(buf[2])
	}
	//fmt type [0 1] MessageLength MessageType
	if fmt < 2 {
		buf, err := chk.read(3)
		if err != nil {
			return err
		}
		chk.readStreamList[csid].MessageLength = uint32(buf[0])<<16 | uint32(buf[1])<<8 | uint32(buf[2])

		buf, err = chk.read(1)
		if err != nil {
			return err
		}
		chk.readStreamList[csid].MessageType = MessageType(buf[0])
	}
	//fmt type 0 MessageStreamID
	if fmt < 1 {
		buf, err := chk.read(4)
		if err != nil {
			return err
		}
		chk.readStreamList[csid].MessageStreamID = binary.LittleEndian.Uint32(buf)
	}
	// 5.3.1.3. Extended Timestamp
	if chk.readStreamList[csid].Timestamp == 0xFFFFFF {
		buf, err := chk.read(4)
		if err != nil {
			return err
		}
		chk.readStreamList[csid].ExtendTimestamp = binary.BigEndian.Uint32(buf)
	}
	return nil
}

// 5.3.1.1. Chunk Basic Header
func (chk *Chunk) readBasicHeader() (int, int, error) {
	bs, err := chk.read(1)
	if err != nil {
		return 0, 0, err
	}
	fmt := int(bs[0] >> 6)
	csid := int(bs[0] & 0x3f)
	switch csid {
	case 0:
		// Chunk stream IDs 64-319 can be encoded in the 2-byte form of the
		// header. ID is computed as (the second byte + 64).
		csidBytes, err := chk.read(1)
		if err != nil {
			return 0, 0, err
		}
		csid = 64 + int(csidBytes[0])
	case 1:
		// 	Chunk stream IDs 64-65599 can be encoded in the 3-byte version of
		//  this field. ID is computed as ((the third byte)*256 + (the second
		//  byte) + 64).
		csidBytes, err := chk.read(2)
		if err != nil {
			return 0, 0, err
		}
		csid = 64 + int(csidBytes[0]) + (int(csidBytes[1]) << 8)
	default:
		// csid keep
	}
	return fmt, csid, nil
}

// 读取一条 Message
// 一条消息可能是多条 chunk，中间可能插其他 csid 的 chunk（如协议控制）
// 按 csid 分别攒，哪个凑满就返回哪个
func (chk *Chunk) readMessage() (Message, error) {
	for {
		fmt, csid, err := chk.readBasicHeader()
		if err != nil {
			return Message{}, err
		}
		if fmt > 3 {
			return Message{}, errors.New("fmt > 3")
		}
		if err := chk.readMsgHeader(fmt, csid); err != nil {
			return Message{}, err
		}
		readed := uint32(len(chk.readStreamList[csid].PayLoad))
		remaining := chk.readStreamList[csid].MessageLength - readed
		if remaining > chk.readChunkSize {
			remaining = chk.readChunkSize
		}
		load, err := chk.read(int(remaining))
		if err != nil {
			return Message{}, err
		}
		chk.readStreamList[csid].PayLoad = append(chk.readStreamList[csid].PayLoad, load...)

		if uint32(len(chk.readStreamList[csid].PayLoad)) >= chk.readStreamList[csid].MessageLength {
			msg := *chk.readStreamList[csid]
			chk.readStreamList[csid].PayLoad = nil
			return msg, nil
		}
	}
}

func (chk *Chunk) Read() (*Message, error) {
	for {
		msg, err := chk.readMessage()
		if err != nil {
			return nil, err
		}

		// 5.4. Protocol Control Messages
		// RTMP Chunk Stream uses message type IDs 1, 2, 3, 5, and 6 for
		// protocol control messages. These messages contain information needed
		// by the RTMP Chunk Stream protocol.
		// These protocol control messages MUST have message stream ID 0 (known
		// as the control stream) and be sent in chunk stream ID 2. Protocol
		// control messages take effect as soon as they are received; their
		// timestamps are ignored.
		switch msg.MessageType {
		case SetChunkSize:
			// 5.4.1. Set Chunk Size (1)
			chk.readChunkSize = binary.BigEndian.Uint32(msg.PayLoad)
			log.Println("set readChunkSize", chk.readChunkSize)
		default:
			return &msg, nil
		}
	}
}
