// Package amf implements Action Message Format AMF0 and AMF3.
// Specs: Adobe AMF0 / AMF3; usage notes from veovera/enhanced-rtmp.
package amf

import "time"

// Value is a decoded AMF value.
type Value interface{}

// AMF0 type markers.
const (
	AMF0Number      = 0x00
	AMF0Boolean     = 0x01
	AMF0String      = 0x02
	AMF0Object      = 0x03
	AMF0Movieclip   = 0x04 // reserved
	AMF0Null        = 0x05
	AMF0Undefined   = 0x06
	AMF0Reference   = 0x07
	AMF0ECMAArray   = 0x08
	AMF0ObjectEnd   = 0x09
	AMF0StrictArray = 0x0a
	AMF0Date        = 0x0b
	AMF0LongString  = 0x0c
	AMF0Unsupported = 0x0d
	AMF0Recordset   = 0x0e // reserved
	AMF0XMLDocument = 0x0f
	AMF0TypedObject = 0x10
	AMF0AVMPlus     = 0x11 // next value is AMF3
)

// AMF3 type markers.
const (
	AMF3Undefined  = 0x00
	AMF3Null       = 0x01
	AMF3False      = 0x02
	AMF3True       = 0x03
	AMF3Integer    = 0x04
	AMF3Double     = 0x05
	AMF3String     = 0x06
	AMF3XMLDoc     = 0x07
	AMF3Date       = 0x08
	AMF3Array      = 0x09
	AMF3Object     = 0x0a
	AMF3XML        = 0x0b
	AMF3ByteArray  = 0x0c
	AMF3VectorInt  = 0x0d
	AMF3VectorUint = 0x0e
	AMF3VectorDbl  = 0x0f
	AMF3VectorObj  = 0x10
	AMF3Dictionary = 0x11
)

// ---------------------------------------------------------------------------
// Shared Go types (same concept in AMF0 and AMF3 uses the same name)
// ---------------------------------------------------------------------------

// Undefined is AMF undefined (distinct from null/nil). AMF0 marker 0x06 / AMF3 0x00.
type Undefined struct{}

func (Undefined) String() string { return "undefined" }

// Object is a dynamic name→value map.
// AMF0 Object (0x03); AMF3 anonymous dynamic object (0x0a).
// Enhanced RTMP: prefer Object when writing; when reading also accept ECMAArray via AsObject.
type Object map[string]Value

// Date is milliseconds since Unix epoch.
// AMF0 Date (0x0b) also has Timezone; AMF3 Date (0x08) leaves Timezone=0.
type Date struct {
	Millis   float64
	Timezone int16 // AMF0 only
}

func (d Date) Time() time.Time {
	sec := int64(d.Millis) / 1000
	nsec := (int64(d.Millis) % 1000) * int64(time.Millisecond)
	return time.Unix(sec, nsec).UTC()
}

// XMLDocument is AMF0 xml-document (0x0f) / AMF3 xml-doc (0x07).
type XMLDocument string

// TypedObject is a classed / traits object.
// AMF0 typed-object (0x10): ClassName + Fields.
// AMF3 object (0x0a) with class name or sealed keys: ClassName + Keys/Values + optional Dynamic Fields.
type TypedObject struct {
	ClassName      string
	Fields         Object   // AMF0 properties; AMF3 dynamic properties
	Keys           []string // AMF3 sealed member names (order preserved)
	Values         []Value  // AMF3 sealed member values
	Dynamic        bool     // AMF3: may have dynamic Fields
	Externalizable bool     // AMF3: body not supported
}

// ---------------------------------------------------------------------------
// AMF0-only types
// ---------------------------------------------------------------------------

// Unsupported is AMF0 unsupported-marker (0x0d).
type Unsupported struct{}

func (Unsupported) String() string { return "unsupported" }

// ECMAArray is AMF0 ECMA Array (0x08), associative.
type ECMAArray map[string]Value

// StrictArray is AMF0 strict (dense) array (0x0a).
type StrictArray []Value

// ---------------------------------------------------------------------------
// AMF3-only types
// ---------------------------------------------------------------------------

// Integer is AMF3 integer (0x04), signed 29-bit.
type Integer int32

// XML is AMF3 xml (0x0b). Distinct from XMLDocument (xml-doc).
type XML string

// ByteArray is AMF3 byte-array (0x0c).
type ByteArray []byte

// Array is AMF3 array (0x09): dense part plus optional associative part.
type Array struct {
	Dense []Value
	Assoc Object
}

// VectorInt is AMF3 vector-int (0x0d).
type VectorInt struct {
	Fixed  bool
	Values []int32
}

// VectorUint is AMF3 vector-uint (0x0e).
type VectorUint struct {
	Fixed  bool
	Values []uint32
}

// VectorDouble is AMF3 vector-double (0x0f).
type VectorDouble struct {
	Fixed  bool
	Values []float64
}

// VectorObject is AMF3 vector-object (0x10).
type VectorObject struct {
	TypeName string
	Fixed    bool
	Values   []Value
}

// DictEntry is one key/value pair in a Dictionary (order preserved).
type DictEntry struct {
	Key   Value
	Value Value
}

// Dictionary is AMF3 dictionary (0x11).
type Dictionary struct {
	WeakKeys bool
	Entries  []DictEntry
}

// AsObject returns an Object view of v.
// Enhanced RTMP: when decoding, accept either Object or ECMAArray.
func AsObject(v Value) (Object, bool) {
	switch o := v.(type) {
	case Object:
		return o, true
	case ECMAArray:
		return Object(o), true
	case map[string]Value:
		return Object(o), true
	case map[string]interface{}:
		out := Object{}
		for k, vv := range o {
			out[k] = vv
		}
		return out, true
	default:
		return nil, false
	}
}
