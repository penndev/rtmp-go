package rtmp

import (
	"errors"
	"fmt"

	"github.com/penndev/rtmp/amf"
)

// 7.2.1.1. connect — client → server
type ConnectCommand struct {
	CommandName   string        `json:"commandName"`
	TransactionID float64       `json:"transactionId"`
	CommandObject ConnectObject `json:"commandObject"`
	UserArguments amf.Object    `json:"userArguments,omitempty"`
}

// Command Object name-value pairs for connect
type ConnectObject struct {
	App            string  `json:"app"`
	FlashVer       string  `json:"flashVer"`
	SwfUrl         string  `json:"swfUrl"`
	TcUrl          string  `json:"tcUrl"`
	Fpad           bool    `json:"fpad"`
	AudioCodecs    float64 `json:"audioCodecs"`
	VideoCodecs    float64 `json:"videoCodecs"`
	VideoFunction  float64 `json:"videoFunction"`
	PageUrl        string  `json:"pageUrl"`
	ObjectEncoding float64 `json:"objectEncoding"`
}

func amfNumber(v amf.Value) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case amf.Integer:
		return float64(x)
	default:
		return 0
	}
}

func ConnectParse(vals []amf.Value) (*ConnectCommand, error) {
	var cmd ConnectCommand
	if len(vals) < 3 {
		return nil, errors.New("connect: need Command Name, Transaction ID, Command Object")
	}
	name, ok := vals[0].(string)
	if !ok {
		return nil, errors.New("connect: Command Name is not string")
	}
	if name != "connect" {
		return nil, fmt.Errorf("connect: Command Name is %q", name)
	}
	cmd.CommandName = name

	switch tid := vals[1].(type) {
	case float64:
		cmd.TransactionID = tid
	case amf.Integer:
		cmd.TransactionID = float64(tid)
	default:
		return nil, errors.New("connect: Transaction ID is not number")
	}

	obj, ok := amf.AsObject(vals[2])
	if !ok {
		return nil, errors.New("connect: Command Object missing")
	}
	if v, ok := obj["app"].(string); ok {
		cmd.CommandObject.App = v
	}
	if v, ok := obj["flashVer"].(string); ok {
		cmd.CommandObject.FlashVer = v
	}
	if v, ok := obj["swfUrl"].(string); ok {
		cmd.CommandObject.SwfUrl = v
	}
	if v, ok := obj["tcUrl"].(string); ok {
		cmd.CommandObject.TcUrl = v
	}
	if v, ok := obj["fpad"].(bool); ok {
		cmd.CommandObject.Fpad = v
	}
	if v, ok := obj["pageUrl"].(string); ok {
		cmd.CommandObject.PageUrl = v
	}
	cmd.CommandObject.AudioCodecs = amfNumber(obj["audioCodecs"])
	cmd.CommandObject.VideoCodecs = amfNumber(obj["videoCodecs"])
	cmd.CommandObject.VideoFunction = amfNumber(obj["videoFunction"])
	cmd.CommandObject.ObjectEncoding = amfNumber(obj["objectEncoding"])

	if len(vals) > 3 {
		if u, ok := amf.AsObject(vals[3]); ok {
			cmd.UserArguments = u
		}
	}
	return &cmd, nil
}

// ConnectBuild 按 7.2.1.1 组装 connect 应答 AMF 值
func ConnectBuild(cmd *ConnectCommand, success bool) []amf.Value {
	oe := cmd.CommandObject.ObjectEncoding
	if oe != 3 {
		oe = 0
	}
	name := "_result"
	info := amf.Object{
		"level":          "status",
		"code":           "NetConnection.Connect.Success",
		"description":    "Connection succeeded.",
		"objectEncoding": oe,
	}
	if !success {
		name = "_error"
		info = amf.Object{
			"level":          "error",
			"code":           "NetConnection.Connect.Failed",
			"description":    "Connection failed.",
			"objectEncoding": oe,
		}
	}
	return []amf.Value{
		name,
		cmd.TransactionID,
		amf.Object{
			"fmsVer":       "FMS/3,0,1,123",
			"capabilities": 31.0,
		},
		info,
	}
}
