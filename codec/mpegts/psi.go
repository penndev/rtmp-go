package mpegts

import "encoding/binary"

const (
	PIDPAT = 0x0000
	PIDPMT = 0x1000

	PIDVideo = 0x0100
	PIDAudio = 0x0101

	// ISO/IEC 13818-1 Table 2-34 — stream_type
	StreamTypeAAC  = 0x0f // ISO/IEC 13818-7 Audio with ADTS
	StreamTypeH264 = 0x1b // AVC
	StreamTypeH265 = 0x24 // HEVC

	tableIDPAT = 0x00
	tableIDPMT = 0x02
)

// PAT is ISO/IEC 13818-1 program_association_section, one program.
type PAT struct {
	TransportStreamID uint16
	ProgramNumber     uint16
	ProgramMapPID     uint16
}

// ElementaryStream is one entry in a program map.
type ElementaryStream struct {
	StreamType    byte
	ElementaryPID uint16
}

// PMT is ISO/IEC 13818-1 TS_program_map_section.
type PMT struct {
	ProgramNumber uint16
	PCRPID        uint16
	Streams       []ElementaryStream
}

func (p PAT) Marshal() []byte {
	body := []byte{
		byte(p.TransportStreamID >> 8), byte(p.TransportStreamID),
		0xc1, // reserved, version 0, current_next_indicator
		0x00, // section_number
		0x00, // last_section_number
		byte(p.ProgramNumber >> 8), byte(p.ProgramNumber),
		0xe0 | byte(p.ProgramMapPID>>8), byte(p.ProgramMapPID),
	}
	return section(tableIDPAT, body)
}

func (p PMT) Marshal() []byte {
	body := []byte{
		byte(p.ProgramNumber >> 8), byte(p.ProgramNumber),
		0xc1,
		0x00,
		0x00,
		0xe0 | byte(p.PCRPID>>8), byte(p.PCRPID),
		0xf0, 0x00, // program_info_length
	}
	for _, es := range p.Streams {
		body = append(body,
			es.StreamType,
			0xe0|byte(es.ElementaryPID>>8), byte(es.ElementaryPID),
			0xf0, 0x00, // ES_info_length
		)
	}
	return section(tableIDPMT, body)
}

// section builds table_id + section_length + body + CRC_32.
func section(tableID byte, body []byte) []byte {
	length := len(body) + 4
	sec := make([]byte, 0, 3+len(body)+4)
	sec = append(sec, tableID, 0xb0|byte(length>>8), byte(length))
	sec = append(sec, body...)
	crc := mpegCRC32(sec)
	var sum [4]byte
	binary.BigEndian.PutUint32(sum[:], crc)
	return append(sec, sum[:]...)
}

// mpegCRC32 is the MPEG-2 CRC in ISO/IEC 13818-1 Annex A.
// Polynomial 0x04C11DB7, init 0xFFFFFFFF, not reflected, xorout 0.
func mpegCRC32(data []byte) uint32 {
	crc := uint32(0xffffffff)
	for _, b := range data {
		crc ^= uint32(b) << 24
		for i := 0; i < 8; i++ {
			if crc&0x80000000 != 0 {
				crc = (crc << 1) ^ 0x04c11db7
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
