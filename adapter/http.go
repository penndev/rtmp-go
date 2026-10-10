package adapter

import (
	_ "embed"
	"html/template"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"
)

//go:embed index.html
var indexHTML string

func init() {
	mime.AddExtensionType(".m3u8", "application/vnd.apple.mpegurl")
	mime.AddExtensionType(".ts", "video/MP2T")
}

type playFunc func(w http.ResponseWriter, r *http.Request, name string)

type Http struct {
	onNames func() []string
	onFlv   playFunc
	onM3u8  playFunc
	onMpd   playFunc
}

func NewHttp() *Http {
	return &Http{}
}

func (s *Http) OnNames(fn func() []string) {
	s.onNames = fn
}

func (s *Http) OnFlv(fn playFunc) {
	s.onFlv = fn
}

func (s *Http) OnM3u8(fn playFunc) {
	s.onM3u8 = fn
}

func (s *Http) OnMpd(fn playFunc) {
	s.onMpd = fn
}

func streamName(p string) string {
	name := strings.TrimSuffix(path.Base(path.Clean(p)), path.Ext(p))
	if name == "" || name == "." || name == ".." {
		return ""
	}
	return name
}

func (s *Http) Listen(addr string) error {
	return http.ListenAndServe(addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := streamName(r.URL.Path)
		var handle playFunc
		switch {
		case r.URL.Path == "/":
			if s.onNames == nil {
				http.NotFound(w, r)
				return
			}
			t, err := template.New("index").Parse(indexHTML)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-cache")
			names := s.onNames()
			items := make([]struct {
				Name string
				HLS  bool
			}, len(names))
			hlsPlayMu.Lock()
			for i, name := range names {
				live := hlsPlay[name]
				items[i].Name = name
				items[i].HLS = live != nil && len(live.segments) > 0
			}
			hlsPlayMu.Unlock()
			if err := t.Execute(w, items); err != nil {
				log.Println(err)
			}
			return
		case strings.HasSuffix(r.URL.Path, ".m3u8"):
			if s.onM3u8 == nil {
				http.Error(w, "not implemented", http.StatusNotImplemented)
				return
			}
			handle = s.onM3u8
		case strings.HasSuffix(r.URL.Path, ".flv"):
			if s.onFlv == nil {
				http.NotFound(w, r)
				return
			}
			handle = s.onFlv
		case strings.HasSuffix(r.URL.Path, ".mpd"):
			if s.onMpd == nil {
				http.Error(w, "not implemented", http.StatusNotImplemented)
				return
			}
			handle = s.onMpd
		default:
			if strings.HasPrefix(r.URL.Path, "/runtime/") {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.StripPrefix("/runtime/", http.FileServer(http.Dir("runtime"))).ServeHTTP(w, r)
				return
			}
			http.NotFound(w, r)
			return
		}
		if name == "" {
			http.NotFound(w, r)
			return
		}
		handle(w, r, name)
	}))
}
