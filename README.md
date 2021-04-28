## rtmp-go
* 本项目只能用于学习与研究，如用于生产项目产生的任何后果作者不负任何责任
* 本项目目前并未进行任何优化与测试，请勿用于学习之外的其他目的谢谢
## 简单启动测试

安装启动：
    
    git clone https://github.com/Penndev/rtmp-go.git
    cd rtmp-go
    go run .

首先打开终端推流： `ffmpeg -re -i input.mp4  -f flv rtmp://localhost:1935/live/room`

然后再次终端播放： `ffplay rtmp://localhost:1935/live/room`

## 支持以下软件并通过测试（mac版本）

Vlc播放rtmp视频地址 [VLC media player](https://www.videolan.org/) 

使用Obs进行推流     [ OBS Studio ](https://obsproject.com/)

安装使用FFmpeg工具包进行rtmp视频推流播放 [ FFmpeg  ](https://ffmpeg.org/)

## 已知但未解决

* Flash不支持 -（ 播放有数据传输，但是画面黑屏）
* 录制功能 - （直播实时录制，http分发，flv，m3u8,回放mp4等）
* 后台管理端 - （制作http回调，流管理）

## PDF reference

Amf数据格式文档 [Amf data structure](https://www.adobe.com/content/dam/acom/en/devnet/pdf/amf0-file-format-specification.pdf)

Rtmp协议详解（握手部分文档缺失【存在握手验证的情况】） [Rtmp specification](https://www.adobe.com/content/dam/acom/en/devnet/rtmp/pdf/rtmp_specification_1.0.pdf)
