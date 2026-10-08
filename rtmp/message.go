package rtmp

import "github.com/penndev/rtmp/codec/flv"

type MessageHeader struct {
	Timestamp       uint32      // 绝对时间（读完一条消息后）
	MessageLength   uint32      // 3 byte
	MessageType     MessageType // 1 byte
	MessageStreamID uint32      // 4 byte
	ExtendTimestamp uint32      // 4 byte

	// 同一 csid 上跨消息保留的读状态
	TimestampDelta     uint32 // 上一次增量；fmt 0 后为 0
	HasExtendTimestamp bool   // 上一片是否带了 5.3.1.3 扩展时间戳
}

type Message struct {
	MessageHeader
	PayLoad []byte
}

func (m *Message) Type() flv.TagType { return flv.TagType(m.MessageType) }
func (m *Message) Timestamp() uint32 { return m.MessageHeader.Timestamp }
func (m *Message) Data() []byte      { return m.PayLoad }
