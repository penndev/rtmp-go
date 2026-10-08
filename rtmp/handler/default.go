package handler

import (
	"log"
	"strings"
)

// DefaultHandler 用 app-stream 当唯一名。
// 推流和播放是否允许由 Stream 判断，这里只打印。
type DefaultHandler struct{}

func NewDefaultHandler() *DefaultHandler {
	return &DefaultHandler{}
}

func (h *DefaultHandler) OnName(app, stream string) string {
	stream, _, _ = strings.Cut(stream, "?")
	return app + "-" + stream
}

func (h *DefaultHandler) OnPublish(app, stream string) bool {
	log.Printf("OnPublish app=%s stream=%s name=%s", app, stream, h.OnName(app, stream))
	return true
}

func (h *DefaultHandler) OnPublishStop(app, stream string) {
	log.Printf("OnPublishStop app=%s stream=%s name=%s", app, stream, h.OnName(app, stream))
}

func (h *DefaultHandler) OnPlay(app, stream string) bool {
	log.Printf("OnPlay app=%s stream=%s name=%s", app, stream, h.OnName(app, stream))
	return true
}

func (h *DefaultHandler) OnPlayStop(app, stream string) {
	log.Printf("OnPlayStop app=%s stream=%s name=%s", app, stream, h.OnName(app, stream))
}
