package flv

import (
	"encoding/binary"
	"errors"

	"github.com/penndev/rtmp/codec/h264"
	"github.com/penndev/rtmp/codec/h265"
)

// Enhanced RTMP v2 — Table: Extended VideoTagHeader
// https://veovera.org/docs/enhanced/enhanced-rtmp-v2

// VideoPacketType is the UB[4] packet type when IsExVideoHeader is set.
type VideoPacketType byte

const (
	VIDEO_PACKET_TYPE_SEQUENCE_START VideoPacketType = 0 // SequenceStart
	VIDEO_PACKET_TYPE_CODED_FRAMES   VideoPacketType = 1 // CodedFrames
	VIDEO_PACKET_TYPE_SEQUENCE_END   VideoPacketType = 2 // SequenceEnd
	// CodedFramesX: CompositionTime offset is implicitly 0, so the SI24 is not sent.
	VIDEO_PACKET_TYPE_CODED_FRAMES_X VideoPacketType = 3
	// Metadata: body is AMF metadata, not video. FrameType is ignored.
	VIDEO_PACKET_TYPE_METADATA VideoPacketType = 4
	// MPEG2TSSequenceStart: bitstream is carried as MPEG-2 TS.
	// Mutually exclusive with SequenceStart.
	VIDEO_PACKET_TYPE_MPEG2TS_SEQUENCE_START VideoPacketType = 5
	// Multitrack: turns on video multitrack mode.
	VIDEO_PACKET_TYPE_MULTITRACK VideoPacketType = 6
	// ModEx: modifier for the current packet. While the packet type is ModEx,
	// read modExDataSize = UI8+1 (or UI16+1 when that value is 256), then the
	// ModEx bytes and a VideoPacketModExType, then the next VideoPacketType.
	// 8–15 are reserved.
	VIDEO_PACKET_TYPE_MODEX VideoPacketType = 7
)

// VideoPacketModExType is the UB[4] modifier carried by VideoPacketType ModEx.
type VideoPacketModExType byte

const (
	// TimestampOffsetNano applies only to the current message and does not change the RTMP timestamp.
	// 1 ms = 1_000_000 ns. modExData is at least 3 bytes and holds at most 999_999 ns.
	VIDEO_PACKET_MODEX_TIMESTAMP_OFFSET_NANO VideoPacketModExType = 0
)

// FourCC is a video codec id: four ASCII characters, on the wire a big-endian UI32.
type FourCC string

const (
	FOURCC_VP8  FourCC = "vp08"
	FOURCC_VP9  FourCC = "vp09"
	FOURCC_AV1  FourCC = "av01"
	FOURCC_AVC  FourCC = "avc1"
	FOURCC_HEVC FourCC = "hvc1"
	FOURCC_VVC  FourCC = "vvc1"
)

// AvMultitrackType is the UB[4] multitrack mode, shared by audio and video.
type AvMultitrackType byte

const (
	AV_MULTITRACK_ONE_TRACK               AvMultitrackType = 0 // OneTrack
	AV_MULTITRACK_MANY_TRACKS             AvMultitrackType = 1 // ManyTracks
	AV_MULTITRACK_MANY_TRACKS_MANY_CODECS AvMultitrackType = 2 // ManyTracksManyCodecs
)

// VideoCommand is the UI8 after a command frame (FrameType == 5).
type VideoCommand byte

const (
	VIDEO_COMMAND_START_SEEK VideoCommand = 0 // start of client-side seeking
	VIDEO_COMMAND_END_SEEK   VideoCommand = 1 // end of client-side seeking
)

