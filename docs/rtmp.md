# 介绍

Adobe’s Real Time Messaging Protoco（RTMP）基于 TCP [RFC0793]  半私有协议，主要用户与服务端双向视频消息传输。  的视频流传输。双向消息多路复用服务，该服务旨在在一对视频流之间传送视频，音频和数据消息的并行流以及相关的定时信息。 交流的同行。 实现通常为不同类别的消息分配不同的优先级，这可能会在传输能力受到限制时影响消息排队到基础流传输的顺序。 本备忘录描述了实时消息协议的语法和操作

# 术语

https://wwwimages2.adobe.com/content/dam/acom/en/devnet/rtmp/pdf/rtmp_specification_1.0.pdf
对照 https://tools.ietf.org/html/rfc2119

# 定义

Payload:   数据包中包含的数据，例如音频样本或压缩视频数据。 

Packet:  数据包由固定的报头和有效载荷数据组成。 一些基础协议可能需要对数据包进行封装才能定义

Port ：“传输协议用来在给定主机内的多个目的地之间进行区分的抽象。TCP/ IP协议使用小的正整数来标识端口。”  

Transport address :   数据包从源传输地址传输到目标传输地址

Message stream: 消息在其中流动的逻辑通信通道

Message stream ID: 每条消息都有一个与之关联的ID，以标识消息在其中流动

Chunk: 消息的一部分。 在通过网络发送消息之前，将消息分成较小的部分并进行交织。 这些块确保跨多个流的所有消息按时间戳排序的端到端传递。

Chunk stream: 允许块在特定方向上流动的逻辑通信通道。 块流可以从客户端传播到服务器，然后反向传输。

Chunk stream ID:  每个块都有一个与之关联的ID，以标识它所流入的块流

Multiplexing: 将单独的音频/视频数据转换为一个连贯的音频/视频流，从而可以同时传输多个视频和音频的过程。

DeMultiplexing: 反向复用的过程，其中将交错的音频和视频数据组合在一起以形成原始音频和视频数据。

Remote Procedure Call (RPC):  一种允许客户端或服务器在对等端调用子例程或过程的请求。

Metadata:  有关数据的描述。 电影的元数据包括电影标题，持续时间，创建日期等。

Application Instance: 客户端通过发送连接请求与之连接的服务器上的应用程序实例。

Action Message Format (AMF):  一种紧凑的二进制格式，用于序列化ActionScript对象图。 AMF有两个版本：AMF 0 [AMF0]和AMF 3 [AMF3]。


# 字节顺序，对齐方式和时间格式

所有整数字段均按网络字节顺序传送，字节零是所示的第一个字节，而位零是字或字段中的最高有效位。 此字节顺序通常称为big-endian。 传输顺序在Internet协议[RFC0791]中进行了详细描述。 除非另有说明，否则本文档中的数字常数以十进制（10为基数）为单位。


除非另有说明，否则RTMP中的所有数据都是字节对齐的。 例如，一个16位字段可能处于奇数字节偏移处。 在指示填充的地方，填充字节的值应为零。


RTMP中的时间戳以相对于未指定历元的整数毫秒为单位给出。 通常，每个流将从0时间戳开始，但这不是必需的，只要两个端点在该时期达成一致即可。 请注意，这意味着跨多个流（尤其是来自单独主机的流）的任何同步都需要RTMP之外的一些其他机制。


由于时间戳长为32位，因此每49天，17小时，2分钟和47.296秒滚动一次。 因为允许流连续运行（可能连续数年），所以RTMP应用程序在处理时间戳时应使用序列号算法[RFC1982]，并且应能够处理环绕。 例如，一个应用程序假设所有相邻时间戳都在2 ^ 31-1毫秒之内，因此10000在4000000000之后，而3000000000在4000000000之后。


相对于先前时间戳，时间戳增量也被指定为毫秒的无符号整数。 时间戳增量可能为24或32位长。

# RTMP块流

本部分指定实时消息协议块流（RTMP块流）。 它为更高级别的多媒体流协议提供多路复用和打包服务。

尽管RTMP块流旨在与实时消息协议一起使用（第6节），但它可以处理任何发送消息流的协议。 每个消息都包含时间戳和有效负载类型标识。 RTMP块流和RTMP一起适用于多种音频视频应用，从一对一和一对多的实时广播到视频点播服务再到交互式会议应用

当与可靠的传输协议（例如TCP [RFC0793]）一起使用时，RTMP块流可跨多个流提供按时间戳排序的所有消息的端到端交付保证。 RTMP块流不提供任何优先级划分或类似形式的控制，但是可以由更高级别的协议用来提供这种优先级划分。 例如，实时视频服务器可能会选择发送慢速客户端的视频消息，以确保根据发送时间或确认每个消息的时间及时接收音频消息。

RTMP块流包括其自己的带内协议控制消息，还提供了用于更高级别协议的机制来嵌入用户控制消息。


# 讯息格式

可以分为多个块以支持多路复用的消息格式取决于更高级别的协议。 但是消息格式应该包含以下字段，这些字段是创建块所必需的。


Timestamp:  消息的时间戳。 该字段可以传输4个字节。

Length: 消息有效负载的长度。 如果无法删除消息头，则应将其包括在长度中。 该字段在块头中占用3个字节。 

Type Id:  一系列类型ID保留用于协议控制消息。 这些传播信息的消息由RTMP块流协议和更高级别的协议处理。 所有其他类型的ID可供更高级别的协议使用，并且被RTMP Chunk Stream视为不透明值。 实际上，在RTMP块流中，没有任何要求将这些值用作类型。 所有（非协议）消息都可以是同一类型，或者应用程序可以使用此字段来区分同时记录的曲目而不是类型。 该字段在块头中占用1个字节

Message Stream ID:  消息流ID可以是任意值。 多路复用到同一块流上的不同消息流将根据其消息流ID进行多路分解。 除此之外，就RTMP块流而言，这是一个不透明的值。 该字段在块头中以小尾数格式占用4个字节。



# 握手  Handshake

RTMP连接从握手开始。 握手不同于协议的其余部分。 它由三个静态大小的块组成，而不是由带有标头的可变大小的块组成。 客户端（发起连接的端点）和服务器分别发送相同的三个块。 为了说明起见，这些块在由客户端发送时将被指定为C0，C1和C2。 服务器发送的S0，S1和S2。

## 握手顺序 Handshake Sequence

握手始于客户端发送C0和C1块。 
客户端必须等到收到S1后再发送C2。 
客户端必须等到收到S2后再发送其他数据。 
服务器必须等到收到C0后再发送S0和S1，也可以等到C1之后。 
服务器必须等到收到C1后再发送S2。 
服务器必须等到收到C2后再发送其他数据。

## C0 and S0 格式

 0 1 2 3 4 5 6 7
 +-+-+-+-+-+-+-+-+
 |   version     |
 +-+-+-+-+-+-+-+-+ 
  C0 and S0 bits

**以下是C0 / S0数据包中的字段：**

Version (8 bits): 在C0中，此字段标识客户端请求的RTMP版本。 在S0中，此字段标识服务器选择的RTMP版本。 本规范定义的版本为3。值0-2是较早的专有产品使用的不赞成使用的值。默认值为0； 

4-31保留用于将来的实现； 和不允许使用32-255（以使RTMP与基于文本的协议区分开，后者始终以可打印字符开头）。 不能识别客户端请求的版本的服务器应以3响应。客户端可以选择降级为版本3或放弃握手。

## C1 and S1 格式

 
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

