package amf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

// Decode3All reads all consecutive AMF3 values from b.
func Decode3All(b []byte) ([]Value, error) {
	d := &Decoder{data: b}
	var out []Value
	for d.off < len(d.data) {
		v, err := d.Decode3()
		if err != nil {
			return out, err
		}
		out = append(out, v)
	}
	return out, nil
}

// Decode3 reads one AMF3 value from b and returns remaining bytes.
func Decode3(b []byte) (Value, []byte, error) {
	d := &Decoder{data: b}
	v, err := d.Decode3()
	if err != nil {
		return nil, b, err
	}
	return v, d.data[d.off:], nil
}

// Encode3All encodes vals as consecutive AMF3 values.
func Encode3All(vals []Value) ([]byte, error) {
	e := &encoder3{}
	var out []byte
	for _, v := range vals {
		b, err := e.encode(v)
		if err != nil {
			return nil, err
		}
		out = append(out, b...)
	}
	return out, nil
}

// Encode3 encodes one value as AMF3.
func Encode3(v Value) ([]byte, error) {
	e := &encoder3{}
	return e.encode(v)
}

type encoder3 struct {
	strings []string
	// objects/traits refs omitted on encode: always write inline (enough for RTMP).
}

// Decode3 reads one AMF3-coded value.
func (d *Decoder) Decode3() (Value, error) {
	marker, err := d.u8()
	if err != nil {
		return nil, err
	}
	switch marker {
	case AMF3Undefined:
		return Undefined{}, nil
	case AMF3Null:
		return nil, nil
	case AMF3False:
		return false, nil
	case AMF3True:
		return true, nil
	case AMF3Integer:
		u, err := d.readU29()
		if err != nil {
			return nil, err
		}
		n := int32(u)
		if u >= 0x10000000 {
			n = int32(u - 0x20000000)
		}
		return Integer(n), nil
	case AMF3Double:
		return d.f64()
	case AMF3String:
		return d.readAMF3StringRaw()
	case AMF3XMLDoc:
		s, err := d.readAMF3StringRaw()
		if err != nil {
			return nil, err
		}
		return XMLDocument(s), nil
	case AMF3Date:
		return d.readAMF3Date()
	case AMF3Array:
		return d.readAMF3Array()
	case AMF3Object:
		return d.readAMF3Object()
	case AMF3XML:
		s, err := d.readAMF3StringRaw()
		if err != nil {
			return nil, err
		}
		return XML(s), nil
	case AMF3ByteArray:
		return d.readAMF3ByteArray()
	case AMF3VectorInt:
		return d.readAMF3VectorInt()
	case AMF3VectorUint:
		return d.readAMF3VectorUint()
	case AMF3VectorDbl:
		return d.readAMF3VectorDouble()
	case AMF3VectorObj:
		return d.readAMF3VectorObject()
	case AMF3Dictionary:
		return d.readAMF3Dictionary()
	default:
		return nil, fmt.Errorf("amf3: unknown marker 0x%02x", marker)
	}
}

func (d *Decoder) readU29() (uint32, error) {
	var v uint32
	for i := 0; i < 3; i++ {
		b, err := d.u8()
		if err != nil {
			return 0, err
		}
		if b < 0x80 {
			return v<<7 | uint32(b), nil
		}
		v = v<<7 | uint32(b&0x7f)
	}
	b, err := d.u8()
	if err != nil {
		return 0, err
	}
	return v<<8 | uint32(b), nil
}

func (d *Decoder) readAMF3StringRaw() (string, error) {
	u, err := d.readU29()
	if err != nil {
		return "", err
	}
	if u&1 == 0 {
		idx := int(u >> 1)
		if idx >= len(d.amf3Strings) {
			return "", fmt.Errorf("amf3: bad string ref %d", idx)
		}
		return d.amf3Strings[idx], nil
	}
	length := int(u >> 1)
	if length == 0 {
		return "", nil
	}
	b, err := d.bytes(length)
	if err != nil {
		return "", err
	}
	s := string(b)
	d.amf3Strings = append(d.amf3Strings, s)
	return s, nil
}

