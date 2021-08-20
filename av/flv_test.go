package av

import (
	"testing"
)

func TestFlv(t *testing.T) {
	tag := Tag{
		tagType:             8,
		timeStreamp:         10,
		timeStreampExtended: 12,
		tagData:             []byte{1, 2, 3, 33, 33, 4, 33, 123, 4},
	}
	t.Log(tag.genByte())

	var flv FLV
	t.Log(flv.genHead("av"))
}
