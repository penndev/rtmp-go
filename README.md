# rtmp

The RTMP specification is now maintained by [Veovera](https://veovera.org/) rather than Adobe. This project follows their Enhanced RTMP extensions. Older builds are on the matching tags.

### Codecs

- Classic RTMP / FLV: Sorenson H.263, H.264
- Enhanced RTMP: H.264 (`avc1`), H.265 (`hvc1`)

### Run

```bash
go run . -rtmp 127.0.0.1:1935 -http 127.0.0.1:8080
```

Open `http://127.0.0.1:8080/` in a browser. The page lists live streams. Each one has a copyable FLV address. HLS appears after a TS segment exists, so H.263 stays FLV only. It shows “No streams” when nothing is publishing. Refresh reloads the page.

The default name is `app-stream`. Publishing to `rtmp://127.0.0.1:1935/live/room` gives:

- `http://127.0.0.1:8080/live-room.flv`
- `http://127.0.0.1:8080/live-room.m3u8`

### Docker

```bash
docker run -d --name rtmp -p 1935:1935 -p 8080:8080 penndev/rtmp:latest
```

`latest` is the newest `v*` release. That release is also tagged with its version, such as `penndev/rtmp:v0.0.2`. FLV and TS files are written to `/app/runtime`.

```bash
docker build -t rtmp .
docker run -d --name rtmp -p 1935:1935 -p 8080:8080 rtmp
```

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

##



---

### 编码

- 经典 RTMP / FLV：Sorenson H.263、H.264
- Enhanced RTMP：H.264（`avc1`）、H.265（`hvc1`）

### 运行

```bash
go run . -rtmp 127.0.0.1:1935 -http 127.0.0.1:8080
```

浏览器打开 `http://127.0.0.1:8080/`。页面列出当前推流，每路都有可复制的 FLV 地址。写出 TS 分片后才出现 HLS，所以 H.263 只有 FLV。没有推流时显示 No streams，点 Refresh 重新加载。

默认名是 `app-stream`。推 `rtmp://127.0.0.1:1935/live/room` 时，页面上的地址是：

- `http://127.0.0.1:8080/live-room.flv`
- `http://127.0.0.1:8080/live-room.m3u8`

### Docker

```bash
docker run -d --name rtmp -p 1935:1935 -p 8080:8080 penndev/rtmp:latest
```

`latest` 是最新的 `v*` 版本。同一次发布还会打上版本标签，例如 `penndev/rtmp:v0.0.2`。FLV 和 TS 写在 `/app/runtime`。

```bash
docker build -t rtmp .
docker run -d --name rtmp -p 1935:1935 -p 8080:8080 rtmp
```

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
