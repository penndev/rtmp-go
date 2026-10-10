package adapter

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/penndev/rtmp/codec/flv"
	"github.com/penndev/rtmp/codec/h264"
	"github.com/penndev/rtmp/codec/mpegts"
)

type tagMsg struct {
	typ  flv.TagType
	ts   uint32
	data []byte
}

func (m tagMsg) Type() flv.TagType { return m.typ }
func (m tagMsg) Timestamp() uint32 { return m.ts }
func (m tagMsg) Data() []byte      { return m.data }

type chanSub struct {
	ch chan any
}

func (s *chanSub) Chan() <-chan any { return s.ch }
func (s *chanSub) Close()           {}

func TestAdapterTs(t *testing.T) {
	rec := h264.AVCDecoderConfigurationRecord{
		ConfigurationVersion:        1,
		AVCProfileIndication:        0x42,
		AVCLevelIndication:          0x1e,
		LengthSizeMinusOne:          3,
		SequenceParameterSetNALUnit: [][]byte{{0x67, 0x42, 0x00, 0x1e}},
		PictureParameterSetNALUnit:  [][]byte{{0x68, 0xce}},
	}
	cfg, err := rec.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	videoSeq := append([]byte{0x17, 0x00, 0x00, 0x00, 0x00}, cfg...)
	nalu := []byte{0x00, 0x00, 0x00, 0x03, 0x65, 0x88, 0x84}
	videoTag := append([]byte{0x17, 0x01, 0x00, 0x00, 0x00}, nalu...)

	sub := &chanSub{ch: make(chan any, 4)}
	sub.ch <- tagMsg{flv.TAG_TYPE_VIDEO, 0, videoSeq}
	sub.ch <- tagMsg{flv.TAG_TYPE_AUDIO, 0, []byte{0xaf, 0x00, 0x12, 0x10}}
	sub.ch <- tagMsg{flv.TAG_TYPE_VIDEO, 40, videoTag}
	sub.ch <- tagMsg{flv.TAG_TYPE_AUDIO, 40, []byte{0xaf, 0x01, 0x11, 0x22}}
	close(sub.ch)

	defer clearHLS("adapter-ts-test")
	AdapterTs("adapter-ts-test", sub)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/runtime/adapter-ts-test.m3u8", nil)
	WriteM3u8(rr, req, "adapter-ts-test")
	if rr.Code != http.StatusOK {
		t.Fatalf("m3u8 %d %s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, "#EXT-X-ENDLIST") || !strings.Contains(body, "adapter-ts-test/0.ts") {
		t.Fatalf("playlist:\n%s", body)
	}
	if _, err := os.Stat(filepath.Join("runtime", "adapter-ts-test.m3u8")); !os.IsNotExist(err) {
		t.Fatal("m3u8 file should not be written")
	}

	raw, err := os.ReadFile(filepath.Join("runtime", "adapter-ts-test", "0.ts"))
	if err != nil || len(raw) == 0 || len(raw)%mpegts.PacketSize != 0 {
		t.Fatalf("ts err=%v len=%d", err, len(raw))
	}
	var video, audio bool
	for off := 0; off < len(raw); off += mpegts.PacketSize {
		pkt, err := mpegts.Parse(raw[off : off+mpegts.PacketSize])
		if err != nil {
			t.Fatal(err)
		}
		if pkt.PID == mpegts.PIDVideo && pkt.PayloadUnitStartIndicator {
			video = true
		}
		if pkt.PID == mpegts.PIDAudio && pkt.PayloadUnitStartIndicator {
			audio = true
		}
	}
	if !video || !audio {
		t.Fatalf("video=%v audio=%v", video, audio)
	}
}

func TestAdapterTsHLS(t *testing.T) {
	rec := h264.AVCDecoderConfigurationRecord{
		ConfigurationVersion:        1,
		AVCProfileIndication:        0x42,
		AVCLevelIndication:          0x1e,
		LengthSizeMinusOne:          3,
		SequenceParameterSetNALUnit: [][]byte{{0x67, 0x42, 0x00, 0x1e}},
		PictureParameterSetNALUnit:  [][]byte{{0x68, 0xce}},
	}
	cfg, err := rec.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	videoSeq := append([]byte{0x17, 0x00, 0x00, 0x00, 0x00}, cfg...)
	nalu := []byte{0x00, 0x00, 0x00, 0x03, 0x65, 0x88, 0x84}
	videoTag := append([]byte{0x17, 0x01, 0x00, 0x00, 0x00}, nalu...)

	sub := &chanSub{ch: make(chan any, 16)}
	done := make(chan struct{})
	go func() {
		AdapterTs("hls-live", sub)
		close(done)
	}()
	defer clearHLS("hls-live")
	defer func() {
		select {
		case <-done:
		default:
			close(sub.ch)
			<-done
		}
	}()

	sub.ch <- tagMsg{flv.TAG_TYPE_VIDEO, 0, videoSeq}
	for _, ts := range []uint32{0, 2000, 4000, 6000, 8000, 10000, 12000, 14000} {
		sub.ch <- tagMsg{flv.TAG_TYPE_VIDEO, ts, videoTag}
	}

	var seq0 int
	deadline := time.Now().Add(2 * time.Second)
	for {
		hlsPlayMu.Lock()
		live := hlsPlay["hls-live"]
		var segs []tsSegment
		var endList bool
		if live != nil {
			segs = append([]tsSegment(nil), live.segments...)
			endList = live.endList
		}
		hlsPlayMu.Unlock()
		if live != nil && !endList && len(segs) == liveSegments && strings.HasSuffix(segs[len(segs)-1].path, "/6.ts") {
			seq0, _ = strconv.Atoi(strings.TrimSuffix(filepath.Base(segs[0].path), ".ts"))
			for _, s := range segs {
				raw, err := os.ReadFile(s.path)
				if err != nil || len(raw) == 0 || len(raw)%mpegts.PacketSize != 0 || raw[0] != mpegts.SyncByte {
					t.Fatalf("seg %s err=%v len=%d", s.path, err, len(raw))
				}
				pkt, err := mpegts.Parse(raw[:mpegts.PacketSize])
				if err != nil {
					t.Fatal(err)
				}
				if pkt.PID != mpegts.PIDPAT {
					t.Fatalf("seg %s first pid %x", s.path, pkt.PID)
				}
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("playlist segs timeout")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if seq0 != 4 {
		t.Fatalf("media sequence %d", seq0)
	}
	if _, err := os.Stat(filepath.Join("runtime", "hls-live", "0.ts")); err != nil {
		t.Fatal(err)
	}

	recw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/runtime/hls-live.m3u8", nil)
	WriteM3u8(recw, req, "hls-live")
	body := recw.Body.String()
	if recw.Code != http.StatusOK || strings.Contains(body, "#EXT-X-ENDLIST") || !strings.Contains(body, "#EXT-X-MEDIA-SEQUENCE:4") {
		t.Fatalf("live playlist %d\n%s", recw.Code, body)
	}

	close(sub.ch)
	<-done
	hlsPlayMu.Lock()
	live := hlsPlay["hls-live"]
	hlsPlayMu.Unlock()
	if live == nil || !live.endList || len(live.segments) != liveSegments || !strings.HasSuffix(live.segments[0].path, "/5.ts") {
		t.Fatalf("ended live=%v", live)
	}
}

func TestAdapterTsFFmpeg(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	dir := t.TempDir()
	in := dir + "/in.flv"
	cmd := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "testsrc=size=160x120:rate=10",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100",
		"-t", "1", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac",
		"-f", "flv", in)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("ffmpeg: %v %s", err, out)
	}
	f, err := os.Open(in)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	sub := &chanSub{ch: make(chan any, 256)}
	err = flv.Reader(f, func(*flv.FLV) func(*flv.Tag) error {
		return func(tag *flv.Tag) error {
			sub.ch <- tagMsg{tag.TagType, tag.Timestamp, append([]byte(nil), tag.Data...)}
			return nil
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	close(sub.ch)
	defer clearHLS("adapter-ts-ffmpeg")
	AdapterTs("adapter-ts-ffmpeg", sub)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/adapter-ts-ffmpeg.m3u8", nil)
	WriteM3u8(rr, req, "adapter-ts-ffmpeg")
	if rr.Code != http.StatusOK {
		t.Fatalf("m3u8 %d %s", rr.Code, rr.Body.String())
	}
	m3u8 := "adapter-ts-ffmpeg.m3u8"
	if err := os.WriteFile(m3u8, rr.Body.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(m3u8)
	probe := exec.Command("ffprobe", "-hide_banner", "-loglevel", "error",
		"-show_entries", "stream=codec_type,codec_name", "-of", "csv=p=0", m3u8)
	out, err := probe.CombinedOutput()
	if err != nil {
		t.Fatalf("ffprobe: %v %s", err, out)
	}
	got := string(out)
	if !strings.Contains(got, "h264") || !strings.Contains(got, "aac") {
		t.Fatalf("streams:\n%s", got)
	}
	dec := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-i", m3u8, "-f", "null", "-")
	if out, err := dec.CombinedOutput(); err != nil {
		t.Fatalf("decode: %v %s", err, out)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".m3u8") {
			name := strings.TrimSuffix(filepath.Base(r.URL.Path), ".m3u8")
			WriteM3u8(w, r, name)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/runtime/") {
			http.StripPrefix("/runtime/", http.FileServer(http.Dir("runtime"))).ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	hls := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error",
		"-i", srv.URL+"/adapter-ts-ffmpeg.m3u8", "-f", "null", "-")
	if out, err := hls.CombinedOutput(); err != nil {
		t.Fatalf("m3u8: %v %s", err, out)
	}
}

func clearHLS(name string) {
	hlsPlayMu.Lock()
	delete(hlsPlay, name)
	hlsPlayMu.Unlock()
	os.RemoveAll(filepath.Join("runtime", name))
	os.Remove(filepath.Join("runtime", name+".m3u8"))
}

func TestADTSHeader(t *testing.T) {
	frame := []byte{0x11}
	out := adts([]byte{0x12, 0x10}, frame)
	if len(out) != 8 || out[0] != 0xff || out[1] != 0xf1 || out[2] != 0x50 || out[7] != 0x11 {
		t.Fatalf("%x", out)
	}
}
