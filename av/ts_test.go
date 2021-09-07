package av

import "testing"

// go test -v -run TestTsHeader  rtmp-go/av
func TestTsHeader(t *testing.T) {
	var th tsPacketHeader
	th.unitStart = true
	th.adaptation = 1
	// th.pid = 1
	// th.scriamble = 2
	// th.counter = 12

	t.Logf("%x", th.genTsPacketHeader())
}
