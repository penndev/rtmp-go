package adapter

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/penndev/rtmp/amf"
	"github.com/penndev/rtmp/codec/flv"
	"github.com/penndev/rtmp/rtmp/stream"
)

func AdapterFlv(name string, sub stream.Subscriber) {
	defer sub.Close()
	dir := fmt.Sprintf("runtime/%s", name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Println(err)
		return
	}
	file, err := os.OpenFile(fmt.Sprintf("%s/%s.flv", dir, name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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

func WriteFlv(w http.ResponseWriter, r *http.Request, sub stream.Subscriber) {
	defer sub.Close()

	w.Header().Set("Content-Type", "video/x-flv")
	w.Header().Set("Cache-Control", "no-cache")
	flusher, _ := w.(http.Flusher)

	fw, err := flv.NewFlv(w, flv.TYPE_FLAGS_AUDIO_VIDEO)
	if err != nil {
		log.Println(err)
		return
	}
	if flusher != nil {
		flusher.Flush()
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case m, ok := <-sub.Chan():
			if !ok {
				return
			}
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
				} else if len(values) == 3 {
					if key, ok := values[1].(string); ok && key == "onMetaData" {
						encoded, err := amf.Encode0("onMetaData", values[2])
						if err != nil {
							log.Println("amf encode:", err)
						} else {
							fdata = encoded
						}
					}
				}
			}
			if err := fw.TagWrite(ftype, msg.Timestamp(), fdata); err != nil {
				log.Println("flv write:", err)
				return
			}
			if flusher != nil {
				flusher.Flush()
			}
		}
	}
}
