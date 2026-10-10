package rtmp

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"testing"
)

func testChunk(in []byte) *Chunk {
	return &Chunk{
		r:               bufio.NewReader(bytes.NewReader(in)),
		readStreamList:  make(map[int]*Message),
		writeStreamList: make(map[int]MessageHeader),
		readChunkSize:   128,
		writeChunkSize:  128,
	}
}

func put24(b *bytes.Buffer, v uint32) {
	b.Write([]byte{byte(v >> 16), byte(v >> 8), byte(v)})
}

func put32(b *bytes.Buffer, v uint32) {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], v)
	b.Write(buf[:])
}

// fmt 0：绝对时间 + 长度 + 类型 + stream id + payload
func chunk0(csid int, ts uint32, msgType byte, payload []byte) []byte {
	var b bytes.Buffer
	b.WriteByte(byte(csid))
	if ts >= 0xFFFFFF {
		put24(&b, 0xFFFFFF)
	} else {
		put24(&b, ts)
	}
	put24(&b, uint32(len(payload)))
	b.WriteByte(msgType)
	var sid [4]byte
	binary.LittleEndian.PutUint32(sid[:], 1)
	b.Write(sid[:])
	if ts >= 0xFFFFFF {
		put32(&b, ts)
	}
	b.Write(payload)
	return b.Bytes()
}

// fmt 1：时间增量 + 长度 + 类型 + payload
func chunk1(csid int, delta uint32, msgType byte, payload []byte) []byte {
	var b bytes.Buffer
	b.WriteByte(byte(csid) | 0x40)
	put24(&b, delta)
	put24(&b, uint32(len(payload)))
	b.WriteByte(msgType)
	b.Write(payload)
	return b.Bytes()
}

// fmt 2：只有时间增量，长度和类型沿用上一条
func chunk2(csid int, delta uint32, payload []byte) []byte {
	var b bytes.Buffer
	b.WriteByte(byte(csid) | 0x80)
	put24(&b, delta)
	b.Write(payload)
	return b.Bytes()
}

// fmt 3：新消息，时间增量沿用上一条
func chunk3(csid int, payload []byte) []byte {
	var b bytes.Buffer
	b.WriteByte(byte(csid) | 0xC0)
	b.Write(payload)
	return b.Bytes()
}

func TestTimestampAudioVideo(t *testing.T) {
	// 视频帧长变化，走 fmt 0/1，增量 40ms。
	// 音频帧长固定，第二条之后走 fmt 3，增量 23ms。
	var raw bytes.Buffer
	raw.Write(chunk0(6, 0, byte(Video), []byte{0x12}))
	raw.Write(chunk0(4, 0, byte(Audio), []byte{0xaf, 0x01}))
	raw.Write(chunk1(6, 40, byte(Video), []byte{0x22, 0x22}))
	raw.Write(chunk2(4, 23, []byte{0xaf, 0x01}))
	raw.Write(chunk1(6, 40, byte(Video), []byte{0x32}))
	raw.Write(chunk3(4, []byte{0xaf, 0x01}))
	raw.Write(chunk3(4, []byte{0xaf, 0x01}))

	chk := testChunk(raw.Bytes())
	var got []string
	for i := 0; i < 7; i++ {
		msg, err := chk.Read()
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, string(rune('0'+msg.MessageType))+"@"+itoa(msg.Timestamp()))
	}
	want := []string{
		"9@0",
		"8@0",
		"9@40",
		"8@23",
		"9@80",
		"8@46",
		"8@69",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("msg %d: got %s want %s", i, got[i], want[i])
		}
	}
}

func TestTimestampSplitChunk(t *testing.T) {
	// 一条视频拆成两片：第一片 fmt 1 带增量，第二片 fmt 3 不能再加一次
	payload := []byte{1, 2, 3, 4, 5, 6}
	var raw bytes.Buffer
	raw.Write(chunk0(6, 1000, byte(Video), []byte{0x01}))

	var first bytes.Buffer
	first.WriteByte(6 | 0x40)
	put24(&first, 40)
	put24(&first, uint32(len(payload)))
	first.WriteByte(byte(Video))
	first.Write(payload[:4])
	raw.Write(first.Bytes())

	var cont bytes.Buffer
	cont.WriteByte(6 | 0xC0)
	cont.Write(payload[4:])
	raw.Write(cont.Bytes())

	chk := testChunk(raw.Bytes())
	chk.readChunkSize = 4

	msg, err := chk.Read()
	if err != nil {
		t.Fatal(err)
	}
	if msg.Timestamp() != 1000 {
		t.Fatalf("first ts %d", msg.Timestamp())
	}
	msg, err = chk.Read()
	if err != nil {
		t.Fatal(err)
	}
	if msg.Timestamp() != 1040 {
		t.Fatalf("split frame ts %d, want 1040", msg.Timestamp())
	}
	if !bytes.Equal(msg.PayLoad, payload) {
		t.Fatalf("payload %v", msg.PayLoad)
	}
}

func TestTimestampExtended(t *testing.T) {
	const ts = 0x01000000
	chk := testChunk(chunk0(6, ts, byte(Video), []byte{0x01, 0x02, 0x03, 0x04, 0x05}))

	// 读入后再按绝对时间写出，fmt 3 分片也要带 extended timestamp，再读回来仍是同一个时间
	msg, err := chk.Read()
	if err != nil {
		t.Fatal(err)
	}
	if msg.Timestamp() != ts {
		t.Fatalf("read ts %d", msg.Timestamp())
	}

	var out bytes.Buffer
	w := &Chunk{
		w:               bufio.NewWriter(&out),
		readStreamList:  make(map[int]*Message),
		writeStreamList: make(map[int]MessageHeader),
		readChunkSize:   3,
		writeChunkSize:  3,
	}
	if err := w.Write(CSIDVideo, 1, msg); err != nil {
		t.Fatal(err)
	}
	r := testChunk(out.Bytes())
	r.readChunkSize = 3
	msg, err = r.Read()
	if err != nil {
		t.Fatal(err)
	}
	if msg.Timestamp() != ts {
		t.Fatalf("roundtrip ts %d", msg.Timestamp())
	}
	if !bytes.Equal(msg.PayLoad, []byte{0x01, 0x02, 0x03, 0x04, 0x05}) {
		t.Fatalf("payload %v", msg.PayLoad)
	}
}

func itoa(v uint32) string {
	if v == 0 {
		return "0"
	}
	var b [10]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	return string(b[i:])
}