Time (4 bytes):  该字段包含一个时间戳，该时间戳应该用作该端点发送的所有将来块的纪元。 可以是0，也可以是任意值。 为了同步多个块流，端点可能希望发送其他块流的时间戳的当前值。

Zero (4 bytes):  该字段必须全为0。


Random data (1528 bytes): 该字段可以包含任意值。 由于每个端点都必须区分对它发起的握手的响应和对等端发起的握手的响应，因此该数据应该发送足够随机的信息。 但是，不需要加密安全的随机性甚至动态值


## C2 and S2 格式

C2和S2数据包的长度为1536个字符byte，几乎分别是S1和C1的回波，由以下字段组成：

Time (4 bytes):  该字段必须包含对等体在S1（对于C2）或C1（对于S2）中发送的时间戳。

Time2（4个字节）：此字段必须包含时间戳，在该时间戳下读取对等方发送的先前数据包（s1或c1）。



## 握手顺序

         +-------------+                           +-------------+
         |    Client   |       TCP/IP Network      |    Server   |
         +-------------+            |              +-------------+
               |                    |                     |
         Uninitialized              |               Uninitialized
               |          C0        |                     |
               |------------------->|         C0          |
               |                    |-------------------->|
               |          C1        |                     |
               |------------------->|         S0          |
               |                    |<--------------------|
               |                    |         S1          |
          Version sent              |<--------------------|
               |          S0        |                     |
               |<-------------------|                     |
               |          S1        |                     |
               |<-------------------|                Version sent
               |                    |         C1          |
               |                    |-------------------->|
               |          C2        |                     |
               |------------------->|         S2          |
               |                    |<--------------------|
            Ack sent                |                  Ack Sent
               |          S2        |                     |
               |<-------------------|                     |
               |                    |         C2          |
               |                    |-------------------->|
          Handshake Done            |               Handshake Done
               |                    |                     |
                    Pictorial Representation of Handshake

**下面描述了握手图中提到的状态：**

Uninitialized:  在此阶段发送协议版本。 客户端和服务器都未初始化。 客户端在数据包C0中发送协议版本。 如果服务器支持该版本，则它将发送S0和S1作为响应。 如果不是，则服务器通过采取适当的措施来响应。 在RTMP中，此操作将终止连接。


Version Sent:  在未初始化状态之后，客户端和服务器都处于“已发送版本”状态。 客户端正在等待数据包S1，服务器正在等待数据包C1。服务器发送数据包S2。 然后状态变为“已发送确认”。 发送确认客户端和服务器分别等待S2和C2。 完成握手：客户端和服务器交换消息。


# 消息块 Chunking

握手后，连接将多路复用一个或多个块流。 每个块流从一个消息流中携带一种类型的消息。 创建的每个块都有一个与之关联的唯一ID，称为块流ID。 块通过网络传输。 传输时，每个块必须在下一个块之前完整发送。 在接收器端，基于块流ID将块组装为消息。

分块允许将较高级别协议中的大型消息分解为较小的消息，例如，防止大型低优先级消息（例如视频）阻止较小的高优先级消息（例如音频或控制）。

分块还允许以较小的开销发送小消息，因为块头包含信息的压缩表示，否则该信息必须包含在消息本身中。

块大小是可配置的。 可以使用“设置块大小”控制消息进行设置 较大的块大小可减少CPU使用率，但也会进行较大的写入，这可能会延迟带宽较低的连接上的其他内容。 较小的块不利于高比特率流传输。 每个方向的块大小均独立保持。

## 块格式 Chunk Format

每个块均由 标题(hander) 和 数据(data) 组成。 标头本身包含三个部分：


+--------------+----------------+--------------------+--------------+ 
| Basic Header | Message Header | Extended Timestamp |  Chunk Data  |  
+--------------+----------------+--------------------+--------------+
|                                                    |
|<------------------- Chunk Header ----------------->|

   


Basic Header (1 to 3 bytes):   该字段编码块流ID和块类型。 块类型确定编码的消息头的格式。 长度完全取决于块流ID，它是一个可变长度字段。

Message Header (0, 3, 7, or 11 bytes):   该字段对有关正在发送的消息的信息进行编码（无论是全部还是部分）。 可以使用块头中指定的块类型来确定长度。

Extended Timestamp (0 or 4 bytes):   在某些情况下，取决于“块消息”标头中的编码时间戳或时间戳增量字段，此字段存在。 

 Chunk Data (variable size):   该块的有效负载，最大为配置的最大块大小。


## 基础头信息 Chunk Basic Header

块基本头对块流ID和块类型进行编码（在下图中由fmt字段表示）。 块类型确定编码的消息头的格式。 块基本标头字段可以是1、2或3个字节，具体取决于块流ID（chunk stream ID）。

一个实现应该使用可以保存ID的最小表示。 ？


该协议最多支持65597个ID为3-65599的流。 
ID 0、1和2被保留。 
值0表示2字节形式，其ID范围为64-319（第二个字节+ 64）。 
值1表示3字节形式，其ID范围为64-65599（（第三个字节）* 256 +第二个字节+ 64）。 
3-63范围内的值表示完整的流ID。 
值为2的块流ID保留用于低级协议控制消息和命令。

块基本头中的0-5位（最低有效位）表示块流ID。

块流ID 2-63可以在此字段的1字节版本中编码。

 0 1 2 3 4 5 6 7 （bits）
+-+-+-+-+-+-+-+-+
|fmt|   cs id   |
+-+-+-+-+-+-+-+-+
Chunk basic header 1


**块流ID 64-319可以头文件的2字节形式编码。 ID计算为（第二个字节+ 64）。**

 0                   10            
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|fmt|     0     |   cs id - 64  |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	Chunk basic header 2


**块流ID 64-65599可以在此字段的3字节版本中进行编码。 ID计算为（（第三个字节）* 256 +（第二个字节）+ 64）。**

 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|fmt|     1     |        cs id - 64             |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	Chunk basic header 3

cs id (6 bits):  该字段包含块流ID，取值范围为2-63。 值0和1用于指示此字段的2字节或3字节版本。

fmt (2 bits): 此字段标识“块邮件标题”使用的四种格式之一。 下一节将说明每种块类型的“块消息头”。

cs id - 64 (8 or 16 bits): 该字段包含块流ID减去64。例如，ID 365将由cs id中的1表示，此处由16位301表示。

值64-319的块流ID可以由2字节或3字节形式的标头表示。


## 块信息头 

块消息头有四种不同的格式，由块基础头中的“ fmt”字段选择。 一个实现应该为每个块消息头使用尽可能紧凑的表示形式。


### Type 0
“fmt” = 0("00")
Chunk Message Header  = 11 bytes
这种类型必须在组块流的开始以及流时间戳向后的时候使用。

时间戳（3个字节）：对于类型0块，消息的绝对时间戳在此处发送。 如果时间戳大于或等于16777215（十六进制0xFFFFFF），则此字段务必为16777215，指示存在扩展时间戳字段以对完整的32位时间戳进行编码。 否则，此字段应为整个时间戳
     
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                   timestamp                   |message length |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|     message length (cont)     |message type id| msg stream id |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|           message stream id (cont)            |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+ 
     Chunk Message Header - Type 0


### Type 1 


“fmt” = 1("01")
Chunk Message Header  = 7 bytes
类型1块标题的长度为7个字节。 消息流ID不包括在内； 该块采用与前面的块相同的流ID。 消息大小可变的流（例如，许多视频格式）应在每个新消息的第一个块之后使用这种格式
     
0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                timestamp delta                |message length |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|     message length (cont)     |message type id|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+ 
    Chunk Message Header - Type 1


### Type 2

