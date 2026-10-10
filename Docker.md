# penndev/rtmp

RTMP live server. Play with RTMP, HTTP-FLV, and HLS.

### Run

```bash
docker run -d --name rtmp -p 1935:1935 -p 8080:8080 penndev/rtmp:latest
```

`latest` is the newest `v*` release. That release is also tagged with its version, such as `penndev/rtmp:v0.0.2`.

`1935` is RTMP. `8080` is HTTP. Open `http://127.0.0.1:8080/`. The page lists live streams. FLV and TS files are written to `/app/runtime`.

Publishing to `rtmp://127.0.0.1:1935/live/room` gives:

- `http://127.0.0.1:8080/live-room.flv`
- `http://127.0.0.1:8080/live-room.m3u8`

### Publish

```bash
ffmpeg -re -i in.mp4 -c:v libx264 -c:a aac -f flv rtmp://127.0.0.1:1935/live/room
```

OBS Studio: Settings → Stream, server `rtmp://127.0.0.1:1935/live/`, stream key `room`.

### Play

```bash
ffplay http://127.0.0.1:8080/live-room.m3u8
ffplay http://127.0.0.1:8080/live-room.flv
ffplay rtmp://127.0.0.1:1935/live/room
```

HLS appears after a TS segment exists, so H.263 stays FLV only.

### Codecs

- Classic RTMP / FLV: Sorenson H.263, H.264
- Enhanced RTMP: H.264 (`avc1`), H.265 (`hvc1`)

Source: https://github.com/penndev/rtmp-go

---

RTMP 直播服务。可以用 RTMP、HTTP-FLV、HLS 播放。

### 运行

```bash
docker run -d --name rtmp -p 1935:1935 -p 8080:8080 penndev/rtmp:latest
```

`latest` 是最新的 `v*` 版本。同一次发布还会打上版本标签，例如 `penndev/rtmp:v0.0.2`。

`1935` 是 RTMP，`8080` 是 HTTP。浏览器打开 `http://127.0.0.1:8080/`，页面列出当前推流。FLV 和 TS 写在 `/app/runtime`。

推 `rtmp://127.0.0.1:1935/live/room` 时，页面上的地址是：

- `http://127.0.0.1:8080/live-room.flv`
- `http://127.0.0.1:8080/live-room.m3u8`

### 推流

```bash
ffmpeg -re -i in.mp4 -c:v libx264 -c:a aac -f flv rtmp://127.0.0.1:1935/live/room
```

OBS Studio：设置 → 直播，服务器 `rtmp://127.0.0.1:1935/live/`，推流码 `room`。

### 播放

```bash
ffplay http://127.0.0.1:8080/live-room.m3u8
ffplay http://127.0.0.1:8080/live-room.flv
ffplay rtmp://127.0.0.1:1935/live/room
```

写出 TS 分片后才出现 HLS，所以 H.263 只有 FLV。

### 编码

- 经典 RTMP / FLV：Sorenson H.263、H.264
- Enhanced RTMP：H.264（`avc1`）、H.265（`hvc1`）

源码：https://github.com/penndev/rtmp-go
