package flv

import (
	"bytes"
	"testing"
)

func TestTagMarshalRoundTrip(t *testing.T) {
	tag := Tag{
		TagType:   TAG_TYPE_AUDIO,
		Timestamp: 40,
		Data:      []byte{0xaf, 0x01, 1, 2, 3},
	}
	raw := tag.Marshal()
	var got Tag
	if err := got.Unmarshal(raw); err != nil {
		t.Fatal(err)
	}
	if got.TagType != tag.TagType || got.Timestamp != tag.Timestamp || !bytes.Equal(got.Data, tag.Data) {
		t.Fatalf("got=%#v", got)
	}
	if got.PreviousTagSize != 11+5 {
		t.Fatalf("prev=%d", got.PreviousTagSize)
	}
}

func TestStreamRemux(t *testing.T) {
	var srcBuf bytes.Buffer
	w, err := NewFlv(&srcBuf, TYPE_FLAGS_AUDIO_VIDEO)
	if err != nil {
		t.Fatal(err)
	}
	_ = w.TagWrite(TAG_TYPE_SCRIPT_DATA, 0, []byte{0x02, 0x00, 0x0a, 'o', 'n', 'M', 'e', 't', 'a', 'D', 'a', 't', 'a'})
	_ = w.TagWrite(TAG_TYPE_AUDIO, 0, []byte{0xaf, 0x00, 0x12, 0x10})
	_ = w.TagWrite(TAG_TYPE_VIDEO, 40, []byte{0x17, 0x00, 0x00, 0x00, 0x00, 0x01})

	var dstBuf bytes.Buffer
	n := 0
	err = Reader(bytes.NewReader(srcBuf.Bytes()), func(f *FLV) func(*Tag) error {
		dst, err := NewFlv(&dstBuf, f.TypeFlags)
		if err != nil {
			t.Fatal(err)
		}
		return func(tag *Tag) error {
			n++
			return dst.WriteTag(tag)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("tags=%d", n)
	}

	var lastTS uint32
	err = Reader(bytes.NewReader(dstBuf.Bytes()), func(f *FLV) func(*Tag) error {
		return func(tag *Tag) error {
			lastTS = tag.Timestamp
			return nil
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if lastTS != 40 {
		t.Fatalf("lastTS=%d", lastTS)
	}
}
