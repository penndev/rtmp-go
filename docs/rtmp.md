* ## 目录
    1. [介绍](#introduction) 
        1. 术语
    2. [贡献者](#contributors)
    3. [定义](#definitions)
    4. [字节顺序，校准，时间格式](#datatime)
    5. [Rtmp Chunk Steam](#chunksteam)
        1. [Message Format](#chunkformat)
        2. [Handshake](#handshake)
            1. 握手序列
            2. C0和S0 数据格式
            3. C1和S1 数据格式
            4. C2和S2 数据格式
        3. Chunking 
            1. chunk 格式
                1. Chunk Basic Header 
                2. Chunk Message Header
                    1. Type 0
                    2. Type 1
                    3. Type 2
                    4. Type 3
                    5. Common Header Fields
            2. Examples
                1. 例 1
                2. 例 2
        4. Protocol Control Messages
            1. Set Chunk Size
            2. Abort Message
            3. Acknowledgement
            4. Window Acknowledgement Size
            5. Set Peer Bandwidth
    6. RTMP Message Formats
        1. RTMP Message Formats
            1. RTMP Message Format 
            2. Message Payload
        2. User Control Messages
    7. RTMP Command Messages 
        1. 消息类型
            1. Command Message (20, 17)
            2. Data Message (18, 15) 
            3. Shared Object Message (19, 16)
            4. Audio Message (8)
            5. Video Message (9) 
            6. Aggregate Message (22)
            7. User Control Message Events
        2. 命令类型
            1. NetConnection Commands
                1. connect 
                2. Call
                3. createStream 

            2. NetStream Commands
                1. play
                2. play2
                3. deleteStream
                4. receiveAudio
                5. receiveVideo
                6. publish
                7. seek
                8. pause
        3. Message Exchange Examples
    8. 参考文献
    * 作者地址



* ## 介绍  <span id="introduction"></span>

Adobe’s Real Time Messaging Protoco（RTMP）基于 TCP [RFC0793]协议，主要用户与服务端双向音视频消息传输。
> 可以用于视频的播放与推流，该协议主要用于媒体的互联网传输。因为Adobe Flash的退出，目前播放使用的比较少，大部分的应用是视频采集推流用。


#### 文档与术语
- 关键字“必须”，“必须”，“必须”，“应”，“应不”，
  “应该”，“不应该”，“推荐”，“不推荐”，“可以”和
  “可选的”  [词语的定义]( https://tools.ietf.org/html/rfc2119 )
- [tcp标准详情](https://tools.ietf.org/html/rfc793 )
- [rtmp官方英文文档]( https://wwwimages2.adobe.com/content/dam/acom/en/devnet/rtmp/pdf/rtmp_specification_1.0.pdf )


* ## 贡献者 <span id="contributors" ></span>
>ps : 咋滴你还想贡献些什么还是咋滴。。。

* ## 定义 <span id="definitions" ></span>

- Payload: 实际需要的数据（视频，音频，控制等..），去除协议冗余后的。 
- Packet: 数据包由固定的报头和有效载荷数据组成。 一些基础协议可能需要对数据包进行封装才能定义
- Port: 传输协议用来的标识端口。
- Transport address: 数据地址链路组合。
- Message stream: 数据的逻辑流通通道。
- Message stream ID: 每条消息都有消息ID用来表示消息属于那条消息。
- Chunk: 消息的分块在通过网络发送消息之前，将消息分成较小的部分并进行传输。确保能够按照先后顺序到达。
- Chunk stream: 块数据的逻辑流通通道。
- Chunk stream ID: 每个块都有一个与之关联的ID用来标识它属于那个块。
- Multiplexing: 音频视频的多路复用。
- DeMultiplexing: 将多路数据的数据整合为原始数据。
- Remote Procedure Call (RPC):  一种允许客户端或服务器在对等端调用子例程或过程的请求。
- Metadata: 有关数据的描述。 电影的元数据 包括电影标题，持续时间，创建日期等。
- Application Instance: 客户端通过发送连接请求后服务器上的应用程序实例。
Action Message Format (AMF):  一种紧凑的二进制格式，用于序列化原始数据。AMF有两个版本：AMF0[AMF0] 和 AMF3[AMF3]。
>网络编程的基本概念，我还是要多学习。。。

* ## 字节顺序，对齐方式和时间格式 <span id="datatime"></span>
    * 所有整数字段均按网络字节顺序传送，字节零是所示的第一个字节，而位零是字或字段中的最高有效位。 此字节顺序通常称为big-endian。
    * 传输顺序在Internet协议[RFC0791]中进行了详细描述。 除非另有说明，否则本文档中的数字常数以十进制（10为基数）为单位。
    * 除非另有说明，否则RTMP中的所有数据都是字节对齐的。 例如，一个16位字段可能处于奇数字节偏移处。 在指示填充的地方，填充字节的值应为零。
    * RTMP中的时间戳以相对于未指定历元的整数毫秒为单位给出。 通常，每个流将从0时间戳开始，但这不是必需的，只要两个端点在该时期达成一致即可。 请注意，这意味着跨多个流（尤其是来自单独主机的流）的任何同步都需要RTMP之外的一些其他机制。
    * 由于时间戳长为32位，因此每49天，17小时，2分钟47.296秒滚动一次。 因为允许流连续运行（可能连续数年），所以RTMP应用程序在处理时间戳时应使用序列号算法[RFC1982]，并且应能够处理环绕。
    * 相对于先前时间戳，时间戳增量也被指定为毫秒的无符号整数。 时间戳增量可能为24或32位长。

* ## RTMP Chunk Steam <span id="chunksteam"></span>
    * 本节指定实时消息协议块流（RTMP块流）。 提供多路复用和打包，为更高级别的多媒体流协议提供服务。
    * While RTMP Chunk Stream was designed to work with the Real Time Messaging Protocol (Section 6), it can handle any protocol that sends a stream of messages. Each message contains timestamp and payload  type identification. RTMP Chunk Stream and RTMP together are  suitable for a wide variety of audio-video applications, from one-to-one and one-to-many live broadcasting to video-on-demand services to interactive conferencing applications.
    * When used with a reliable transport protocol such as TCP [RFC0793], RTMP Chunk Stream provides guaranteed timestamp-ordered end-to-end delivery of all messages, across multiple streams. RTMP Chunk Stream does not provide any prioritization or similar forms of control, but  can be used by higher-level protocols to provide such prioritization.  For example, a live video server might choose to drop video messages  for a slow client to ensure that audio messages are received in a  timely fashion, based on either the time to send or the time to  acknowledge each message.
    * RTMP Chunk Stream includes its own in-band protocol control messages,  and also offers a mechanism for the higher-level protocol to embed user control messages.

    > ps： 没理解，大概是依靠时间戳优化视频传输的延迟？？

#### Message Format  <span id="chunkformat"></span>

* 由多个chunk组成的 message 包含一下字段。
    - Timestamp:  消息的时间戳。 该字段可以传输4个字节。
    - Length: 消息有效负载的长度。 如果无法删除消息头，则应将其包括在长度中。 该字段在块头中占用3个字节。 
    - Type Id:  一系列类型ID保留用于协议控制消息。
    - Message Stream ID:  消息流ID可以是任意值。 多路复用到同一块流上的不同消息流将根据其消息流ID进行多路分解。



#### 握手  Handshake  <span id="handshake"></span>

- RTMP连接从握手开始。 握手不同于协议的其余部分。 它由三个静态大小的块组成，而不是由带有标头的可变大小的块组成。 客户端（发起连接的端点）和服务器分别发送相同的三个块。 为了说明起见，这些块在由客户端发送时将被指定为C0，C1和C2。 服务器发送的S0，S1和S2。

- 握手顺序 Handshake Sequence
    * 握手始于客户端发送C0和C1块。 
    * 客户端必须等到收到S1后再发送C2。 
    * 客户端必须等到收到S2后再发送其他数据。 
    * 服务器必须等到收到C0后再发送S0和S1，也可以等到C1之后。 
    * 服务器必须等到收到C1后再发送S2。 
    * 服务器必须等到收到C2后再发送其他数据。

- C0 and S0 格式
```bash
    0 1 2 3 4 5 6 7
    +-+-+-+-+-+-+-+-+
    |   version     |
    +-+-+-+-+-+-+-+-+ 
    C0 and S0 bits
```

- 以下是C0,S0数据包中的字段：

    * Version (8 bits): 在C0中，此字段标识客户端请求的RTMP版本。 在S0中，此字段标识服务器选择的RTMP版本。 本规范定义的版本为3。值0-2是较早的专有产品使用的不赞成使用的值。默认值为0； 

    * 4-31保留用于将来的实现； 和不允许使用32-255（以使RTMP与基于文本的协议区分开，后者始终以可打印字符开头）。 不能识别客户端请求的版本的服务器应以3响应。客户端可以选择降级为版本3或放弃握手。

- C1 and S1 格式
```bash
     0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
    |                        time (4 bytes)                         |
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
    |                        zero (4 bytes)                         |
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
    |                        random bytes                           |
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
    |                         random bytes                          |
    |                            (cont)                             |
    |                             ....                              |
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
                             C1 and S1 bits
```
* C1 S1 的字段 
    - Time (4 bytes):  该字段包含一个时间戳，该时间戳应该用作该端点发送的所有将来块的纪元。 可以是0，也可以是任意值。 为了同步多个块流，端点可能希望发送其他块流的时间戳的当前值。
    - Zero (4 bytes):  该字段必须全为0。
    - Random data (1528 bytes): 该字段可以包含任意值。 由于每个端点都必须区分对它发起的握手的响应和对等端发起的握手的响应，因此该数据应该发送足够随机的信息。 但是，不需要加密安全的随机性甚至动态值



* C2 and S2 格式
```bash

  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 |                      time (4 bytes)                           |
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 |                      time2 (4 bytes)                          |
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 |                        random echo                            |
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 |                           random echo                         |
 |                            (cont)                             |
 |                             ....                              |
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 C2 and S2 bits
```
* C2 and S2 字段
    - C2和S2数据包的长度为1536个字符byte，几乎分别是S1和C1的回波，由以下字段组成：
    - Time (4 bytes):  该字段必须包含对等体在S1（对于C2）或C1（对于S2）中发送的时间戳。
    - Time2（4个字节）：此字段必须包含时间戳，在该时间戳下读取对等方发送的先前数据包（s1或c1）。
    - 随机回波（1528字节）：此字段必须包含随机数据S1（对于C2）或S2（对于C1）中的对等方发送的字段。 同时 可以将time和time2字段与当前字段一起使用 时间戳作为带宽和/或延迟的快速估计连接，但这可能没用。


##### 握手顺序

    +-------------+                           +-------------+
    |    Client   |       TCP/IP Network      |    Server   |
    +-------------+          |                +-------------+
        |                    |                     |
    Uninitialized            |               Uninitialized
        |          C0        |                     |
        |------------------->|         C0          |
        |                    |-------------------->|
        |          C1        |                     |
        |------------------->|         S0          |
        |                    |<--------------------|
        |                    |         S1          |
    Version sent             |<--------------------|
        |          S0        |                     |
        |<-------------------|                     |
        |          S1        |                     |
        |<-------------------|                Version sent
        |                    |         C1          |
        |                    |-------------------->|
        |          C2        |                     |
        |------------------->|         S2          |
        |                    |<--------------------|
    Ack sent                 |                  Ack Sent
        |          S2        |                     |
        |<-------------------|                     |
        |                    |         C2          |
        |                    |-------------------->|
    Handshake Done           |               Handshake Done
        |                    |                     |
            Pictorial Representation of Handshake

**下面描述了握手图中提到的状态：**

Uninitialized:  在此阶段发送协议版本。 客户端和服务器都未初始化。 客户端在数据包C0中发送协议版本。 如果服务器支持该版本，则它将发送S0和S1作为响应。 如果不是，则服务器通过采取适当的措施来响应。 在RTMP中，此操作将终止连接。

Version Sent:  在未初始化状态之后，客户端和服务器都处于“已发送版本”状态。 客户端正在等待数据包S1，服务器正在等待数据包C1。服务器发送数据包S2。 然后状态变为“已发送确认”。 发送确认客户端和服务器分别等待S2和C2。 完成握手：客户端和服务器交换消息。




#### 消息块 Chunking  <span id="chunking"></span>
- 握手后，连接将多路复用一个或多个块流。 每个块流从一个消息流中携带一种类型的消息。 创建的每个块都有一个与之关联的唯一ID，称为块流ID。 块通过网络传输。 传输时，每个块必须在下一个块之前完整发送。 在接收器端，基于块流ID将块组装为消息。

- 分块允许将较高级别协议中的大型消息分解为较小的消息，例如，防止大型低优先级消息（例如视频）阻止较小的高优先级消息（例如音频或控制）。

- 分块还允许以较小的开销发送小消息，因为块头包含信息的压缩表示，否则该信息必须包含在消息本身中。

- 块大小是可配置的。 可以使用“设置块大小”控制消息进行设置 较大的块大小可减少CPU使用率，但也会进行较大的写入，这可能会延迟带宽较低的连接上的其他内容。 较小的块不利于高比特率流传输。 每个方向的块大小均独立保持。

##### 块格式 Chunk Format

每个块均由 标题(hander) 和 数据(data) 组成。 标头本身包含三个部分：
```bash
    +--------------+----------------+--------------------+--------------+ 
    | Basic Header | Message Header | Extended Timestamp |  Chunk Data  |
    +--------------+----------------+--------------------+--------------+
    |                                                    |
    |<------------------- Chunk Header ----------------->|
```
- Basic Header (1 to 3 bytes): 该字段编码块流ID和块类型。 块类型确定编码的消息头的格式。 长度完全取决于块流ID，它是一个可变长度字段。
- Message Header (0, 3, 7, or 11 bytes): 该字段对有关正在发送的消息的信息进行编码（无论是全部还是部分）。 可以使用basic header中指定的块类型来确定长度。
- Extended Timestamp (0 or 4 bytes): 在某些情况下，取决于“块消息”标头中的编码时间戳或时间戳增量字段，此字段存在。 
- Chunk Data (variable size): 该块的有效负载，最大为配置的最大块大小。


##### 基础头信息 Chunk Basic Header
```bash
    0 1 2 3 4 5 6 7 （bits）
    +-+-+-+-+-+-+-+-+
    |fmt|   cs id   |
    +-+-+-+-+-+-+-+-+
    Chunk basic header 1
```
- basic header 可以是 1，2，3个字节。取决于csid
- 该协议最多支持65597个ID为3-65599的流。 
- csid ID 0、1和2被保留。
```bash
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
    |fmt|     0     |   csid - 64  |
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
        Chunk basic header 2
``` 
- 值0表示2字节形式，其ID范围为64-319（第二个字节+ 64）。 
```bash
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
    |fmt|     1     |        csid - 64             |
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
        Chunk basic header 3
```
- 值1表示3字节形式，其ID范围为64-65599（（第三个字节）* 256 + 第二个字节+ 64）。
- 值为2的块流ID保留用于低级协议控制消息和命令。 
- 3-63范围内的值表示完整的流ID。 

#### 块信息头 

块消息头有四种不同的格式，由块基础头中的“ fmt”字段选择。 一个实现应该为每个块消息头使用尽可能紧凑的表示形式。

##### Type 0  “fmt” = 0("00")
 Chunk Message Header = 11 bytes 这种类型必须在组块流的开始以及流时间戳向后的时候使用。
```bash
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
    |                   timestamp                   |message length |
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
    |     message length (cont)     |message type id| msg stream id |
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
    |           message stream id (cont)            |
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+ 
        Chunk Message Header - Type 0
```
时间戳记（3个字节）：对于类型0块，绝对时间戳记为消息发送到这里。 如果时间戳大于或等于16777215（十六进制0xFFFFFF），此字段必须为16777215，指示存在扩展时间戳字段编码完整的32位时间戳。 否则，此字段应是整个时间戳。



##### Type 1 “fmt” = 1("01")
Chunk Message Header  = 7 bytes 类型1块标题的长度为7个字节。 消息流ID不包括在内；该块采用与前面的块相同的流ID。 消息大小可变的流（例如，许多视频格式）应在每个新消息的第一个块之后使用这种格式
```bash     
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
    |                timestamp delta                |message length |
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
    |     message length (cont)     |message type id|
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+ 
        Chunk Message Header - Type 1
```

##### Type 2 “fmt” = 2("10") 
Chunk Message Header  = 3 bytes 既不包括流ID，也不包括消息长度。 该块与先前的块具有相同的流ID和消息长度。 具有固定大小消息的流（例如，某些音频和数据格式）应在每个消息的第一块之后使用这种格式。
```bash
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 | timestamp delta |
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```
##### Type 3 “fmt” = 3("11")

流ID，消息长度和时间戳增量字段不存在。此类型的块从相同的块流ID的前一个块中获取值。当单个消息被分割成块时，消息中除第一个消息外的所有块都应该使用这种类型。 请参见示例2（type=1）。
由大小完全相同，流ID和时间间隔完全相同的消息组成的流应在类型2的一个块之后的所有块中使用此类型。请参见示例1（type=0）。
如果第一条消息和第二条消息之间的增量与第一条消息的时间戳相同，则类型3的块可以立即跟随类型0的块，因为不需要块2的块来注册增量 。 
如果类型3组块紧随类型0组块，则此类型3组块的时间戳增量与类型0组块的时间戳相同。

##### 通用标题字段 
- timestamp delta (3 bytes): 对于fmt=1或fmt=2块，将在此处发送前一个块的时间戳和当前块的时间戳之间的差。 如果增量大于或等于16777215（十六进制0xFFFFFF），则此字段务必为16777215，指示存在扩展时间戳字段以对完整的32位增量进行编码。 否则，该字段应为实际增量。

- message length (3 bytes): 对于type-0或type-1块，消息的长度在此处发送。message lenght 与chunk payload长度是不一样的。

- message type id (1 byte): 对于type-0或type-1块，消息的类型在此处发送。

- message stream id (4 bytes): 对于type-0的chunk，将存储stream id。 消息流ID以Little-Endian格式存储。 通常，同一chunk csid的所有消息都将来自同一stream id。Typically, all messages in the same chunk stream will  come from the same message stream. While it is possible to  multiplex separate message streams into the same chunk stream,  this defeats the benefits of the header compression. However, if  one message stream is closed and another one subsequently opened,  there is no reason an existing chunk stream cannot be reused by  sending a new type-0 chunk.

>  英文部分不懂。只有看源码了。

- 扩展时间戳 扩展时间戳字段用于编码时间戳或大于16777215（0xFFFFFF）的时间戳增量；
 也就是说，对于不适合类型0、1，或2块的24位字段的时间戳或时间戳变化量。 该字段编码完整的32位时间戳或时间戳增量。 通过将类型0块的时间戳字段或类型1或2块的时间戳增量字段设置为16777215（0xFFFFFF）来指示此字段的存在。 当相同块流ID的最新Type 0、1或2块指示存在扩展时间戳字段时，此字段以Type 3块形式出现。



### 协议控制消息 Protocol Control Messages  <span id="controlmessage"></span>

[所有的协议消息类型](http://assets.processon.com/chart_image/5dfc3247e4b00cdf4f0ce846.png)

RTMP块流将消息类型ID 1、2、3、5和6用于协议控制消息。 这些消息包含RTMP块流协议所需的信息。 这些协议控制消息必须具有消息流ID 0（称为控制流），并以组块流ID 2发送。协议控制消息一接收到就生效； 它们的时间戳将被忽略。

- Set Chunk Size (1)

最大块大小默认为128字节，但是客户端或服务器可以更改此值，并使用此消息更新其对等方。 例如，假设客户端要发送131字节的音频数据，并且块大小为128。在这种情况下，客户端可以将此消息发送到服务器，以通知它现在块大小为131字节。 然后，客户端可以在单个块中发送音频数据。

最大块大小应至少为128个字节，并且必须至少为1个字节。 每个方向的最大块大小独立保持。

- Abort Message (2)

协议控制消息2（中止消息）用于通知对等端是否正在等待块完成消息，然后通过块流丢弃部分接收的消息。 对等方接收块流ID作为此协议消息的有效负载。 当关闭时，应用程序可以发送此消息，以指示不需要进一步处理消息。

块流ID（32位）：此字段保存块流ID，其当前消息将被丢弃。

- Acknowledgement (3)

客户端或服务器必须在接收到等于窗口大小的字节后，向对等方发送确认。 窗口大小是发送方在未收到接收方确认的情况下发送的最大字节数。 该消息指定序列号，该序列号是到目前为止接收到的字节数。


- Window Acknowledgement Size (5)

客户端或服务器发送此消息，以通知对等方在两次发送确认之间使用的窗口大小。 在发送方发送窗口大小字节之后，发送方期望来自其对等方的确认。 自从上一次发送确认以来，接收方必须在收到指示的字节数后，或者从会话开始时（如果尚未发送确认），在收到指示的字节数后发送确认 。

- Set Peer Bandwidth (6)

客户端或服务器发送此消息以限制其对等方的输出带宽。 接收此消息的对等方通过将已发送但未确认的数据量限制为此消息中指示的窗口大小来限制其输出带宽。 如果窗口大小与发送给该消息发送者的最后一个窗口大小不同，则收到此消息的对等方应以窗口确认大小消息作为响应。

