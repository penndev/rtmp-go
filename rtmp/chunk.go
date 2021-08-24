package rtmp

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
)

//默认ChunkSize
var DefaultChunkSize = 128

//写入的ChunkSize
var SetChunkSize = 4096

// MsgHeader header信息
type MsgHeader struct {
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
	// Value 2 保留用作控制协议。
	csid uint32 //6 bit

	// ChunkStreamID
	// 不同的 ChunkStreamID 会有不同的 MsgHeader
	// 因为多路复用，可能会随时恢复之前读取的状态。
	rChkList map[uint32]*MsgHeader

	// ChunkStreamID
	// 不同的 ChunkStreamID 会有不同的 MsgHeader
	// 因为多路复用，可能会随时恢复之前读取的状态。
	wChkList map[uint32]*MsgHeader

	// 读写IO
	r *bufio.Reader
	w *bufio.Writer

	// read size 从客户端共读取了多少字节的数据（不包括握手）
	rSize uint

	//rtmp 协议读取 chunk 限制大小
	rChkSize uint32

	//rtmp 协议发送 chunk 限制大小
	wChkSize uint32
}

//从bufio读取数据并进行字节统计
func (chk *Chunk) Read(l int) ([]byte, error) {
	buf := make([]byte, l)
	l, err := io.ReadFull(chk.r, buf)
	chk.rSize += uint(l)
	return buf, err
}

func (chk *Chunk) Write(buf []byte) error {
	_, err := chk.w.Write(buf)

	// c.rwByteSize.write += l
	//l
	return err
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
		chk.csid = uint32(64 + csid[0] + csid[1]*255)
	}
	return nil
}

// 制作基础消息头
func (chk *Chunk) rspBasicHeader(basicFmt byte, csid uint32) []byte {
	load := make([]byte, 1)
	if csid < 64 {
		load[0] = byte(csid + uint32(basicFmt<<6))
	} else if csid < 320 {
		load[0] = byte(0 + (basicFmt << 6))
		load = append(load, byte(csid-64))
	} else if csid < 65510 {
		load[0] = byte(1 + (basicFmt << 6))
		Second := csid - 64
		if Second > 255 {
			load = append(load, byte(Second%256), byte(Second/256))
		} else {
			load = append(load, byte(Second), 0)
		}
	}
	return load
}

