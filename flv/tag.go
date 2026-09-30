package flv

import (
	"encoding/binary"
	"errors"
	"io"
)

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
