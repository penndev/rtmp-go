package rtmp

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"log"
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

//ClientFullKey ->sha256key 0:30
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

	// zeroFull := binary.BigEndian.Uint32(buf[5:9]) // hand shake type
	// if zeroFull == 0 {                            // 复杂握手
	// <-S0+S1+S2
	S2 := buf[1:1537]
	c.Write(buf)
	c.Write(S2)
	// } else { //简单握手。
	// 	c.Write(genS0S1S2(buf))
	// 	log.Println("debug --001 复杂握手flash 会握手失败，待完善！！！")
	// }

	// C2->
	_, err = c.Read(make([]byte, 1536))

	return err
}

// get c1 digest content
func getC1Digest(c1 []byte) []byte {
	// 12 = 4 time 4 zerofill 4offset
	offset := (int(c1[8])+int(c1[9])+int(c1[10])+int(c1[11]))%728 + 12
	digest := c1[offset : offset+32]

	h := hmac.New(sha256.New, ClientFullKey[:30])
	h.Write(c1[:offset])
	h.Write(c1[offset+32:])
	newDig := h.Sum(nil)

	if bytes.Equal(digest, newDig) {
		return newDig
	}

	// 776 = 764 + 12   772=764+8
	offset = (int(c1[772])+int(c1[773])+int(c1[774])+int(c1[775]))%728 + 776
	digest = c1[offset : offset+32]

	h = hmac.New(sha256.New, ClientFullKey[:30])
	h.Write(c1[:offset])
	h.Write(c1[offset+32:])
	newDig = h.Sum(nil)

	if !bytes.Equal(digest, newDig) {
		log.Println("rtmp complex handshake failed. ")
	}
	return newDig
}

// GenS012
func genS0S1S2(buf []byte) []byte {
	s012 := append(make([]byte, 1), VERSION)

	s1 := append(s012, buf[1:5]...)
	s1 = append(s012, S1Full...)
	s1 = append(s012, buf[9:]...)
	offset := (int(s1[8])+int(s1[9])+int(s1[10])+int(s1[11]))%728 + 12
	h := hmac.New(sha256.New, ServerFullKey[:36])
	h.Write(s1[:offset])
	h.Write(s1[offset+32:])
	s1Digest := h.Sum(nil)

	s012 = append(s012, s1[:offset]...)
	s012 = append(s012, s1Digest...)
	s012 = append(s012, s1[offset+32:]...)

	c1 := buf[1:]
	c1Digest := getC1Digest(c1)
	h = hmac.New(sha256.New, c1Digest)
	h.Write(c1[:1504])
	s2Digest := h.Sum(nil)

	s012 = append(s012, c1[:1504]...)
	s012 = append(s012, s2Digest...)

	return s012
}
