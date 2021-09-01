package rtmp

import "rtmp-go/amf"

func respConnect(b bool) []byte {
	if !b {
		return amf.Encode([]amf.Value{"_error", 1, nil, nil})
	}
	repVer := make(map[string]amf.Value)
	repVer["fmsVer"] = "FMS/3,0,1,123"
	repVer["capabilities"] = 31
	repStatus := make(map[string]amf.Value)
	repStatus["level"] = "status"
	repStatus["code"] = "NetConnection.Connect.Success"
	repStatus["description"] = "Connection succeeded."
	repStatus["objectEncoding"] = 3
	return amf.Encode([]amf.Value{"_result", 1, repVer, repStatus})
}

func respCreateStream(b bool, transaId int, streamId int) []byte {
	return amf.Encode([]amf.Value{"_result", transaId, nil, streamId})
}

func respPublish(b bool) []byte {
	res := make(map[string]amf.Value)
	res["level"] = "status"
	if b {
		res["code"] = "NetStream.Publish.Start"
	} else {
		res["code"] = "NetStream.Publish.BadName"
	}
	res["description"] = "Start publishing"
	return amf.Encode([]amf.Value{"onStatus", 0, nil, res})
}

func respPlay(b bool) []byte {
	res := make(map[string]amf.Value)
	res["level"] = "status"
	if b {
		res["code"] = "NetStream.Play.Start"
	} else {
		res["code"] = "NetStream.Play.Failed"
	}
	res["description"] = "Start playing"
	return amf.Encode([]amf.Value{"onStatus", 0, nil, res})
}
