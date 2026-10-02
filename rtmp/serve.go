// https://veovera.org/docs/legacy/rtmp-v1-0-spec.pdf

package rtmp

import (
	"log"
	"net"

	"github.com/penndev/rtmp/amf"
)

type Serve struct {
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

	if err := conn.CreateStreamReply(cs, DefaultNetStreamID); err != nil {
		log.Printf("%s CreateStreamReply fail err[%s]", nc.RemoteAddr(), err.Error())
		return
	}

	for {
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
			continue
		}
		if err != nil {
			log.Printf("%s decode command fail err[%s]", nc.RemoteAddr(), err.Error())
			return
		}
		if len(values) == 0 {
			log.Printf("%s empty command message", nc.RemoteAddr())
			continue
		}
		name, _ := values[0].(string)

		switch name {
		case "publish":
			if len(values) >= 4 {
				if s, ok := values[3].(string); ok {
					conn.Stream = s
				}
			}
			if err := srv.handlePublish(conn); err != nil {
				log.Printf("%s handlePublish fail err[%s]", nc.RemoteAddr(), err.Error())
			} else {
				log.Printf("%s handlePublish Finsh", nc.RemoteAddr())
			}
			return
		case "play":
			if len(values) >= 4 {
				if s, ok := values[3].(string); ok {
					conn.Stream = s
				}
			}
			if err := srv.handlePlay(conn); err != nil {
				log.Printf("%s handlePlay fail err[%s]", nc.RemoteAddr(), err.Error())
			} else {
				log.Printf("%s handlePlay Finsh", nc.RemoteAddr())
			}
			return
		default:
			log.Printf("%s ignore netstream command: %s", nc.RemoteAddr(), name)
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
func New() *Serve {
	return &Serve{}
}