func (d *Decoder) readAMF3Date() (Date, error) {
	u, err := d.readU29()
	if err != nil {
		return Date{}, err
	}
	if u&1 == 0 {
		idx := int(u >> 1)
		if idx >= len(d.amf3Objects) {
			return Date{}, fmt.Errorf("amf3: bad date ref %d", idx)
		}
		dt, ok := d.amf3Objects[idx].(Date)
		if !ok {
			return Date{}, errors.New("amf3: date ref not a Date")
		}
		return dt, nil
	}
	ms, err := d.f64()
	if err != nil {
		return Date{}, err
	}
	dt := Date{Millis: ms}
	d.amf3Objects = append(d.amf3Objects, dt)
	return dt, nil
}

func (d *Decoder) readAMF3ByteArray() (ByteArray, error) {
	u, err := d.readU29()
	if err != nil {
		return nil, err
	}
	if u&1 == 0 {
		idx := int(u >> 1)
		if idx >= len(d.amf3Objects) {
			return nil, fmt.Errorf("amf3: bad bytearray ref %d", idx)
		}
		ba, ok := d.amf3Objects[idx].(ByteArray)
		if !ok {
			return nil, errors.New("amf3: bytearray ref type mismatch")
		}
		return ba, nil
	}
	length := int(u >> 1)
	b, err := d.bytes(length)
	if err != nil {
		return nil, err
	}
	ba := ByteArray(append([]byte(nil), b...))
	d.amf3Objects = append(d.amf3Objects, ba)
	return ba, nil
}

func (d *Decoder) readAMF3Array() (Array, error) {
	u, err := d.readU29()
	if err != nil {
		return Array{}, err
	}
	if u&1 == 0 {
		idx := int(u >> 1)
		if idx >= len(d.amf3Objects) {
			return Array{}, fmt.Errorf("amf3: bad array ref %d", idx)
		}
		arr, ok := d.amf3Objects[idx].(Array)
		if !ok {
			return Array{}, errors.New("amf3: array ref type mismatch")
		}
		return arr, nil
	}
	denseLen := int(u >> 1)
	arr := Array{Assoc: Object{}}
	refIdx := len(d.amf3Objects)
	d.amf3Objects = append(d.amf3Objects, arr)

	for {
		name, err := d.readAMF3StringRaw()
		if err != nil {
			return Array{}, err
		}
		if name == "" {
			break
		}
		v, err := d.Decode3()
		if err != nil {
			return Array{}, err
		}
		arr.Assoc[name] = v
	}
	if len(arr.Assoc) == 0 {
		arr.Assoc = nil
	}

	arr.Dense = make([]Value, 0, denseLen)
	for i := 0; i < denseLen; i++ {
		v, err := d.Decode3()
		if err != nil {
			return Array{}, err
		}
		arr.Dense = append(arr.Dense, v)
	}
	d.amf3Objects[refIdx] = arr
	return arr, nil
}

