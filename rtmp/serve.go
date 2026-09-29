package rtmp

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/penndev/rtmp/amf"
)

type Serve struct {
	mu   sync.RWMutex
	Addr string
}

func (srv *Serve) handle(nc net.Conn) {
	log.Printf("rtmp handle %s", nc.RemoteAddr())
	defer func() {
		nc.Close()
		if err := recover(); err != nil {
			log.Printf("%s: %s", "recover: ", err)
		}
	}()

	conn := NewConn(nc)
	// check rtmp handshake
	if err := conn.Handshake(); err != nil {
		log.Printf("%s ServeHandShake fail err[%s]", nc.RemoteAddr(), err.Error())
		return
	}
	cmd, err := conn.Connect()
	if err != nil {
		log.Printf("%s handleStream fail err[%s]", nc.RemoteAddr(), err.Error())
		return
	}
	if err := conn.ConnectReply(cmd, true); err != nil {
		log.Printf("%s ConnectReply fail err[%s]", nc.RemoteAddr(), err.Error())
		return
	}

	if err := conn.SetWindowAcknowledgementSize(DEFAULT_WINDOW_ACK_SIZE); err != nil {
		log.Printf("%s SetWindowAcknowledgementSize fail err[%s]", nc.RemoteAddr(), err.Error())
		return
	}
	if err := conn.SetBandwidth(DEFAULT_PEER_BANDWIDTH); err != nil {
		log.Printf("%s SetBandwidth fail err[%s]", nc.RemoteAddr(), err.Error())
		return
	}
	if err := conn.SetChunkSize(PREFERRED_CHUNK_SIZE); err != nil {
		log.Printf("%s SetChunkSize fail err[%s]", nc.RemoteAddr(), err.Error())
		return
	}

	cs, err := conn.CreateStream()
	if err != nil {
		log.Printf("%s CreateStream fail err[%s]", nc.RemoteAddr(), err.Error())
		return
	}

	if err := conn.CreateStreamReply(cs, DEFAULT_STREAM_ID); err != nil {
		log.Printf("%s CreateStreamReply fail err[%s]", nc.RemoteAddr(), err.Error())
		return
	}
	log.Printf("stream=%s", conn.Stream)

	msg, err := conn.Read()
	if err != nil {
		log.Printf("%s Read fail err[%s]", nc.RemoteAddr(), err.Error())
		return
	}

	var values []amf.Value
	switch msg.MessageType {
	case AMF0CommandMessage:
		values, err = amf.Decode0(msg.PayLoad)
	case AMF3CommandMessage:
		values, err = amf.Decode3(msg.PayLoad)
	default:
		log.Printf("%s unexpected message type=%d", nc.RemoteAddr(), msg.MessageType)
		return
	}
	if err != nil {
		log.Printf("%s decode command fail err[%s]", nc.RemoteAddr(), err.Error())
		return
	}
	if len(values) == 0 {
		log.Printf("%s empty command message", nc.RemoteAddr())
		return
	}
	name, _ := values[0].(string)
	if len(values) >= 4 {
		if s, ok := values[3].(string); ok && s != "" && !strings.HasSuffix(conn.Stream, s) {
			conn.Stream += s
		}
	}

	switch name {
	case "publish":
		if err := srv.handlePublish(conn); err != nil {
			log.Printf("%s handlePublish fail err[%s]", nc.RemoteAddr(), err.Error())
		}
	case "play":
		if err := srv.handlePlay(conn); err != nil {
			log.Printf("%s handlePlay fail err[%s]", nc.RemoteAddr(), err.Error())
		}
	default:
		log.Printf("%s unknown netstream command: %s", nc.RemoteAddr(), name)
	}
}

// publish: reply onStatus + Stream Begin, then receive AV and print type
func (srv *Serve) handlePublish(conn *Conn) error {
	if err := conn.PublishReply(true); err != nil {
		return err
	}
	if err := conn.StreamBegin(uint32(conn.StreamID)); err != nil {
		return err
	}
	log.Printf("%s publishing stream=%s", conn.nc.RemoteAddr(), conn.Stream)

	for {
		msg, err := conn.Read()
		if err != nil {
			return err
		}
		switch msg.MessageType {
		case Audio:
			log.Printf("audio ts=%d len=%d", msg.Timestamp, len(msg.PayLoad))
		case Video:
			log.Printf("video ts=%d len=%d", msg.Timestamp, len(msg.PayLoad))
		case AMF0DataMessage, AMF3DataMessage:
			log.Printf("data type=%d ts=%d len=%d", msg.MessageType, msg.Timestamp, len(msg.PayLoad))
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
			case "FCUnpublish":
				fallthrough
			case "deleteStream":
				return fmt.Errorf("publisher closed: %s", cmd)
			default:
				log.Printf("publish command ignored: %s", cmd)
			}
		default:
			log.Printf("publish ignore type=%d len=%d", msg.MessageType, len(msg.PayLoad))
		}
	}
}

// play: finish play handshake, then wait (media send left empty)
func (srv *Serve) handlePlay(conn *Conn) error {
	if err := conn.PlayReply(true); err != nil {
		return err
	}
	if err := conn.StreamBegin(uint32(conn.StreamID)); err != nil {
		return err
	}
	log.Printf("%s playing stream=%s (media send empty)", conn.nc.RemoteAddr(), conn.Stream)

	// wait for video data — sending left empty for now
	for {
		if _, err := conn.Read(); err != nil {
			return err
		}
	}
}

// rtmp server listen
func (srv *Serve) Listen(address string) error {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer ln.Close()
	for {
		nc, err := ln.Accept()
		if err != nil {
			return err
		}
		go srv.handle(nc)
	}
}

// create new rtmp serve
func NewRtmp() *Serve {
	s := &Serve{}
	return s
}
