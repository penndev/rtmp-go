package handler

import "testing"

func TestAuthHandler(t *testing.T) {
	h := NewAuthHandler("admin", "123456")
	if got := h.OnName("app", "stream?user=admin&password=123456"); got != "app-stream" {
		t.Fatalf("name=%s", got)
	}
	if got := h.OnName("app?keep", "stream"); got != "app?keep-stream" {
		t.Fatalf("name=%s", got)
	}
	if !h.OnPublish("app", "stream?user=admin&password=123456") {
		t.Fatal("publish")
	}
	if h.OnPublish("app", "stream?user=admin&password=bad") {
		t.Fatal("bad password")
	}
	if h.OnPlay("app?user=admin&password=123456", "stream") {
		t.Fatal("query must be on stream")
	}
	if !h.OnPlay("app", "stream?user=ad%6Din&password=123456") {
		t.Fatal("play")
	}
}
