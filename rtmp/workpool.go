package rtmp

import (
	"fmt"
	"rtmp-go/av"
	"time"
)

// chan 传输内容
type Pack struct {
	Type    byte
	Time    uint32
	Content []byte
}

type WorkPool struct {
	playList map[string]([]chan Pack)
}

func (wp *WorkPool) ClosePublish(room string, push chan Pack) {
	close(push)
}

func (wp *WorkPool) Publish(room string, push chan Pack) {

	s := fmt.Sprint(time.Now().Unix())
	var flv av.FLV
	flv.GenFlv(room + s)

	go func() {
		for pck := range push {
			flv.AddTag(pck.Type, pck.Time, pck.Content)
			for _, py := range wp.playList[room] {
				py <- pck
			}
		}
		defer flv.Close()
	}()
}

func newWorkPool() *WorkPool {
	wp := &WorkPool{
		playList: make(map[string]([]chan Pack)),
	}
	return wp
}
