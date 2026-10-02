// Remux runtime/app%2Fstream.flv -> runtime/app%2Fstream.out.flv (streaming).
//
//	go run ./flv/cmd
package main

import (
	"log"
	"os"

	"github.com/penndev/rtmp/codec/flv"
	"github.com/penndev/rtmp/codec/h264"
)

func main() {
	const inPath = "runtime/app-stream.flv"
	const outPath = "runtime/app-stream.out.flv"

	in, err := os.Open(inPath)
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()

	out, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	n := 0
	err = flv.Reader(in, func(f *flv.FLV) func(*flv.Tag) error {
		// header ready on f — open destination with same TypeFlags
		dst, err := flv.NewFlv(out, f.TypeFlags)
		if err != nil {
			log.Fatal(err)
		}

		var avc *h264.AVCDecoderConfigurationRecord
		return func(tag *flv.Tag) error {
			n++
			if n < 50 {
				ftag, err := flv.FormatTag(tag.TagType, tag.Timestamp, tag.Data, avc)
				if err != nil {
					log.Println("format tag:", err)
					return dst.WriteTag(tag)
				}
				if ftag.TagType == flv.TAG_TYPE_VIDEO &&
					ftag.CodecID == flv.CODEC_ID_AVC &&
					ftag.AVCPacketType == flv.AVC_PACKET_TYPE_SEQUENCE_HEADER {
					log.Printf("%+v\n", ftag.AVCDecoderConfigurationRecord)
					marshal, err := ftag.AVCDecoderConfigurationRecord.Marshal()
					if err != nil {
						log.Println("marshal:", err)
						return dst.WriteTag(tag)
					}
					log.Printf("% X\n", marshal)
					log.Printf("% X\n", tag.Data)
					avc = &ftag.AVCDecoderConfigurationRecord
				}
				if ftag.AVCPacketType == flv.AVC_PACKET_TYPE_NALU {
					types := make([]h264.NalUnitType, len(ftag.Nalu))
					for i, n := range ftag.Nalu {
						types[i] = n.NalUnitType
					}
					log.Printf("nalus=%v\n", types)
				}
			}
			return dst.WriteTag(tag)
		}
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("remux %s -> %s tags=%d", inPath, outPath, n)
}
