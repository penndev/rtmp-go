# rtmp

> rtmp的话语权已经从Adobe转到一个公益组织《https://veovera.org/》本项目也进行跟踪该组织的拓展标准进行重构。如果你仍然需要学习老版本的rtmp，请查看对应的tag。



### 效果演示

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