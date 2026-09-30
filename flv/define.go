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
