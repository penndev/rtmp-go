package main

import (
	"log"
	"net"
)

func main() {
	listen, err := net.Listen("tcp", ":1935")
	if err != nil {
		log.Fatal(err)
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		handle1(conn)
	}
}

// func handle(conn net.Conn) {
// 	defer conn.Close()

// 	log.Println("new client ---> ", conn.RemoteAddr().String())

// 	chunkSteamHandshake(conn)

// 	log.Println("client Handshake finish !")

// 	var chunkSteam chunk
// 	for {

// 		chunkSteam.initialization(conn)

// 		bufflen := make([]byte, chunkSteam.messagelength)
// 		len, err := conn.Read(bufflen)
// 		if err != nil {
// 			log.Fatal(len, err)
// 		}

// 		bufflen = bufflen[:len]
// 		if chunkSteam.messagelength != len {
// 			log.Println("want and had not same", chunkSteam.messagelength, len)
// 		}

// 		if chunkSteam.messagetypeid == 20 {

// 			CommandAMF0(bufflen, conn)

// 		} else if chunkSteam.messagetypeid == 18 {
// 			bufflenTmp := make([]byte, 1024)
// 			len, _ := conn.Read(bufflenTmp)

// 			log.Println(len,bufflenTmp, chunkSteam.messagetypeid, chunkSteam.messagelength)
// 			Data(bufflen)
// 		} else {
// 			log.Println("[had other chunk type]", chunkSteam.messagetypeid, chunkSteam.messagelength)
// 		}

// 	}
// }

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
