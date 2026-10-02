package flv

import (
	"encoding/binary"
	"errors"
	"io"
)

// FLV streams tags in one direction: either read (r) or write (w).
// E.2 header fields live on FLV itself.
type FLV struct {
	// Signature UI8[3]: always 'F' 'L' 'V'
	Signature [3]byte
	// Version UI8: File version (for example, 0x01 for FLV version 1)
	Version byte
	// TypeFlags: TypeFlagsAudio bit2, TypeFlagsVideo bit0
	TypeFlags byte
	// DataOffset UI32: The length of this header in bytes
	DataOffset uint32
	// PreviousTagSize tracks the last tag size (UI32 after each tag).
	PreviousTagSize uint32

	r io.Reader
	w io.Writer
}

func (f *FLV) headerWire() []byte {
	b := make([]byte, 9)
	copy(b[0:3], f.Signature[:])
	b[3] = f.Version
	b[4] = f.TypeFlags
	binary.BigEndian.PutUint32(b[5:9], f.DataOffset)
	return b
}

func (f *FLV) readHeader(r io.Reader) error {
	var raw [9]byte
	if _, err := io.ReadFull(r, raw[:]); err != nil {
		return err
	}
	copy(f.Signature[:], raw[0:3])
	f.Version = raw[3]
	f.TypeFlags = raw[4]
	f.DataOffset = binary.BigEndian.Uint32(raw[5:9])
	if string(f.Signature[:]) != SIGNATURE {
		return errors.New("flv: bad signature")
	}
	if f.DataOffset < 9 {
		return errors.New("flv: DataOffset < 9")
	}
	if f.DataOffset > 9 {
		if _, err := io.CopyN(io.Discard, r, int64(f.DataOffset-9)); err != nil {
			return err
		}
	}
	return nil
}

// NewFlv creates a write-only FLV, writes header + PreviousTagSize0.
func NewFlv(w io.Writer, typeFlags byte) (*FLV, error) {
	f := &FLV{
		Signature:  [3]byte{'F', 'L', 'V'},
		Version:    VERSION,
		TypeFlags:  typeFlags,
		DataOffset: DATA_OFFSET,
		w:          w,
	}
	if _, err := w.Write(f.headerWire()); err != nil {
		return nil, err
	}
	var prev0 [4]byte // PreviousTagSize0: Always 0
	if _, err := w.Write(prev0[:]); err != nil {
		return nil, err
	}
	return f, nil
}

// Reader parses the FLV header from r, calls onHeader with the FLV (header ready),
// then streams tags into the handler returned by onHeader.
//
//	flv.Reader(r, func(f *flv.FLV) func(*flv.Tag) error {
//	    return func(tag *flv.Tag) error { ... }
//	})
func Reader(r io.Reader, onHeader func(*FLV) func(*Tag) error) error {
	f := &FLV{r: r}
	if err := f.readHeader(r); err != nil {
		return err
	}
	var prev0 [4]byte
	if _, err := io.ReadFull(r, prev0[:]); err != nil {
		return err
	}
	if binary.BigEndian.Uint32(prev0[:]) != 0 {
		return errors.New("flv: PreviousTagSize0 must be 0")
	}
	handler := onHeader(f)
	if handler == nil {
		return nil
	}
	for {
		var tag Tag
		_, err := tag.ReadFrom(f.r)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil
		}
		if err != nil {
			return err
		}
		f.PreviousTagSize = tag.PreviousTagSize
		if err := handler(&tag); err != nil {
			return err
		}
	}
}

// WriteTag writes one tag to the writer (sets DataSize / PreviousTagSize).
func (f *FLV) WriteTag(tag *Tag) error {
	if f.w == nil {
		return errors.New("flv: not a writer")
	}
	tag.DataSize = uint32(len(tag.Data))
	tag.PreviousTagSize = 11 + tag.DataSize
	if _, err := tag.WriteTo(f.w); err != nil {
		return err
	}
	f.PreviousTagSize = tag.PreviousTagSize
	return nil
}

// TagWrite builds a tag from type/timestamp/payload and writes it.
func (f *FLV) TagWrite(tagType byte, timestamp uint32, data []byte) error {
	return f.WriteTag(&Tag{
		TagType:   tagType,
		Timestamp: timestamp,
		Data:      data,
	})
}
