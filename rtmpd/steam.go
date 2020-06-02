package rtmp

import (
	"log"

	"github.com/pennfly/rtmp-go/av"
)

//Steam 处理流数据
func (c *Connnect) Steam() error {
	var flv av.FLV
	flv.GenFlv()
	// 处理消息
	var err error
	var VideoTime uint32
	for {
		msg, err := c.ReadMsg()
		if err != nil {
			break
		}

		switch mType := int(msg.MessageTypeID); mType {
		case 20, 17: //AMF-encoded commands
			CommandMessage(&msg)
		case 18, 15: // Metadata
			flv.AddTag(mType, 0, msg.Payload[16:])
		case 8, 9: // Video data
			VideoTime += msg.Timestamp
			log.Println("Video time --- ", msg.Timestamp, VideoTime)
			flv.AddTag(mType, VideoTime, msg.Payload)
		// 	log.Println("Video time --- ", msg.Timestamp, VideoTime)
		// 	// flv.AddTag(mType, VideoTime, msg.Payload)
		// case 8: // Audio data
		// 	AudioTime += msg.Timestamp
		// 	log.Println("Audio time", msg.Timestamp, AudioTime)
		// 	// flv.AddTag(mType, AudioTime, msg.Payload)
		default:
			//log.Println("Rtmp Msg: [typeid,lengthe]", msg.MessageTypeID, msg.MessageLength)
			log.Println("Oop had cant case Message type id:", mType)
		}
	}
	flv.Close()
	return err
}
