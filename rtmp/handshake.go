package rtmp

import (
	"encoding/binary"
	"errors"
)

// Handshake Order
// C0 + C1 -> Server
// S0 + S1 + S2 ->Client
// C2 -> Server

//Handshake rtmp 简单握手流程
func Handshake(c *Connnect) error {
	buf, err := c.ReadBuffer(1537)
	if err != nil {
		return err
	}
	// RTMP VERSION == 3
	if buf[0] != Version {
		return errors.New("rtmp version is not support")
	}

	//判断是否 Fill Zero
	if binary.BigEndian.Uint32(buf[5:9]) != 0 {
		return errors.New("rtmp complex handshake not support")
	}

	// <-S0+S1+S2
	S01T := append(buf, buf[1:5]...)
	S2 := append(buf[1:5], buf[9:1537]...)
	S012 := append(S01T, S2...)
	c.WriteBuffer(S012)

	// C2->
	_, err = c.ReadBuffer(1536)
	return err
}
