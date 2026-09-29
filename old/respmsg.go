package rtmp

import "github.com/penndev/rtmp/amf"

func respConnect(b bool) []byte {
	if !b {
		out, _ := amf.Encode0([]amf.Value{"_error", 1, nil, nil})
		return out
	}
	repVer := amf.Object{
		"fmsVer":       "FMS/3,0,1,123",
		"capabilities": 31,
	}
	repStatus := amf.Object{
		"level":          "status",
		"code":           "NetConnection.Connect.Success",
		"description":    "Connection succeeded.",
		"objectEncoding": 3,
	}
	out, _ := amf.Encode0([]amf.Value{"_result", 1, repVer, repStatus})
	return out
}

func respCreateStream(_ bool, tranId int, streamId int) []byte {
	out, _ := amf.Encode0([]amf.Value{"_result", tranId, nil, streamId})
	return out
}

func respPublish(b bool) []byte {
	res := amf.Object{
		"level":       "status",
		"description": "Start publishing",
	}
	if b {
		res["code"] = "NetStream.Publish.Start"
	} else {
		res["code"] = "NetStream.Publish.BadName"
	}
	out, _ := amf.Encode0([]amf.Value{"onStatus", 0, nil, res})
	return out
}

func respPlay(b bool) []byte {
	res := amf.Object{
		"level":       "status",
		"description": "Start playing",
	}
	if b {
		res["code"] = "NetStream.Play.Start"
	} else {
		res["code"] = "NetStream.Play.Failed"
	}
	out, _ := amf.Encode0([]amf.Value{"onStatus", 0, nil, res})
	return out
}
