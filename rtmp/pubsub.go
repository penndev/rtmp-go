package rtmp

import (
	"sync"
	"time"
)

// 存储视频元数据
type metaInfo struct {
	ready int
	meta  Pack
	audit Pack
	video Pack
}

type PubSub struct {
	mu         sync.RWMutex       // 读写锁
	timeout    time.Duration      // 发布超时时间
	buffer     int                // 订阅队列的缓存大小
	subscriber map[chan Pack]bool // 订阅者
	mediaInfo  metaInfo           //拓展信息
}

func (ps *PubSub) sendPack(ch chan Pack, pk Pack) {
	// 还是会有退出的线程发送报错
	select {
	case ch <- pk:
	case <-time.After(ps.timeout):
	}
}

// 判断是否存在
func (ps *PubSub) Publish(pk Pack) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	// 处理解码信息
	if ps.mediaInfo.ready < 7 {
		if pk.MessageTypeID == 18 || pk.MessageTypeID == 15 {
			ps.mediaInfo.meta = pk
			ps.mediaInfo.ready |= 1
		} else if pk.MessageTypeID == 8 {
			ps.mediaInfo.audit = pk
			ps.mediaInfo.ready |= 2
		} else if pk.MessageTypeID == 9 {
			ps.mediaInfo.video = pk
			ps.mediaInfo.ready |= 4
		}
	}

	for sub := range ps.subscriber {
		go ps.sendPack(sub, pk)
	}
}

// 中途订阅。
func (ps *PubSub) Subscription() chan Pack {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ch := make(chan Pack, ps.buffer)
	ps.subscriber[ch] = true
	// 处理编码信息
	go func() {
		ch <- ps.mediaInfo.meta
		ch <- ps.mediaInfo.video
		ch <- ps.mediaInfo.audit
	}()
	return ch
}

// 中途退出订阅。
func (ps *PubSub) SubscriptionClose(ch chan Pack) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.subscriber, ch)
}

func (ps *PubSub) Close() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for sub := range ps.subscriber {
		delete(ps.subscriber, sub)
		close(sub)
	}
}
