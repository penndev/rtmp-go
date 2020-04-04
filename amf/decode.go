package amf

import (
	"encoding/binary"
	"log"
)

// Value 通用类型
type Value interface{}

func Decode(b []byte) []Value {
	var r []Value

	for len(b) > 0 {
		var val Value
		switch b[0] {
		case TypeNumber:
			val = ReadNumber(b[1:9])
			b = b[9:]
		case TypeBoolean:
			val = ReadBoolean(b[1:2])
			b = b[2:]
		case TypeString:
			end := 3 + int(binary.BigEndian.Uint16(b[1:3]))
			val = ReadString(b[3:end])
			b = b[end:]
		case TypeObject:
			var end int
			for i, v := range b {
				if v == TypeObjectEnd {
					if b[i-1] == b[i-2] && int(b[i-1]) == 0 {
						end = i + 1
						break
					}
				}
			}
			val = ReadObject(b[1 : end-3])
			b = b[end:]
		case TypeNull:
			val = ReadNull()
			b = b[1:]
		case TypeEcmaArray:
			var end int
			val, end = ReadEcmaObject(b[1:])
			b = b[end:]
		default:
			log.Println("rtmp amf-遇到未处理的数据类型:")
			val = nil
		}

		if val == nil {
			break
		}

		r = append(r, val)

	}

	return r
}
