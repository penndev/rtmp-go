package rtmp

import (
	"errors"
	"fmt"
	"log"
	"rtmp-go/amf"
)

type Pack struct {
	ChunkMessageHeader
	PayLoad []byte
}

func callConnect(pk Pack, client func(Pack)) {
	client(pk)
}

func callClose(close func()) {
	close()
}

func netConnectionCommand(chk *Chunk, conn *Conn) error {
	read := 0
	for {
		pk, err := chk.handlesMsg()
		if err != nil {
			return err
		}
		if pk.MessageTypeID != 20 {
			return errors.New("netConnectionCommand err: cant handle type id" + fmt.Sprint(pk.MessageTypeID))
		}
		item := amf.Decode(pk.PayLoad)
		switch item[0] {
		case "connect":
			read = 1
			media, ok := item[2].(map[string]amf.Value)
			if !ok {
				return errors.New("netConnectionCommand connect err:) catn find media")
			}
			app, ok := media["app"].(string)
			if !ok {
				return errors.New("netConnectionCommand connect err:) cant find app")
			}
			stu := conn.onConnect(app)
			chk.setChunkSize(SetChunkSize)
			chk.sendMsg(20, 3, respConnect(stu))
			if !stu {
				return errors.New("netConnectionCommand connect err:) cat conntect app " + app)
			}
			chk.setWindowAcknowledgementSize(2500000)
		case "createStream":
			tranId, ok := item[1].(float64)
			if !ok {
				return errors.New("netConnectionCommand createStream err:) cant find tranid")
			}
			chk.sendMsg(20, 3, respCreateStream(true, int(tranId), DefaultStreamID))
			if read == 1 {
				read = 2
			} else {
				return errors.New("netConnectionCommand err:) not do connect action")
			}
		case "releaseStream":
		case "FCPublish":
		default:
			log.Println("netConnectionCommand err: cant handle command->", item[0])
		}
		if read == 2 {
			break
		}
	}
	return nil
}

func netStreamCommand(chk *Chunk, conn *Conn) error {
	for {
		pk, err := chk.handlesMsg()
		if err != nil {
			return err
		}
		if pk.MessageTypeID != 20 {
			return errors.New("netStreamCommand err: cant handle type id" + fmt.Sprint(pk.MessageTypeID))
		}
		item := amf.Decode(pk.PayLoad)
		switch item[0] {
		case "publish":
			streamId, ok := item[1].(float64)
			if !ok {
				return errors.New("netStreamCommand err: streamId error")
			}
			streamType, ok := item[4].(string)
			if !ok || streamType != "live" {
				return errors.New("netStreamCommand err: streamType error")
			}
			streamName, ok := item[3].(string)
			if !ok {
				return errors.New("netStreamCommand err: streamName error")
			}
			status := conn.onPublish(streamName)
			chk.sendMsg(20, 3, respPublish(status))
			if !status {
				return errors.New("netStreamCommand err: streamname checkout fail")
			}
			chk.setStreamBegin(uint32(streamId))
			return nil
		case "play":
			streamName, ok := item[3].(string)
			if !ok {
				return errors.New("netStreamCommand play err: streamName error")
			}
			status := conn.onPlay(streamName)
			chk.sendMsg(20, 3, respPlay(status))
			if !status {
				return errors.New("netStreamCommand play err: streamname checkout fail")
			}
			return nil
		}

	}
}

//阻塞处理 AVPackChan 收到的消息。
func netHandleCommand(chk *Chunk, conn *Conn, app *App) error {
	go handle(chk, conn)
	var client func(Pack)
	var close func()
	if conn.IsPublish {
		stream := app.addPublish(conn.App, conn.Stream)
		readyIng := 0
		client = func(pk Pack) {
			if readyIng < 7 { //初始化视频关键(解码)信息。
				stream.setMeta(pk, &readyIng)
				return
			}
			stream.setPack(pk)
		}
		close = func() {
			// 这里不设置状态。
			app.delPublish(conn.App, conn.Stream)
		}
	} else {
		stream, ok := app.addPlay(conn.App, conn.Stream, conn.AVPackChan) // 初始化流不存在。
		if !ok {
			log.Println("Play stream not found:", conn.App, conn.Stream)
			conn.IsStoped = true //禁止下面阻塞读
		}
		readyIng := 0
		client = func(pk Pack) {
			// 必须初始化关键帧。====
			if readyIng == 0 { //初始化视频关键(解码)信息。
				stream.getMeta(chk)
				readyIng = 1
			}
			chk.sendPack(DefaultStreamID, pk)
		}
		close = func() {
			chk.setStreamEof(DefaultStreamID)
			app.delPlay(conn.App, conn.Stream, conn.AVPackChan)
			if !conn.IsStoped { //如果是推送端关闭的。
				conn.IsStoped = true // 防止read协程再次close PackChan。
			}
		}
	}
	// 如果流已经被停止了
	// 或者客户端直接断开连接了
	// 则直接停止当前线程。
	if !conn.IsStoped {
		for avpack := range conn.AVPackChan {
			callConnect(avpack, client)
		}
	}
	callClose(close)
	return nil
}

// 阻塞读chunk消息。
func handle(chk *Chunk, conn *Conn) {
	for {
		pk, err := chk.handlesMsg()
		if err != nil {
			//如果是客户端主动关闭,标记状态，并通知chan对端。
			if !conn.IsStoped {
				conn.IsStoped = true
				// 如果是推送端关闭了
				close(conn.AVPackChan)
			}
			return
		}
		switch pk.MessageTypeID {
		case 8, 9, 15, 18:
			//不允许向已关闭的chan传输数据。
			if conn.IsStoped {
				return
			}
			conn.AVPackChan <- pk
		case 20:
			item := amf.Decode(pk.PayLoad)
			switch item[0] {
			case "FCUnpublish":
			case "deleteStream":
				// 主动关闭
				conn.IsStoped = true
				close(conn.AVPackChan)
				conn.onClose()
				return
			default:
				log.Println("未遇到的消息(type 20)->", item[0])
			}
		default:
			log.Println("未遇到的type->", pk.MessageTypeID)
		}
	}
}