“fmt” = 2("10") 
Chunk Message Header  = 3 bytes
既不包括流ID，也不包括消息长度。 该块与先前的块具有相同的流ID和消息长度。 具有固定大小消息的流（例如，某些音频和数据格式）应在每个消息的第一块之后使用这种格式。



### Type 3

“fmt” = 3("11")

流ID，消息长度和时间戳增量字段不存在。  
此类型的块从相同的块流ID的前一个块中获取值。
当单个消息被分割成块时，消息中除第一个消息外的所有块都应该使用这种类型。 请参见示例2（第5.3.2.2节）。
由大小完全相同，流ID和时间间隔完全相同的消息组成的流应在类型2的一个块之后的所有块中使用此类型。请参见示例1（第5.3.2.1节）。
如果第一条消息和第二条消息之间的增量与第一条消息的时间戳相同，则类型3的块可以立即跟随类型0的块，因为不需要块2的块来注册增量 。 
如果类型3组块紧随类型0组块，则此类型3组块的时间戳增量与类型0组块的时间戳相同。

### 通用标题字段 

块消息头中每个字段的描述：

timestamp delta (3 bytes): 对于fmt=1或fmt=2块，将在此处发送前一个块的时间戳和当前块的时间戳之间的差。 如果增量大于或等于16777215（十六进制0xFFFFFF），则此字段务必为16777215，指示存在扩展时间戳字段以对完整的32位增量进行编码。 否则，该字段应为实际增量。

message length (3 bytes): 对于类型0或类型1块，消息的长度在此处发送。 请注意，这通常与块有效负载的长度不同。 块有效载荷长度是除最后一个块以外的所有块的最大块大小，最后一个块的其余部分（对于小消息，可能是整个长度）。

message type id (1 byte): 对于类型0或类型1块，消息的类型在此处发送。

message stream id (4 bytes): 对于类型0块，将存储消息流ID。 消息流ID以Little-Endian格式存储。 通常，同一块流中的所有消息都将来自同一消息流。 尽管可以将单独的消息流多路复用到相同的块流中，但这使标头压缩的好处无法实现。 但是，如果关闭了一个消息流，然后又打开了另一个消息流，则没有理由不能通过发送新的Type-0块来重用现有的块流。

## 扩展时间戳 Extended Timestamp

Extended Timestamp字段用于编码大于16777215（0xFFFFFF）的时间戳或时间戳增量； 也就是说，对于不适合类型0、1，或2块的24位字段的时间戳或时间戳变化量。 该字段编码完整的32位时间戳或时间戳增量。 通过将类型0块的时间戳字段或类型1或2块的时间戳增量字段设置为16777215（0xFFFFFF）来指示此字段的存在。 当相同块流ID的最新Type 0、1或2块指示存在扩展时间戳字段时，此字段以Type 3块形式出现。


# 协议控制消息 Protocol Control Messages

http://assets.processon.com/chart_image/5dfc3247e4b00cdf4f0ce846.png

RTMP块流将消息类型ID 1、2、3、5和6用于协议控制消息。 这些消息包含RTMP块流协议所需的信息。 这些协议控制消息必须具有消息流ID 0（称为控制流），并以组块流ID 2发送。协议控制消息一接收到就生效； 它们的时间戳将被忽略。

## Set Chunk Size (1)

最大块大小默认为128字节，但是客户端或服务器可以更改此值，并使用此消息更新其对等方。 例如，假设客户端要发送131字节的音频数据，并且块大小为128。在这种情况下，客户端可以将此消息发送到服务器，以通知它现在块大小为131字节。 然后，客户端可以在单个块中发送音频数据。

最大块大小应至少为128个字节，并且必须至少为1个字节。 每个方向的最大块大小独立保持。

## Abort Message (2)

协议控制消息2（中止消息）用于通知对等端是否正在等待块完成消息，然后通过块流丢弃部分接收的消息。 对等方接收块流ID作为此协议消息的有效负载。 当关闭时，应用程序可以发送此消息，以指示不需要进一步处理消息。

块流ID（32位）：此字段保存块流ID，其当前消息将被丢弃。

##  Acknowledgement (3)

客户端或服务器必须在接收到等于窗口大小的字节后，向对等方发送确认。 窗口大小是发送方在未收到接收方确认的情况下发送的最大字节数。 该消息指定序列号，该序列号是到目前为止接收到的字节数。


## Window Acknowledgement Size (5)

客户端或服务器发送此消息，以通知对等方在两次发送确认之间使用的窗口大小。 在发送方发送窗口大小字节之后，发送方期望来自其对等方的确认。 自从上一次发送确认以来，接收方必须在收到指示的字节数后，或者从会话开始时（如果尚未发送确认），在收到指示的字节数后发送确认 。

## Set Peer Bandwidth (6)

客户端或服务器发送此消息以限制其对等方的输出带宽。 接收此消息的对等方通过将已发送但未确认的数据量限制为此消息中指示的窗口大小来限制其输出带宽。 如果窗口大小与发送给该消息发送者的最后一个窗口大小不同，则收到此消息的对等方应以窗口确认大小消息作为响应。


# RTMP Message Formats   


本节指定使用较低级传输层（例如RTMP块流）在网络上的实体之间传输的RTMP消息的格式。 尽管RTMP旨在与RTMP块流一起使用，但它可以使用任何其他传输协议来发送消息。 RTMP块流和RTMP一起适用于多种音频视频应用程序，从一对一和一对多的实时广播到视频点播服务再到交互式会议应用程序。


服务器和客户端通过网络发送RTMP消息以相互通信。 消息可以包括音频，视频，数据或任何其他消息。 RTMP消息分为两部分，标题和有效负载。


## Message Header

消息头包含以下内容：


Message Type: 一个字节的字段代表消息类型。 一系列类型ID（1-6）保留用于协议控制消息。

Length:  三字节字段，表示有效载荷的大小（以字节为单位）。 它以大端格式设置。

Timestamp: 包含消息时间戳的四字节字段。 这4个字节按big-endian顺序打包。

Message Stream Id: 标识消息流的三字节字段。 这些字节以big-endian格式设置。


   0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 
  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
  | Message Type  |                Payload length                 |
  |   (1 byte)    |                 (3 bytes)                     |
  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
  |                       Timestamp                               |
  |                       (4 bytes)                               | 
  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
  |                Stream ID                      | 
  |                (3 bytes)                      |     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
                    Message Header


## Message Payload 

消息的另一部分是有效负载，它是消息中包含的实际数据。 例如，可能是一些音频样本或压缩的视频数据。 有效负载格式和解释超出了本文档的范围。


##  User Control Messages (4)

RTMP对用户控制消息使用消息类型ID 4。 这些消息包含RTMP流传输层使用的信息。 RTMP块流协议使用ID为1、2、3、5和6的协议消息。

用户控制消息应该使用消息流ID 0（称为控制流），并且在通过RTMP块流发送时，应在块流ID 2上发送。用户控制消息在从流中接收时是有效的。 它们的时间戳将被忽略。

客户端或服务器发送此消息，以将用户控制事件通知给对等方。 该消息携带事件类型和事件数据

消息数据的前2个字节用于标识事件类型。 事件类型后面是事件数据。 事件数据字段的大小是可变的。 但是，在消息必须通过RTMP块流层的情况下，最大块大小（第5.4.1节）应该足够大，以允许这些消息适合单个块。





# RTMP Command Messages

本节描述服务器和客户端之间交换以相互通信的不同类型的消息和命令

