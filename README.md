# rtmp-go

实现了rtmp协议，可进行视频rtmp推流,rtmp播放,存储直播flv流视频功能

目前新的直播技术为 https://webrtc.org/ 面向更好的未来(rtmp可进行学习)

下一版本准备实现 http-flv http-hls 播放功能预计版本(1.0)时间为(21年10月中旬左右)

## 使用指南

推流软件:  [**obs**](https://obsproject.com/zh-cn) 使用obs进行推流

播放器软件:  [**vlc**](https://www.videolan.org/vlc/) 万能播放器，不做介绍。

基于浏览器的JS播放器:  [**videojs**](https://videojs.com/) 可 flv,mp4,hls,等等协议的播放 (浏览器不能实现rtmp流播放)

鼻祖全能王：[**ffmpeg**](https://www.ffmpeg.org/) 可以进行推流，播放，格式转换等等超多超强大的功能


## PDF reference 

nginx rtmp拓展 https://github.com/arut/nginx-rtmp-module

Amf数据格式文档 [Amf data structure](https://www.adobe.com/content/dam/acom/en/devnet/pdf/amf0-file-format-specification.pdf)

Rtmp协议详解（握手部分文档缺失【存在握手验证的情况】） [Rtmp specification](https://www.adobe.com/content/dam/acom/en/devnet/rtmp/pdf/rtmp_specification_1.0.pdf)

## License

rtmp-go 为本人[pennilessfor@gmail.com](mailto:pennilessfor@gmail.com)学习研究制作，不存在任何限制。希望可以和更多伙伴互相学习交流