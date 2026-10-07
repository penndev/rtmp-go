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
const (
	// 8 = audio
	TAG_TYPE_AUDIO byte = 8
	// 9 = video
	TAG_TYPE_VIDEO byte = 9
	// 18 = script data
	TAG_TYPE_SCRIPT_DATA byte = 18
)

// E.4.3.1 VideoTagHeader — CodecID UB[4]
// Enhanced RTMP (V2)
const (
	CODEC_ID_SORENSON_H263   byte = 2
	CODEC_ID_SCREEN_VIDEO    byte = 3
	CODEC_ID_ON2_VP6         byte = 4
	CODEC_ID_ON2_VP6_ALPHA   byte = 5
	CODEC_ID_SCREEN_VIDEO_V2 byte = 6
	CODEC_ID_AVC             byte = 7
)

// E.4.2.1 AudioTagHeader — SoundFormat UB[4]
const (
	SOUND_FORMAT_LINEAR_PCM_PE    byte = 0  // Linear PCM, platform endian
	SOUND_FORMAT_ADPCM            byte = 1  // ADPCM
	SOUND_FORMAT_MP3              byte = 2  // MP3
	SOUND_FORMAT_LINEAR_PCM_LE    byte = 3  // Linear PCM, little endian
	SOUND_FORMAT_NELLYMOSER_16KHZ byte = 4  // Nellymoser 16-kHz mono
	SOUND_FORMAT_NELLYMOSER_8KHZ  byte = 5  // Nellymoser 8-kHz mono
	SOUND_FORMAT_NELLYMOSER       byte = 6  // Nellymoser
	SOUND_FORMAT_G711_A_LAW       byte = 7  // G.711 A-law logarithmic PCM
	SOUND_FORMAT_G711_MU_LAW      byte = 8  // G.711 mu-law logarithmic PCM
	SOUND_FORMAT_RESERVED         byte = 9  // reserved
	SOUND_FORMAT_AAC              byte = 10 // AAC
	SOUND_FORMAT_SPEEX            byte = 11 // Speex
	SOUND_FORMAT_MP3_8KHZ         byte = 14 // MP3 8-Khz
	SOUND_FORMAT_DEVICE_SPECIFIC  byte = 15 // Device-specific sound
)

// E.4.2.1 AudioTagHeader — SoundRate UB[2]
const (
	SOUND_RATE_5_5KHZ byte = 0 // 5.5 kHz
	SOUND_RATE_11KHZ  byte = 1 // 11 kHz
	SOUND_RATE_22KHZ  byte = 2 // 22 kHz
	SOUND_RATE_44KHZ  byte = 3 // 44 kHz
)

// E.4.2.1 AudioTagHeader — SoundSize UB[1]
const (
	SOUND_SIZE_8BIT  byte = 0 // snd8Bit
	SOUND_SIZE_16BIT byte = 1 // snd16Bit
)

// E.4.2.1 AudioTagHeader — SoundType UB[1]
const (
	SOUND_TYPE_MONO   byte = 0 // sndMono
	SOUND_TYPE_STEREO byte = 1 // sndStereo
)

// E.4.2.1 AudioTagHeader — AACPacketType UI8 (SoundFormat == AAC)
const (
	AAC_PACKET_TYPE_SEQUENCE_HEADER byte = 0 // AAC sequence header
	AAC_PACKET_TYPE_RAW             byte = 1 // AAC raw
)

// E.4.3.1 VideoTagHeader — FrameType UB[4]
const (
	FRAME_TYPE_KEY                byte = 1 // key frame (for AVC, a seekable frame)
	FRAME_TYPE_INTER              byte = 2 // inter frame (for AVC, a non-seekable frame)
	FRAME_TYPE_DISPOSABLE_INTER   byte = 3 // disposable inter frame (H.263 only)
	FRAME_TYPE_GENERATED_KEY      byte = 4 // generated key frame (reserved for server use only)
	FRAME_TYPE_VIDEO_INFO_COMMAND byte = 5 // video info/command frame
)

// E.4.3.1 VideoTagHeader — AVCPacketType UI8 (CodecID == AVC)
const (
	AVC_PACKET_TYPE_SEQUENCE_HEADER byte = 0 // AVC sequence header
	AVC_PACKET_TYPE_NALU            byte = 1 // AVC NALU
	AVC_PACKET_TYPE_END_OF_SEQUENCE byte = 2 // AVC end of sequence
)

type TagReader interface {
	Type() byte
	Timestamp() uint32
	Data() []byte
}
