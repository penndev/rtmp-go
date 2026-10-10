package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/penndev/rtmp/adapter"
	"github.com/penndev/rtmp/pubsub"
	"github.com/penndev/rtmp/rtmp"
	"github.com/penndev/rtmp/rtmp/handler"
	"github.com/penndev/rtmp/rtmp/stream"
)

func main() {
	rtmpAddr := flag.String("rtmp", "127.0.0.1:1935", "RTMP listen address")
	httpAddr := flag.String("http", "127.0.0.1:8080", "HTTP listen address")
	user := flag.String("user", "", "publish username, empty uses the default handler")
	password := flag.String("password", "", "publish password")
	flag.Parse()

	// rtmp
	hub := pubsub.NewHubStream()
	var h handler.Handler = handler.NewDefaultHandler()
	if *user != "" {
		h = handler.NewAuthHandler(*user, *password)
	}
	rtmpSrv := rtmp.New(h, hub)
	hub.AfterPublish = func(name string, sub stream.Subscriber) {
		go adapter.AdapterFlv(name, sub)
		// hls ts generation
		tsSub, err := hub.Play(name)
		if err != nil {
			log.Println(err)
			return
		}
		go adapter.AdapterTs(name, tsSub)
	}

	// http
	httpSrv := adapter.NewHttp()
	httpSrv.OnNames(func() []string {
		return hub.Names()
	})
	httpSrv.OnFlv(func(w http.ResponseWriter, r *http.Request, name string) {
		sub, err := hub.Play(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		adapter.WriteFlv(w, r, sub)
	})
	httpSrv.OnM3u8(func(w http.ResponseWriter, r *http.Request, name string) {
		adapter.WriteM3u8(w, r, name)
	})

	go func() {
		log.Printf("http on http://%s", *httpAddr)
		log.Fatal(httpSrv.Listen(*httpAddr))
	}()

	log.Printf("rtmp on rtmp://%s", *rtmpAddr)
	log.Fatal(rtmpSrv.Listen(*rtmpAddr))
}
