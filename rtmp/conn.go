package rtmp

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"github.com/penndev/rtmp/amf"
)

type Conn struct {
	Chunk
	nc       net.Conn
	App      string  // from connect
	Stream   string  // stream name from FCPublish / publish / play
	StreamID float64 // NetStream ID from createStream; 0 is reserved for NetConnection
}

func (c *Conn) Connect() (*ConnectCommand, error) {
	msg, err := c.Read()
	if err != nil {
		return nil, err
	}
	var values []amf.Value
	switch msg.MessageType {
	case AMF0CommandMessage:
		values, err = amf.Decode0(msg.PayLoad)
	case AMF3CommandMessage:
		values, err = amf.Decode3(msg.PayLoad)
	default:
		return nil, fmt.Errorf("unknown message type: %d", msg.MessageType)
	}
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, errors.New("empty command message")
	}

	name, _ := values[0].(string)
	if name != "connect" {
		return nil, fmt.Errorf("unknown command: %s", name)
	}
	cmd, err := ConnectParse(values)
	if err != nil {
		return nil, err
	}
	c.App = cmd.CommandObject.App
	return cmd, nil
}

func (c *Conn) ConnectReply(cmd *ConnectCommand, success bool) error {
	payload, err := amf.Encode0(ConnectBuild(cmd, success)...)
	if err != nil {
		return err
	}
	// Command messages (20): typically chunk stream ID 3, message stream ID 0
	return c.Write(CSIDCommand, 0, &Message{
		MessageHeader: MessageHeader{MessageType: AMF0CommandMessage},
		PayLoad:       payload,
	})
}

// 5.4.1. Set Chunk Size (1)
// The maximum chunk size defaults to 128 bytes, but the client or the
// server can change this value, and updates its peer using this
// message.
func (c *Conn) SetChunkSize(size uint32) error {
	// chunk size (31 bits): This field holds the new maximum chunk size,
	// in bytes, which will be used for all of the sender’s subsequent
	// chunks until further notice. Valid sizes are 1 to 2147483647
	// (0x7FFFFFFF) inclusive; the first bit MUST be zero.
	if size < 1 || size > 0x7FFFFFFF {
		return errors.New("SetChunkSize: size must be 1..2147483647")
	}
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, size)
	// Protocol control messages MUST have message stream ID 0 and be
	// sent in chunk stream ID 2 (5.4).
	if err := c.Write(CSIDProtocolControl, 0, &Message{
		MessageHeader: MessageHeader{MessageType: SetChunkSize},
		PayLoad:       payload,
	}); err != nil {
		return err
	}
	c.writeChunkSize = size
	return nil
}

// 5.4.4. Window Acknowledgement Size (5)
// The client or the server sends this message to inform the peer of
// the window size to use between sending acknowledgements.
func (c *Conn) SetWindowAcknowledgementSize(size uint32) error {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, size)
	return c.Write(CSIDProtocolControl, 0, &Message{
		MessageHeader: MessageHeader{MessageType: WindowAcknowledgementSize},
		PayLoad:       payload,
	})
}

// 5.4.5. Set Peer Bandwidth (6)
// The client or the server sends this message to limit the output
// bandwidth of its peer.
// Limit Type: 0 - Hard, 1 - Soft, 2 - Dynamic.
func (c *Conn) SetBandwidth(size uint32) error {
	payload := make([]byte, 5)
	binary.BigEndian.PutUint32(payload[:4], size)
	payload[4] = 2 // Dynamic
	return c.Write(CSIDProtocolControl, 0, &Message{
		MessageHeader: MessageHeader{MessageType: SetPeerBandwidth},
		PayLoad:       payload,
	})
}

func (c *Conn) CreateStream() (*CreateStreamCommand, error) {
	for {
		msg, err := c.Read()
		if err != nil {
			return nil, err
		}
		var values []amf.Value
		switch msg.MessageType {
		case AMF0CommandMessage:
			values, err = amf.Decode0(msg.PayLoad)
		case AMF3CommandMessage:
			values, err = amf.Decode3(msg.PayLoad)
		default:
			return nil, fmt.Errorf("unknown message type: %d", msg.MessageType)
		}
		if err != nil {
			return nil, err
		}
		if len(values) == 0 {
			return nil, errors.New("empty command message")
		}
		name, _ := values[0].(string)
		switch name {
		case "createStream":
			return CreateStreamParse(values)
		case "FCPublish":
			// commandName, transactionId, null, streamName
			if len(values) >= 4 {
				if s, ok := values[3].(string); ok {
					c.Stream = s
				}
			}
			continue
		case "releaseStream", "FCUnpublish":
			continue
		default:
			return nil, fmt.Errorf("unknown command: %s", name)
		}
	}
}