func TagEnhancedVideo(t *Tag, data []byte) error {
	// ExVideoTagHeader. Low 4 bits are VideoPacketType, not CodecID.
	off := 1
	t.VideoPacketType = VideoPacketType(data[0] & 0x0f)
	for t.VideoPacketType == VIDEO_PACKET_TYPE_MODEX {
		if off >= len(data) {
			return errors.New("flv: truncated video modex")
		}
		modExSize := int(data[off]) + 1
		off++
		if modExSize == 256 {
			if off+2 > len(data) {
				return errors.New("flv: truncated video modex size")
			}
			modExSize = int(binary.BigEndian.Uint16(data[off:off+2])) + 1
			off += 2
		}
		if off+modExSize > len(data) {
			return errors.New("flv: truncated video modex")
		}
		modExData := data[off : off+modExSize]
		off += modExSize
		if off >= len(data) {
			return errors.New("flv: truncated video modex type")
		}
		modExType := VideoPacketModExType(data[off] >> 4)
		t.VideoPacketType = VideoPacketType(data[off] & 0x0f)
		off++
		if modExType != VIDEO_PACKET_MODEX_TIMESTAMP_OFFSET_NANO {
			return errors.New("flv: unknown video modex type")
		}
		if len(modExData) < 3 {
			return errors.New("flv: video modex timestamp too short")
		}
		t.VideoTimestampNanoOffset = uint32(modExData[0])<<16 | uint32(modExData[1])<<8 | uint32(modExData[2])
	}
	if t.VideoPacketType > VIDEO_PACKET_TYPE_MODEX {
		return errors.New("flv: unknown video packet type")
	}

	// Command frame has a UI8 and no body. Metadata ignores FrameType.
	if t.VideoPacketType != VIDEO_PACKET_TYPE_METADATA && t.FrameType == FRAME_TYPE_VIDEO_INFO_COMMAND {
		if off >= len(data) {
			return errors.New("flv: truncated video command")
		}
		t.VideoCommand = VideoCommand(data[off])
		if t.VideoCommand != VIDEO_COMMAND_START_SEEK && t.VideoCommand != VIDEO_COMMAND_END_SEEK {
			return errors.New("flv: unknown video command")
		}
		return nil
	}

	if t.VideoPacketType == VIDEO_PACKET_TYPE_MULTITRACK {
		if off >= len(data) {
			return errors.New("flv: truncated video multitrack")
		}
		t.IsVideoMultitrack = true
		t.AvMultitrackType = AvMultitrackType(data[off] >> 4)
		inner := VideoPacketType(data[off] & 0x0f)
		off++
		if inner == VIDEO_PACKET_TYPE_MULTITRACK || inner > VIDEO_PACKET_TYPE_MODEX {
			return errors.New("flv: invalid multitrack video packet type")
		}
		t.VideoPacketType = inner
		if t.AvMultitrackType > AV_MULTITRACK_MANY_TRACKS_MANY_CODECS {
			return errors.New("flv: unknown video multitrack type")
		}
		if t.AvMultitrackType != AV_MULTITRACK_MANY_TRACKS_MANY_CODECS {
			if off+4 > len(data) {
				return errors.New("flv: truncated video fourcc")
			}
			t.FourCC = FourCC(data[off : off+4])
			off += 4
		}
		// Each track carries its own body. Stop after the shared header.
		if t.AvMultitrackType != AV_MULTITRACK_ONE_TRACK {
			return nil
		}
		if off >= len(data) {
			return errors.New("flv: truncated video track id")
		}
		t.VideoTrackID = data[off]
		off++
	} else {
		if off+4 > len(data) {
			return errors.New("flv: truncated video fourcc")
		}
		t.FourCC = FourCC(data[off : off+4])
		off += 4
	}

	body := data[off:]
	switch t.VideoPacketType {
	case VIDEO_PACKET_TYPE_METADATA, VIDEO_PACKET_TYPE_SEQUENCE_END, VIDEO_PACKET_TYPE_MPEG2TS_SEQUENCE_START:
		return nil
	case VIDEO_PACKET_TYPE_SEQUENCE_START:
		var err error
		switch t.FourCC {
		case FOURCC_AVC:
			t.AVCDecoderConfigurationRecord, err = h264.FormatAVCDecoderConfigurationRecord(body)
		case FOURCC_HEVC:
			t.HEVCDecoderConfigurationRecord, err = h265.FormatHEVCDecoderConfigurationRecord(body)
		}
		return err
	case VIDEO_PACKET_TYPE_CODED_FRAMES:
		if t.FourCC == FOURCC_AVC || t.FourCC == FOURCC_HEVC || t.FourCC == FOURCC_VVC {
			if len(body) < 3 {
				return errors.New("flv: truncated composition time")
			}
			cts := uint32(body[0])<<16 | uint32(body[1])<<8 | uint32(body[2])
			if cts&0x800000 != 0 {
				cts |= 0xff000000
			}
			t.CompositionTime = int32(cts)
		}
		return nil
	case VIDEO_PACKET_TYPE_CODED_FRAMES_X:
		return nil
	default:
		return errors.New("flv: unknown video packet type")
	}
}
