// 当前代码都是根据rtmp协议文档进行实现
// https://www.adobe.com/content/dam/acom/en/devnet/rtmp/pdf/rtmp_specification_1.0.pdf

package rtmp

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
)

// 默认ChunkSize
const DefaultChunkSize = 128

// 需要设置的chunksize写入的ChunkSize
// 最小为128
const SetChunkSize = 4096

// csid Value 2 保留用作控制协议。
const ChunkControlID = 2

// csid value 3 默认都用作message消息流
const ChunkMessageID = 3

// caid value 4 默认用作新的媒体传输流使用
const ChunkAVPackID = 4

const DefaultStreamID = 5

// rtmp message 消息头信息
// 因为需要存储多路复用所以不包含消息体 payload.
// 阅读官方详细说明  5.3.1.2.5. Common Header Fields  15 page
type ChunkMessageHeader struct {
	Timestamp       uint32 // 3 byte
	MessageLength   uint32 // 3 byte
	MessageTypeID   byte   // 1 byte
	MessageStreamID uint32 // 4 byte
	ExtendTimestamp uint32 // 4 byte
}

// Chunk 处理rtmp中底层的流数据
// rtmp 协议文档中每个rtmp被分割为多个chunk
// 阅读官方详细说明   5.3. Chunking   11 page
type Chunk struct {
	fmt  byte   //2 bit
	csid uint32 //6 bit

	// 多路复用读写的 csid 用
	// 不同的 ChunkStreamID 会有不同的 MsgHeader
	readChunkList  map[uint32]*ChunkMessageHeader
	writeChunkList map[uint32]*ChunkMessageHeader

	// 读写IO
	r *bufio.Reader
	w *bufio.Writer

	//rtmp chunk 协议读取 chunk 限制大小
	readChunkSize  uint32
	writeChunkSize uint32
}

//从net.conn 读取数据，阻塞型函数
func (chk *Chunk) Read(l int) ([]byte, error) {
	buf := make([]byte, l)
	_, err := io.ReadFull(chk.r, buf)
	return buf, err
}

//写入到net.conn数据，并不一定会发送
func (chk *Chunk) Write(buf []byte) error {
	_, err := chk.w.Write(buf)
	return err
}

