package amf

import "log"

// Encode 对数据进行编码
func Encode(val []Value) []byte {
	var res []byte
	for _, v := range val {
		switch iType := v.(type) {
		case string:
			res = append(res, WriteString(iType)...)
		case float64:
			res = append(res, WriteNumber(iType)...)
		case int:
			res = append(res, WriteNumber(float64(iType))...)
		case map[string]Value:
			res = append(res, WriteObject(iType)...)
		case nil:
			res = append(res, WriteNull(iType)...)
		default:
			log.Println("rtmp amf-Encode 遇到未处理的数据类型：", val, v, iType)
		}
	}
	return res
}
