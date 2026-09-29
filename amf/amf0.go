package amf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

// Decode0 reads all consecutive AMF0 values from b.
// AMF0AVMPlus (0x11) switches to AMF3 for the next value.
func Decode0(b []byte) ([]Value, error) {
	d := &Decoder{data: b}
	var out []Value
	for d.off < len(d.data) {
		v, err := d.Decode0()
		if err != nil {
			return out, err
		}
		out = append(out, v)
	}
	return out, nil
}

// Decoder holds byte cursor and AMF0/AMF3 reference tables for one message.
type Decoder struct {
	data []byte
	off  int

	// AMF0 object references (0-based index into this table).
	amf0Refs []Value

	// AMF3 reference tables.
	amf3Strings []string
	amf3Objects []Value
	amf3Traits  []amf3Trait
}

type amf3Trait struct {
	className      string
	externalizable bool
	dynamic        bool
	keys           []string
}

func (d *Decoder) remain() int { return len(d.data) - d.off }

func (d *Decoder) need(n int) error {
	if d.remain() < n {
		return errors.New("amf: unexpected end of data")
	}
	return nil
}

func (d *Decoder) u8() (byte, error) {
	if err := d.need(1); err != nil {
		return 0, err
	}
	v := d.data[d.off]
	d.off++
	return v, nil
}

func (d *Decoder) u16() (uint16, error) {
	if err := d.need(2); err != nil {
		return 0, err
	}
	v := binary.BigEndian.Uint16(d.data[d.off:])
	d.off += 2
	return v, nil
}

func (d *Decoder) u32() (uint32, error) {
	if err := d.need(4); err != nil {
		return 0, err
	}
	v := binary.BigEndian.Uint32(d.data[d.off:])
	d.off += 4
	return v, nil
}

func (d *Decoder) f64() (float64, error) {
	if err := d.need(8); err != nil {
		return 0, err
	}
	bits := binary.BigEndian.Uint64(d.data[d.off:])
	d.off += 8
	return math.Float64frombits(bits), nil
}

func (d *Decoder) bytes(n int) ([]byte, error) {
	if err := d.need(n); err != nil {
		return nil, err
	}
	b := d.data[d.off : d.off+n]
	d.off += n
	return b, nil
}

// Decode0 reads one AMF0-coded value.
func (d *Decoder) Decode0() (Value, error) {
	marker, err := d.u8()
	if err != nil {
		return nil, err
	}
	switch marker {
	case AMF0Number:
		return d.f64()
	case AMF0Boolean:
		b, err := d.u8()
		if err != nil {
			return nil, err
		}
		return b != 0, nil
	case AMF0String:
		return d.readAMF0String()
	case AMF0Object:
		return d.readAMF0Object()
	case AMF0Movieclip:
		return nil, errors.New("amf0: movieclip reserved")
	case AMF0Null:
		return nil, nil
	case AMF0Undefined:
		return Undefined{}, nil
	case AMF0Reference:
		idx, err := d.u16()
		if err != nil {
			return nil, err
		}
		if int(idx) >= len(d.amf0Refs) {
			return nil, fmt.Errorf("amf0: bad reference %d", idx)
		}
		return d.amf0Refs[idx], nil
	case AMF0ECMAArray:
		return d.readAMF0ECMAArray()
	case AMF0ObjectEnd:
		return nil, errors.New("amf0: unexpected object-end")
	case AMF0StrictArray:
		return d.readAMF0StrictArray()
	case AMF0Date:
		return d.readAMF0Date()
	case AMF0LongString:
		return d.readAMF0LongString()
	case AMF0Unsupported:
		return Unsupported{}, nil
	case AMF0Recordset:
		return nil, errors.New("amf0: recordset reserved")
	case AMF0XMLDocument:
		s, err := d.readAMF0LongString()
		if err != nil {
			return nil, err
		}
		return XMLDocument(s), nil
	case AMF0TypedObject:
		return d.readAMF0TypedObject()
	case AMF0AVMPlus:
		// Enhanced RTMP / AMF0: next value is AMF3-coded.
		return d.Decode3()
	default:
		return nil, fmt.Errorf("amf0: unknown marker 0x%02x", marker)
	}
}