func (d *Decoder) readAMF3Object() (Value, error) {
	u, err := d.readU29()
	if err != nil {
		return nil, err
	}
	if u&1 == 0 {
		idx := int(u >> 1)
		if idx >= len(d.amf3Objects) {
			return nil, fmt.Errorf("amf3: bad object ref %d", idx)
		}
		return d.amf3Objects[idx], nil
	}

	u >>= 1
	var tr amf3Trait
	if u&1 == 0 {
		idx := int(u >> 1)
		if idx >= len(d.amf3Traits) {
			return nil, fmt.Errorf("amf3: bad trait ref %d", idx)
		}
		tr = d.amf3Traits[idx]
	} else {
		u >>= 1
		tr.externalizable = u&1 == 1
		u >>= 1
		tr.dynamic = u&1 == 1
		u >>= 1
		nKeys := int(u)
		name, err := d.readAMF3StringRaw()
		if err != nil {
			return nil, err
		}
		tr.className = name
		tr.keys = make([]string, nKeys)
		for i := 0; i < nKeys; i++ {
			k, err := d.readAMF3StringRaw()
			if err != nil {
				return nil, err
			}
			tr.keys[i] = k
		}
		d.amf3Traits = append(d.amf3Traits, tr)
	}

	if tr.externalizable {
		return nil, fmt.Errorf("amf3: externalizable class %q not supported", tr.className)
	}

	// Anonymous dynamic, no sealed keys → plain Object (shared name with AMF0 Object).
	if tr.className == "" && len(tr.keys) == 0 && tr.dynamic {
		o := Object{}
		refIdx := len(d.amf3Objects)
		d.amf3Objects = append(d.amf3Objects, o)
		for {
			name, err := d.readAMF3StringRaw()
			if err != nil {
				return nil, err
			}
			if name == "" {
				break
			}
			v, err := d.Decode3()
			if err != nil {
				return nil, err
			}
			o[name] = v
		}
		d.amf3Objects[refIdx] = o
		return o, nil
	}
	if tr.className == "" && len(tr.keys) == 0 && !tr.dynamic {
		o := Object{}
		d.amf3Objects = append(d.amf3Objects, o)
		return o, nil
	}

	to := TypedObject{
		ClassName:      tr.className,
		Dynamic:        tr.dynamic,
		Externalizable: tr.externalizable,
		Keys:           tr.keys,
		Values:         make([]Value, len(tr.keys)),
	}
	refIdx := len(d.amf3Objects)
	d.amf3Objects = append(d.amf3Objects, to)

	for i := range tr.keys {
		v, err := d.Decode3()
		if err != nil {
			return nil, err
		}
		to.Values[i] = v
	}
	if tr.dynamic {
		to.Fields = Object{}
		for {
			name, err := d.readAMF3StringRaw()
			if err != nil {
				return nil, err
			}
			if name == "" {
				break
			}
			v, err := d.Decode3()
			if err != nil {
				return nil, err
			}
			to.Fields[name] = v
		}
	}
	d.amf3Objects[refIdx] = to
	return to, nil
}

func (d *Decoder) readAMF3VectorInt() (Value, error) {
	u, err := d.readU29()
	if err != nil {
		return nil, err
	}
	if u&1 == 0 {
		idx := int(u >> 1)
		if idx >= len(d.amf3Objects) {
			return nil, fmt.Errorf("amf3: bad vector-int ref %d", idx)
		}
		return d.amf3Objects[idx], nil
	}
	n := int(u >> 1)
	fixedB, err := d.u8()
	if err != nil {
		return nil, err
	}
	v := VectorInt{Fixed: fixedB != 0, Values: make([]int32, n)}
	for i := 0; i < n; i++ {
		if err := d.need(4); err != nil {
			return nil, err
		}
		v.Values[i] = int32(binary.BigEndian.Uint32(d.data[d.off:]))
		d.off += 4
	}
	d.amf3Objects = append(d.amf3Objects, v)
	return v, nil
}

func (d *Decoder) readAMF3VectorUint() (Value, error) {
	u, err := d.readU29()
	if err != nil {
		return nil, err
	}
	if u&1 == 0 {
		idx := int(u >> 1)
		if idx >= len(d.amf3Objects) {
			return nil, fmt.Errorf("amf3: bad vector-uint ref %d", idx)
		}
		return d.amf3Objects[idx], nil
	}
	n := int(u >> 1)
	fixedB, err := d.u8()
	if err != nil {
		return nil, err
	}
	v := VectorUint{Fixed: fixedB != 0, Values: make([]uint32, n)}
	for i := 0; i < n; i++ {
		if err := d.need(4); err != nil {
			return nil, err
		}
		v.Values[i] = binary.BigEndian.Uint32(d.data[d.off:])
		d.off += 4
	}
	d.amf3Objects = append(d.amf3Objects, v)
	return v, nil
}

