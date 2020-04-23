package rtmp

import "log"

//Steam 处理流数据
func (c *Connnect) Steam() error {
	// 处理消息
	for {
		msg, err := c.ReadMsg()
		if err != nil {
			return err
		}

		switch mType := int(msg.MessageTypeID); mType {
		case 20, 17: //AMF-encoded commands
			CommandMessage(&msg)
		case 18, 15: // Metadata
		case 9: // Video data
		case 8: // Audio data
		default:
			log.Println("Oop had cant case Message type id:", mType)
		}
	}
}