服务器和客户端之间交换的消息的不同类型包括：用于发送音频数据的音频消息，用于发送视频数据的视频消息，用于发送任何用户数据的数据消息，共享对象消息和命令消息。 共享对象消息提供了一种通用方法来管理多个客户端和服务器之间的分布式数据。 命令消息在客户端和服务器之间携带AMF编码的命令。 客户端或服务器可以通过使用命令消息传递给对等方的流请求远程过程调用（RPC）。

## Types of Messages

服务器和客户端通过网络发送消息以相互通信。 消息可以是任何类型，包括音频消息，视频消息，命令消息，共享对象消息，数据消息和用户控制消息


## Command Message (20, 17)

命令消息在客户端和服务器之间携带AMF编码的命令。 这些消息的AMF0编码的消息类型值为20，AMF3编码的消息类型值为17。 发送这些消息以执行某些操作，例如对等方上的connect，createStream，发布，播放，暂停。 命令消息（如onstatus，result等）用于通知发送方所请求命令的状态。 命令消息由命令名称，事务ID和包含相关参数的命令对象组成。 客户端或服务器可以通过使用命令消息传递给对等方的流请求远程过程调用（RPC）。

## Data Message (18, 15)

客户端或服务器发送此消息，以将元数据或任何用户数据发送到对等方。 元数据包括有关数据（音频，视频等）的详细信息，例如创建时间，持续时间，主题等。 这些消息的AMF0消息类型值为18，AMF3消息类型值为15。


## Shared Object Message (19, 16)

共享库是在多个客户端，实例等之间同步的Flash对象（名称/值对的集合）。 AMF0的消息类型19和AMF3的消息类型16保留用于共享对象事件。 每条消息可以包含多个事件


支持以下事件类型：




## Audio Message (8)

客户端或服务器发送此消息以将音频数据发送到对等方。 消息类型值为8保留给音频消息


## Video Message (9)

客户端或服务器发送此消息以将视频数据发送到对等方。 消息类型值为9保留给视频消息。


## Aggregate Message (22)

聚合消息是一条单个消息，其中包含使用第6.1节中描述的格式的一系列RTMP子消息。 消息类型22用于汇总消息

聚合消息的消息流ID会覆盖聚合内部的子消息的消息流ID

聚合消息的时间戳和第一个子消息的时间戳之间的差异是用于将子消息的时间戳重新规范化为流时标的偏移量。 偏移量将添加到每个子消息的时间戳中，以达到标准化的流时间。 第一个子消息的时间戳应与聚合消息的时间戳相同，因此偏移量应为零

后退指针包含上一条消息的大小，包括其标题。 包含它以匹配FLV文件的格式，并用于向后搜索

使用聚合消息具有以下性能优势：

块流最多可以发送一个块中的单个完整消息。 因此，增加块大小并使用聚合消息会减少发送的块数量

子消息可以连续存储在内存中。 进行系统调用以在网络上发送数据时，效率更高。

## User Control Message Events

客户端或服务器发送此消息，以将用户控制事件通知给对等方。 有关消息格式的信息，请参见第6.2节。


## Types of Commands

客户端和服务器交换AMF编码的命令。 发送方发送一条命令消息，该消息由命令名称，事务ID和包含相关参数的命令对象组成。 例如，connect命令包含“ app”参数，该参数告诉客户端连接到的服务器应用程序名称。 接收方处理该命令，并以相同的事务ID发送回响应。 响应字符串可以是_result， \_error或方法名称，例如verifyClient或contactExternalServer。


\_result或_error的命令字符串表示响应。 事务ID指示响应所引用的未完成命令。 它与IMAP和许多其他协议中的标记相同。 命令字符串中的方法名称表示发送方正在尝试在接收方运行方法。


以下类对象用于发送各种命令：

NetConnection一个对象，是服务器和客户端之间连接的高级表示。

NetStream一个对象，表示通过其发送音频流，视频流和其他相关数据的通道。 我们还发送诸如播放，暂停等命令，这些命令控制数据流。


## NetConnection Commands

NetConnection管理客户端应用程序和服务器之间的双向连接。 此外，它还支持异步远程方法调用。

可以在NetConnection上发送以下命令：

   o  connect   
   o  call   
   o  close   
   o  createStream


### connect 

客户端将connect命令发送到服务器，以请求连接到服务器应用程序实例。

从客户端到服务器的命令结构如下：

 +----------------+---------+---------------------------------------+
 |  Field Name    |  Type   |           Description                 |
 +--------------- +---------+---------------------------------------+
 | Command Name   | String  | Name of the command. Set to "connect".|
 +----------------+---------+---------------------------------------+
 | Transaction ID | Number  | Always set to 1.                      |
 +----------------+---------+---------------------------------------+
 | Command Object | Object  | Command information object which has  |
 |                |         | the name-value pairs.                 |
 +----------------+---------+---------------------------------------+
 | Optional User  | Object  | Any optional information              |
 | Arguments      |         |                                       |
 +----------------+---------+---------------------------------------+

以下是connect命令的Command 
Object中使用的名称/值对的描述

+-----------+--------+-----------------------------+----------------+
| Property  |  Type  |        Description          | Example Value  |
+-----------+--------+-----------------------------+----------------+
|   app     | String | The Server application name |    testapp     |
|           |        | the client is connected to. |                |
+-----------+--------+-----------------------------+----------------+
| flashver  | String | Flash Player version. It is |    FMSc/1.0    |
|           |        | the same string as returned |                |
|           |        | by the ApplicationScript    |                |
|           |        | getversion () function.     |                |
+-----------+--------+-----------------------------+----------------+
|  swfUrl   | String | URL of the source SWF file  | file://C:/     |
|           |        | making the connection.      | FlvPlayer.swf  |
+-----------+--------+-----------------------------+----------------+
|  tcUrl    | String | URL of the Server.          | rtmp://local   |
|           |        | It has the following format.| host:1935/test |
|           |        | protocol://servername:port/ | app/instance1  |
|           |        | appName/appInstance         |                |
+-----------+--------+-----------------------------+----------------+
|  fpad     | Boolean| True if proxy is being used.| true or false  |
+-----------+--------+-----------------------------+----------------+
|audioCodecs| Number | Indicates what audio codecs | SUPPORT_SND    |
|           |        | the client supports.        | \_MP3          |
+-----------+--------+-----------------------------+----------------+
|videoCodecs| Number | Indicates what video codecs | SUPPORT_VID    |
|           |        | are supported.              | \_SORENSON     |
+-----------+--------+-----------------------------+----------------+
|videoFunct-| Number | Indicates what special video| SUPPORT_VID    |
|ion        |        | functions are supported.    | \_CLIENT_SEEK  |
+-----------+--------+-----------------------------+----------------+
|  pageUrl  | String | URL of the web page from    | http://        |
|           |        | where the SWF file was      | somehost/      |
|           |        | loaded.                     | sample.html    |
+-----------+--------+-----------------------------+----------------+
| object    | Number | AMF encoding method.        |     AMF3       |
| Encoding  |        |                             |                |
+-----------+--------+-----------------------------+----------------+

audioCodecs属性的标志值：

