package rtmp

import (
	"log"

	"github.com/penndev/rtmp/amf"
	"github.com/penndev/rtmp/pubsub"
)

// publish: reply + Stream Begin, then fan-out AV via pubsub until publisher closes
func (srv *Serve) handlePublish(conn *Conn) error {
	if err := conn.PublishReply(true); err != nil {
		return err
	}
	if err := conn.StreamBegin(uint32(conn.StreamID)); err != nil {
		return err
	}
	path := conn.App + "/" + conn.Stream
	log.Printf(
		"%s publishing path=%s conn.App=%s conn.Stream=%s",
		conn.nc.RemoteAddr(), path, conn.App, conn.Stream)
	topic := pubsub.PubTopic(path)

	defer topic.Close()
	// flv to runtime file
	go AdapterFlv(path, topic.Subscribe())

	for {
		msg, err := conn.Read()
		if err != nil {
			return err
		}
		switch msg.MessageType {
		case Audio, Video:
			topic.Publish(&pubsub.Message{Data: msg})
		case AMF0DataMessage, AMF3DataMessage:
			topic.Publish(&pubsub.Message{Data: msg})
		case AMF0CommandMessage, AMF3CommandMessage:
			var values []amf.Value
			switch msg.MessageType {
			case AMF0CommandMessage:
				values, err = amf.Decode0(msg.PayLoad)
			case AMF3CommandMessage:
				values, err = amf.Decode3(msg.PayLoad)
			}
			if err != nil {
				return err
			}
			if len(values) == 0 {
				continue
			}
			cmd, _ := values[0].(string)
			switch cmd {
			case "FCUnpublish", "deleteStream":
				return nil
			default:
				log.Printf("publish command ignored: %s", cmd)
			}
		default:
			log.Printf("publish ignore type=%d len=%d", msg.MessageType, len(msg.PayLoad))
		}
	}
}
