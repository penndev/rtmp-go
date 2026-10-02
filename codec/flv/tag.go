package flv

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/penndev/rtmp/codec/h264"
)

// E.4.2.1 Audio Tag Header
type AudioTagHeader struct {
	// SoundFormat UB[4]
	SoundFormat byte
	// SoundRate UB[2]
	SoundRate byte
	// SoundSize UB[1]
	SoundSize byte
	// SoundType UB[1]
	SoundType byte
	// AACPacketType UI8: only if SoundFormat == 10 (AAC)
	AACPacketType byte
}

// E.4.3.1 Video Tag Header
type VideoTagHeader struct {
	// FrameType UB[4]
	FrameType byte
	// CodecID UB[4]
	CodecID byte
	// AVCPacketType UI8: only if CodecID == 7 (AVC)
	AVCPacketType byte
	// CompositionTime SI24: if AVCPacketType == 1 Composition time offset, else 0
	CompositionTime int32

	AVCDecoderConfigurationRecord h264.AVCDecoderConfigurationRecord
}

// E.4.1 FLV Tag (+ trailing PreviousTagSize in Marshal/Unmarshal).
type Tag struct {
	// Reserved UB[2]: Reserved for FMS, should be 0
	Reserved byte
	// Filter UB[1]: 0 = No pre-processing required
	Filter byte
	// TagType UB[5]: 8 = audio, 9 = video, 18 = script data
	TagType byte
	// DataSize UI24: Number of bytes after StreamID to end of tag
	DataSize uint32
	// Timestamp SI32 ms = (TimestampExtended<<24) | Timestamp UI24
	Timestamp uint32
	// StreamID UI24: Always 0
	StreamID uint32
	// Data: tag body after StreamID
	Data []byte
	// PreviousTagSize UI32 after this tag: 11 + DataSize for FLV version 1
	PreviousTagSize uint32

	AudioTagHeader
	VideoTagHeader
	Nalu []h264.Nalu
}

// Marshal encodes FLVTAG + PreviousTagSize into bytes.
func (t Tag) Marshal() []byte {
	dataLen := len(t.Data)
	out := make([]byte, 11+dataLen+4)
	out[0] = (t.Reserved&0x03)<<6 | (t.Filter&0x01)<<5 | (t.TagType & 0x1f)
	out[1] = byte(dataLen >> 16)
	out[2] = byte(dataLen >> 8)
	out[3] = byte(dataLen)
	ts := t.Timestamp
	out[4] = byte(ts >> 16)
	out[5] = byte(ts >> 8)
	out[6] = byte(ts)
	out[7] = byte(ts >> 24)
	// StreamID UI24: Always 0 (bytes 8-10)
	copy(out[11:], t.Data)
	prev := t.PreviousTagSize
	if prev == 0 {
		prev = uint32(11 + dataLen)
	}
	binary.BigEndian.PutUint32(out[11+dataLen:], prev)
	return out
}

// Unmarshal decodes one FLVTAG + PreviousTagSize from b.
func (t *Tag) Unmarshal(b []byte) error {
	if len(b) < 15 {
		return errors.New("flv: tag too short")
	}
	t.Reserved = (b[0] >> 6) & 0x03
	t.Filter = (b[0] >> 5) & 0x01
	t.TagType = b[0] & 0x1f
	t.DataSize = uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	t.Timestamp = uint32(b[7])<<24 | uint32(b[4])<<16 | uint32(b[5])<<8 | uint32(b[6])
	t.StreamID = uint32(b[8])<<16 | uint32(b[9])<<8 | uint32(b[10])
	need := int(11 + t.DataSize + 4)
	if len(b) < need {
		return errors.New("flv: tag truncated")
	}
	t.Data = append([]byte(nil), b[11:11+t.DataSize]...)
	t.PreviousTagSize = binary.BigEndian.Uint32(b[11+t.DataSize:])
	return nil
}

