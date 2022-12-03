# rtmp-go

基于rtmp1.0协议开发的直播服务器，支持rtmp、http-flv播放。直播全程录制等功能。

## rtmp push

#### install

```
go install github.com/penndev/rtmp-go@latest
rtmp-go
```

直接 [下载](./releases) 可执行文件

### ffmpeg

```
> ffmpeg -re -i <filename.mp4> -vcodec h264 -acodec aac -f flv rtmp://localhost/live/room
```
### obs

```
OBS Studio > 设置 > 直播 > 服务器 rtmp://127.0.0.1:1935/live/  
OBS Studio > 设置 > 直播 > 推流码 room
```
## play

#### 播放地址
查看控制台输出播放地址
或者查看 rtmp topic 中的 map key

### ffplay 

#### ffplay rtmp play

```
> ffplay rtmp://localhost/<urlpath>
```

### vlc media player

#### vlc rtmp play
```
vlc > 媒体 > 打开网络串流 > 网络 > rtmp://localhost/<urlpath>
```

## Reference 

RTMP [Rtmp specification 1.0](./docs/rtmp_specification_1.0.pdf)

AMF [Action Message Format [0,3]](./docs/amf0-file-format-specification.pdf)

FLV [Video File Format Specification version 10](./docs/video_file_format_spec_v10.pdf)
