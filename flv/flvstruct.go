package flv

import (
	"encoding/binary"
	"io"
)

//FLV header
const (
	// 固定文件头 0x46 0x4c 0x56
	Signature string = "FLV"
	// FLV 版本，固定为1
	Version byte = 1
)

//FLV Flags  掩码位判断视频包含的内容。
// "00000001" 1-只有视频
// "00000100" 4-只有音频
// "00000101" 5-有视频有音频
var Flags = map[string]byte{"v": 1, "a": 4, "av": 5}

// 固定偏移位置
var DataOffset = []byte{0, 0, 0, 9}

//FLV 结构体
type FLV struct {
	// 上一个FLV Tag size
	PreviousTagSize uint32
	// FLV存储时间为累加时间
	Timestamp uint32

	w io.Writer
}

// FLV Tag 结构体
type Tag struct {
	tagType             byte   //1 byte;8audio,9video,18scripts;
	timeStreamp         uint32 //3 byte;时间戳
	timeStreampExtended byte   //1 byte;
	tagData             []byte
}

func NewFlv(w io.Writer) *FLV {
	flv := &FLV{
		w: w,
	}
	// 写FLV文件头
	flv.w.Write(NewHeader("av"))
	return flv
}

// 生成FLV header
func NewHeader(flags string) []byte {
	hd := make([]byte, 13)
	copy(hd, []byte(Signature))
	hd[3] = Version
	hd[4] = Flags[flags]
	copy(hd[5:9], DataOffset)
	return hd
}

// FLV 写入Tag Byte
func (flv *FLV) TagWrite(tagType byte, timeStreamp uint32, timeStreampExtended byte, tagData []byte) {
	var tag Tag
	tag.tagType = tagType
	flv.Timestamp += timeStreamp
	tag.timeStreamp = flv.Timestamp
	// 写真实的拓展时间 - 还未实现
	tag.timeStreampExtended = timeStreampExtended
	tag.tagData = tagData
	flv.w.Write(tag.genByte())
}

// Tag genByte  根据tag内容 生成byte字节码
func (t Tag) genByte() []byte {
	dataLen := len(t.tagData)
	// tagDataSize | 1,2,3 ] 没有uint 24 差评
	tag := make([]byte, 11+dataLen+4)
	binary.BigEndian.PutUint32(tag[:4], uint32(dataLen))
	// tagType | 0 ] 后写因为上一步被覆盖
	tag[0] = t.tagType
	// tagTimestreamp | 4,5,6 ]
	binary.BigEndian.PutUint32(tag[4:8], t.timeStreamp)
	copy(tag[4:7], tag[5:8])
	// tagTimestampExtend | 7 ]
	tag[7] = t.timeStreampExtended
	//SteamID always put zero ｜ 8,9,10 ]
	copy(tag[11:], t.tagData)
	binary.BigEndian.PutUint32(tag[11+dataLen:], uint32(dataLen+11))
	return tag
}
