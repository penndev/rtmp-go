package rtmp

import (
	"log"
	"net"
)

//rtmp 握手
func chunkSteamHandshake(c net.Conn) (err error) {
	buf := make([]byte, 1537)
	_, err = c.Read(buf)
	if err != nil {
		return err
	}
	// C0+C1->
	if buf[0] != 3 {
		log.Println("rtmp version is not right!")
	} else {
		// <-S0+S1+S2
		mg := buf[0:]
		for _, value := range buf[1:1537] {
			mg = append(mg, value)
		}
		c.Write(mg)
	}
	// throw C2->

	c2 := make([]byte, 1536)
	_, err = c.Read(c2)
	return err
}
