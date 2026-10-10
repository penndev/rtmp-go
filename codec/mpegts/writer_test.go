package mpegts

import (
	"bytes"
	"testing"
)

func TestMPEGCRC32(t *testing.T) {
	if got := mpegCRC32([]byte("123456789")); got != 0x0376e6e7 {
		t.Fatalf("crc=%08x", got)
	}
	pat := PAT{TransportStreamID: 1, ProgramNumber: 1, ProgramMapPID: PIDPMT}.Marshal()
	if mpegCRC32(pat[:len(pat)-4]) != uint32(pat[len(pat)-4])<<24|uint32(pat[len(pat)-3])<<16|uint32(pat[len(pat)-2])<<8|uint32(pat[len(pat)-1]) {
		t.Fatal("PAT crc mismatch")
	}
}

func TestPTSRoundTrip(t *testing.T) {
	pts := Clock90(1000, 0)
	buf := appendTimestamp(nil, 0x20, pts)
	if parseTimestamp(buf) != pts {
		t.Fatalf("pts %d != %d", parseTimestamp(buf), pts)
	}
	dts := Clock90(40, -10)
	buf = appendTimestamp(nil, 0x10, dts)
	if parseTimestamp(buf) != dts {
		t.Fatalf("dts %d != %d", parseTimestamp(buf), dts)
	}
}

func TestWriteVideoAudio(t *testing.T) {
	var buf bytes.Buffer
	ts := NewTS(&buf)
	ts.SetAudioStreamType(StreamTypeAAC)
	annex := []byte{0, 0, 0, 1, 0x09, 0xf0, 0, 0, 0, 1, 0x67, 0x42, 0, 0, 0, 1, 0x68, 0xce, 0, 0, 0, 1, 0x65, 0x88}
	if err := ts.WriteVideo(40, 0, annex, true); err != nil {
		t.Fatal(err)
	}
	adts := []byte{0xff, 0xf1, 0x50, 0x80, 0x01, 0x1f, 0xfc, 0x11}
	if err := ts.WriteAudio(40, adts); err != nil {
		t.Fatal(err)
	}
	// A second video frame must not repeat PAT when the program is unchanged.
	if err := ts.WriteVideo(80, 0, annex, false); err != nil {
		t.Fatal(err)
	}
	raw := buf.Bytes()
	if len(raw)%PacketSize != 0 {
		t.Fatalf("len=%d", len(raw))
	}

	var sawPAT, sawPMT, sawVideo, sawAudio bool
	var videoCC, audioCC []byte
	for off := 0; off < len(raw); off += PacketSize {
		pkt, err := Parse(raw[off : off+PacketSize])
		if err != nil {
			t.Fatal(err)
		}
		switch pkt.PID {
		case PIDPAT:
			sawPAT = true
			sec := psiSection(t, pkt.Payload)
			if sec[0] != 0x00 {
				t.Fatalf("table id %x", sec[0])
			}
			pmtPID := uint16(sec[10]&0x1f)<<8 | uint16(sec[11])
			if pmtPID != PIDPMT {
				t.Fatalf("pmt pid %x", pmtPID)
			}
		case PIDPMT:
			sawPMT = true
			sec := psiSection(t, pkt.Payload)
			if sec[0] != 0x02 {
				t.Fatalf("pmt table %x", sec[0])
			}
			streams := sec[12 : len(sec)-4]
			if len(streams) != 10 || streams[0] != StreamTypeH264 || streams[5] != StreamTypeAAC {
				t.Fatalf("streams %x", streams)
			}
		case PIDVideo:
			videoCC = append(videoCC, pkt.ContinuityCounter)
			if pkt.PayloadUnitStartIndicator && bytes.HasPrefix(pkt.Payload, []byte{0, 0, 1, streamIDVideo}) {
				sawVideo = true
				got := parseTimestamp(pkt.Payload[9:14])
				if got != Clock90(40, 0) && got != Clock90(80, 0) {
					t.Fatalf("video pts %d", got)
				}
				if pkt.AdaptationField == nil || !pkt.AdaptationField.PCRFlag {
					t.Fatal("video packet missing PCR")
				}
			}
		case PIDAudio:
			audioCC = append(audioCC, pkt.ContinuityCounter)
			if pkt.PayloadUnitStartIndicator && bytes.HasPrefix(pkt.Payload, []byte{0, 0, 1, streamIDAudio}) {
				sawAudio = true
				if parseTimestamp(pkt.Payload[9:14]) != Clock90(40, 0) {
					t.Fatalf("audio pts")
				}
			}
		default:
			t.Fatalf("pid %x", pkt.PID)
		}
	}
	if !sawPAT || !sawPMT || !sawVideo || !sawAudio {
		t.Fatalf("pat=%v pmt=%v video=%v audio=%v", sawPAT, sawPMT, sawVideo, sawAudio)
	}
	checkCC(t, videoCC)
	checkCC(t, audioCC)
}

func psiSection(t *testing.T, payload []byte) []byte {
	t.Helper()
	if len(payload) < 4 || payload[0] != 0 {
		t.Fatalf("psi pointer %#v", payload)
	}
	sec := payload[1:]
	n := 3 + int(sec[1]&0x0f)<<8 + int(sec[2])
	if n > len(sec) {
		t.Fatalf("section length %d", n)
	}
	return sec[:n]
}

func checkCC(t *testing.T, cc []byte) {
	t.Helper()
	for i := 1; i < len(cc); i++ {
		if cc[i] != (cc[i-1]+1)&0x0f {
			t.Fatalf("continuity %v", cc)
		}
	}
}
