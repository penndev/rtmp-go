package main

import (
	"bytes"
	"log"
	"net"
)

func handle1(conn net.Conn) {
	defer conn.Close()

	chunkSteamHandshake(conn)

	chunkSteamControMessage(conn)

}

func checkErr(err error) {

	if err != nil {
		log.Fatal(err)
	}

}

func ctrl(chunk chunk, buf *bytes.Reader, conn net.Conn) {

}

func chunkSteamControMessage(buf net.Conn) {
	var chunk chunk
	for {
		chunk.initialization(buf)

		if chunk.messagetypeid == 20 || chunk.messagetypeid == 18 {
			CommandAMF0(chunk.messagebody, buf)
		} else {
			log.Println("[had other chunk type]", chunk.fmt, chunk.csid, chunk.messagetypeid, chunk.messagelength)
		}
	}

}
