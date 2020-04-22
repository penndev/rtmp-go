package amf

import (
	"encoding/binary"
	"log"
	"math"
)

//
const (
	TypeNumber        = 0x00
	TypeBoolean       = 0x01
	TypeString        = 0x02
	TypeObject        = 0x03
	TypeMovieclip     = 0x04 //保留类型未使用; reserved, not supported
	TypeNull          = 0x05
	TypeUndefined     = 0x06
	TypeReference     = 0x07
	TypeEcmaArray     = 0x08
	TypeObjectEnd     = 0x09 //对象结尾
	TypeStrictArray   = 0x0a
	TypeDate          = 0x0b
	TypeLongString    = 0x0c
	TypeUnsupported   = 0x0d
	TypeRecordset     = 0x0e //保留类型未使用; reserved, not supported xml-document-marker     =0x0f
	TypeTypedObject   = 0x10
	TypeAvmplusObject = 0x11 //切换到amf3
)

// Value 通用类型
type Value interface{}

// ReadNumber 读取double的值 8个字节
func ReadNumber(bytes []byte) float64 {
	bits := binary.BigEndian.Uint64(bytes[0:8])
	float := math.Float64frombits(bits)
	return float
}

// ReadBoolean 读取amf布尔值 1个字节
func ReadBoolean(bytes []byte) bool {
	if bytes[0] != 0 {
		return true
	}
	return false
}

// ReadString 读取utf8字符 变长字节
func ReadString(bytes []byte) string {
	str := string(bytes)
	return str
}

// ReadObject 读取对象类型
func ReadObject(bytes []byte) map[string]Value {

	obj := make(map[string]Value)

	for len(bytes) > 0 {
		var val Value
		var end int

		vStart := 2 + int(binary.BigEndian.Uint16(bytes[0:2]))
		key := ReadString(bytes[2:vStart])

		switch bytes[vStart] {
		case TypeNumber:
			end = vStart + 9
			val = ReadNumber(bytes[vStart+1 : end])
		case TypeString:
			start := vStart + 3
			end = start + int(binary.BigEndian.Uint16(bytes[vStart+1:start]))
			val = ReadString(bytes[start:end])
		default:
			log.Println("遇到未处理的数据类型ReadObject:", bytes)
			val = nil
		}

		if val == nil {
			break
		}

		obj[key] = val
		bytes = bytes[end:]
	}
	return obj
}

//ReadEcmaObject 读取定长类型
func ReadEcmaObject(bytes []byte) (map[string]Value, int) {
	obj := make(map[string]Value)
	lenght := int(binary.BigEndian.Uint32(bytes[0:4]))
	leng := 4
	bytes = bytes[4:]

	for lenght > 0 {

		var val Value
		var end int

		vStart := 2 + int(binary.BigEndian.Uint16(bytes[0:2]))
		key := ReadString(bytes[2:vStart])

		switch bytes[vStart] {
		case TypeNumber:
			end = vStart + 9
			val = ReadNumber(bytes[vStart+1 : end])
		case TypeString:
			start := vStart + 3
			end = start + int(binary.BigEndian.Uint16(bytes[vStart+1:start]))
			val = ReadString(bytes[start:end])
		case TypeBoolean:
			start := vStart + 1
			end = start + 1

			val = ReadBoolean(bytes[start:end])
		default:
			log.Println("ReadEcmaObject:", bytes)
			val = nil
		}

		if val == nil {
			break
		}

		obj[key] = val
		leng += end
		lenght--
		bytes = bytes[end:]
	}
	return obj, leng
}

// ReadNull 读取空类型
func ReadNull() bool {
	return false
}

// WriteNumber 写入数据
func WriteNumber(s int) []byte {
	var val []byte
	val = append(val, 0x00)

	bits := math.Float64bits(float64(s))
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, bits)
	val = append(val, bytes...)
	return val
}

// WriteString 写字符串数据
func WriteString(s string) []byte {
	var val []byte
	val = append(val, 0x02)
	sbyte := []byte(s)
	sLen := len(sbyte)
	val = append(val, byte(sLen/256), byte(sLen%56))
	val = append(val, sbyte...)
	return val
}

// WriteObject 写入对象
func WriteObject(objs map[string]Value) []byte {

	var res []byte
	res = append(res, 0x03)
	for i, v := range objs {
		res = append(res, WriteString(i)[1:]...)
		switch iType := v.(type) {
		case string:
			res = append(res, WriteString(iType)...)
		case int:
			res = append(res, WriteNumber(iType)...)
		default:
			log.Println("rtmp amf- WriteObject 遇到未处理的数据类型:", iType)
		}
	}
	return append(res, 0, 0, TypeObjectEnd)
}

// WriteNull 写入
func WriteNull(v Value) []byte {
	return []byte{5}
}
