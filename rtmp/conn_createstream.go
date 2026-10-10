package rtmp

import (
	"errors"
	"fmt"

	"github.com/penndev/rtmp/amf"
)

// 7.2.2.1. createStream — client → server
// Command Name / Transaction ID / Command Object (null)
type CreateStreamCommand struct {
	CommandName   string  `json:"commandName"`
	TransactionID float64 `json:"transactionId"`
}

func CreateStreamParse(vals []amf.Value) (*CreateStreamCommand, error) {
	var cmd CreateStreamCommand
	if len(vals) < 2 {
		return nil, errors.New("createStream: need Command Name, Transaction ID")
	}
	name, ok := vals[0].(string)
	if !ok {
		return nil, errors.New("createStream: Command Name is not string")
	}
	if name != "createStream" {
		return nil, fmt.Errorf("createStream: Command Name is %q", name)
	}
	cmd.CommandName = name

	switch tid := vals[1].(type) {
	case float64:
		cmd.TransactionID = tid
	case amf.Integer:
		cmd.TransactionID = float64(tid)
	default:
		return nil, errors.New("createStream: Transaction ID is not number")
	}
	return &cmd, nil
}
