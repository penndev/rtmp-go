package mpegts

import (
	"net/url"
	"strconv"

	"github.com/penndev/rtmp-go/rtmp"
)

var defaultH264HZ = 90

func Adapterts(topic string, ch <-chan rtmp.Pack) {
	filename := "runtime/" + url.QueryEscape(topic) + ".ts"

	t := &TsPack{}
	t.NewTs(filename)
	defer delete(cache, topic)
	var tslen uint32 // single tsfile sum(dts)
	for pk := range ch {
		// gen new ts file (dts 5*second)
		if tslen > 5000 {
			var extinf = ExtInf{
				Inf:  tslen,
				File: filename,
			}
			// file add the hls cache
			if v, ok := cache[topic]; ok {
				cache[topic] = append(v, extinf)
			} else {
				cache[topic] = []ExtInf{extinf}
			}
			filename = "runtime/" + url.QueryEscape(topic) + strconv.Itoa(int(t.DTS)) + ".ts"
			t.NewTs(filename)
			tslen = 0
		}

		t.FlvTag(pk.MessageTypeID, pk.Timestamp, byte(pk.ExtendTimestamp), pk.PayLoad)

		tslen += pk.Timestamp
	}
}
