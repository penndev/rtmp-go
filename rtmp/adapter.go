package rtmp

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/penndev/rtmp/codec/flv"
	"github.com/penndev/rtmp/pubsub"
)

// AdapterFlv subscribes to topic and writes runtime/<path>.flv until the
// subscription is closed.
func AdapterFlv(path string, sub *pubsub.Subscription) {
	defer sub.Close()

	u, err := url.Parse(path)
	if err != nil {
		log.Println(err)
		return
	}
	name := strings.ReplaceAll(strings.Trim(u.Path, "/"), "/", "-")
	file, err := os.OpenFile(fmt.Sprintf("runtime/%s.flv", name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	w, err := flv.NewFlv(file, flv.TYPE_FLAGS_AUDIO_VIDEO)
	if err != nil {
		log.Println(err)
		return
	}
	for m := range sub.Chan() {
		msg, ok := m.Data.(*Message)
		if !ok || msg == nil {
			continue
		}
		if err := w.TagWrite(msg.Type(), msg.Timestamp(), msg.Data()); err != nil {
			log.Println("flv write:", err)
			return
		}
	}
}
