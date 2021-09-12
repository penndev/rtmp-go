package av

import (
	"encoding/binary"
	"log"
	"os"
)

const (
	Signature string = "FLV"
	Version   byte   = 1
)

// Flags  掩码位判断视频包含的内容。
// "00000001" 1-只有视频
// "00000100" 4-只有音频
// "00000101" 5-有视频有音频
var Flags = map[string]byte{"v": 1, "a": 4, "av": 5}

var DataOffset = []byte{0, 0, 0, 9}

//Tag 包含的详情
type Tag struct {
	tagType             byte   //1 byte;8audio,9video,18scripts;
	timeStreamp         uint32 //3 byte;时间戳
	timeStreampExtended byte   //1 byte;
	tagData             []byte
}

// Tag genByte
// 根据tag内容
// 生成byte字节码
func (t Tag) genByte() []byte {
	dataLen := len(t.tagData)
	tag := make([]byte, 11+dataLen+4)
	// tagDataSize 3[]byte
	binary.BigEndian.PutUint32(tag[:4], uint32(dataLen))
	// type 1 byte
	tag[0] = t.tagType
	// timeStreamp 3[]byte
	binary.BigEndian.PutUint32(tag[4:8], t.timeStreamp)
	copy(tag[4:7], tag[5:8])
	// tagTimestampExtend 1 byte
	tag[7] = t.timeStreampExtended
	// 9 10 11 tagSteamID ｜ 8 - 9 - 10 | put 0
	// put tagdata
	copy(tag[11:], t.tagData)
	binary.BigEndian.PutUint32(tag[11+dataLen:], uint32(dataLen+11))
	return tag
}

//FLV flv的数据结构
type FLV struct {
	File            *os.File
	PreviousTagSize uint32
	Timestamp       uint32

	testh264 *os.File
}

//GenHeader FLV 生成文件头。
func (f *FLV) genHead(flags string) []byte {
	hd := make([]byte, 13)
	copy(hd, []byte(Signature))
	hd[3] = Version
	hd[4] = Flags[flags]
	copy(hd[5:9], DataOffset)
	// Last PreviousTagSize had put 4 byte zero
	return hd
}

var sps = false

//AddTag FLV 生成写入tag
func (f *FLV) AddTag(tagType byte, timeStreamp uint32, tagData []byte) {
	f.Timestamp += timeStreamp
	var tag Tag
	tag.tagType = tagType
	tag.timeStreamp = f.Timestamp
	tag.timeStreampExtended = 0
	tag.tagData = tagData
	f.File.Write(tag.genByte())
	if tagType == 9 {
		if !sps {
			sps = true
			return
		}
		// f.testh264.Write()
		allnual := tagData[5:]
		n := 0
		for {
			if n >= len(allnual) {
				break
			}
			ulen := binary.BigEndian.Uint32(allnual[n : n+4])
			n += 4
			f.testh264.Write([]byte{0, 0, 0, 1})
			len := int(ulen)
			f.testh264.Write(allnual[n : n+len])
			n += len
			// 去读真实的nalu数据
		}

	}
}

//Close 管理flv连接
func (f *FLV) Close() {
	f.File.Close()
}

//GenFlv 生成新的flv
func (f *FLV) GenFlv(name string) error {
	var err error
	f.File, err = os.OpenFile(name+".flv", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
	}
	f.testh264, err = os.OpenFile(name+".h264", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
	}

	f.File.Write(f.genHead("av"))
	return nil
}
