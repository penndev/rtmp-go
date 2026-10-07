package rtmp

import (
	"log"

	"github.com/penndev/rtmp/pubsub"
)

// play: reply + Stream Begin, send cached meta, forward AV;
// on topic close send Stream EOF + Play.Stop
func (srv *Serve) handlePlay(conn *Conn) error {

	if err := conn.StreamBegin(uint32(conn.StreamID)); err != nil {
		return err
	}

	log.Printf(
		"%s playing name=%s conn.App=%s conn.Stream=%s",
		conn.nc.RemoteAddr(), conn.Name(), conn.App, conn.Stream)

	sid := uint32(conn.StreamID)

	sub := pubsub.SubTopic(conn.Name())
	defer func() {
		sub.Close()
	}()

	errCh := make(chan error, 1)
	go func() {
		for {
			if _, err := conn.Read(); err != nil {
				errCh <- err
				return
			}
		}
	}()

	for {
		select {
		case m, ok := <-sub.Chan():
			if !ok {
				conn.StreamEOF(sid)
				conn.PlayStop()
				return nil
			}
			msg, ok := m.Data.(*Message)
			if !ok || msg == nil {
				continue
			}
			csid := CSIDData
			switch msg.MessageType {
			case Audio:
				csid = CSIDAudio
			case Video:
				csid = CSIDVideo
			}
			if err := conn.Write(csid, sid, msg); err != nil {
				return err
			}
		case err := <-errCh:
			return err
		}
	}
}
