package flv

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/penndev/rtmp-go/rtmp"
)

func AdapterFlv(name string, ch <-chan rtmp.Pack) {
	flvFile, err := os.OpenFile("runtime/"+url.QueryEscape(name)+".flv", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
	}
	defer flvFile.Close()
	flv := NewFlv(flvFile)
	for pk := range ch {
		flv.TagWrite(pk.MessageTypeID, pk.Timestamp, byte(pk.ExtendTimestamp), pk.PayLoad)
	}
}

func HandleFlv(subTopic func(string) (*rtmp.PubSub, bool)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		param := r.URL.Query()
		topic := param.Get("topic")
		if subscriber, ok := subTopic(topic); ok {
			flv := NewFlv(w)
			ch := subscriber.Subscription()
			defer subscriber.SubscriptionClose(ch)
			for pk := range ch {
				flv.TagWrite(pk.MessageTypeID, pk.Timestamp, byte(pk.ExtendTimestamp), pk.PayLoad)
			}
		} else {
			http.NotFound(w, r)
		}
	}
}
