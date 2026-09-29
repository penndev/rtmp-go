package rtmp

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
)

type MessageHeader struct {
	Timestamp       uint32      // 3 byte
	MessageLength   uint32      // 3 byte
	MessageType     MessageType // 1 byte
	MessageStreamID uint32      // 4 byte
	ExtendTimestamp uint32      // 4 byte
}

type Message struct {
	MessageHeader
	PayLoad []byte
}

type Chunk struct {
	r *bufio.Reader
	w *bufio.Writer

	readStreamList  map[int]*Message
	writeStreamList map[int]MessageHeader

	readChunkSize  uint32
	writeChunkSize uint32

	// 5.4.3 / 5.4.4 acknowledgement accounting (uint32 wraps at 2^32)
	bytesReceived uint32 // bytes read so far
	bytesSent     uint32 // bytes written so far
	ackWindowSize uint32 // peer Window Acknowledgement Size; 0 = do not ack
	lastAckBytes  uint32 // bytesReceived at last Acknowledgement we sent
}

// 从net.conn 读取数据，阻塞型函数
func (chk *Chunk) read(l int) ([]byte, error) {
	buf := make([]byte, l)
	_, err := io.ReadFull(chk.r, buf)
	if err != nil {
		return nil, err
	}
	chk.bytesReceived += uint32(l)
	// 5.4.4: The receiving peer MUST send an Acknowledgement after
	// receiving the indicated number of bytes since the last
	// Acknowledgement was sent, or beginning of the session if no
	// Acknowledgement has yet been sent.
	if chk.ackWindowSize > 0 && chk.bytesReceived-chk.lastAckBytes >= chk.ackWindowSize {
		payload := make([]byte, 4)
		binary.BigEndian.PutUint32(payload, chk.bytesReceived)
		if err := chk.sendMsg(Acknowledgement, 2, 0, payload); err != nil {
			return nil, err
		}
		chk.lastAckBytes = chk.bytesReceived
	}
	return buf, nil
}

// 写入到net.conn数据，并不一定会发送
func (chk *Chunk) write(buf []byte) error {
	n, err := chk.w.Write(buf)
	chk.bytesSent += uint32(n)
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

// rtmp message filter control messages
func (chk *Chunk) Read() (*Message, error) {
	for {
		msg, err := chk.readMessage()
		if err != nil {
			return nil, err
		}

		switch msg.MessageType {
		case SetChunkSize:
			// 5.4.1. Set Chunk Size (1)
			// chunk size (31 bits): Valid sizes are 1 to 2147483647 (0x7FFFFFFF).
			// The first bit MUST be zero.
			if len(msg.PayLoad) < 4 {
				return nil, errors.New("SetChunkSize: payload too short")
			}
			size := binary.BigEndian.Uint32(msg.PayLoad) & 0x7FFFFFFF
			if size < 1 {
				return nil, errors.New("SetChunkSize: size must be at least 1")
			}
			chk.readChunkSize = size
		case AbortMessage:
			// 5.4.2. Abort Message (2)
			// chunk stream ID (32 bits): discard the partially received message.
			if len(msg.PayLoad) < 4 {
				return nil, errors.New("AbortMessage: payload too short")
			}
			csid := int(binary.BigEndian.Uint32(msg.PayLoad))
			if m, ok := chk.readStreamList[csid]; ok {
				m.PayLoad = nil
			}
		case Acknowledgement:
			// 5.4.3. Acknowledgement (3)
			// sequence number (32 bits): This field holds the number of
			// bytes received so far (by the peer) — compare with bytesSent.
			if len(msg.PayLoad) < 4 {
				return nil, errors.New("Acknowledgement: payload too short")
			}
			seq := binary.BigEndian.Uint32(msg.PayLoad)
			if seq > chk.bytesSent {
				return nil, errors.New("Acknowledgement: sequence exceeds bytes sent")
			}
		case WindowAcknowledgementSize:
			// 5.4.4. Window Acknowledgement Size (5)
			// The receiving peer MUST send an Acknowledgement after
			// receiving the indicated number of bytes since the last
			// Acknowledgement was sent.
			if len(msg.PayLoad) < 4 {
				return nil, errors.New("WindowAcknowledgementSize: payload too short")
			}
			chk.ackWindowSize = binary.BigEndian.Uint32(msg.PayLoad)
		default:
			return &msg, nil
		}
	}
}

// writeBasicHeader 5.3.1.1
func writeBasicHeader(fmt byte, csid int) []byte {
	if csid < 64 {
		return []byte{byte(csid) | (fmt << 6)}
	}
	if csid < 320 {
		return []byte{fmt << 6, byte(csid - 64)}
	}
	n := csid - 64
	return []byte{1 | (fmt << 6), byte(n), byte(n >> 8)}
}

// writeMsgHeader 5.3.1.2
func writeMsgHeader(fmt byte, h MessageHeader) []byte {
	var buf []byte
	if fmt < 3 {
		ts := h.Timestamp
		if ts >= 0xFFFFFF {
			ts = 0xFFFFFF
		}
		buf = append(buf, byte(ts>>16), byte(ts>>8), byte(ts))
	}
	if fmt < 2 {
		ml := h.MessageLength
		buf = append(buf, byte(ml>>16), byte(ml>>8), byte(ml), byte(h.MessageType))
	}
	if fmt < 1 {
		sid := make([]byte, 4)
		binary.LittleEndian.PutUint32(sid, h.MessageStreamID)
		buf = append(buf, sid...)
	}
	if h.Timestamp >= 0xFFFFFF {
		ext := make([]byte, 4)
		binary.BigEndian.PutUint32(ext, h.ExtendTimestamp)
		buf = append(buf, ext...)
	}
	return buf
}

// sendMsg 拼完整后一次写出（csid 须 >= 2）
func (chk *Chunk) sendMsg(msgType MessageType, csid int, streamID uint32, payload []byte) error {
	if csid < 2 {
		return errors.New("sendMsg: csid must be >= 2")
	}
	writeFmt := byte(1)
	if _, ok := chk.writeStreamList[csid]; !ok {
		writeFmt = 0
	}
	h := MessageHeader{
		Timestamp:       chk.writeStreamList[csid].Timestamp,
		ExtendTimestamp: chk.writeStreamList[csid].ExtendTimestamp,
		MessageType:     msgType,
		MessageLength:   uint32(len(payload)),
		MessageStreamID: streamID,
	}
	chk.writeStreamList[csid] = h

	out := make([]byte, 0, 16+len(payload))
	writed := 0
	for writed < len(payload) {
		n := int(chk.writeChunkSize)
		if remain := len(payload) - writed; remain < n {
			n = remain
		}
		if writed == 0 {
			out = append(out, writeBasicHeader(writeFmt, csid)...)
			out = append(out, writeMsgHeader(writeFmt, h)...)
		} else {
			out = append(out, writeBasicHeader(3, csid)...)
		}
		out = append(out, payload[writed:writed+n]...)
		writed += n
	}
	if err := chk.write(out); err != nil {
		return err
	}
	return chk.w.Flush()
}