+----------------------+----------------------------+--------------+
|      Codec Flag      |          Usage             |     Value    |
+----------------------+----------------------------+--------------+
|  SUPPORT_SND_NONE    | Raw sound, no compression  |    0x0001    |
+----------------------+----------------------------+--------------+
|  SUPPORT_SND_ADPCM   | ADPCM compression          |    0x0002    |
+----------------------+----------------------------+--------------+
|  SUPPORT_SND_MP3     | mp3 compression            |    0x0004    |
+----------------------+----------------------------+--------------+
|  SUPPORT_SND_INTEL   | Not used                   |    0x0008    |
+----------------------+----------------------------+--------------+
|  SUPPORT_SND_UNUSED  | Not used                   |    0x0010    |
+----------------------+----------------------------+--------------+
|  SUPPORT_SND_NELLY8  | NellyMoser at 8-kHz        |    0x0020    |
|                      | compression                |              |
+----------------------+----------------------------+--------------+
|  SUPPORT_SND_NELLY   | NellyMoser compression     |    0x0040    |
|                      | (5, 11, 22, and 44 kHz)    |              |
+----------------------+----------------------------+--------------+
|  SUPPORT_SND_G711A   | G711A sound compression    |    0x0080    |
|                      | (Flash Media Server only)  |              |
+----------------------+----------------------------+--------------+
|  SUPPORT_SND_G711U   | G711U sound compression    |    0x0100    |
|                      | (Flash Media Server only)  |              |
+----------------------+----------------------------+--------------+
|  SUPPORT_SND_NELLY16 | NellyMouser at 16-kHz      |    0x0200    |
|                      | compression                |              |
+----------------------+----------------------------+--------------+
|  SUPPORT_SND_AAC     | Advanced audio coding      |    0x0400    |
|                      | (AAC) codec                |              |
+----------------------+----------------------------+--------------+
|  SUPPORT_SND_SPEEX   | Speex Audio                |    0x0800    |
+----------------------+----------------------------+--------------+
|  SUPPORT_SND_ALL     | All RTMP-supported audio   |    0x0FFF    |
|                      | codecs                     |              |
+----------------------+----------------------------+--------------+

videoCodecs属性的标志值：

+----------------------+----------------------------+--------------+
|      Codec Flag      |            Usage           |    Value     |
+----------------------+----------------------------+--------------+
|  SUPPORT_VID_UNUSED  | Obsolete value             |    0x0001    |
+----------------------+----------------------------+--------------+
|  SUPPORT_VID_JPEG    | Obsolete value             |    0x0002    |    +----------------------+----------------------------+--------------+
| SUPPORT_VID_SORENSON | Sorenson Flash video       |    0x0004    |
+----------------------+----------------------------+--------------+
| SUPPORT_VID_HOMEBREW | V1 screen sharing          |    0x0008    |
+----------------------+----------------------------+--------------+
| SUPPORT_VID_VP6 (On2)| On2 video (Flash 8+)       |    0x0010    |
+----------------------+----------------------------+--------------+
| SUPPORT_VID_VP6ALPHA | On2 video with alpha       |    0x0020    |
| (On2 with alpha      | channel                    |              |
| channel)             |                            |              |
+----------------------+----------------------------+--------------+
| SUPPORT_VID_HOMEBREWV| Screen sharing version 2   |    0x0040    |
| (screensharing v2)   | (Flash 8+)                 |              |
+----------------------+----------------------------+--------------+
| SUPPORT_VID_H264     | H264 video                 |    0x0080    |
+----------------------+----------------------------+--------------+
| SUPPORT_VID_ALL      | All RTMP-supported video   |    0x00FF    |
|                      | codecs                     |              |
+----------------------+----------------------------+--------------+

videoFunction属性的标志值：

+----------------------+----------------------------+--------------+
|    Function Flag     |           Usage            |     Value    |
+----------------------+----------------------------+--------------+
| SUPPORT_VID_CLIENT   | Indicates that the client  |       1      |
| \_SEEK               | can perform frame-accurate |              |
|                      | seeks.                     |              |
+----------------------+----------------------------+--------------+

对象编码属性的值：

+----------------------+----------------------------+--------------+
|    Encoding Type     |           Usage            |    Value     |
+----------------------+----------------------------+--------------+
|        AMF0          | AMF0 object encoding       |      0       |
|                      | supported by Flash 6 and   |              |
|                      | later                      |              |
+----------------------+----------------------------+--------------+
|        AMF3          | AMF3 encoding from         |      3       |
|                      | Flash 9 (AS3)              |              |
+----------------------+----------------------------+--------------+

从服务器到客户端的命令结构如下：

+--------------+----------+----------------------------------------+
| Field Name   |   Type   |             Description                |
+--------------+----------+----------------------------------------+
| Command Name |  String  | \_result or \_error; indicates whether |
|              |          | the response is result or error.       |
+--------------+----------+----------------------------------------+
| Transaction  |  Number  | Transaction ID is 1 for connect        |
| ID           |          | responses                              |
|              |          |                                        |
+--------------+----------+----------------------------------------+
| Properties   |  Object  | Name-value pairs that describe the     |
|              |          | properties(fmsver etc.) of the         |
|              |          | connection.                            |
+--------------+----------+----------------------------------------+
| Information  |  Object  | Name-value pairs that describe the     |
|              |          | response from|the server. ’code’,      |
|              |          | ’level’, ’description’ are names of few|
|              |          | among such information.                |
+--------------+----------+----------------------------------------+


connect命令中的消息流

+--------------+                              +-------------+
|    Client    |             |                |    Server   |
+------+-------+             |                +------+------+
    |              Handshaking done               |
    |                     |                       |
    |                     |                       |
    |                     |                       |
    |                     |                       |
    |----------- Command Message(connect) ------->|
    |                                             |
    |<------- Window Acknowledgement Size --------|
    |                                             |
    |<----------- Set Peer Bandwidth -------------|
    |                                             |
    |-------- Window Acknowledgement Size ------->|
    |                                             |
    |<------ User Control Message(StreamBegin) ---|
    |                                             |
    |<------------ Command Message ---------------|
    |       (\_result- connect response)          |
    |                                             |


执行命令期间的消息流为：

1.客户端将connect命令发送到服务器，以请求与服务器应用程序实例连接。

2.收到连接命令后，服务器将协议消息“窗口确认大小”发送给客户端。 服务器还连接到connect命令中提到的应用程序。

3.服务器将协议消息“设置对等带宽”发送给客户端。

4.客户端在处理了协议消息“设置对等带宽”后，将协议消息“窗口确认大小”发送到服务器。

5.服务器将另一种类型为User Control Message（StreamBegin）的协议消息发送到客户端。

6.服务器发送结果命令消息，通知客户端连接状态（成功/失败）。 该命令指定事务ID（对于connect命令，始终等于1）服务器版本（字符串）。 此外，它还规范了其他与连接响应有关的信息，例如级别（字符串），代码（字符串），描述（字符串），对象编码（数字）等。


### Call 

NetConnection对象的调用方法在接收端运行远程过程调用（RPC）。 调用的RPC名称作为参数传递给call命令。

从发送者到接收者的命令结构如下：

+--------------+----------+----------------------------------------+
|Field Name    |   Type   |             Description                |
+--------------+----------+----------------------------------------+
| Procedure    |  String  | Name of the remote procedure that is   |
| Name         |          | called.                                |
+--------------+----------+----------------------------------------+
| Transaction  |  Number  | If a response is expected we give a    |
|              |          | transaction Id. Else we pass a value of|
| ID           |          | 0                                      |
+--------------+----------+----------------------------------------+
| Command      |  Object  | If there exists any command info this  |
| Object       |          | is set, else this is set to null type. |
+--------------+----------+----------------------------------------+
| Optional     |  Object  | Any optional arguments to be provided  |
| Arguments    |          |                                        |
+--------------+----------+----------------------------------------+

响应的命令结构如下：

+--------------+----------+----------------------------------------+
| Field Name   |   Type   |             Description                |
+--------------+----------+----------------------------------------+
| Command Name |  String  | Name of the command.                   |
|              |          |                                        |
+--------------+----------+----------------------------------------+
| Transaction  |  Number  | ID of the command, to which the        |
| ID           |          | response belongs.                      |
+--------------+----------+----------------------------------------+
| Command      |  Object  | If there exists any command info this  |
| Object       |          | is set, else this is set to null type. |
+--------------+----------+----------------------------------------+
| Response     | Object   | Response from the method that was      |
|              |          | called.                                |
+------------------------------------------------------------------+

