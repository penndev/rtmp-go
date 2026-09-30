// Remux runtime/app%2Fstream.flv -> runtime/app%2Fstream.out.flv (streaming).
//
//	go run ./flv/cmd
package main

import (
	"log"
	"os"

	"github.com/penndev/rtmp/flv"
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
		return func(tag *flv.Tag) error {
			n++
			return dst.WriteTag(tag)
		}
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("remux %s -> %s tags=%d", inPath, outPath, n)
}
