package mpegts

import "errors"

// ISO/IEC 13818-1 Table 2-2 — Transport Stream packet. One packet is 188 bytes.
const PacketSize = 188

const (
	SyncByte = 0x47

	// adaptation_field_control
	AFCPayload           = 0x01
	AFCAdaptation        = 0x02
	AFCAdaptationPayload = 0x03
)

// Packet is one transport packet.
type Packet struct {
	SyncByte byte

	TransportErrorIndicator    bool
	PayloadUnitStartIndicator  bool
	TransportPriority          bool
	PID                        uint16 // 13 bits
	TransportScramblingControl byte   // 2 bits
	AdaptationFieldControl     byte   // 2 bits
	ContinuityCounter          byte   // 4 bits

	AdaptationField *AdaptationField
	Payload         []byte
}

// AdaptationField is ISO/IEC 13818-1 Table 2-6.
// Length is adaptation_field_length: the number of bytes following it.
type AdaptationField struct {
	Length byte

	DiscontinuityIndicator            bool
	RandomAccessIndicator             bool
	ElementaryStreamPriorityIndicator bool
	PCRFlag                           bool
	// PCR is program_clock_reference_base in units of 90 kHz. The 27 MHz extension is 0.
	PCR uint64
}

func (p Packet) Marshal() ([]byte, error) {
	if p.SyncByte == 0 {
		p.SyncByte = SyncByte
	}
	afc := p.AdaptationFieldControl
	if afc == 0 {
		if p.AdaptationField != nil && len(p.Payload) > 0 {
			afc = AFCAdaptationPayload
		} else if p.AdaptationField != nil {
			afc = AFCAdaptation
		} else {
			afc = AFCPayload
		}
	}
	buf := make([]byte, 4, PacketSize)
	buf[0] = p.SyncByte
	if p.TransportErrorIndicator {
		buf[1] |= 0x80
	}
	if p.PayloadUnitStartIndicator {
		buf[1] |= 0x40
	}
	if p.TransportPriority {
		buf[1] |= 0x20
	}
	buf[1] |= byte((p.PID >> 8) & 0x1f)
	buf[2] = byte(p.PID)
	buf[3] = (afc&0x03)<<4 | (p.ContinuityCounter & 0x0f)
	if p.AdaptationField != nil {
		buf = p.AdaptationField.append(buf)
	}
	buf = append(buf, p.Payload...)
	if len(buf) != PacketSize {
		return nil, errors.New("mpegts: packet is not 188 bytes")
	}
	return buf, nil
}

func (a AdaptationField) append(dst []byte) []byte {
	dst = append(dst, a.Length)
	if a.Length == 0 {
		return dst
	}
	flags := byte(0)
	if a.DiscontinuityIndicator {
		flags |= 0x80
	}
	if a.RandomAccessIndicator {
		flags |= 0x40
	}
	if a.ElementaryStreamPriorityIndicator {
		flags |= 0x20
	}
	if a.PCRFlag {
		flags |= 0x10
	}
	dst = append(dst, flags)
	used := 1
	if a.PCRFlag {
		dst = appendPCR(dst, a.PCR)
		used += 6
	}
	for used < int(a.Length) {
		dst = append(dst, 0xff)
		used++
	}
	return dst
}

// program_clock_reference is 33 bits of base plus a 9-bit extension.
func appendPCR(dst []byte, base uint64) []byte {
	base &= 0x1ffffffff
	return append(dst,
		byte(base>>25),
		byte(base>>17),
		byte(base>>9),
		byte(base>>1),
		byte((base&1)<<7)|0x7e,
		0,
	)
}

func Parse(b []byte) (Packet, error) {
	if len(b) != PacketSize {
		return Packet{}, errors.New("mpegts: packet is not 188 bytes")
	}
	if b[0] != SyncByte {
		return Packet{}, errors.New("mpegts: bad sync byte")
	}
	p := Packet{
		SyncByte:                   b[0],
		TransportErrorIndicator:    b[1]&0x80 != 0,
		PayloadUnitStartIndicator:  b[1]&0x40 != 0,
		TransportPriority:          b[1]&0x20 != 0,
		PID:                        uint16(b[1]&0x1f)<<8 | uint16(b[2]),
		TransportScramblingControl: (b[3] >> 6) & 0x03,
		AdaptationFieldControl:     (b[3] >> 4) & 0x03,
		ContinuityCounter:          b[3] & 0x0f,
	}
	off := 4
	if p.AdaptationFieldControl == AFCAdaptation || p.AdaptationFieldControl == AFCAdaptationPayload {
		if off >= len(b) {
			return Packet{}, errors.New("mpegts: truncated adaptation field")
		}
		af, n, err := parseAdaptation(b[off:])
		if err != nil {
			return Packet{}, err
		}
		p.AdaptationField = &af
		off += n
	}
	if p.AdaptationFieldControl == AFCPayload || p.AdaptationFieldControl == AFCAdaptationPayload {
		p.Payload = append([]byte(nil), b[off:]...)
	}
	return p, nil
}

func parseAdaptation(b []byte) (AdaptationField, int, error) {
	length := int(b[0])
	af := AdaptationField{Length: b[0]}
	if length == 0 {
		return af, 1, nil
	}
	if len(b) < 1+length {
		return AdaptationField{}, 0, errors.New("mpegts: truncated adaptation field")
	}
	flags := b[1]
	af.DiscontinuityIndicator = flags&0x80 != 0
	af.RandomAccessIndicator = flags&0x40 != 0
	af.ElementaryStreamPriorityIndicator = flags&0x20 != 0
	af.PCRFlag = flags&0x10 != 0
	if af.PCRFlag {
		if length < 7 {
			return AdaptationField{}, 0, errors.New("mpegts: truncated PCR")
		}
		af.PCR = uint64(b[2])<<25 | uint64(b[3])<<17 | uint64(b[4])<<9 | uint64(b[5])<<1 | uint64(b[6]>>7)
	}
	return af, 1 + length, nil
}
