package rtmp

//  Version (8 bits): In C0, this field identifies the RTMP version
//  requested by the client. In S0, this field identifies the RTMP
//  version selected by the server. The version defined by this
//  specification is 3. Values 0-2 are deprecated values used by
//  earlier proprietary products; 4-31 are reserved for future
//  implementations; and 32-255 are not allowed (to allow
//  distinguishing RTMP from text-based protocols, which always start
//  with a printable character). A server that does not recognize the
//  client’s requested version SHOULD respond with 3. The client MAY
//  choose to degrade to version 3, or to abandon the handsha
const VERSION = 0x03

//  Tthe maximum chunk size defaults to 128 bytes, but the client or the
//  server can change this value, and updates its peer using this
//  message. For example, suppose a client wants to send 131 bytes of
//  audio data and the chunk size is 128. In this case, the client can
//  send this message to the server to notify it that the chunk size is
//  now 131 bytes. The client can then send the audio data in a single
//  chunk.
const DEFAULT_CHUNK_SIZE = 128

// MessageType 消息类型
// 见 RTMP 规范 5.4 / 6.x / 7.x
type MessageType byte

const (
	// 5.4. Protocol Control Messages
	SetChunkSize              MessageType = 1
	AbortMessage              MessageType = 2
	Acknowledgement           MessageType = 3
	UserControl               MessageType = 4
	WindowAcknowledgementSize MessageType = 5
	SetPeerBandwidth          MessageType = 6

	// 7. Audio / Video
	Audio MessageType = 8
	Video MessageType = 9

	// Data / Shared Object / Command
	AMF3DataMessage    MessageType = 15
	AMF3SharedObject   MessageType = 16
	AMF3CommandMessage MessageType = 17
	AMF0DataMessage    MessageType = 18
	AMF0SharedObject   MessageType = 19
	AMF0CommandMessage MessageType = 20
	AggregateMessage   MessageType = 22
)