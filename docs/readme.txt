第一步，握手 handshake
说明：




第二步，进行 connect

=====================================================================

        +--------------+                +-------------+
        | Client       |        |       |    Server   |
        +------+-------+        |       +------+------+
        |                Handshaking done             |
        |                       |                     |
        |                       |                     |
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
        |       (_result- connect response)           |
        | 

======================================================================
                                       
说明：




第三步，传输流

传输流推送端整体流程

=======================================================================

            +--------------------+        +-----------+
            | Publisher Client |     |       | Server |
            +----------+---------+   |    +-----+-----+
                |            Handshaking Done         |
                |                    |                |
                |                    |                |
        ---+----|-----  Command Message(connect)----->|
                |                    |                |
              | |<----- Window Acknowledge Size ------|
     Connect  | |                                     |
              | |<-------Set Peer BandWidth ----------|
              | |                                     |
              | |------ Window Acknowledge Size ----->|
              | |                                     |
              | |<------User Control(StreamBegin)-----|
              | |                                     |
       ---+---- |<---------Command Message -----------|
                |     (_result- connect response)     |
                |                                     |
       ---+---- |--- Command Message(createStream)--->|
       Create | |                                     |
       Stream | |                                     |
       ---+---- |<------- Command Message ------------|
                |   (_result- createStream response)  |
                |                                     |
       ---+---- |---- Command Message(publish) ------>|
              | |                                     |
              | |<------User Control(StreamBegin)-----|
              | |                                     |
              | |-----Data Message (Metadata)-------->|
              | |                                     |
    Publishing| |------------ Audio Data ------------>|
      Content | |                                     |
              | |------------ SetChunkSize ---------->|
              | |                                     |
              | |<----------Command Message ----------|
              | |      (_result- publish result)      |
              | |                                     |
              | |------------- Video Data ----------->|
              | |                                     |
              | |                                     |
              | |    Until the stream is complete     |
              | |                                     |
             Message flow in publishing a video stream| 

=======================================================================