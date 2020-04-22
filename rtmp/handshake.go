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
	buf, err := c.Read(1537)
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
	S2 := buf[1:1537]
	S012 := append(buf, S2...)
	c.WriteBuffer(S012)

	// C2->
	_, err = c.Read(1536)
	return err
}
