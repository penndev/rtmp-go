package adapter

import (
	"fmt"
	"math"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
)

const liveSegments = 3
const liveItemDuration = 2000

type tsSegment struct {
	path     string
	duration float64
}

type hlsLive struct {
	segments []tsSegment
	endList  bool
}

func (h *hlsLive) push(seg tsSegment) {
	h.segments = append(h.segments, seg)
	if len(h.segments) > liveSegments {
		h.segments = append([]tsSegment(nil), h.segments[len(h.segments)-liveSegments:]...)
	}
}

var (
	hlsPlayMu sync.Mutex
	hlsPlay   = map[string]*hlsLive{}
)

func WriteM3u8(w http.ResponseWriter, r *http.Request, name string) {
	hlsPlayMu.Lock()
	live := hlsPlay[name]
	var segs []tsSegment
	var endList bool
	if live != nil {
		segs = append([]tsSegment(nil), live.segments...)
		endList = live.endList
	}
	hlsPlayMu.Unlock()
	if len(segs) == 0 {
		http.NotFound(w, r)
		return
	}

	target := 1
	for _, s := range segs {
		sec := int(math.Ceil(s.duration))
		if sec > target {
			target = sec
		}
	}
	seq, _ := strconv.Atoi(strings.TrimSuffix(path.Base(segs[0].path), path.Ext(segs[0].path)))
	var b strings.Builder
	fmt.Fprintf(&b, "#EXTM3U\n")
	fmt.Fprintf(&b, "#EXT-X-VERSION:3\n")
	fmt.Fprintf(&b, "#EXT-X-TARGETDURATION:%d\n", target)
	fmt.Fprintf(&b, "#EXT-X-MEDIA-SEQUENCE:%d\n", seq)
	fmt.Fprintf(&b, "#EXT-X-INDEPENDENT-SEGMENTS\n")
	for _, s := range segs {
		fmt.Fprintf(&b, "#EXTINF:%.3f,\n%s\n", s.duration, s.path)
	}
	if endList {
		b.WriteString("#EXT-X-ENDLIST\n")
	}
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, _ = w.Write([]byte(b.String()))
}
