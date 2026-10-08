// https://veovera.org/docs/legacy/rtmp-v1-0-spec.pdf

package rtmp

import (
	"log"
	"net"

	"github.com/penndev/rtmp/amf"
	"github.com/penndev/rtmp/rtmp/handler"
	"github.com/penndev/rtmp/rtmp/stream"
)

type Serve struct {
	Addr    string
	Handler handler.Handler
	Stream  stream.Stream
}

func (srv *Serve) handle(nc net.Conn) {
	log.Printf("rtmp handle %s", nc.RemoteAddr())
	defer func() {
		nc.Close()
		if err := recover(); err != nil {
			log.Printf("%s: %s", "recover: ", err)
		}
	}()

	if srv.Handler == nil {
		srv.Handler = handler.NewDefaultHandler()
	}
	if srv.Stream == nil {
		log.Printf("%s stream is nil", nc.RemoteAddr())
		return
	}

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
			topic, err := srv.Stream.Publish(srv.Handler.OnName(conn.App, conn.Stream))
			if err != nil {
				log.Printf("%s publish rejected, %s app=%s stream=%s", nc.RemoteAddr(), err, conn.App, conn.Stream)
				if err := conn.PublishReply(false); err != nil {
					log.Printf("%s PublishReply fail err[%s]", nc.RemoteAddr(), err.Error())
				}
				return
			}
			defer topic.Close()
			if !srv.Handler.OnPublish(conn.App, conn.Stream) {
				log.Printf("%s publish rejected by handler app=%s stream=%s", nc.RemoteAddr(), conn.App, conn.Stream)
				if err := conn.PublishReply(false); err != nil {
					log.Printf("%s PublishReply fail err[%s]", nc.RemoteAddr(), err.Error())
				}
				return
			}
			defer srv.Handler.OnPublishStop(conn.App, conn.Stream)
			if err := conn.PublishReply(true); err != nil {
				log.Printf("%s PublishReply fail err[%s]", nc.RemoteAddr(), err.Error())
				return
			}
			if err := srv.handlePublish(conn, topic); err != nil {
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
			sub, err := srv.Stream.Play(srv.Handler.OnName(conn.App, conn.Stream))
			if err != nil {
				log.Printf("%s play rejected, %s app=%s stream=%s", nc.RemoteAddr(), err, conn.App, conn.Stream)
				if err := conn.PlayReply(false); err != nil {
					log.Printf("%s PlayReply fail err[%s]", nc.RemoteAddr(), err.Error())
				}
				return
			}
			defer sub.Close()
			if !srv.Handler.OnPlay(conn.App, conn.Stream) {
				log.Printf("%s play rejected by handler app=%s stream=%s", nc.RemoteAddr(), conn.App, conn.Stream)
				if err := conn.PlayReply(false); err != nil {
					log.Printf("%s PlayReply fail err[%s]", nc.RemoteAddr(), err.Error())
				}
				return
			}
			defer srv.Handler.OnPlayStop(conn.App, conn.Stream)
			if err := conn.PlayReply(true); err != nil {
				log.Printf("%s PlayReply fail err[%s]", nc.RemoteAddr(), err.Error())
				return
			}
			if err := srv.handlePlay(conn, sub); err != nil {
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
	if srv.Handler == nil {
		srv.Handler = handler.NewDefaultHandler()
	}
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
func New(h handler.Handler, streams stream.Stream) *Serve {
	return &Serve{Handler: h, Stream: streams}
}
