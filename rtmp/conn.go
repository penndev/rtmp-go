package rtmp

type Conn struct {
	App    string
	Stream string

	// 主播端 true
	// 播放端 false
	Publish bool
	Closed  bool

	AVPackChan chan Pack
	CloseChan  chan bool
}

// 根据返回值处理连接是否继续
// return true 继续下一步
func (c *Conn) onConnect(app string) bool {
	c.App = app
	return true
}

func (c *Conn) onPublish(stream string) bool {
	//验证密钥。
	c.Publish = true
	return true
}

func (c *Conn) onPlay(stream string) bool {
	return true
}

func (c *Conn) onClose() {
	c.Closed = true
	c.CloseChan <- true
}

func newConn() *Conn {
	return &Conn{
		AVPackChan: make(chan Pack),
		CloseChan:  make(chan bool),
	}
}