// ReadFrom reads one FLVTAG + PreviousTagSize from r.
func (t *Tag) ReadFrom(r io.Reader) (int64, error) {
	var hdr [11]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return 0, err
	}
	n := int64(11)
	t.Reserved = (hdr[0] >> 6) & 0x03
	t.Filter = (hdr[0] >> 5) & 0x01
	t.TagType = hdr[0] & 0x1f
	t.DataSize = uint32(hdr[1])<<16 | uint32(hdr[2])<<8 | uint32(hdr[3])
	t.Timestamp = uint32(hdr[7])<<24 | uint32(hdr[4])<<16 | uint32(hdr[5])<<8 | uint32(hdr[6])
	t.StreamID = uint32(hdr[8])<<16 | uint32(hdr[9])<<8 | uint32(hdr[10])
	t.Data = make([]byte, t.DataSize)
	if _, err := io.ReadFull(r, t.Data); err != nil {
		return n, err
	}
	n += int64(t.DataSize)
	var prev [4]byte
	if _, err := io.ReadFull(r, prev[:]); err != nil {
		return n, err
	}
	n += 4
	t.PreviousTagSize = binary.BigEndian.Uint32(prev[:])
	return n, nil
}

// WriteTo writes FLVTAG + PreviousTagSize to w.
func (t *Tag) WriteTo(w io.Writer) (int64, error) {
	t.DataSize = uint32(len(t.Data))
	if t.PreviousTagSize == 0 {
		t.PreviousTagSize = 11 + t.DataSize
	}
	raw := t.Marshal()
	n, err := w.Write(raw)
	return int64(n), err
}

func FormatTag(tagType byte, timestamp uint32, data []byte, avcDecoderConfigurationRecord *h264.AVCDecoderConfigurationRecord) (*Tag, error) {
	t := &Tag{
		TagType:   tagType,
		Timestamp: timestamp,
		DataSize:  uint32(len(data)),
		Data:      data,
	}
	switch tagType {
	case TAG_TYPE_AUDIO:
		if len(data) < 1 {
			return nil, errors.New("flv: audio tag too short")
		}
		t.AudioTagHeader = AudioTagHeader{
			SoundFormat: data[0] >> 4,
			SoundRate:   (data[0] >> 2) & 0x03,
			SoundSize:   (data[0] >> 1) & 0x01,
			SoundType:   data[0] & 0x01,
		}
		if t.SoundFormat == SOUND_FORMAT_AAC {
			if len(data) < 2 {
				return nil, errors.New("flv: AAC audio tag too short")
			}
			t.AACPacketType = data[1]
		}
	case TAG_TYPE_VIDEO:
		if len(data) < 1 {
			return nil, errors.New("flv: video tag too short")
		}
		t.VideoTagHeader = VideoTagHeader{
			FrameType: data[0] >> 4,
			CodecID:   data[0] & 0x0f,
		}
		if t.CodecID == CODEC_ID_AVC {
			if len(data) < 5 {
				return nil, errors.New("flv: AVC video tag too short")
			}
			t.AVCPacketType = data[1]
			cts := uint32(data[2])<<16 | uint32(data[3])<<8 | uint32(data[4])
			if cts&0x800000 != 0 {
				cts |= 0xff000000 // sign-extend SI24
			}
			t.CompositionTime = int32(cts)
			switch t.AVCPacketType {
			case AVC_PACKET_TYPE_SEQUENCE_HEADER:
				var err error
				t.AVCDecoderConfigurationRecord, err = h264.FormatAVCDecoderConfigurationRecord(data[5:])
				if err != nil {
					return nil, err
				}
			case AVC_PACKET_TYPE_NALU:
				if avcDecoderConfigurationRecord == nil {
					return nil, errors.New("flv: missing AVCDecoderConfigurationRecord")
				}
				size := int(avcDecoderConfigurationRecord.LengthSizeMinusOne) + 1
				payload := data[5:]
				for len(payload) >= size {
					l := 0
					for i := 0; i < size; i++ {
						l = l<<8 | int(payload[i])
					}
					payload = payload[size:]
					if len(payload) < l {
						return nil, errors.New("flv: truncated NAL unit")
					}
					n, err := h264.FormatNalu(payload[:l])
					if err != nil {
						return nil, err
					}
					t.Nalu = append(t.Nalu, n)
					payload = payload[l:]
				}
			}
		}
	case TAG_TYPE_SCRIPT_DATA:
	default:
		return nil, errors.New("flv: unknown tag type")
	}
	return t, nil
}
