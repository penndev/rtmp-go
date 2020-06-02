package rtmp

const (
	// Version rtmp handshake version is 0x03
	Version = 3

	// ChunkDefaultSize The maximum chunk size defaults to 128 bytes
	//	The maximum chunk size SHOULD be at least 128 bytes, and MUST be at
	//	least 1 byte. The maximum chunk size is maintained independently for
	//	each direction.
	ChunkDefaultSize = 128
)

// RTMP Chunk Stream uses message type IDs 1, 2, 3, 5, and 6 for protocol control messages.
// These messages contain information needed  by the RTMP Chunk Stream protocol.
const (

	/*
		Protocol control message 1, Set Chunk Size, is used to notify the
		peer of a new maximum chunk size.
		The maximum chunk size defaults to 128 bytes, but the client or the
		server can change this value, and updates its peer using this
		message. For example, suppose a client wants to send 131 bytes of
		audio data and the chunk size is 128. In this case, the client can
		send this message to the server to notify it that the chunk size is
		now 131 bytes. The client can then send the audio data in a single chunk.

		chunk size (31 bits): This field holds the new maximum chunk size,
		in bytes, which will be used for all of the sender’s subsequent
		chunks until further notice. Valid sizes are 1 to 2147483647
		(0x7FFFFFFF) inclusive; however, all sizes greater than 16777215
		(0xFFFFFF) are equivalent since no chunk is larger than one
		message, and no message is larger than 16777215 bytes.
	*/
	SetChunkSize = 0x01

	/*
				Protocol control message 2, Abort Message, is used to notify the peer
				if it is waiting for chunks to complete a message, then to discard
				the partially received message over a chunk stream. The peer
				receives the chunk stream ID as this protocol message’s payload. An
				application may send this message when closing in order to indicate
				that further processing of the messages is not required.

				chunk stream ID (32 bits): This field holds the chunk stream ID,
		 		whose current message is to be discarded.
	*/
	AbortMessage = 0x02

	/*
		The client or the server MUST send an acknowledgment to the peer
		after receiving bytes equal to the window size. The window size is
		the maximum number of bytes that the sender sends without receiving
		acknowledgment from the receiver. This message specifies the
		sequence number, which is the number of the bytes received so far.

		sequence number (32 bits): This field holds the number of bytes
		received so far.
	*/
	Acknowledgement = 0x03

	/*
		The client or the server sends this message to inform the peer of the
		window size to use between sending acknowledgments. The sender
		expects acknowledgment from its peer after the sender sends window
		size bytes. The receiving peer MUST send an Acknowledgement
		(Section 0x03) after receiving the indicated number of bytes since
		the last Acknowledgement was sent, or from the beginning of the
		session if no Acknowledgement has yet been sent.

		Payload for the ‘Window Acknowledgement Size’ protocol message
	*/
	WindowAcknowledgementSize = 0x5

	/*
		The client or the server sends this message to limit the output
		bandwidth of its peer. The peer receiving this message limits its
		output bandwidth by limiting the amount of sent but unacknowledged
		data to the window size indicated in this message. The peer
		receiving this message SHOULD respond with a Window Acknowledgement
		Size message if the window size is different from the last one sent
		to the sender of this message.

		Payload for the ‘Set Peer Bandwidth’ protocol message
		The Limit Type is one of the following values:
		0 - Hard: The peer SHOULD limit its output bandwidth to the
		indicated window size.
		1 - Soft: The peer SHOULD limit its output bandwidth to the the
		window indicated in this message or the limit already in effect,
		whichever is smaller.
		2 - Dynamic: If the previous Limit Type was Hard, treat this message
		as though it was marked Hard, otherwise ignore this message.
	*/
	SetPeerBandwidth = 0x06
)

// RTMP Command Messages
// This section describes the different types of messages and commands
// that are exchanged between the server and the client to communicate
// with each other.
const (
	/*
		The different types of messages that are exchanged between the server
		and the client include audio messages for sending the audio data,
		video messages for sending video data, data messages for sending any
		user data, shared object messages, and command messages. Shared
		object messages provide a general purpose way to manage distributed
		data among multiple clients and a server. Command messages carry the
		AMF encoded commands between the client and the server. A client or
		a server can request Remote Procedure Calls (RPC) over streams that
		are communicated using the command messages to the peer.

		The server and the client send messages over the network to
		communicate with each other. The messages can be of any type which
		includes audio messages, video messages, command messages, shared
		object messages, data messages, and user control messages.
	*/

	/*
		Command messages carry the AMF-encoded commands between the client
		and the server. These messages have been assigned message type value
		of 20 for AMF0 encoding and message type value of 17 for AMF3
		encoding. These messages are sent to perform some operations like
		connect, createStream, publish, play, pause on the peer. Command
		messages like onstatus, result etc. are used to inform the sender
		about the status of the requested commands. A command message
		consists of command name, transaction ID, and command object that
		contains related parameters. A client or a server can request Remote
		Procedure Calls (RPC) over streams that are communicated using the
		command messages to the peer.
		Command Message (20, 17)
	*/
	CommandMessageAMF0 = 20
	CommandMessageAMF3 = 17

	/*
		 	Data Message (18, 15)
			The client or the server sends this message to send Metadata or any
			user data to the peer. Metadata includes details about the
			data(audio, video etc.) like creation time, duration, theme and so
			on. These messages have been assigned message type value of 18 for
			AMF0 and message type value of 15 for AMF3.
	*/
	DataMessageAMF0 = 18
	DataMessageAMF3 = 15

	/*
		Shared Object Message (19, 16)
		A shared object is a Flash object (a collection of name value pairs)
		that are in synchronization across multiple clients, instances, and
		so on. The message types 19 for AMF0 and 16 for AMF3 are reserved
		for shared object events. Each message can contain multiple events
	*/
	SharedObjectMessageAMF0 = 19
	SharedObjectMessageAMF3 = 16

	// The client or the server sends this message to send audio data to the
	// peer. The message type value of 8 is reserved for audio messages.
	AudioMessage = 8

	// 	The client or the server sends this message to send video data to the
	// peer. The message type value of 9 is reserved for video messages.
	VideoMessage = 9

	// An aggregate message is a single message that contains a series of
	// RTMP sub-messages using the format described in Section 6.1. Message
	// type 22 is used for aggregate messages.
	AggregateMessage = 22
)