func (d *Decoder) readAMF3VectorDouble() (Value, error) {
	u, err := d.readU29()
	if err != nil {
		return nil, err
	}
	if u&1 == 0 {
		idx := int(u >> 1)
		if idx >= len(d.amf3Objects) {
			return nil, fmt.Errorf("amf3: bad vector-double ref %d", idx)
		}
		return d.amf3Objects[idx], nil
	}
	n := int(u >> 1)
	fixedB, err := d.u8()
	if err != nil {
		return nil, err
	}
	v := VectorDouble{Fixed: fixedB != 0, Values: make([]float64, n)}
	for i := 0; i < n; i++ {
		f, err := d.f64()
		if err != nil {
			return nil, err
		}
		v.Values[i] = f
	}
	d.amf3Objects = append(d.amf3Objects, v)
	return v, nil
}

func (d *Decoder) readAMF3VectorObject() (Value, error) {
	u, err := d.readU29()
	if err != nil {
		return nil, err
	}
	if u&1 == 0 {
		idx := int(u >> 1)
		if idx >= len(d.amf3Objects) {
			return nil, fmt.Errorf("amf3: bad vector-object ref %d", idx)
		}
		return d.amf3Objects[idx], nil
	}
	n := int(u >> 1)
	fixedB, err := d.u8()
	if err != nil {
		return nil, err
	}
	typeName, err := d.readAMF3StringRaw()
	if err != nil {
		return nil, err
	}
	v := VectorObject{TypeName: typeName, Fixed: fixedB != 0, Values: make([]Value, n)}
	refIdx := len(d.amf3Objects)
	d.amf3Objects = append(d.amf3Objects, v)
	for i := 0; i < n; i++ {
		item, err := d.Decode3()
		if err != nil {
			return nil, err
		}
		v.Values[i] = item
	}
	d.amf3Objects[refIdx] = v
	return v, nil
}

func (d *Decoder) readAMF3Dictionary() (Value, error) {
	u, err := d.readU29()
	if err != nil {
		return nil, err
	}
	if u&1 == 0 {
		idx := int(u >> 1)
		if idx >= len(d.amf3Objects) {
			return nil, fmt.Errorf("amf3: bad dictionary ref %d", idx)
		}
		return d.amf3Objects[idx], nil
	}
	n := int(u >> 1)
	weak, err := d.u8()
	if err != nil {
		return nil, err
	}
	dict := Dictionary{WeakKeys: weak != 0, Entries: make([]DictEntry, 0, n)}
	refIdx := len(d.amf3Objects)
	d.amf3Objects = append(d.amf3Objects, dict)
	for i := 0; i < n; i++ {
		k, err := d.Decode3()
		if err != nil {
			return nil, err
		}
		v, err := d.Decode3()
		if err != nil {
			return nil, err
		}
		dict.Entries = append(dict.Entries, DictEntry{Key: k, Value: v})
	}
	d.amf3Objects[refIdx] = dict
	return dict, nil
}

// ---------------------------------------------------------------------------
// AMF3 encode
// ---------------------------------------------------------------------------

func EncodeU29(v uint32) []byte {
	if v <= 0x7f {
		return []byte{byte(v)}
	}
	if v <= 0x3fff {
		return []byte{byte((v>>7)&0x7f | 0x80), byte(v & 0x7f)}
	}
	if v <= 0x1fffff {
		return []byte{
			byte((v>>14)&0x7f | 0x80),
			byte((v>>7)&0x7f | 0x80),
			byte(v & 0x7f),
		}
	}
	return []byte{
		byte((v>>22)&0x7f | 0x80),
		byte((v>>15)&0x7f | 0x80),
		byte((v>>8)&0x7f | 0x80),
		byte(v),
	}
}

func (e *encoder3) writeString(s string) []byte {
	if s == "" {
		return []byte{0x01}
	}
	for i, prev := range e.strings {
		if prev == s {
			return EncodeU29(uint32(i << 1))
		}
	}
	e.strings = append(e.strings, s)
	sb := []byte(s)
	out := EncodeU29(uint32(len(sb)<<1 | 1))
	return append(out, sb...)
}