### createStream 

客户端将此命令发送到服务器以创建用于消息通信的逻辑通道。音频，视频和元数据的发布是通过使用createStream命令创建的流通道进行的。

NetConnection是默认通信通道，其流ID为0。协议和一些命令消息（包括createStream）使用默认通信通道。

从客户端到服务器的命令结构如下：

+--------------+----------+----------------------------------------+
| Field Name   |   Type   |             Description                |
+--------------+----------+----------------------------------------+
| Command Name |  String  | Name of the command. Set to            |
|              |          | "createStream".                        |
+--------------+----------+----------------------------------------+
| Transaction  |  Number  | Transaction ID of the command.         |
| ID           |          |                                        |
+--------------+----------+----------------------------------------+
| Command      |  Object  | If there exists any command info this  |
| Object       |          | is set, else this is set to null type. |
+--------------+----------+----------------------------------------+

从服务器到客户端的命令结构如下：

+--------------+----------+----------------------------------------+
| Field Name   |   Type   |             Description                |
+--------------+----------+----------------------------------------+
| Command Name |  String  | \_result or \_error; indicates whether |
|              |          | the response is result or error.       |
+--------------+----------+----------------------------------------+
| Transaction  |  Number  | ID of the command that response belongs|
| ID           |          | to.                                    |
+--------------+----------+----------------------------------------+
| Command      |  Object  | If there exists any command info this  |
| Object       |          | is set, else this is set to null type. |
+--------------+----------+----------------------------------------+
| Stream       |  Number  | The return value is either a stream ID |
| ID           |          | or an error information object.        |
+--------------+----------+----------------------------------------+




## NetStream Commands

NetStream定义了流音频，视频和数据消息可以通过该通道流过将客户端连接到服务器的NetConnection的通道。 一个NetConnection对象可以为多个数据流支持多个NetStreams。

客户端可以在NetStream上向服务器发送以下命令：

 o  play
 o  play2
 o  deleteStream
 o  closeStream
 o  receiveAudio
 o  receiveVideo
 o  publish
 o  seek
 o  pause

 服务器使用“ onStatus”命令将NetStream状态更新发送给客户端：

 +--------------+----------+----------------------------------------+
 | Field Name   |   Type   |             Description                |
 +--------------+----------+----------------------------------------+
 | Command Name |  String  | The command name "onStatus".           |
 +--------------+----------+----------------------------------------+
 | Transaction  |  Number  | Transaction ID set to 0.               |
 | ID           |          |                                        |
 +--------------+----------+----------------------------------------+
 | Command      |  Null    | There is no command object for         |
 | Object       |          | onStatus messages.                     |
 +--------------+----------+----------------------------------------+
 | Info Object  | Object   | An AMF object having at least the      |
 |              |          | following three properties: "level"    |
 |              |          | (String): the level for this message,  |
 |              |          | one of "warning", "status", or "error";|
 |              |          | "code" (String): the message code, for |
 |              |          | example "NetStream.Play.Start"; and    |
 |              |          | "description" (String): a human-       |
 |              |          | readable description of the message.   |
 |              |          | The Info object MAY contain other      |
 |              |          | properties as appropriate to the code. |
 +--------------+----------+----------------------------------------+

 NetStream状态消息命令的格式。

### play

客户端将此命令发送到服务器以播放流。 也可以使用此命令多次创建播放列表。

如果要创建在不同的直播或录制的流之间切换的动态播放列表，请多次调用播放并传递false进行重置。 相反，如果要立即播放指定的流，清除排队等待播放的任何其他流，请传递true进行重置。

从客户端到服务器的命令结构如下：

