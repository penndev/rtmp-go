# rtmp-go

实现了rtmp协议，可进行视频rtmp推流, rtmp播放, http-flv播放, 存储直播flv文件等功能

## 演示流程
第一步
```
#如果你已经配置好了go开发环境可直接执行go get -u
> go install github.com/penndev/rtmp-go@latest

#然后直接运行命令
> rtmp-go

> 如果你本地没有配置go环境则[点我下载可执行文件](https://github.com/penndev/rtmp-go/releases)
```
第二步

需下载并安装下面的相关工具  obs(推流工具，主播工具)   vlc(播放器，用户端)

1. 在obs中设置推流地址为 `rtmp://127.0.0.1:1935/live/room` 注意实际IP

2. 推流后程序会输出可播放的rtmp播放 与 http-flv播放地址，同时在 runtime 中生成 flv 文件

3. 根据 2 获取的url在vlc中进行播放

**注意**

- 目前主要处理了h264的编码格式，其他格式可能会抛异常

- win10安装docker 可能 1935 端口启动失败

- 如果需要我帮助或者讨论，请在底部邮件中联系我。

- 如果你遇到其他问题，欢迎并感谢提交[issues](https://github.com/penndev/rtmp-go/issues/new)


## 相关工具

推流软件:  [**obs**](https://obsproject.com/zh-cn) 使用obs进行推流

播放器软件:  [**vlc**](https://www.videolan.org/vlc/) 万能播放器，不做介绍。

全能王：[**ffmpeg**](https://www.ffmpeg.org/) 可以进行推流，播放，格式转换等等超多超强大的功能


基于浏览器的JS播放器:  [**videojs**](https://videojs.com/) 可 flv,mp4,hls,等等协议的播放 (浏览器不能实现rtmp流播放)

## Reference 

RTMP 协议标准（复杂握手部分缺失）[Rtmp specification 1.0](https://www.adobe.com/content/dam/acom/en/devnet/rtmp/pdf/rtmp_specification_1.0.pdf)

AMF 数据结构 [Action Message Format [0,3]](https://www.adobe.com/content/dam/acom/en/devnet/pdf/amf0-file-format-specification.pdf)

FLV 数据结构 [Video File Format Specification version 10](https://www.adobe.com/content/dam/acom/en/devnet/flv/video_file_format_spec_v10.pdf)

## License

rtmp-go 为本人[pennilessfor@gmail.com](mailto:pennilessfor@gmail.com)学习研究制作，不存在任何限制。希望可以和更多伙伴互相学习交流
