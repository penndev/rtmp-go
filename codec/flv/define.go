// Package flv implements the FLV file format.
// See Video File Format Specification Version 10.1:
// https://veovera.org/docs/legacy/video-file-format-v10-1-spec.pdf
package flv

// E.2 The FLV header
const (
	// Signature byte always 'F' (0x46) 'L' (0x4C) 'V' (0x56)
	SIGNATURE = "FLV"

	// File version (for example, 0x01 for FLV version 1)
	VERSION byte = 1

	// DataOffset UI32: The length of this header in bytes
	DATA_OFFSET uint32 = 9

	// TypeFlagsAudio UB[1]: 1 = Audio tags are present
	TYPE_FLAGS_AUDIO byte = 0x04
	// TypeFlagsVideo UB[1]: 1 = Video tags are present
	TYPE_FLAGS_VIDEO byte = 0x01
	// Audio and video tags are present
	TYPE_FLAGS_AUDIO_VIDEO = TYPE_FLAGS_AUDIO | TYPE_FLAGS_VIDEO
)

// E.4.1 FLV Tag — TagType UB[5]
type TagType byte

const (
	// 8 = audio
	TAG_TYPE_AUDIO TagType = 8
	// 9 = video
	TAG_TYPE_VIDEO TagType = 9
	// 18 = script data
	TAG_TYPE_SCRIPT_DATA TagType = 18
)

// E.4.3.1 VideoTagHeader — CodecID UB[4]
// Enhanced RTMP (V2)
type CodecID byte

const (
	CODEC_ID_SORENSON_H263   CodecID = 2
	CODEC_ID_SCREEN_VIDEO    CodecID = 3
	CODEC_ID_ON2_VP6         CodecID = 4
	CODEC_ID_ON2_VP6_ALPHA   CodecID = 5
	CODEC_ID_SCREEN_VIDEO_V2 CodecID = 6
	CODEC_ID_AVC             CodecID = 7
)

// E.4.2.1 AudioTagHeader — AACPacketType UI8 (SoundFormat == AAC)
type AACPacketType byte

const (
	AAC_PACKET_TYPE_SEQUENCE_HEADER AACPacketType = 0 // AAC sequence header
	AAC_PACKET_TYPE_RAW             AACPacketType = 1 // AAC raw
)

// E.4.3.1 VideoTagHeader — FrameType UB[4]
type FrameType byte

const (
	FRAME_TYPE_KEY                FrameType = 1 // key frame (for AVC, a seekable frame)
	FRAME_TYPE_INTER              FrameType = 2 // inter frame (for AVC, a non-seekable frame)
	FRAME_TYPE_DISPOSABLE_INTER   FrameType = 3 // disposable inter frame (H.263 only)
	FRAME_TYPE_GENERATED_KEY      FrameType = 4 // generated key frame (reserved for server use only)
	FRAME_TYPE_VIDEO_INFO_COMMAND FrameType = 5 // video info/command frame
)

// E.4.3.1 VideoTagHeader — AVCPacketType UI8 (CodecID == AVC)
type AVCPacketType byte

const (
	AVC_PACKET_TYPE_SEQUENCE_HEADER AVCPacketType = 0 // AVC sequence header
	AVC_PACKET_TYPE_NALU            AVCPacketType = 1 // AVC NALU
	AVC_PACKET_TYPE_END_OF_SEQUENCE AVCPacketType = 2 // AVC end of sequence
)

type TagReader interface {
	Type() TagType
	Timestamp() uint32
	Data() []byte
}
