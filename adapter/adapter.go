package adapter

import (
	"fmt"
	"log"
	"os"

	"github.com/penndev/rtmp/amf"
	"github.com/penndev/rtmp/codec/flv"
	"github.com/penndev/rtmp/rtmp/stream"
)

func AdapterFlv(path string, sub stream.Subscriber) {
	defer sub.Close()
	file, err := os.OpenFile(fmt.Sprintf("runtime/%s.flv", path), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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
		msg, ok := m.(flv.TagReader)
		if !ok || msg == nil {
			continue
		}
		ftype := msg.Type()
		fdata := msg.Data()
		if ftype == flv.TAG_TYPE_SCRIPT_DATA {
			values, err := amf.Decode0(fdata)
			if err != nil {
				log.Println("amf decode:", err)
			}
			if len(values) != 3 || values[1].(string) != "onMetaData" {
				log.Println("onMetaData not found", len(values), values)
			} else {
				fdata, err = amf.Encode0("onMetaData", values[2])
				if err != nil {
					log.Println("amf encode:", err)
				}
			}
		}

		if err := w.TagWrite(ftype, msg.Timestamp(), fdata); err != nil {
			log.Println("flv write:", err)
			return
		}

	}
}
