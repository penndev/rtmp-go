package rtmp

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"time"
)

// 5.2. Handshake . . . . . . . . . . . . . . . . . . . . . . . . 7
// 5.2.1. Handshake Sequence . . . . . . . . . . . . . . . . . . 7
// 5.2.2. C0 and S0 Format . . . . . . . . . . . . . . . . . . . 7
// 5.2.3. C1 and S1 Format . . . . . . . . . . . . . . . . . . . 8
// 5.2.4. C2 and S2 Format . . . . . . . . . . . . . . . . . . . 8
// 5.2.5. Handshake Diagram . . . . . . . . . . . . . . . . . . 10
func (c *Conn) Handshake() error {
	c.nc.SetReadDeadline(time.Now().Add(10 * time.Second))
	defer c.nc.SetReadDeadline(time.Time{})

	// C0: version(1)
	c0 := make([]byte, 1)
	_, err := io.ReadFull(c.nc, c0)
	if err != nil {
		return err
	}
	if c0[0] != VERSION {
		return errors.New("rtmp version is not support")
	}

	// S0 + S1
	// S0: version(1)
	// S1: time(4) + zero(4) + random(1528)
	s0s1 := make([]byte, 1537)
	s0s1[0] = VERSION
	s1rand := s0s1[9:1537]
	if _, err = rand.Read(s1rand); err != nil {
		return err
	}
	start := time.Now()
	n, err := c.nc.Write(s0s1)
	if err != nil {
		return err
	}
	if n != 1537 {
		return errors.New("write s1 failed")
	}

	// C1: time(4) + zero(4) + random(1528)
	c1 := make([]byte, 1536)
	_, err = io.ReadFull(c.nc, c1)
	if err != nil {
		return err
	}

	// // Flash Player 9 RTMP-e
	// if !bytes.Equal(c1[4:8], []byte{0, 0, 0, 0}) {
	// 	todo: handle flash player 9 rtmp-e
	// }

	// S2: time = C1.time, time2 = 从发出 S1 到此刻的毫秒, random = C1.random
	s2 := make([]byte, 1536)
	copy(s2[0:4], c1[0:4])
	s2Time2 := uint32(time.Since(start).Milliseconds())
	binary.BigEndian.PutUint32(s2[4:8], s2Time2)
	copy(s2[8:], c1[8:])
	n, err = c.nc.Write(s2)
	if err != nil {
		return err
	}
	if n != len(s2) {
		return errors.New("write s2 failed")
	}

	// C2: time(4) + time2(4) + random echo(1528)
	c2 := make([]byte, 1536)
	_, err = io.ReadFull(c.nc, c2)
	if err != nil {
		return err
	}
	// C2.time 应回显 S1.time。当前 S1.time 为 0。
	if !bytes.Equal(c2[0:4], s0s1[1:5]) {
		return errors.New("c2 time is not s1 time")
	}
	if !bytes.Equal(c2[8:], s1rand) {
		return errors.New("c2 random is not s1 random")
	}

	return nil
}
