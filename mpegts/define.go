package mpegts

var VideoMark byte = 0xe0
var AudioMark byte = 0xc0

func hexPts(dpvalue uint32) []byte {
	dphex := make([]byte, 5)
	dphex[0] = 0x31 | byte(dpvalue>>29)
	hp := uint16((dpvalue>>15)&0x7fff)*2 + 1
	dphex[1] = byte(hp >> 8)
	dphex[2] = byte(hp & 0xff)
	he := (dpvalue&0x7fff)*2 + 1
	dphex[3] = byte(he >> 8)
	dphex[4] = byte(he & 0xff)
	return dphex
}

func hexDts(dpvalue uint32) []byte {
	dphex := make([]byte, 5)
	dphex[0] = 0x11 | byte(dpvalue>>29)
	hp := ((dpvalue>>15)&0x7fff)*2 + 1
	dphex[1] = byte(hp >> 8)
	dphex[2] = byte(hp & 0xff)
	he := (dpvalue&0x7fff)*2 + 1
	dphex[3] = byte(he >> 8)
	dphex[4] = byte(he & 0xff)
	return dphex
}

func hexPcr(dts uint32) []byte {
	adapt := make([]byte, 7)
	adapt[0] = 0x50
	adapt[1] = byte(dts >> 25)
	adapt[2] = byte(dts>>17) & 0xff
	adapt[3] = byte(dts>>9) & 0xff
	adapt[4] = byte(dts>>1) & 0xff
	adapt[5] = byte((dts&0x1)<<7) | 0x7e
	return adapt
}

func SDT() []byte {
	bt := make([]byte, 188)
	for i := range bt {
		bt[i] = 0xff
	}
	copy(bt[0:45], []byte{
		0x47, 0x40, 0x11, 0x10,
		0x00, 0x42, 0xF0, 0x25, 0x00, 0x01, 0xC1, 0x00, 0x00, 0xFF,
		0x01, 0xFF, 0x00, 0x01, 0xFC, 0x80, 0x14, 0x48, 0x12, 0x01,
		0x06, 0x46, 0x46, 0x6D, 0x70, 0x65, 0x67, 0x09, 0x53, 0x65,
		0x72, 0x76, 0x69, 0x63, 0x65, 0x30, 0x31, 0x77, 0x7C, 0x43,
		0xCA})
	return bt
}

func PAT() []byte {
	bt := make([]byte, 188)
	for i := range bt {
		bt[i] = 0xff
	}
	copy(bt[0:21], []byte{
		0x47, 0x40, 0x00, 0x10,
		0x00,
		0x00, 0xB0, 0x0D, 0x00, 0x01, 0xC1, 0x00, 0x00, 0x00, 0x01,
		0xF0, 0x00, 0x2A, 0xB1, 0x04, 0xB2})
	return bt
}

func PMT() []byte {
	bt := make([]byte, 188)
	for i := range bt {
		bt[i] = 0xff
	}
	copy(bt[0:31], []byte{
		0x47, 0x50, 0x00, 0x10,
		0x00,
		0x02, 0xB0, 0x17, 0x00, 0x01, 0xC1, 0x00, 0x00, 0xE1, 0x00,
		0xF0, 0x00, 0x1B, 0xE1, 0x00, 0xF0, 0x00, 0x0F, 0xE1, 0x01,
		0xF0, 0x00, 0x2F, 0x44, 0xB9, 0x9B})
	return bt
}

// 首先使用nalu数据组合成es数据
// pes header https://dvd.sourceforge.net/dvdinfo/pes-hdr.html
func PES(mtype byte, pts uint32, dts uint32) []byte {
	header := make([]byte, 9)
	copy(header[0:3], []byte{0, 0, 1})
	header[3] = mtype
	header[6] = 0x80
	if pts > 0 {
		if dts > 0 {
			header[7] = 0xc0
			header[8] = 0x0a
			header = append(header, hexPts(pts)...)
			header = append(header, hexDts(dts)...)
		} else {
			header[7] = 0x80
			header[8] = 0x05
			header = append(header, hexPts(pts)...)
		}
	}
	return header
}
