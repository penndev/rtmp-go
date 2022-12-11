package mpegts

import (
	"encoding/binary"
	"io"
	"log"
)

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

type TsPack struct {
	VideoContinuty byte
	AudioContinuty byte
	DTS            uint32
	w              io.Writer
}

func (t *TsPack) toHead(adapta, mixed bool, mtype byte) []byte {
	tsHead := make([]byte, 4)
	tsHead[0] = 0x47
	if adapta {
		tsHead[1] |= 0x40
	}
	if mtype == VideoMark {
		tsHead[1] |= 1
		tsHead[2] |= 0
		tsHead[3] |= t.VideoContinuty
		t.VideoContinuty = (t.VideoContinuty + 1) % 16
		// log.Println(t.VideoContinuty)
	} else if mtype == AudioMark {
		tsHead[1] |= 1
		tsHead[2] |= 1
		tsHead[3] |= t.AudioContinuty
		t.AudioContinuty = (t.AudioContinuty + 1) % 16
	}
	if adapta || mixed {
		tsHead[3] |= 0x30
	} else {
		tsHead[3] |= 0x10
	}
	return tsHead
}

func (t *TsPack) toPack(mtype byte, pes []byte) {
	adapta := true
	mixed := false
	for {
		pesLen := len(pes)
		if pesLen <= 0 {
			break
		}
		if pesLen < 184 {
			mixed = true
		}
		cPack := make([]byte, 188)
		for i := range cPack {
			cPack[i] = 0xff
		}
		copy(cPack[0:4], t.toHead(adapta, mixed, mtype))
		if mixed {
			fillLen := 183 - pesLen
			cPack[4] = byte(fillLen)
			if fillLen > 0 {
				cPack[5] = 0
			}
			copy(cPack[fillLen+5:188], pes[:pesLen])
			pes = pes[pesLen:]
		} else if adapta {
			// 获取pcr
			cPack[4] = 7
			copy(cPack[5:12], hexPcr(t.DTS*uint32(defaultH264HZ)))
			copy(cPack[12:188], pes[0:176])
			pes = pes[176:]
		} else {
			copy(cPack[4:188], pes[0:184])
			pes = pes[184:]
		}
		adapta = false
		t.w.Write(cPack)
	}
}

func (t *TsPack) videoTag(tagData []byte) {
	codecID := tagData[0] & 0x0f
	if codecID != 7 {
		log.Println("遇到了不是h264的视频数据", codecID)
	}
	compositionTime := binary.BigEndian.Uint32([]byte{0, tagData[2], tagData[3], tagData[4]})
	nalu := []byte{}
	if tagData[1] == 0 { //avc IDR frame | flv sps pps
		spsLen := int(binary.BigEndian.Uint16(tagData[11:13]))
		sps := tagData[13 : 13+spsLen]
		spsnalu := append([]byte{0, 0, 0, 1}, sps...)
		nalu = append(nalu, spsnalu...)
		ppsLen := int(binary.BigEndian.Uint16(tagData[14+spsLen : 16+spsLen]))
		pps := tagData[16+spsLen : 16+spsLen+ppsLen]
		ppsnalu := append([]byte{0, 0, 0, 1}, pps...)
		nalu = append(nalu, ppsnalu...)
	} else if tagData[1] == 1 { //avc nalu
		readed := 5
		for len(tagData) > (readed + 5) {
			readleng := int(binary.BigEndian.Uint32(tagData[readed : readed+4]))
			readed += 4
			nalu = append(nalu, []byte{0, 0, 0, 1}...)
			nalu = append(nalu, tagData[readed:readed+readleng]...)
			readed += readleng
		}
	} //else panic
	dts := t.DTS * uint32(defaultH264HZ)
	pts := dts + compositionTime*uint32(defaultH264HZ)
	pes := PES(VideoMark, pts, dts)
	t.toPack(VideoMark, append(pes, nalu...))
}

func (t *TsPack) audioTag(tagData []byte) {
	soundFormat := (tagData[0] & 0xf0) >> 4
	if soundFormat != 10 {
		log.Println("遇到了不是aac的音频数据")
	}
	if tagData[1] == 1 {
		tagData = tagData[2:]
		adtsHeader := []byte{0xff, 0xf1, 0x4c, 0x80, 0x00, 0x00, 0xfc}
		adtsLen := uint16(((len(tagData) + 7) << 5) | 0x1f)
		binary.BigEndian.PutUint16(adtsHeader[4:6], adtsLen)
		adts := append(adtsHeader, tagData...)
		pts := t.DTS * uint32(defaultH264HZ)
		pes := PES(AudioMark, pts, 0)
		t.toPack(AudioMark, append(pes, adts...))
	}

}

func (t *TsPack) FlvTag(tagType byte, timeStreamp uint32, timeStreampExtended byte, tagData []byte) {
	// 组合视频戳
	dts := uint32(timeStreampExtended)*16777216 + timeStreamp
	t.DTS += dts
	// 判断是视频还是音频
	if tagType == 9 {
		t.videoTag(tagData)
	} else if tagType == 8 {
		t.audioTag(tagData)
	}
}
