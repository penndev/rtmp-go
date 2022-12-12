package mpegts

import (
	"encoding/binary"
	"io"
	"log"
	"os"
)

type TsPack struct {
	VideoContinuty byte
	AudioContinuty byte
	DTS            uint32
	IDR            []byte
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
		t.IDR = tagData
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

	if tagType == 9 {
		t.videoTag(tagData)
	} else if tagType == 8 {
		t.audioTag(tagData)
	}
}

func (t *TsPack) NewTs(filename string) {
	var err error
	if t.w, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644); err != nil {
		log.Println(err)
	}
	t.w.Write(SDT())
	t.w.Write(PAT())
	t.w.Write(PMT())
	if len(t.IDR) > 0 {
		t.videoTag(t.IDR)
	}

}
