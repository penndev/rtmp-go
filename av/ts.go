package av

import (
	"encoding/binary"
	"log"
	"os"
)

//首先进行分包
func Ts() {
	file, err := os.OpenFile("live.ts", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Println(err)
	}

	tsPackage := make([]byte, 188)
	for i := 0; i < 1; i++ {
		file.Read(tsPackage)
		log.Println(tsPackage)
	}

	newfile, err := os.OpenFile("genlive.ts", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Println(err)
	}
	// 写入新文件。
	for {
		n, err := file.Read(tsPackage)
		if err != nil {
			log.Println(n, err)
			break
		}
		newfile.Write(tsPackage)
	}

}

// 0000-0000 0000-0000 0000-0000 0000-0000
// 1--- ---- 2345 ---- ---- ---- 6-7- 8---
// 1:  0x47 表示文件类型。
// 2： 错误标识 如果为1则表示有错误。
// 3： 是否有附属字段
// 4： 优先级 如果为1则有更高优先级。
// 5： 表示PID
// 6： 是否加密
// 7： 适配域控制 [0 保留 1只有Payload数据 2只有适配域 3两种都有]
// 8： 同一Pid下的递增值
type tsPacketHeader struct {
	// syncByte   byte
	tshErr     bool
	unitStart  bool
	priority   bool
	pid        uint16
	scriamble  uint8
	adaptation uint8
	counter    uint8
}

func (tsh *tsPacketHeader) genTsPacketHeader() []byte {
	hd := make([]byte, 4)
	hd[0] = 0x47
	if tsh.tshErr {
		hd[1] |= 0x80
	}
	if tsh.unitStart {
		hd[1] |= 0x40
	}
	if tsh.priority {
		hd[1] |= 0x20
	}
	//13位。PID
	pid := make([]byte, 2)
	binary.BigEndian.PutUint16(pid, tsh.pid)
	hd[1] |= pid[0]
	hd[2] = pid[1]
	//-
	hd[3] |= tsh.scriamble << 6
	hd[3] |= tsh.adaptation << 4
	hd[3] |= tsh.counter
	log.Println(hd)
	return hd
}

func genPAT() []byte {

	return nil
}
