package mpegts

const (
	streamIDVideo = 0xe0
	streamIDAudio = 0xc0
)

// PES is a packetized elementary stream header plus payload.
// ISO/IEC 13818-1 Table 2-21. PTS and DTS are in units of 90 kHz.
type PES struct {
	StreamID byte
	PTS      uint64
	DTS      uint64
	HasDTS   bool
	Data     []byte
}

func (p PES) Marshal() []byte {
	headerDataLen := 5
	if p.HasDTS {
		headerDataLen = 10
	}
	// Bytes after PES_packet_length. 0 means the length is unbounded, which
	// video packets in a transport stream are allowed to use.
	rest := 3 + headerDataLen + len(p.Data)
	packetLength := rest
	if packetLength > 0xffff || p.StreamID == streamIDVideo {
		packetLength = 0
	}
	buf := make([]byte, 0, 6+3+headerDataLen+len(p.Data))
	buf = append(buf, 0x00, 0x00, 0x01, p.StreamID)
	buf = append(buf, byte(packetLength>>8), byte(packetLength))
	buf = append(buf, 0x80) // '10'
	if p.HasDTS {
		buf = append(buf, 0xc0) // PTS_DTS_flags = '11'
	} else {
		buf = append(buf, 0x80) // PTS_DTS_flags = '10'
	}
	buf = append(buf, byte(headerDataLen))
	buf = appendPTS(buf, p.PTS, p.HasDTS)
	if p.HasDTS {
		buf = appendDTS(buf, p.DTS)
	}
	return append(buf, p.Data...)
}

func appendPTS(dst []byte, pts uint64, withDTS bool) []byte {
	prefix := byte(0x20) // '0010'
	if withDTS {
		prefix = 0x30 // '0011'
	}
	return appendTimestamp(dst, prefix, pts)
}

func appendDTS(dst []byte, dts uint64) []byte {
	return appendTimestamp(dst, 0x10, dts) // '0001'
}

func appendTimestamp(dst []byte, prefix byte, ts uint64) []byte {
	ts &= 0x1ffffffff
	var b [5]byte
	b[0] = prefix | byte(((ts>>30)&0x07)<<1) | 0x01
	b[1] = byte(ts >> 22)
	b[2] = byte((ts>>14)&0xfe) | 0x01
	b[3] = byte(ts >> 7)
	b[4] = byte((ts<<1)&0xfe) | 0x01
	return append(dst, b[:]...)
}

func parseTimestamp(b []byte) uint64 {
	return uint64((b[0]>>1)&0x07)<<30 |
		uint64(b[1])<<22 |
		uint64(b[2]>>1)<<15 |
		uint64(b[3])<<7 |
		uint64(b[4]>>1)
}

// Clock90 converts a millisecond timestamp to a 90 kHz clock.
func Clock90(ms uint32, cts int32) uint64 {
	v := int64(ms) + int64(cts)
	if v < 0 {
		v = 0
	}
	return (uint64(v) * 90) & 0x1ffffffff
}
