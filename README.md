# rtmp-go

基于rtmp1.0协议开发的直播服务器

- 推流协议 `rtmp`
- 拉流（播放）
    - rtmp 
    - http-fly
    - hls (m3u8)

## 直播录制功能

> 录制文件存放`runtime`目录下

## 推流

- 使用ffmpeg进行rtmp推流
    ```bash
    ffmpeg -re -i <in.mp4> -vcodec h264 -acodec aac -f flv rtmp://localhost/live/room
    ```

- 使用obs studio进行rtmp推流
    1. 进入 OBS Studio > **设置** > **直播**
    2. 输入 **服务器**: `rtmp://127.0.0.1:1935/live/`
    3. 输入 **推流码** `room`
  

## 播放

**播放地址为 `rtmp Serve` 中 Topic 的key组成** (不同的推流工具组成的key可能会有不同，请留意控制台输出)

使用 ffmpeg 播放器播放
```
> ffplay <urlpath>
```

_或者使用其他支持相关视频格式的播放器进行播放_
