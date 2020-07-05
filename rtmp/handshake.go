package rtmp

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"net"
)

//VERSION rtmp version
const VERSION = 0x03

//S1Full S1 full key
var S1Full = []byte{0x0C, 0x00, 0x0D, 0x0E}

//ServerFullKey ->sha256key 0:36
var ServerFullKey = []byte{
	'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
	'F', 'l', 'a', 's', 'h', ' ', 'M', 'e', 'd', 'i', 'a', ' ',
	'S', 'e', 'r', 'v', 'e', 'r', ' ',
	'0', '0', '1',

	0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
	0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
	0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
}

var ClientFullKey = []byte{
	'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
	'F', 'l', 'a', 's', 'h', ' ', 'P', 'l', 'a', 'y', 'e', 'r', ' ',
	'0', '0', '1',
	0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
	0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
	0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
}

//CS1 | 4字节时间戳time | 4字节全0二进制串 | 1528字节随机二进制串 |
//CS2 | 4字节时间戳time | 4字节time2 | 1528字节随机二进制串 |
// C0 + C1 -> Server
// S0 + S1 + S2 -> Client
// C2 -> Server

//ServeHandShake rtmp Server handshake。
func ServeHandShake(c net.Conn) error {
	buf := make([]byte, 1537)
	_, err := c.Read(buf)
	if err != nil {
		return err
	}
	// RTMP VERSION == 3
	if buf[0] != VERSION {
		return errors.New("rtmp version is not support")
	}

	zeroFull := binary.BigEndian.Uint32(buf[5:9]) // hand shake type
	if zeroFull == 0 {                            // 复杂握手
		// <-S0+S1+S2
		S2 := buf[1:1537]
		c.Write(buf)
		c.Write(S2)
	} else { //复杂握手。
		s012 := genS0S1S2(buf)
		c.Write(s012)
	}

	// C2->
	_, err = c.Read(make([]byte, 1536))

	return err
}

func getOffset(CS1 []byte, l bool) int {
	start := 8
	if l {
		start = 772
	}
	orOffset := int(CS1[start]) + int(CS1[start+1]) + int(CS1[start+2]) + int(CS1[start+3])
	return orOffset%728 + 4 + start
}

//计算 digest的方法
func isDigest(c1 []byte, key []byte, offset int) []byte {
	digest := c1[offset : offset+32]
	h := hmac.New(sha256.New, key)
	h.Write(c1[:offset])
	h.Write(c1[offset+32:])
	newDig := h.Sum(nil)
	if bytes.Equal(digest, newDig) {
		return newDig
	}
	return nil
}

//提取C1 or S1  32位 摘要
func resSC1Digest(c1 []byte, key []byte) []byte {
	offset := getOffset(c1, true)
	if i := isDigest(c1, key, offset); i != nil {
		return i
	}

	offset = getOffset(c1, false)
	if i := isDigest(c1, key, offset); i != nil {
		return i
	}

	return c1[offset : offset+32]
}

func genSC1(SC1 []byte, key []byte) []byte {
	offset := getOffset(SC1, true)
	digest := isDigest(SC1, key, offset)
	copy(SC1[offset:], digest)
	return SC1
}

func genSC2(SC2 []byte, key []byte) []byte {
	// rand.Read(p)
	offset := len(SC2) - 32
	digest := isDigest(SC2, key, offset)
	copy(SC2[offset:], digest)
	return SC2
}

// GenS012
func genS0S1S2(buf []byte) []byte {
	C1 := buf[1:1537]
	c1Digest := resSC1Digest(C1, ClientFullKey[:30])
	S1 := genSC1(C1, ServerFullKey[:36])
	c1Time := C1[0:4]
	copy(C1[4:], c1Time)

	S2 := genSC2(C1, isDigest(ServerFullKey, c1Digest, 0))

	S012 := []byte{VERSION}
	S012 = append(S012, S1...)
	S012 = append(S012, S2...)
	return S012
}