func (d *Decoder) readAMF0String() (string, error) {
	n, err := d.u16()
	if err != nil {
		return "", err
	}
	b, err := d.bytes(int(n))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (d *Decoder) readAMF0LongString() (string, error) {
	n, err := d.u32()
	if err != nil {
		return "", err
	}
	b, err := d.bytes(int(n))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (d *Decoder) readAMF0Object() (Object, error) {
	obj := Object{}
	d.amf0Refs = append(d.amf0Refs, obj)
	for {
		key, err := d.readAMF0String()
		if err != nil {
			return nil, err
		}
		if key == "" {
			end, err := d.u8()
			if err != nil {
				return nil, err
			}
			if end != AMF0ObjectEnd {
				return nil, errors.New("amf0: object missing end marker")
			}
			return obj, nil
		}
		v, err := d.Decode0()
		if err != nil {
			return nil, err
		}
		obj[key] = v
	}
}

func (d *Decoder) readAMF0ECMAArray() (ECMAArray, error) {
	// Count is approximate (Flash often lies); read until object-end.
	if _, err := d.u32(); err != nil {
		return nil, err
	}
	arr := ECMAArray{}
	d.amf0Refs = append(d.amf0Refs, arr)
	for {
		key, err := d.readAMF0String()
		if err != nil {
			return nil, err
		}
		if key == "" {
			end, err := d.u8()
			if err != nil {
				return nil, err
			}
			if end != AMF0ObjectEnd {
				return nil, errors.New("amf0: ecma-array missing end marker")
			}
			return arr, nil
		}
		v, err := d.Decode0()
		if err != nil {
			return nil, err
		}
		arr[key] = v
	}
}

func (d *Decoder) readAMF0StrictArray() (StrictArray, error) {
	n, err := d.u32()
	if err != nil {
		return nil, err
	}
	arr := make(StrictArray, 0, n)
	d.amf0Refs = append(d.amf0Refs, arr)
	for i := uint32(0); i < n; i++ {
		v, err := d.Decode0()
		if err != nil {
			return nil, err
		}
		arr = append(arr, v)
	}
	// Fix reference to point at final slice header (append may reallocate).
	d.amf0Refs[len(d.amf0Refs)-1] = arr
	return arr, nil
}

func (d *Decoder) readAMF0Date() (Date, error) {
	ms, err := d.f64()
	if err != nil {
		return Date{}, err
	}
	tz, err := d.u16()
	if err != nil {
		return Date{}, err
	}
	return Date{Millis: ms, Timezone: int16(tz)}, nil
}

func (d *Decoder) readAMF0TypedObject() (TypedObject, error) {
	name, err := d.readAMF0String()
	if err != nil {
		return TypedObject{}, err
	}
	fields := Object{}
	to := TypedObject{ClassName: name, Fields: fields}
	d.amf0Refs = append(d.amf0Refs, to)
	for {
		key, err := d.readAMF0String()
		if err != nil {
			return TypedObject{}, err
		}
		if key == "" {
			end, err := d.u8()
			if err != nil {
				return TypedObject{}, err
			}
			if end != AMF0ObjectEnd {
				return TypedObject{}, errors.New("amf0: typed-object missing end marker")
			}
			return to, nil
		}
		v, err := d.Decode0()
		if err != nil {
			return TypedObject{}, err
		}
		fields[key] = v
	}
}

// Encode0 writes vals as consecutive AMF0 values.
// Enhanced RTMP: prefer Object (not ECMAArray) when creating data.
func Encode0(vals []Value) ([]byte, error) {
	var out []byte
	for _, v := range vals {
		b, err := encode0(v)
		if err != nil {
			return nil, err
		}
		out = append(out, b...)
	}
	return out, nil
}

// encode0 encodes one value as AMF0.
func encode0(v Value) ([]byte, error) {
	if v == nil {
		return []byte{AMF0Null}, nil
	}
	switch x := v.(type) {
	case Undefined:
		return []byte{AMF0Undefined}, nil
	case Unsupported:
		return []byte{AMF0Unsupported}, nil
	case bool:
		if x {
			return []byte{AMF0Boolean, 1}, nil
		}
		return []byte{AMF0Boolean, 0}, nil
	case float64:
		return encodeAMF0Number(x), nil
	case float32:
		return encodeAMF0Number(float64(x)), nil
	case int:
		return encodeAMF0Number(float64(x)), nil
	case int8:
		return encodeAMF0Number(float64(x)), nil
	case int16:
		return encodeAMF0Number(float64(x)), nil
	case int32:
		return encodeAMF0Number(float64(x)), nil
	case int64:
		return encodeAMF0Number(float64(x)), nil
	case uint:
		return encodeAMF0Number(float64(x)), nil
	case uint8:
		return encodeAMF0Number(float64(x)), nil
	case uint16:
		return encodeAMF0Number(float64(x)), nil
	case uint32:
		return encodeAMF0Number(float64(x)), nil
	case uint64:
		return encodeAMF0Number(float64(x)), nil
	case string:
		return encodeAMF0String(x), nil
	case Object:
		return encodeAMF0Object(x)
	case map[string]Value:
		return encodeAMF0Object(Object(x))
	case map[string]interface{}:
		o := Object{}
		for k, vv := range x {
			o[k] = vv
		}
		return encodeAMF0Object(o)
	case ECMAArray:
		return encodeAMF0ECMAArray(x)
	case StrictArray:
		return encodeAMF0StrictArray(x)
	case []Value:
		return encodeAMF0StrictArray(StrictArray(x))
	case []interface{}:
		a := make(StrictArray, len(x))
		for i, vv := range x {
			a[i] = vv
		}
		return encodeAMF0StrictArray(a)
	case Date:
		return encodeAMF0Date(x), nil
	case XMLDocument:
		return encodeAMF0XMLDocument(string(x)), nil
	case TypedObject:
		return encodeAMF0TypedObject(x)
	case Integer:
		return encodeAMF0Number(float64(x)), nil
	default:
		return nil, fmt.Errorf("amf0: cannot encode %T", v)
	}
}

func encodeAMF0Number(f float64) []byte {
	b := make([]byte, 9)
	b[0] = AMF0Number
	binary.BigEndian.PutUint64(b[1:], math.Float64bits(f))
	return b
}

func encodeAMF0UTF8(s string) []byte {
	sb := []byte(s)
	if len(sb) > 0xffff {
		// caller should use long string
		sb = sb[:0xffff]
	}
	b := make([]byte, 2+len(sb))
	binary.BigEndian.PutUint16(b, uint16(len(sb)))
	copy(b[2:], sb)
	return b
}

func encodeAMF0String(s string) []byte {
	sb := []byte(s)
	if len(sb) > 0xffff {
		b := make([]byte, 5+len(sb))
		b[0] = AMF0LongString
		binary.BigEndian.PutUint32(b[1:], uint32(len(sb)))
		copy(b[5:], sb)
		return b
	}
	b := make([]byte, 1+2+len(sb))
	b[0] = AMF0String
	binary.BigEndian.PutUint16(b[1:], uint16(len(sb)))
	copy(b[3:], sb)
	return b
}

func encodeAMF0Object(obj Object) ([]byte, error) {
	out := []byte{AMF0Object}
	for k, v := range obj {
		out = append(out, encodeAMF0UTF8(k)...)
		ev, err := encode0(v)
		if err != nil {
			return nil, err
		}
		out = append(out, ev...)
	}
	out = append(out, 0x00, 0x00, AMF0ObjectEnd)
	return out, nil
}

func encodeAMF0ECMAArray(arr ECMAArray) ([]byte, error) {
	out := make([]byte, 5)
	out[0] = AMF0ECMAArray
	binary.BigEndian.PutUint32(out[1:], uint32(len(arr)))
	for k, v := range arr {
		out = append(out, encodeAMF0UTF8(k)...)
		ev, err := encode0(v)
		if err != nil {
			return nil, err
		}
		out = append(out, ev...)
	}
	out = append(out, 0x00, 0x00, AMF0ObjectEnd)
	return out, nil
}

func encodeAMF0StrictArray(arr StrictArray) ([]byte, error) {
	out := make([]byte, 5)
	out[0] = AMF0StrictArray
	binary.BigEndian.PutUint32(out[1:], uint32(len(arr)))
	for _, v := range arr {
		ev, err := encode0(v)
		if err != nil {
			return nil, err
		}
		out = append(out, ev...)
	}
	return out, nil
}

func encodeAMF0Date(d Date) []byte {
	b := make([]byte, 11)
	b[0] = AMF0Date
	binary.BigEndian.PutUint64(b[1:], math.Float64bits(d.Millis))
	binary.BigEndian.PutUint16(b[9:], uint16(d.Timezone))
	return b
}

func encodeAMF0XMLDocument(s string) []byte {
	sb := []byte(s)
	b := make([]byte, 5+len(sb))
	b[0] = AMF0XMLDocument
	binary.BigEndian.PutUint32(b[1:], uint32(len(sb)))
	copy(b[5:], sb)
	return b
}

func encodeAMF0TypedObject(t TypedObject) ([]byte, error) {
	out := []byte{AMF0TypedObject}
	out = append(out, encodeAMF0UTF8(t.ClassName)...)
	// AMF0 uses Fields; if only AMF3 sealed Keys/Values present, merge into Fields for wire.
	fields := t.Fields
	if fields == nil {
		fields = Object{}
	}
	if len(t.Keys) > 0 {
		fields = Object{}
		for k, v := range t.Fields {
			fields[k] = v
		}
		for i, k := range t.Keys {
			if i < len(t.Values) {
				fields[k] = t.Values[i]
			}
		}
	}
	for k, v := range fields {
		out = append(out, encodeAMF0UTF8(k)...)
		ev, err := encode0(v)
		if err != nil {
			return nil, err
		}
		out = append(out, ev...)
	}
	out = append(out, 0x00, 0x00, AMF0ObjectEnd)
	return out, nil
}
