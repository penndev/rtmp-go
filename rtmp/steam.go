package rtmp

import (
	"errors"
	"log"

	"rtmp-go/amf"
)

func (m *Message) respPublish(steam amf.Value) {

	m.Conn.Stream = steam.(string)
	m.Server.addPool(m.Conn.App, m.Conn.Stream)

	res := make(map[string]amf.Value)
	res["level"] = "status"
	res["code"] = "NetStream.Publish.Start"
	res["description"] = "Start publishing"
	resp := []amf.Value{"onStatus", 0, nil, res}

	c := Chunk{
		MessageTypeID: 20,
		SteamID:       4,
		Payload:       amf.Encode(resp),
		Conn:          m.Conn,
	}
	c.SendChunk()

	m.Conn.IsPusher = true
}

func (m *Message) respPlay(steam amf.Value) error {
	m.Conn.Stream = steam.(string)

	if _, ok := m.Server.WorkPool[m.Conn.App][m.Conn.Stream]; !ok {
		// m.respDelete
		m.Conn.Close()
		return errors.New("rtmp live dont set" + m.Conn.App + m.Conn.Stream)
	}

	m.streamBegin()

	res := make(map[string]amf.Value)
	res["level"] = "status"
	res["code"] = "NetStream.Play.Start"
	res["description"] = "Start playing"
	resp := []amf.Value{"onStatus", 0, nil, res}
	m.Conn.SteamID = m.Chunk.SteamID
	c := Chunk{
		MessageTypeID: 20,
		SteamID:       m.Chunk.SteamID,
		Payload:       amf.Encode(resp),
		Conn:          m.Conn,
	}
	c.SendChunk()

	go m.Playing()

	return nil
}

func (m *Message) respCreateSteam(nmb amf.Value) {
	t, ok := nmb.(float64)
	if !ok {
		t = 0
	}
	repByte := amf.Encode([]amf.Value{"_result", int(t), nil, int(t)})
	c := Chunk{
		MessageTypeID: 20,
		SteamID:       m.Chunk.SteamID,
		Payload:       repByte,
		Conn:          m.Conn,
	}
	c.SendChunk()
}

func (m *Message) respConnect(amfObj amf.Value) {
	m.setChunkSize()
	//set conn app
	app, ok := amfObj.(map[string]amf.Value)
	if !ok {
		m.Conn.Close()
	}
	m.Conn.App = app["app"].(string)
	//resp connect
	var arrSour []amf.Value
	repVer := make(map[string]amf.Value)
	repVer["fmsVer"] = "FMS/3,0,1,123"
	repVer["capabilities"] = 31
	repStatus := make(map[string]amf.Value)
	repStatus["level"] = "status"
	repStatus["code"] = "NetConnection.Connect.Success"
	repStatus["description"] = "Connection succeeded."
	repStatus["objectEncoding"] = 3
	// _error or _result
	repSour := append(arrSour, "_result", 1, repVer, repStatus)
	c := Chunk{
		MessageTypeID: 20,
		SteamID:       m.Chunk.SteamID,
		Payload:       amf.Encode(repSour),
		Conn:          m.Conn,
	}
	c.SendChunk()
}

func (m *Message) deleteStream() {
	if !m.Conn.IsPusher {
		if play, ok := m.Server.WorkPool[m.Conn.App][m.Conn.Stream].Player[m.Conn.remoteAddr]; ok {
			delete(m.Server.WorkPool[m.Conn.App][m.Conn.Stream].Player, m.Conn.remoteAddr)
			close(play)
		}
	}

	m.Conn.Close()
}

func (m *Message) setMetadata() {
	item := amf.Decode(m.Chunk.Payload)
	meta := item[2]
	resp := []amf.Value{"onMetaData", meta}

	metaItem, ok := meta.(map[string]amf.Value)
	if !ok {
		m.Conn.Close()
	}

	wp := m.Server.WorkPool[m.Conn.App][m.Conn.Stream]
	wp.Metadata = amf.Encode(resp)
	switch vid := metaItem["videocodecid"].(type) {
	case string:
		if vid == "avc1" {
			wp.videocodecid = 7
		}
	case float64:
		wp.videocodecid = int(vid)
	}

}

// Playing 加入播放队列
func (m *Message) Playing() {
	pool := m.Server.WorkPool[m.Conn.App][m.Conn.Stream]
	play := make(chan Chunk)
	pool.Player[m.Conn.remoteAddr] = play
	//send metadata
	c := Chunk{
		MessageTypeID: 18,
		SteamID:       4,
		Payload:       pool.Metadata,
		Conn:          m.Conn,
	}
	c.SendChunk()

	if pool.videocodecid == 7 {
		v := Chunk{
			MessageTypeID: 9,
			SteamID:       4,
			Payload:       pool.VideoCode,
			Conn:          m.Conn,
		}
		v.SendChunk()
		a := Chunk{
			MessageTypeID: 8,
			SteamID:       4,
			Payload:       pool.AudioCode,
			Conn:          m.Conn,
		}
		a.SendChunk()
	}
	//
	start := false
	for {
		x := <-play
		//first video mast keyframe
		if start == false {
			if x.MessageTypeID == 9 {
				k := x.Payload[0]
				if k>>4 == 1 {
					start = true
				} else {
					continue
				}
			} else {
				continue
			}
		}

		x.Conn = m.Conn
		if err := x.SendChunk(); err != nil {
			log.Println(err)
			break
		}
	}
}
