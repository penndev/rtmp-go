package rtmp

import (
	"sync"
	"time"
)

// 存储视频元数据
type metaPack struct {
	meta  []byte
	audit []byte
	video []byte
}

type PubSub struct {
	mu         sync.RWMutex       // 读写锁
	timeout    time.Duration      // 发布超时时间
	buffer     int                // 订阅队列的缓存大小
	subscriber map[chan Pack]bool // 订阅者
	// mediameta  metaPack           //拓展信息
}

func (ps *PubSub) sendPack(ch chan Pack, pk Pack) {
	select {
	case ch <- pk:
	case <-time.After(ps.timeout):
	}
}

// 判断是否存在
func (ps *PubSub) Publish(pk Pack) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for sub := range ps.subscriber {
		go ps.sendPack(sub, pk)
	}
}

func (ps *PubSub) Subscription() <-chan Pack {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ch := make(chan Pack, ps.buffer)
	ps.subscriber[ch] = true
	return ch
}

func (ps *PubSub) Close() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for sub := range ps.subscriber {
		delete(ps.subscriber, sub)
		close(sub)
	}
}
