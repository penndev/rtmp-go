# rtmp-go

实现了rtmp协议，可进行视频rtmp推流,rtmp播放,http-flv播放, 存储直播flv流视频功能

## 演示流程


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