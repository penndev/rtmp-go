package av

import (
	"encoding/binary"
	"log"
	"os"
)

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
var aac = false

//AddTag FLV 生成写入tag
func (f *FLV) AddTag(tagType byte, timeStreamp uint32, tagData []byte) {
	f.Timestamp += timeStreamp
	var tag Tag
	tag.tagType = tagType
	tag.timeStreamp = f.Timestamp
	tag.timeStreampExtended = 0
	tag.tagData = tagData
	f.File.Write(tag.genByte())

	if tagType == 8 {
		if !aac {
			aac = true
			return
		}

		adtsh := []byte{0xff, 0xf1, 0x4c, 0x80}
		f.testaac.Write(adtsh)
		tmp := make([]byte, 2)
		tmpnull := uint16(len(tagData) + 5)
		tmpnull = tmpnull << 5
		tmpnull = tmpnull | 0x1f
		binary.BigEndian.PutUint16(tmp, tmpnull)
		f.testaac.Write(tmp)

		f.testaac.Write([]byte{0xfc})

		adts := tagData[2:]
		f.testaac.Write(adts)

	}

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
	f.File, err = os.OpenFile("runtime/"+name+".flv", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
	}
	f.testh264, err = os.OpenFile("runtime/"+name+".h264", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
	}

	f.testaac, err = os.OpenFile("runtime/"+name+".aac", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
	}

	f.File.Write(f.genHead("av"))
	return nil
}