func (e *encoder3) encode(v Value) ([]byte, error) {
	if v == nil {
		return []byte{AMF3Null}, nil
	}
	switch x := v.(type) {
	case Undefined:
		return []byte{AMF3Undefined}, nil
	case bool:
		if x {
			return []byte{AMF3True}, nil
		}
		return []byte{AMF3False}, nil
	case Integer:
		return e.encodeInteger(int32(x)), nil
	case int32:
		return e.encodeInteger(x), nil
	case int:
		if x >= -0x10000000 && x <= 0x0fffffff {
			return e.encodeInteger(int32(x)), nil
		}
		return encodeAMF3Double(float64(x)), nil
	case int64:
		if x >= -0x10000000 && x <= 0x0fffffff {
			return e.encodeInteger(int32(x)), nil
		}
		return encodeAMF3Double(float64(x)), nil
	case float64:
		return encodeAMF3Double(x), nil
	case float32:
		return encodeAMF3Double(float64(x)), nil
	case string:
		out := []byte{AMF3String}
		return append(out, e.writeString(x)...), nil
	case XMLDocument:
		out := []byte{AMF3XMLDoc}
		return append(out, e.writeString(string(x))...), nil
	case XML:
		out := []byte{AMF3XML}
		return append(out, e.writeString(string(x))...), nil
	case Date:
		return e.encodeDate(x), nil
	case ByteArray:
		return e.encodeByteArray(x), nil
	case Object:
		return e.encodeObject(x)
	case map[string]Value:
		return e.encodeObject(Object(x))
	case map[string]interface{}:
		o := Object{}
		for k, vv := range x {
			o[k] = vv
		}
		return e.encodeObject(o)
	case TypedObject:
		return e.encodeTypedObject(x)
	case Array:
		return e.encodeArray(x)
	case StrictArray:
		// treat as dense AMF3 Array
		return e.encodeArray(Array{Dense: []Value(x)})
	case []Value:
		return e.encodeArray(Array{Dense: x})
	case VectorInt:
		return e.encodeVectorInt(x), nil
	case VectorUint:
		return e.encodeVectorUint(x), nil
	case VectorDouble:
		return e.encodeVectorDouble(x), nil
	case VectorObject:
		return e.encodeVectorObject(x)
	case Dictionary:
		return e.encodeDictionary(x)
	default:
		return nil, fmt.Errorf("amf3: cannot encode %T", v)
	}
}

func (e *encoder3) encodeInteger(n int32) []byte {
	u := uint32(n)
	if n < 0 {
		u = uint32(n) & 0x1fffffff
	}
	out := []byte{AMF3Integer}
	return append(out, EncodeU29(u)...)
}

func encodeAMF3Double(f float64) []byte {
	b := make([]byte, 9)
	b[0] = AMF3Double
	binary.BigEndian.PutUint64(b[1:], math.Float64bits(f))
	return b
}

func (e *encoder3) encodeDate(d Date) []byte {
	out := []byte{AMF3Date, 0x01} // inline
	fb := make([]byte, 8)
	binary.BigEndian.PutUint64(fb, math.Float64bits(d.Millis))
	return append(out, fb...)
}

func (e *encoder3) encodeByteArray(ba ByteArray) []byte {
	out := []byte{AMF3ByteArray}
	out = append(out, EncodeU29(uint32(len(ba)<<1|1))...)
	return append(out, ba...)
}

func (e *encoder3) encodeObject(obj Object) ([]byte, error) {
	// anonymous dynamic: U29 = 0x0b
	out := []byte{AMF3Object, 0x0b}
	out = append(out, e.writeString("")...) // empty class name
	for k, v := range obj {
		out = append(out, e.writeString(k)...)
		ev, err := e.encode(v)
		if err != nil {
			return nil, err
		}
		out = append(out, ev...)
	}
	out = append(out, e.writeString("")...) // end dynamic
	return out, nil
}

