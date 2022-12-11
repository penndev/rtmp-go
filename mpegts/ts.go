package mpegts

import (
	"log"
	"net/url"
	"os"

	"github.com/penndev/rtmp-go/rtmp"
)

var defaultH264HZ = 90

func Adapterts(name string, ch <-chan rtmp.Pack) {
	rtimefile, err := os.OpenFile("runtime/"+url.QueryEscape(name)+".ts", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
	}
	defer rtimefile.Close()
	mts := &TsPack{
		w: rtimefile,
	}
	mts.w.Write(SDT())
	mts.w.Write(PAT())
	mts.w.Write(PMT())
	for pk := range ch {
		mts.FlvTag(pk.MessageTypeID, pk.Timestamp, byte(pk.ExtendTimestamp), pk.PayLoad)
	}

}
