package adapter

import "github.com/penndev/rtmp/pubsub"

func Capture(name string, topic *pubsub.Topic) {
	go AdapterFlv(name, topic.Subscribe())
}