func (e *encoder3) encodeTypedObject(t TypedObject) ([]byte, error) {
	if t.Externalizable {
		return nil, fmt.Errorf("amf3: cannot encode externalizable %q", t.ClassName)
	}
	nKeys := len(t.Keys)
	if nKeys != len(t.Values) {
		return nil, errors.New("amf3: TypedObject Keys/Values length mismatch")
	}
	dyn := t.Dynamic || len(t.Fields) > 0
	// U29: count<<4 | dyn<<3 | ext<<2 | traits<<1 | obj
	u := uint32(nKeys)<<4 | 0x03 // traits inline + object inline
	if dyn {
		u |= 1 << 3
	}
	out := []byte{AMF3Object}
	out = append(out, EncodeU29(u)...)
	out = append(out, e.writeString(t.ClassName)...)
	for _, k := range t.Keys {
		out = append(out, e.writeString(k)...)
	}
	for _, v := range t.Values {
		ev, err := e.encode(v)
		if err != nil {
			return nil, err
		}
		out = append(out, ev...)
	}
	if dyn {
		for k, v := range t.Fields {
			out = append(out, e.writeString(k)...)
			ev, err := e.encode(v)
			if err != nil {
				return nil, err
			}
			out = append(out, ev...)
		}
		out = append(out, e.writeString("")...)
	}
	return out, nil
}

func (e *encoder3) encodeArray(a Array) ([]byte, error) {
	out := []byte{AMF3Array}
	out = append(out, EncodeU29(uint32(len(a.Dense)<<1|1))...)
	for k, v := range a.Assoc {
		out = append(out, e.writeString(k)...)
		ev, err := e.encode(v)
		if err != nil {
			return nil, err
		}
		out = append(out, ev...)
	}
	out = append(out, e.writeString("")...)
	for _, v := range a.Dense {
		ev, err := e.encode(v)
		if err != nil {
			return nil, err
		}
		out = append(out, ev...)
	}
	return out, nil
}

func (e *encoder3) encodeVectorInt(v VectorInt) []byte {
	out := []byte{AMF3VectorInt}
	out = append(out, EncodeU29(uint32(len(v.Values)<<1|1))...)
	if v.Fixed {
		out = append(out, 1)
	} else {
		out = append(out, 0)
	}
	for _, n := range v.Values {
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(n))
		out = append(out, b...)
	}
	return out
}

func (e *encoder3) encodeVectorUint(v VectorUint) []byte {
	out := []byte{AMF3VectorUint}
	out = append(out, EncodeU29(uint32(len(v.Values)<<1|1))...)
	if v.Fixed {
		out = append(out, 1)
	} else {
		out = append(out, 0)
	}
	for _, n := range v.Values {
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, n)
		out = append(out, b...)
	}
	return out
}

func (e *encoder3) encodeVectorDouble(v VectorDouble) []byte {
	out := []byte{AMF3VectorDbl}
	out = append(out, EncodeU29(uint32(len(v.Values)<<1|1))...)
	if v.Fixed {
		out = append(out, 1)
	} else {
		out = append(out, 0)
	}
	for _, f := range v.Values {
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, math.Float64bits(f))
		out = append(out, b...)
	}
	return out
}

func (e *encoder3) encodeVectorObject(v VectorObject) ([]byte, error) {
	out := []byte{AMF3VectorObj}
	out = append(out, EncodeU29(uint32(len(v.Values)<<1|1))...)
	if v.Fixed {
		out = append(out, 1)
	} else {
		out = append(out, 0)
	}
	out = append(out, e.writeString(v.TypeName)...)
	for _, item := range v.Values {
		ev, err := e.encode(item)
		if err != nil {
			return nil, err
		}
		out = append(out, ev...)
	}
	return out, nil
}

func (e *encoder3) encodeDictionary(dict Dictionary) ([]byte, error) {
	out := []byte{AMF3Dictionary}
	out = append(out, EncodeU29(uint32(len(dict.Entries)<<1|1))...)
	if dict.WeakKeys {
		out = append(out, 1)
	} else {
		out = append(out, 0)
	}
	for _, ent := range dict.Entries {
		ek, err := e.encode(ent.Key)
		if err != nil {
			return nil, err
		}
		out = append(out, ek...)
		ev, err := e.encode(ent.Value)
		if err != nil {
			return nil, err
		}
		out = append(out, ev...)
	}
	return out, nil
}
