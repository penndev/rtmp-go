package flv

import (
	"log"
	"os"
	"strings"

	"github.com/penndev/rtmp/pubsub"
)

// tagMsg is implemented by *rtmp.Message (avoids an import cycle).
type tagMsg interface {
	TagType() byte
	TagTimestamp() uint32
	TagData() []byte
}

// AdapterFlv subscribes to topic and writes runtime/<path>.flv until the
// subscription is closed.
func AdapterFlv(path string, sub *pubsub.Subscription) {
	defer sub.Close()

	file, err := os.OpenFile("runtime/"+strings.ReplaceAll(path, "/", "-")+".flv", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	w, err := NewFlv(file, TYPE_FLAGS_AUDIO_VIDEO)
	if err != nil {
		log.Println(err)
		return
	}
	for m := range sub.Chan() {
		msg, ok := m.Data.(tagMsg)
		if !ok || msg == nil {
			continue
		}
		if err := w.TagWrite(msg.TagType(), msg.TagTimestamp(), msg.TagData()); err != nil {
			log.Println("flv write:", err)
			return
		}
	}
}