// 7.2.1.3. createStream — server response
// _result or _error, Transaction ID, Command Object (null), Stream ID
func (c *Conn) CreateStreamReply(cmd *CreateStreamCommand, streamID float64) error {
	if streamID == 0 {
		return errors.New("createStream reply: Stream ID must not be 0 (reserved for NetConnection)")
	}
	c.StreamID = streamID
	payload, err := amf.Encode0("_result", cmd.TransactionID, nil, streamID)
	if err != nil {
		return err
	}
	// createStream uses the default communication channel (message stream ID 0)
	return c.Write(CSIDCommand, 0, &Message{
		MessageHeader: MessageHeader{MessageType: AMF0CommandMessage},
		PayLoad:       payload,
	})
}

// 7.2.2.6. publish — server response onStatus
func (c *Conn) PublishReply(success bool) error {
	res := amf.Object{
		"level":       "status",
		"description": "Start publishing",
	}
	if success {
		res["code"] = "NetStream.Publish.Start"
	} else {
		res["code"] = "NetStream.Publish.BadName"
	}
	payload, err := amf.Encode0("onStatus", 0.0, nil, res)
	if err != nil {
		return err
	}
	return c.Write(CSIDCommand, uint32(c.StreamID), &Message{
		MessageHeader: MessageHeader{MessageType: AMF0CommandMessage},
		PayLoad:       payload,
	})
}

// 7.2.2.1. play — server response onStatus
func (c *Conn) PlayReply(success bool) error {
	res := amf.Object{
		"level":       "status",
		"description": "Start playing",
	}
	if success {
		res["code"] = "NetStream.Play.Start"
	} else {
		res["code"] = "NetStream.Play.Failed"
	}
	payload, err := amf.Encode0("onStatus", 0, nil, res)
	if err != nil {
		return err
	}
	return c.Write(CSIDCommand, uint32(c.StreamID), &Message{
		MessageHeader: MessageHeader{MessageType: AMF0CommandMessage},
		PayLoad:       payload,
	})
}

// 6.2. User Control Messages — Stream Begin (event type 0)
func (c *Conn) StreamBegin(streamID uint32) error {
	payload := make([]byte, 6)
	binary.BigEndian.PutUint32(payload[2:], streamID)
	return c.Write(CSIDProtocolControl, 0, &Message{
		MessageHeader: MessageHeader{MessageType: UserControl},
		PayLoad:       payload,
	})
}

// 6.2. User Control Messages — Stream EOF (event type 1)
func (c *Conn) StreamEOF(streamID uint32) error {
	payload := make([]byte, 6)
	binary.BigEndian.PutUint16(payload[:2], 1)
	binary.BigEndian.PutUint32(payload[2:], streamID)
	return c.Write(CSIDProtocolControl, 0, &Message{
		MessageHeader: MessageHeader{MessageType: UserControl},
		PayLoad:       payload,
	})
}

// NetStream.Play.Stop — notify player the stream has ended
func (c *Conn) PlayStop() error {
	res := amf.Object{
		"level":       "status",
		"code":        "NetStream.Play.Stop",
		"description": "Stopped playing stream.",
	}
	payload, err := amf.Encode0("onStatus", 0.0, nil, res)
	if err != nil {
		return err
	}
	return c.Write(CSIDCommand, uint32(c.StreamID), &Message{
		MessageHeader: MessageHeader{MessageType: AMF0CommandMessage},
		PayLoad:       payload,
	})
}

func NewConn(nc net.Conn) *Conn {
	return &Conn{
		nc: nc,
		Chunk: Chunk{
			r:               bufio.NewReader(nc),
			w:               bufio.NewWriter(nc),
			readStreamList:  make(map[int]*Message),
			writeStreamList: make(map[int]MessageHeader),
			readChunkSize:   DEFAULT_CHUNK_SIZE,
			writeChunkSize:  DEFAULT_CHUNK_SIZE,
		},
	}
}
