# rtmp-go

基于rtmp1.0协议开发的直播服务器，支持rtmp、http-flv播放。直播全程录制等功能。

## 安装

可直接运行`go install github.com/penndev/rtmp-go@latest`安装 或  [下载](./releases) 可执行文件


## 推流

使用ffmpeg进行rtmp推流
```
> ffmpeg -re -i <filename.mp4> -vcodec h264 -acodec aac -f flv rtmp://localhost/live/room
```


使用obs studio进行rtmp推流
```
OBS Studio > 设置 > 直播 > 服务器 rtmp://127.0.0.1:1935/live/
OBS Studio > 设置 > 直播 > 推流码 room
```

## 播放

**播放地址为 `rtmp Serve 中 Topic 的key组成 ** (不同的工具组成的key不同，请留意观察控制台输出)


使用 ffmpeg 播放器播放
```
> ffplay <urlpath>
```
或者使用其他支持相关视频格式的播放器进行播放


## Reference 

RTMP [Rtmp specification 1.0](./docs/rtmp_specification_1.0.pdf)

AMF [Action Message Format [0,3]](./docs/amf0-file-format-specification.pdf)

FLV [Video File Format Specification version 10](./docs/video_file_format_spec_v10.pdf)
