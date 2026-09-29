package flv

import (
	"log"
	"net/url"
	"os"

	"github.com/penndev/rtmp/pubsub"
)

// 订阅里 Data 需要能取出 type / 绝对时间戳 / payload（*rtmp.Message 满足）
type tagMsg interface {
	TagType() byte
	TagTimestamp() uint32
	TagData() []byte
}

// AdapterFlv 订阅 topic，写成 runtime/<path>.flv；sub 关闭或 topic 结束后返回
func AdapterFlv(path string, sub *pubsub.Subscription) {
	defer sub.Close()

	if err := os.MkdirAll("runtime", 0755); err != nil {
		log.Println(err)
		return
	}
	f, err := os.OpenFile("runtime/"+url.QueryEscape(path)+".flv", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	w := NewFlv(f)
	for m := range sub.Chan() {
		msg, ok := m.Data.(tagMsg)
		if !ok || msg == nil {
			continue
		}
		w.TagWrite(msg.TagType(), msg.TagTimestamp(), msg.TagData())
	}
}
