package main

import (
	"io"
	"log"
	"net/http"

	"github.com/pennfly/rtmp-go/rtmp"
)

func main() {
	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, world!\n")
	}
	http.HandleFunc("/hello", helloHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))

	rtmp.Serve()
}
