package mpegts

import (
	"net/url"
	"strconv"

	"github.com/penndev/rtmp-go/rtmp"
)

var defaultH264HZ = 90

func Adapter(topic string, ch <-chan rtmp.Pack) {
	filename := "runtime/" + url.QueryEscape(topic) + ".ts"
	t := &TsPack{}
	t.NewTs(filename)
	defer delete(cache, topic)
	var timeMS uint32 // single tsFile sum(dts)
	for pk := range ch {
		// gen new ts file (dts 5*second)
		if timeMS > 5000 {
			var extInf = ExtInf{
				Inf:  timeMS,
				File: filename,
			}
			// file add the hls cache
			if v, ok := cache[topic]; ok {
				cache[topic] = append(v, extInf)
			} else {
				cache[topic] = []ExtInf{extInf}
			}
			filename = "runtime/" + url.QueryEscape(topic) + strconv.Itoa(int(t.DTS)) + ".ts"
			t.NewTs(filename)
			timeMS = 0
		}
		t.FlvTag(pk.MessageTypeID, pk.Timestamp, byte(pk.ExtendTimestamp), pk.PayLoad)
		timeMS += pk.Timestamp
	}
}
