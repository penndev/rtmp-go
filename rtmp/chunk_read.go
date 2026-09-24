package rtmp

import (
	"encoding/binary"
	"errors"
)

// 5.3.1.2. Chunk Message Header
func (chk *Chunk) readMsgHeader() error {
	if _, ok := chk.readChunkList[chk.csid]; !ok {
		chk.readChunkList[chk.csid] = &ChunkMessageHeader{}
	}
	if chk.fmt > 3 {
		return errors.New("fmt > 3")
	}
	//fmt type=[0 1 2]  have Timestamp
	if chk.fmt < 3 {
		buf, err := chk.Read(3)
		if err != nil {
			return err
		}
		chk.readChunkList[chk.csid].Timestamp = uint32(buf[0])<<16 | uint32(buf[1])<<8 | uint32(buf[2])
	}
	//fmt type [0 1] MessageLength MessageType
	if chk.fmt < 2 {
		buf, err := chk.Read(3)
		if err != nil {
			return err
		}
		chk.readChunkList[chk.csid].MessageLength = uint32(buf[0])<<16 | uint32(buf[1])<<8 | uint32(buf[2])

		buf, err = chk.Read(1)
		if err != nil {
			return err
		}
		chk.readChunkList[chk.csid].MessageTypeID = buf[0]
	}
	//fmt type 0 MessageStreamID
	if chk.fmt < 1 {
		buf, err := chk.Read(4)
		if err != nil {
			return err
		}
		chk.readChunkList[chk.csid].MessageStreamID = binary.LittleEndian.Uint32(buf)
	}

	//	5.3.1.3. Extended Timestamp
	if chk.readChunkList[chk.csid].Timestamp == 0xFFFFFF {
		buf, err := chk.Read(4)
		if err != nil {
			return err
		}
		chk.readChunkList[chk.csid].ExtendTimestamp = binary.BigEndian.Uint32(buf)
	}
	return nil
}

// 5.3.1.1. Chunk Basic Header
func (chk *Chunk) readBasicHeader() error {
	bs, err := chk.Read(1)
	if err != nil {
		return err
	}
	chk.fmt = bs[0] >> 6
	chk.csid = uint32(bs[0] & 0x3f)
	switch chk.csid {
	case 0:
		// Chunk stream IDs 64-319 can be encoded in the 2-byte form of the
		// header. ID is computed as (the second byte + 64).
		csid, err := chk.Read(1)
		if err != nil {
			return err
		}
		chk.csid = 64 + uint32(csid[0])
	case 1:
		// 		Chunk stream IDs 64-65599 can be encoded in the 3-byte version of
		//  this field. ID is computed as ((the third byte)*256 + (the second
		//  byte) + 64).
		csid, err := chk.Read(2)
		if err != nil {
			return err
		}
		chk.csid = 64 + uint32(csid[0]) + (uint32(csid[1]) << 8)
	default:
		// chk.csid keep
	}
	return nil
}

// 读取一条 Message
// 一条消息可能是多条chunk消息
// 返回一条原始的 Message
func (chk *Chunk) readMsg() ([]byte, error) {
	var readedLen uint32 = 0
	var payload []byte
	for {
		if err := chk.readBasicHeader(); err != nil {
			return nil, err
		}
		if err := chk.readMsgHeader(); err != nil {
			return nil, err
		}

		//处理剩余未读字节数
		remaining := chk.readChunkList[chk.csid].MessageLength - readedLen
		if remaining > chk.readChunkSize {
			remaining = chk.readChunkSize
		}
		//\本次读取多少数据。
		load, err := chk.Read(int(remaining))
		if err != nil {
			return nil, err
		}
		//叠加内容体
		payload = append(payload, load...)
		readedLen += remaining
		//读取数据够数了。break =，panic >
		if readedLen >= chk.readChunkList[chk.csid].MessageLength {
			break
		}
	}
	return payload, nil
}
