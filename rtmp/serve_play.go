package rtmp

import (
	"log"
)

// play: reply + Stream Begin, send cached meta, forward AV;
// on topic close send Stream EOF + Play.Stop
func (srv *Serve) handlePlay(conn *Conn) error {
	if err := conn.PlayReply(true); err != nil {
		return err
	}
	if err := conn.StreamBegin(uint32(conn.StreamID)); err != nil {
		return err
	}

	path := conn.App + "/" + conn.Stream
	log.Printf("%s playing path=%s", conn.nc.RemoteAddr(), path)

	sid := uint32(conn.StreamID)
	srv.mu.RLock()
	meta := srv.meta[path]
	srv.mu.RUnlock()
	if meta != nil {
		if err := conn.Write(5, sid, meta); err != nil {
			return err
		}
	}

	sub := srv.broker.Topic(path).Subscribe()
	defer sub.Close()

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
				_ = conn.StreamEOF(sid)
				_ = conn.PlayStop()
				return nil
			}
			msg, ok := m.Data.(*Message)
			if !ok || msg == nil {
				continue
			}
			csid := 5
			switch msg.MessageType {
			case Audio:
				csid = 4
			case Video:
				csid = 6
			}
			if err := conn.Write(csid, sid, msg); err != nil {
				return err
			}
		case err := <-errCh:
			return err
		}
	}
}
