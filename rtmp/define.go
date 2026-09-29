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

// The return value is either a stream ID or an error information object.
// NetConnection is the default communication channel, which has a stream ID 0.
const DEFAULT_STREAM_ID = 1.0

// Preferred chunk size after handshake. Valid sizes are 1 to 2147483647
// (0x7FFFFFFF inclusive); the first bit MUST be zero (5.4.1).
const PREFERRED_CHUNK_SIZE uint32 = 65536

// 5.4.4. Window Acknowledgement Size (5)
// The client or the server sends this message to inform the peer of
// the window size to use between sending acknowledgements. The
// receiving peer MUST send an Acknowledgement (Section 5.4.3) after
// receiving the indicated number of bytes since the last
// Acknowledgement was sent, or beginning of the session if no
// Acknowledgement has yet been sent.
const DEFAULT_WINDOW_ACK_SIZE uint32 = 2500000

// 5.4.5. Set Peer Bandwidth (6)
// The client or the server sends this message to limit the output
// bandwidth of its peer. The peer receiving this message SHOULD
// limit its output bandwidth by limiting the amount of outstanding
// (unacknowledged) data to the window size indicated.
const DEFAULT_PEER_BANDWIDTH uint32 = 2500000

type MessageType byte

const (
	// 5.4. Protocol Control Messages
	// RTMP Chunk Stream uses message type IDs 1, 2, 3, 5, and 6 for
	// protocol control messages. These messages contain information needed
	// by the RTMP Chunk Stream protocol.
	// These protocol control messages MUST have message stream ID 0 (known
	// as the control stream) and be sent in chunk stream ID 2. Protocol
	// control messages take effect as soon as they are received; their
	// timestamps are ignored.
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

