package av

import (
	"log"
	"net/url"
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

//AddTag FLV 生成写入tag
func (f *FLV) AddTag(tagType byte, timeStreamp uint32, tagData []byte) {
	f.Timestamp += timeStreamp
	var tag Tag
	tag.tagType = tagType
	tag.timeStreamp = f.Timestamp
	tag.timeStreampExtended = 0
	tag.tagData = tagData
	f.File.Write(tag.genByte())
}

//Close 管理flv连接
func (f *FLV) Close() {
	f.File.Close()
}

//GenFlv 生成新的flv
func (f *FLV) GenFlv(name string) error {
	var err error
	f.File, err = os.OpenFile("runtime/"+url.QueryEscape(name)+".flv", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
	}
	f.File.Write(f.genHead("av"))
	return nil
}