//读取Rtmp基础消息头
//并进行分析Chunk[fmt,csid]
func (chk *Chunk) readBasicHeader() error {
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
func (chk *Chunk) writeBasicHeader(basicFmt byte, csid uint32) []byte {
	load := make([]byte, 3)
	if csid < 64 {
		load[0] = byte(csid + uint32(basicFmt<<6))
		return load[:1]
	} else if csid < 320 {
		load[0] = byte(0 + (basicFmt << 6))
		load[1] = byte(csid - 64)
		return load[:2]
	} else if csid < 65510 {
		load[0] = byte(1 + (basicFmt << 6))
		Second := csid - 64
		if Second > 255 {
			load[1] = byte(Second % 256)
			load[2] = byte(Second / 256)
		} else {
			load[1] = byte(Second)
		}
		return load
	}
	log.Println("chunk csid 不合理，请检查")
	return nil
}

// 根据 fmt 来处理获取message Header
func (chk *Chunk) readMsgHeader() error {
	var err error
	var tmp []byte
	if chk.fmt == 3 {
		return nil
	}
	if _, ok := chk.readChunkList[chk.csid]; !ok {
		chk.readChunkList[chk.csid] = &ChunkMessageHeader{}
	}
	//fmt type=[0 1 2]  have Timestamp
	if chk.fmt < 3 {
		if tmp, err = chk.Read(3); err != nil {
			return err
		}
		timestamp := make([]byte, 4)
		copy(timestamp[1:], tmp)
		chk.readChunkList[chk.csid].Timestamp = binary.BigEndian.Uint32(timestamp)
	}
	//fmt type [0 1] MessageLength MessageType
	if chk.fmt < 2 {
		messagelength := make([]byte, 4)
		if tmp, err = chk.Read(3); err != nil {
			return err
		}
		copy(messagelength[1:], tmp)
		chk.readChunkList[chk.csid].MessageLength = binary.BigEndian.Uint32(messagelength)

		if tmp, err = chk.Read(1); err != nil {
			return err
		}
		chk.readChunkList[chk.csid].MessageTypeID = tmp[0]
	}
	//fmt type 0 MessageStreamID
	if chk.fmt < 1 {
		if tmp, err = chk.Read(4); err != nil {
			return err
		}
		chk.readChunkList[chk.csid].MessageStreamID = binary.LittleEndian.Uint32(tmp)
	}
	//判断时间拓展字段是否存在
	if chk.readChunkList[chk.csid].Timestamp == 0xFFFFFF {
		if tmp, err = chk.Read(4); err != nil {
			return err
		}
		chk.readChunkList[chk.csid].ExtendTimestamp = binary.BigEndian.Uint32(tmp)
	}
	return nil
}

// 制作消息头
func (chk *Chunk) rspMsgHeader(basicFmt byte, csid uint32) []byte {
	if basicFmt > 2 {
		return nil
	}
	headByte := make([]byte, 15)
	headLen := 0
	// Type 0 - 1 - 2 had Timestamp
	if basicFmt < 3 {
		readTime := make([]byte, 4)
		binary.BigEndian.PutUint32(readTime, chk.writeChunkList[csid].Timestamp)
		copy(headByte[:3], readTime[1:])
		headLen = 3
	}
	// type 0 - 1 had MessageLength and MessageType
	if basicFmt < 2 {
		readLen := make([]byte, 4)
		binary.BigEndian.PutUint32(readLen, chk.writeChunkList[csid].MessageLength)
		copy(headByte[3:6], readLen[1:])
		headByte[6] = chk.writeChunkList[csid].MessageTypeID
		headLen = 7
	}
	// type 0 had steam id
	if basicFmt < 1 {
		readSteamid := make([]byte, 4)
		binary.BigEndian.PutUint32(readSteamid, chk.writeChunkList[csid].MessageStreamID)
		copy(headByte[7:11], readSteamid)
		headLen = 11
	}

	if chk.writeChunkList[csid].Timestamp == 0xFFFFFF {
		readExtedtime := make([]byte, 4)
		binary.BigEndian.PutUint32(readExtedtime, chk.writeChunkList[csid].ExtendTimestamp)
		copy(headByte[12:15], readExtedtime)
		headLen = 15
	}
	return headByte[:headLen]
}

// 读取一条 Message
// 一条消息可能是多条chunk消息
// 返回一条原始的 Message
func (chk *Chunk) readMsg() ([]byte, error) {
	readedLen := uint32(0)
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

// 返回消息体
// 对底层控制协议进行处理不直接返回。
// 直接返回消息体
func (chk *Chunk) handlesMsg() (Pack, error) {
	//读取原始数据
	payload, err := chk.readMsg()
	if err != nil {
		return Pack{}, err
	}
	// 协议控制消息。
	switch int(chk.readChunkList[chk.csid].MessageTypeID) {
	case 1:
		chk.readChunkSize = binary.BigEndian.Uint32(payload)
	case 2:
		log.Println("Abort Message (2)")
	case 3:
		// Acknowledgement (3) // 收到字节数对照
		log.Println("Acknowledgement (3)", payload)
	case 5:
		// Window Acknowledgement Size (5) // 发送数据对照
		log.Println("Window Acknowledgement Size (5)")
	case 6:
		// Set Peer Bandwidth (6) //限制传送速率
		log.Println("Set Peer Bandwidth (6)")
	default:
		var pk Pack
		pk.PayLoad = payload
		pk.ChunkMessageHeader = *chk.readChunkList[chk.csid]
		return pk, nil
	}
	return chk.handlesMsg()
}

func (chk *Chunk) sendMsg(MessageTypeID byte, csid uint32, Payload []byte) error {
	if csid < 2 {
		return errors.New("sendMsg err:) scsid cant < 2, 0 and 1 cant used")
	}

	writefmt := byte(1)
	if _, ok := chk.writeChunkList[csid]; !ok {
		chk.writeChunkList[csid] = &ChunkMessageHeader{}
		writefmt = 0
	}
	payloadLen := uint32(len(Payload))
	chk.writeChunkList[csid].MessageTypeID = MessageTypeID
	chk.writeChunkList[csid].MessageLength = payloadLen
	//制作基础头
	writed := uint32(0)
	for {
		//本次需要发送多少字节。
		currenLen := chk.writeChunkSize
		//剩余多少字节数据需要发送
		remaining := payloadLen - writed
		if remaining < currenLen {
			currenLen = remaining
		}
		//是否第一次发送
		if writed == 0 {
			chk.Write(chk.writeBasicHeader(writefmt, csid))
			chk.Write(chk.rspMsgHeader(writefmt, csid))
		} else {
			chk.Write(chk.writeBasicHeader(3, csid))
		}
		chk.Write(Payload[writed:(writed + currenLen)])
		chk.w.Flush()
		//数据发送完成
		writed += currenLen
		if writed >= payloadLen {
			break
		}
	}
	return nil
}

// - 底层协议控制传送速率
func (chk *Chunk) setWindowAcknowledgementSize(size uint32) {
	sizeByte := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeByte, size)
	chk.sendMsg(5, 2, sizeByte)
}

// no message is larger than 16777215 bytes.
func (chk *Chunk) setChunkSize(size uint32) error {
	if size > 16777215 {
		return errors.New("setChunkSize err:) set chunk size cant > 16777215")
	}
	sizeByte := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeByte, size)
	chk.sendMsg(1, 2, sizeByte)
	chk.writeChunkSize = size
	return nil
}

func (chk *Chunk) setStreamBegin(streamID uint32) error {
	streamContent := make([]byte, 6)
	binary.BigEndian.PutUint32(streamContent[2:], streamID)
	return chk.sendMsg(4, ChunkControlID, streamContent)
}

// 创建 Chunk Stream
func newChunk(c net.Conn) *Chunk {
	return &Chunk{
		r:              bufio.NewReader(c),
		w:              bufio.NewWriter(c),
		readChunkSize:  uint32(DefaultChunkSize),
		writeChunkSize: uint32(DefaultChunkSize),
		readChunkList:  make(map[uint32]*ChunkMessageHeader),
		writeChunkList: make(map[uint32]*ChunkMessageHeader),
	}
}
