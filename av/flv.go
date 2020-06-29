package av

import (
	"encoding/binary"
	"log"
	"os"
)

const Signature = "FLV"
const Version byte = 1

//Flags //"00000001" 1-只有视频 //"00000100" 4-只有音频  //"00000101" 5-有视频有音频
var Flags = map[string]byte{"v": 1, "a": 4, "av": 5}

var DataOffset = []byte{0, 0, 0, 9}

//Tag 包含的详情
type Tag struct {
	tagType             int    //1 byte;8audio,9video,18scripts;
	dataSize            int    //3 byte;data的长度
	timeStreamp         uint32 //3 byte;时间戳
	timeStreampExtended int    //1 byte;
	streamID            int    //3 byte default [000];
	tagData             []byte
}

//Tag genByte 描述tag参数，生成byte
func (t Tag) genByte() []byte {
	tmpV := make([]byte, 4)
	var tag []byte
	//首先写入 tagType 1[]byte
	tag = append(tag, byte(t.tagType))

	//接着写入 tagDataSize 3[]byte
	binary.BigEndian.PutUint32(tmpV, uint32(t.dataSize))
	tag = append(tag, tmpV[1:]...)

	//接着写入 tagTimestamp 3[]byte
	binary.BigEndian.PutUint32(tmpV, t.timeStreamp)
	tag = append(tag, tmpV[1:]...)

	//接着写入 tagTimestampExtend 1[]byte
	tag = append(tag, 0)
	//写入固定的 tagSteamID  3[]byte
	tag = append(tag, 0, 0, 0)
	//写入 tagData define‘s[]byte
	tag = append(tag, t.tagData...)
	//Last PreviousTagSize 4[]byte
	binary.BigEndian.PutUint32(tmpV, uint32(t.dataSize+11))
	tag = append(tag, tmpV...)
	return tag
}

//FLV flv的数据结构
type FLV struct {
	File            *os.File
	PreviousTagSize uint32
}

//GenHeader FLV 生成文件头。
func (f *FLV) genHead(flags string) {
	var header []byte
	header = append(header, []byte(Signature)...)  //头部固定值
	header = append(header, Version, Flags[flags]) //Version版本。Flags。
	header = append(header, []byte(DataOffset)...) //DataOffset
	header = append(header, 0, 0, 0, 0)            // Last PreviousTagSize
	f.File.Write(header)
}

//AddTag FLV 生成写入tag
func (f *FLV) AddTag(tagType int, timeStreamp uint32, tagData []byte) {

	//Tag Numb
	var tag Tag
	tag.tagType = tagType
	tag.dataSize = len(tagData)
	tag.timeStreamp = timeStreamp
	tag.timeStreampExtended = 0
	tag.tagData = tagData
	//Gen Byte
	//log.Println(tag.genByte())
	f.File.Write(tag.genByte())
}

//Close 管理flv连接
func (f *FLV) Close() {
	f.File.Close()
}

//GenFlv 生成新的flv
func (f *FLV) GenFlv(name string) error {
	// -
	var err error
	f.File, err = os.OpenFile(name+".flv", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
	}
	//
	f.PreviousTagSize = 0
	// -
	f.genHead("av")
	return nil
}