// 根据 fmt 来处理获取message Header
func (chk *Chunk) reqMsgHeader() error {
	//fmt type = 3
	if chk.fmt == 3 {
		return nil
	}

	if _, ok := chk.rChkList[chk.csid]; !ok {
		chk.rChkList[chk.csid] = &MsgHeader{}
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

// 制作消息头
func (chk *Chunk) rspMsgHeader(basicFmt byte, csid uint32) []byte {
	if basicFmt > 2 {
		return nil
	}
	var headArr []byte
	// Type 0 - 1 - 2 had Timestamp
	if basicFmt < 3 {
		readTime := make([]byte, 4)
		binary.BigEndian.PutUint32(readTime, chk.wChkList[csid].Timestamp)
		headArr = readTime[1:]
	}
	// type 0 - 1 had MessageLength and MessageType
	if basicFmt < 2 {
		readLen := make([]byte, 4)
		binary.BigEndian.PutUint32(readLen, chk.wChkList[csid].MessageLength)
		headArr = append(headArr, readLen[1:]...)
		headArr = append(headArr, chk.wChkList[csid].MessageTypeID)
	}
	// type 0 had steam id
	if basicFmt < 1 {
		readSteamid := make([]byte, 4)
		binary.BigEndian.PutUint32(readSteamid, chk.wChkList[csid].MessageStreamID)
		headArr = append(headArr, readSteamid...)
	}

	if chk.wChkList[csid].Timestamp == 0xFFFFFF {
		readExtedtime := make([]byte, 4)
		binary.BigEndian.PutUint32(readExtedtime, chk.wChkList[csid].ExtendTimestamp)
		headArr = append(headArr, readExtedtime...)
	}
	return headArr
}

// 读取一条 Message
// 一条消息可能是多条chunk消息
// 返回一条原始的 Message
func (chk *Chunk) reqMsg() ([]byte, error) {
	readedLen := uint32(0)
	payload := []byte{}
	for {
		if err := chk.reqBasicHeader(); err != nil {
			return nil, err
		}
		if err := chk.reqMsgHeader(); err != nil {
			return nil, err
		}
		//处理剩余未读字节数
		remaining := chk.rChkList[chk.csid].MessageLength - readedLen
		if remaining > chk.rChkSize {
			remaining = chk.rChkSize
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
		if readedLen >= chk.rChkList[chk.csid].MessageLength {
			break
		}
	}
	return payload, nil
}

// 返回消息体
// 对底层控制协议进行处理不直接返回。
// 直接返回消息体
func (chk *Chunk) readMsg() ([]byte, error) {
	//读取原始数据
	payload, err := chk.reqMsg()
	if err != nil {
		return nil, err
	}
	// 协议控制消息。
	switch int(chk.rChkList[chk.csid].MessageTypeID) {
	case 1:
		//Set Chunk Size (1) //设置chunk大小
		chk.rChkSize = binary.BigEndian.Uint32(payload)
	case 2:
		//Abort Message (2) //中止消息。
		log.Println("Abort Message (2)")
	case 3:
		//  Acknowledgement (3) // 收到字节数对照

		log.Println("Acknowledgement (3)", payload)
	case 5:
		// Window Acknowledgement Size (5) // 发送数据对照
		log.Println("Window Acknowledgement Size (5)")
	case 6:
		//  Set Peer Bandwidth (6) //限制传送速率
		log.Println("Set Peer Bandwidth (6)")
	default:
		return payload, nil

	}
	return chk.readMsg()
}

func (chk *Chunk) sendMsg(MessageTypeID byte, csid uint32, Payload []byte) error {
	if csid < 2 {
		return errors.New("csid cant < 2, 0 and 1 cant used")
	}

	writefmt := byte(1)
	if _, ok := chk.wChkList[csid]; !ok {
		chk.wChkList[csid] = &MsgHeader{}
		writefmt = 0
	}
	payloadLen := uint32(len(Payload))
	chk.wChkList[csid].MessageTypeID = MessageTypeID
	chk.wChkList[csid].MessageLength = payloadLen

	//制作基础头
	writed := uint32(0)
	for {
		//本次需要发送多少字节。
		currenLen := chk.wChkSize
		//剩余多少字节数据需要发送
		remaining := payloadLen - writed
		if remaining < currenLen {
			currenLen = remaining
		}

		//是否第一次发送
		if writed == 0 {
			chk.Write(chk.rspBasicHeader(writefmt, csid))
			chk.Write(chk.rspMsgHeader(writefmt, csid))
		} else {
			chk.Write(chk.rspBasicHeader(3, csid))
		}
		chk.Write(Payload[writed:(writed + currenLen)])

		//数据发送完成
		writed += currenLen
		if writed >= payloadLen {
			break
		}
	}
	return chk.w.Flush()
}

func (chk *Chunk) sendAv(MessageTypeID byte, csid uint32, timestamp uint32, Payload []byte) error {
	if csid < 2 {
		return errors.New("csid cant < 2, 0 and 1 cant used")
	}

	writefmt := byte(1)
	if _, ok := chk.wChkList[csid]; !ok {
		chk.wChkList[csid] = &MsgHeader{}
		writefmt = 0
	}
	payloadLen := uint32(len(Payload))
	chk.wChkList[csid].MessageTypeID = MessageTypeID
	chk.wChkList[csid].MessageLength = payloadLen
	chk.wChkList[csid].Timestamp = timestamp
	chk.wChkList[csid].MessageStreamID = 4

	//制作基础头
	writed := uint32(0)
	for {
		//本次需要发送多少字节。
		currenLen := chk.wChkSize
		//剩余多少字节数据需要发送
		remaining := payloadLen - writed

		if remaining < currenLen {
			currenLen = remaining
		}
		// log.Println("剩余字节->", remaining, "已写字节->", payloadLen, "本次写字节->", currenLen)
		//是否第一次发送
		if writed == 0 {
			genBH := chk.rspBasicHeader(writefmt, csid)
			// log.Println("基础头->", genBH)
			chk.Write(genBH)
			genMH := chk.rspMsgHeader(writefmt, csid)
			// log.Println("消息头->", genMH)
			chk.Write(genMH)
		} else {
			genMH := chk.rspBasicHeader(2, csid)
			// log.Println("基础->", genMH)
			chk.Write(genMH)
		}
		chk.Write(Payload[writed:(writed + currenLen)])
		//数据发送完成
		writed += currenLen
		if writed >= payloadLen {
			break
		}
	}
	return chk.w.Flush()
}

// 创建 Chunk Stream
func newChunk(c *net.Conn) *Chunk {
	return &Chunk{
		r:        bufio.NewReader(*c),
		w:        bufio.NewWriter(*c),
		rChkSize: uint32(DefaultChunkSize),
		wChkSize: uint32(DefaultChunkSize),
		rChkList: make(map[uint32]*MsgHeader),
		wChkList: make(map[uint32]*MsgHeader),
	}
}
