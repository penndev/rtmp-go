package handler

import (
	"log"
	"net/url"
	"strings"
)

// AuthHandler 校验 rtmp://host/app/stream?user=...&password=...
// 用户名和密码由 NewAuthHandler 设置。查询参数不进入唯一名。
type AuthHandler struct {
	DefaultHandler
	User     string
	Password string
}

func NewAuthHandler(user, password string) *AuthHandler {
	return &AuthHandler{User: user, Password: password}
}

func (h *AuthHandler) OnPublish(app, stream string) bool {
	name := h.OnName(app, stream)
	if !h.allow(stream) {
		log.Printf("OnPublish auth failed name=%s", name)
		return false
	}
	log.Printf("OnPublish name=%s", name)
	return true
}

func (h *AuthHandler) allow(stream string) bool {
	_, raw, ok := strings.Cut(stream, "?")
	if !ok {
		return false
	}
	q, err := url.ParseQuery(raw)
	if err != nil {
		return false
	}
	return q.Get("user") == h.User && q.Get("password") == h.Password
}
