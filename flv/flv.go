package flv

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/penndev/rtmp-go/rtmp"
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

func Adapterflv(name string, ch <-chan rtmp.Pack) {
	var flv FLV
	flv.GenFlv(name)
	for pk := range ch {
		flv.AddTag(pk.MessageTypeID, pk.Timestamp, pk.PayLoad)
	}
	flv.Close()
}

func Handleflv(subtop func(string) (*rtmp.PubSub, bool)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		param := r.URL.Query()
		topic := param.Get("topic")
		log.Println(topic)
		if subscriber, ok := subtop(topic); ok {
			flv := NewFlv(w)
			ch := subscriber.Subscription()
			defer subscriber.SubscriptionExit(ch)
			for pk := range ch {
				flv.TagWrite(pk.MessageTypeID, pk.Timestamp, byte(pk.ExtendTimestamp), pk.PayLoad)
			}
		} else {
			http.NotFound(w, r)
		}
	}
}