+--------------+----------+-----------------------------------------+
| Field Name   |   Type   |             Description                 |
+--------------+----------+-----------------------------------------+
| Command Name |  String  | Name of the command. Set to "play".     |
+--------------+----------+-----------------------------------------+
| Transaction  |  Number  | Transaction ID set to 0.                |
| ID           |          |                                         |
+--------------+----------+-----------------------------------------+
| Command      |   Null   | Command information does not exist.     |
| Object       |          | Set to null type.                       |
+--------------+----------+-----------------------------------------+
| Stream Name  |  String  | Name of the stream to play.             |
|              |          | To play video (FLV) files, specify the  |
|              |          | name of the stream without a file       |
|              |          | extension (for example, "sample"). To   |
|              |          | play back MP3 or ID3 tags, you must     |
|              |          | precede the stream name with mp3:       |
|              |          | (for example, "mp3:sample". To play     |
|              |          | H.264/AAC files, you must precede the   |
|              |          | stream name with mp4: and specify the   |
|              |          | file extension. For example, to play the|
|              |          | file sample.m4v,specify "mp4:sample.m4v"|
|              |          |                                         |
+--------------+----------+-----------------------------------------+
| Start        |  Number  | An optional parameter that specifies    |
|              |          | the start time in seconds. The default  |
|              |          | value is -2, which means the subscriber |
|              |          | first tries to play the live stream     |
|              |          | specified in the Stream Name field. If a|
|              |          | live stream of that name is not found,it|
|              |          | plays the recorded stream of the same   |
|              |          | name. If there is no recorded stream    |
|              |          | with that name, the subscriber waits for|
|              |          | a new live stream with that name and    |
|              |          | plays it when available. If you pass -1 |
|              |          | in the Start field, only the live stream|
|              |          | specified in the Stream Name field is   |
|              |          | played. If you pass 0 or a positive     |
|              |          | number in the Start field, a recorded   |
|              |          | stream specified in the Stream Name     |
|              |          | field is played beginning from the time |
|              |          | specified in the Start field. If no     |
|              |          | recorded stream is found, the next item |
|              |          | in the playlist is played.              |
+--------------+----------+-----------------------------------------+
| Duration     |  Number  | An optional parameter that specifies the|
|              |          | duration of playback in seconds. The    |
|              |          | default value is -1. The -1 value means |
|              |          | a live stream is played until it is no  |
|              |          | longer available or a recorded stream is|
|              |          | played until it ends. If you pass 0, it |
|              |          | plays the single frame since the time   |
|              |          | specified in the Start field from the   |
|              |          | beginning of a recorded stream. It is   |
|              |          | assumed that the value specified in     |
|              |          | the Start field is equal to or greater  |
|              |          | than 0. If you pass a positive number,  |
|              |          | it plays a live stream for              |
|              |          | the time period specified in the        |
|              |          | Duration field. After that it becomes   |
|              |          | available or plays a recorded stream    |
|              |          | for the time specified in the Duration  |
|              |          | field. (If a stream ends before the     |
|              |          | time specified in the Duration field,   |
|              |          | playback ends when the stream ends.)    |
|              |          | If you pass a negative number other     |
|              |          | than -1 in the Duration field, it       |
|              |          | interprets the value as if it were -1.  |
+--------------+----------+-----------------------------------------+
| Reset        | Boolean  | An optional Boolean value or number     |
|              |          | that specifies whether to flush any     |
|              |          | previous playlist.                      |
+--------------+----------+-----------------------------------------+

播放命令中的消息流

         +-------------+                            +----------+
         | Play Client |             |              |   Server |
         +------+------+             |              +-----+----+
                |        Handshaking and Application       |
                |             connect done                 |
                |                    |                     |
                |                    |                     |
                |                    |                     |
                |                    |                     |
       ---+---- |------Command Message(createStream) ----->|
    Create|     |                                          |
    Stream|     |                                          |
       ---+---- |<---------- Command Message --------------|
                |     (_result- createStream response)     |
                |                                          |
       ---+---- |------ Command Message (play) ----------->|
          |     |                                          |
          |     |<------------  SetChunkSize --------------|
          |     |                                          |
          |     |<---- User Control (StreamIsRecorded) ----|
      Play|     |                                          |
          |     |<---- UserControl (StreamBegin) ----------|
          |     |                                          |
          |     |<--Command Message(onStatus-play reset) --|
          |     |                                          |
          |     |<--Command Message(onStatus-play start) --|
          |     |                                          |
          |     |<-------------Audio Message---------------|
          |     |                                          |
          |     |<-------------Video Message---------------|
          |     |                    |                     |                 
                                     |
                  继续接收音频和视频流，直到完成播放命令中的消息流
 
执行命令期间的消息流为：

1.客户端从服务器成功接收到createStream命令的结果后，发送play命令。

2.服务器在接收到play命令后，发送协议消息以设置块大小。

3.服务器发送另一条协议消息（用户控件），以指定事件“ StreamIsRecorded”和该消息中的流ID。 该消息的前2个字节中包含事件类型，后4个字节中包含流ID。

4.服务器发送另一条协议消息（用户控件），该消息指定事件“ StreamBegin”，以向客户端指示流媒体的开始。

5.如果客户端发送的播放命令成功，则服务器将发送onStatus命令消息NetStream.Play.Start和NetStream.Play.Reset。 仅当客户端发送的播放命令设置了重置标志时，服务器才发送NetStream.Play.Reset。 如果找不到要播放的流，则服务器发送onStatus消息NetStream.Play.StreamNotFound。

此后，服务器发送音频和视频数据，客户端播放

### play2

与play命令不同，play2可以切换到其他比特率流，而无需更改播放内容的时间轴。 服务器维护客户端可以在play2中请求的所有受支持的比特率的多个文件。

从客户端到服务器的命令结构如下：

+--------------+----------+----------------------------------------+
| Field Name   |   Type   |             Description                |
+--------------+----------+----------------------------------------+
| Command Name |  String  | Name of the command, set to "play2".   |
+--------------+----------+----------------------------------------+
| Transaction  |  Number  | Transaction ID set to 0.               |
| ID           |          |                                        |
+--------------+----------+----------------------------------------+
| Command      |   Null   | Command information does not exist.    |
| Object       |          | Set to null type.                      |
+--------------+----------+----------------------------------------+
| Parameters   |  Object  | An AMF encoded object whose properties |
|              |          | are the public properties described    |
|              |          | for the flash.net.NetStreamPlayOptions |
|              |          | ActionScript object.                   |
+--------------+----------+----------------------------------------+

NetStreamPlayOptions对象的公共属性在《 ActionScript 3语言参考》 [AS3]中进行了描述。

下图显示了该命令的消息流。

           +--------------+                          +-------------+           
           | Play2 Client |              |           |    Server   |
           +--------+-----+              |           +------+------+
                    |      Handshaking and Application      |
                    |               connect done            |
                    |                    |                  |
                    |                    |                  |
                    |                    |                  |
                    |                    |                  |
           ---+---- |---- Command Message(createStream) --->|
       Create |     |                                       |
       Stream |     |                                       |
           ---+---- |<---- Command Message (_result) -------|
                    |                                       |
           ---+---- |------ Command Message (play) -------->|
              |     |                                       |
              |     |<------------ SetChunkSize ------------|
              |     |                                       |
              |     |<--- UserControl (StreamIsRecorded)----|
         Play |     |                                       |
              |     |<------- UserControl (StreamBegin)-----|
              |     |                                       |
              |     |<--Command Message(onStatus-playstart)-|
              |     |                                       |
              |     |<---------- Audio Message -------------|
              |     |                                       |
              |     |<---------- Video Message -------------|
              |     |                                       |
                    |                                       |
           ---+---- |-------- Command Message(play2) ------>|
              |     |                                       |
              |     |<------- Audio Message (new rate) -----|
        Play2 |     |                                       |
              |     |<------- Video Message (new rate) -----|
              |     |                    |                  |
              |     |                    |                  |
              |  Keep receiving audio and video stream till finishes
                                         |
                     Message flow in the play2 command

### deleteStream 

当NetStream对象被销毁时，NetStream发送deleteStream命令

从客户端到服务器的命令结构如下：

    +--------------+----------+----------------------------------------+
    | Field Name   |   Type   |             Description                |
    +--------------+----------+----------------------------------------+
    | Command Name |  String  | Name of the command, set to            |
    |              |          | "deleteStream".                        |
    +--------------+----------+----------------------------------------+
    | Transaction  |  Number  | Transaction ID set to 0.               |
    | ID           |          |                                        |
    +--------------+----------+----------------------------------------+
    | Command      |  Null    | Command information object does not    |
    | Object       |          | exist. Set to null type.               |
    +--------------+----------+----------------------------------------+
    | Stream ID    |  Number  | The ID of the stream that is destroyed |
    |              |          | on the server.                         |
    +--------------+----------+----------------------------------------+

服务器不发送任何响应。

### receiveAudio

NetStream发送receiveAudio消息来通知服务器是否将音频发送到客户端。

从客户端到服务器的命令结构如下：

    +--------------+----------+----------------------------------------+
    | Field Name   |   Type   |             Description                |
    +--------------+----------+----------------------------------------+
    | Command Name |  String  | Name of the command, set to            |
    |              |          | "receiveAudio".                        |
    +--------------+----------+----------------------------------------+
    | Transaction  |  Number  | Transaction ID set to 0.               |
    | ID           |          |                                        |
    +--------------+----------+----------------------------------------+
    | Command      |  Null    | Command information object does not    |
    | Object       |          | exist. Set to null type.               |
    +--------------+----------+----------------------------------------+
    | Bool Flag    |  Boolean | true or false to indicate whether to   |
    |              |          | receive audio or not.                  |
    +--------------+----------+----------------------------------------+

如果在bool标志设置为false的情况下发送了receiveAudio命令，则服务器不会发送任何响应。 如果此标志设置为true，则服务器将以状态消息NetStream.Seek.Notify和NetStream.Play.Start进行响应。

### receiveVideo

NetStream发送receiveVideo消息来通知服务器是否将视频发送到客户端。 

从客户端到服务器的命令结构如下：

    +--------------+----------+----------------------------------------+
    | Field Name   |   Type   |             Description                |
    +--------------+----------+----------------------------------------+
    | Command Name |  String  | Name of the command, set to            |
    |              |          | "receiveVideo".                        |
    +--------------+----------+----------------------------------------+
    | Transaction  |  Number  | Transaction ID set to 0.               |
    | ID           |          |                                        |
    +--------------+----------+----------------------------------------+
    | Command      |  Null    | Command information object does not    |
    | Object       |          | exist. Set to null type.               |
    +--------------+----------+----------------------------------------+
    | Bool Flag    |  Boolean | true or false to indicate whether to   |
    |              |          | receive video or not.                  |
    +--------------+----------+----------------------------------------+

如果在bool标志设置为false的情况下发送了receiveVideo命令，则服务器不会发送任何响应。 如果此标志设置为true，则服务器将以状态消息NetStream.Seek.Notify和NetStream.Play.Start进行响应。


### publish

客户端发送publish命令将命名流发布到服务器。 使用此名称，任何客户端都可以播放此流并接收发布的音频，视频和数据消息。

从客户端到服务器的命令结构如下：

    +--------------+----------+----------------------------------------+
    | Field Name   |   Type   |             Description                |
    +--------------+----------+----------------------------------------+
    | Command Name |  String  | Name of the command, set to "publish". |
    +--------------+----------+----------------------------------------+
    | Transaction  |  Number  | Transaction ID set to 0.               |
    | ID           |          |                                        |
    +--------------+----------+----------------------------------------+
    | Command      |  Null    | Command information object does not    |
    | Object       |          | exist. Set to null type.               |
    +--------------+----------+----------------------------------------+
    | Publishing   |  String  | Name with which the stream is          |
    | Name         |          | published.                             |
    +--------------+----------+----------------------------------------+
    | Publishing   |  String  | Type of publishing. Set to "live",     |
    | Type         |          | "record", or "append".                 |
    |              |          | record: The stream is published and the|
    |              |          | data is recorded to a new file.The file|
    |              |          | is stored on the server in a           |
    |              |          | subdirectory within the directory that |
    |              |          | contains the server application. If the|
    |              |          | file already exists, it is overwritten.|
    |              |          | append: The stream is published and the|
    |              |          | data is appended to a file. If no file |
    |              |          | is found, it is created.               |
    |              |          | live: Live data is published without   |
    |              |          | recording it in a file.                |
    +--------------+----------+----------------------------------------+

服务器使用onStatus命令进行响应，以标记发布的开始。

### seek 

客户端发送搜索命令以在媒体文件或播放列表中搜索偏移量（以毫秒为单位）。

从客户端到服务器的命令结构如下：

    +--------------+----------+----------------------------------------+
    | Field Name   |   Type   |             Description                |
    +--------------+----------+----------------------------------------+
    | Command Name |  String  | Name of the command, set to "seek".    |
    +--------------+----------+----------------------------------------+
    | Transaction  |  Number  | Transaction ID set to 0.               |
    | ID           |          |                                        |
    +--------------+----------+----------------------------------------+
    | Command      |  Null    | There is no command information object |
    | Object       |          | for this command. Set to null type.    |
    +--------------+----------+----------------------------------------+
    | milliSeconds |  Number  | Number of milliseconds to seek into    |
    |              |          | the playlist.                          |
    +--------------+----------+----------------------------------------+

查找成功时，服务器将发送状态消息NetStream.Seek.Notify。 如果失败，它将返回_error消息。


### pause

客户端发送暂停命令告诉服务器暂停或开始播放。

从客户端到服务器的命令结构如下：

    +--------------+----------+----------------------------------------+
    | Field Name   |   Type   |             Description                |
    +--------------+----------+----------------------------------------+
    | Command Name |  String  | Name of the command, set to "pause".   |
    +--------------+----------+----------------------------------------+
    | Transaction  |  Number  | There is no transaction ID for this    |
    | ID           |          | command. Set to 0.                     |
    +--------------+----------+----------------------------------------+
    | Command      |  Null    | Command information object does not    |
    | Object       |          | exist. Set to null type.               |
    +--------------+----------+----------------------------------------+
    |Pause/Unpause |  Boolean | true or false, to indicate pausing or  |
    | Flag         |          | resuming play                          |
    +--------------+----------+----------------------------------------+
    | milliSeconds |  Number  | Number of milliseconds at which the    |
    |              |          | the stream is paused or play resumed.  |
    |              |          | This is the current stream time at the |
    |              |          | Client when stream was paused. When the|
    |              |          | playback is resumed, the server will   |
    |              |          | only send messages with timestamps     |
    |              |          | greater than this value.               |
    +--------------+----------+----------------------------------------+

服务器在流暂停时发送状态消息NetStream.Pause.Notify。 流处于未暂停状态时发送NetStream.Unpause.Notify。 如果失败，它将返回_error消息。


## Message Exchange Examples

以下是一些示例，说明使用RTMP进行消息交换。

### Publish Recorded Video 

此示例说明了发布者如何发布流，然后将视频流传输到服务器。 其他客户可以订阅此发布的流并播放视频。


            +--------------------+                     +-----------+
            |  Publisher Client  |        |            |    Server |
            +----------+---------+        |            +-----+-----+
                       |           Handshaking Done          |
                       |                  |                  |
                       |                  |                  |
              ---+---- |----- Command Message(connect) ----->|
                 |     |                                     |
                 |     |<----- Window Acknowledge Size ------|
         Connect |     |                                     |
                 |     |<-------Set Peer BandWidth ----------|
                 |     |                                     |
                 |     |------ Window Acknowledge Size ----->|
                 |     |                                     |
                 |     |<------User Control(StreamBegin)-----|
                 |     |                                     |
              ---+---- |<---------Command Message -----------|
                       |   (_result- connect response)       |
                       |                                     |
              ---+---- |--- Command Message(createStream)--->|
          Create |     |                                     |
          Stream |     |                                     |
              ---+---- |<------- Command Message ------------|
                       | (_result- createStream response)    |
                       |                                     |
              ---+---- |---- Command Message(publish) ------>|
                 |     |                                     |
                 |     |<------User Control(StreamBegin)-----|
                 |     |                                     |
                 |     |-----Data Message (Metadata)-------->|
                 |     |                                     |
       Publishing|     |------------ Audio Data ------------>|
         Content |     |                                     |
                 |     |------------ SetChunkSize ---------->|
                 |     |                                     |
                 |     |<----------Command Message ----------|
                 |     |      (_result- publish result)      |
                 |     |                                     |
                 |     |------------- Video Data ----------->|
                 |     |                  |                  |
                 |     |                  |                  |
                       |    Until the stream is complete     |
                       |                  |                  |
                     Message flow in publishing a video stream 



### Broadcast a Shared Object Message

此示例说明了在创建和更改共享库期间交换的消息。 它还说明了共享对象消息广播的过程。

                    +----------+                       +----------+                    
                    |  Client  |           |           |  Server  |
                    +-----+----+           |           +-----+----+
                          |   Handshaking and Application    |
                          |          connect done            |
                          |                |                 |
                          |                |                 |
                          |                |                 |
                          |                |                 |
      Create and ---+---- |---- Shared Object Event(Use)---->|
      connect       |     |                                  |
      Shared Object |     |                                  |
                 ---+---- |<---- Shared Object Event---------|
                          |       (UseSuccess,Clear)         |
                          |                                  |
                 ---+---- |------ Shared Object Event ------>|
      Shared object |     |         (RequestChange)          |
      Set Property  |     |                                  |
                 ---+---- |<------ Shared Object Event ------|
                          |            (Success)             |
                          |                                  |
                 ---+---- |------- Shared Object Event ----->|
       Shared object|     |           (SendMessage)          |
       Message      |     |                                  |
       Broadcast ---+---- |<------- Shared Object Event -----|
                          |           (SendMessage)          |
                                            |
                                            |
                              Shared object message broadcast



### Publish Metadata from Recorded Stream

本示例描述了用于发布元数据的消息交换。

              +------------------+                       +---------+
              | Publisher Client |         |             |   FMS   |
              +---------+--------+         |             +----+----+
                        |     Handshaking and Application     |
                        |            connect done             |
                        |                  |                  |
                        |                  |                  |
                ---+--- |---Command Messsage(createStream) -->|
            Create |    |                                     |
            Stream |    |                                     |
                ---+--- |<---------Command Message------------|
                        |   (_result - command response)      |
                        |                                     |
                ---+--- |---- Command Message(publish) ------>|
        Publishing |    |                                     |
          metadata |    |<------ UserControl(StreamBegin)-----|
         from file |    |                                     |
                   |    |-----Data Message (Metadata) ------->|
                        |                                     |
                                        |
                                Publishing metadata