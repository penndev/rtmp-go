package rtmp

import "rtmp-go/av"

// chan 传输内容
type Pack struct {
	Type    byte
	Time    uint32
	Content []byte
}

// 播放端操作
type Play struct {
	AV chan Pack
}

func (py *Play) pull(pk Pack) {
	py.AV <- pk
}

// 推送端操作
type Publish struct {
	AV chan Pack
}

func (ph *Publish) run(wp *WorkPool) {
	for av := range ph.AV {
		for _, play := range wp.getPlayer() {
			play.pull(av)
		}
	}
}

func (ph *Publish) push(pk *Pack) {
	ph.AV <- *pk
}

type WorkPool struct {
	play []*Play
}

// 启动一个新连接。
func (wp *WorkPool) start(av *Publish) error {
	go av.run(wp)
	return nil
}

func (wp *WorkPool) getPlayer() []*Play {
	return wp.play
}

func (wp *WorkPool) addPlayer(py *Play) {
	wp.play = append(wp.play, py)
}

func newWorkPool() *WorkPool {
	w := &WorkPool{}

	py := &Play{
		AV: OnPublish(),
	}
	w.addPlayer(py)
	return w
}

func OnPublish() chan Pack {
	var flv av.FLV
	flv.GenFlv("room")

	var flvStream chan Pack

	go func() {
		defer flv.Close()
		for pk := range flvStream {
			flv.AddTag(int(pk.Type), pk.Time, pk.Content)
		}
	}()
	return flvStream
}
