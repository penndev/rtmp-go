package amf_test

import (
	"reflect"
	"testing"

	"github.com/penndev/rtmp/amf"
)

func roundTrip3(t *testing.T, v amf.Value) amf.Value {
	t.Helper()
	b, err := amf.Encode3(v)
	if err != nil {
		t.Fatalf("Encode3(%T): %v", v, err)
	}
	vals, err := amf.Decode3(b)
	if err != nil {
		t.Fatalf("Decode3(%T): %v\nhex=%x", v, err, b)
	}
	if len(vals) != 1 {
		t.Fatalf("Decode3(%T): got %d values %#v", v, len(vals), vals)
	}
	return vals[0]
}

func TestAMF3_Null(t *testing.T) {
	if got := roundTrip3(t, nil); got != nil {
		t.Fatalf("got %#v", got)
	}
}

func TestAMF3_Undefined(t *testing.T) {
	if got := roundTrip3(t, amf.Undefined{}); got != (amf.Undefined{}) {
		t.Fatalf("got %#v", got)
	}
}

func TestAMF3_Boolean(t *testing.T) {
	for _, want := range []bool{true, false} {
		if got := roundTrip3(t, want); got != want {
			t.Fatalf("got %v", got)
		}
	}
}

func TestAMF3_Integer(t *testing.T) {
	for _, want := range []amf.Integer{0, 1, 127, 128, 16383, -1, -0x10000000} {
		got, ok := roundTrip3(t, want).(amf.Integer)
		if !ok || got != want {
			t.Fatalf("want %d got %v (%T)", want, got, got)
		}
	}
}

func TestAMF3_Double(t *testing.T) {
	if got := roundTrip3(t, 3.14); got != 3.14 {
		t.Fatalf("got %v", got)
	}
}

func TestAMF3_String(t *testing.T) {
	if got := roundTrip3(t, ""); got != "" {
		t.Fatalf("empty got %q", got)
	}
	if got := roundTrip3(t, "hi"); got != "hi" {
		t.Fatalf("got %q", got)
	}
	// string reference within one message
	b, err := amf.Encode3All([]amf.Value{"hi", "hi"})
	if err != nil {
		t.Fatal(err)
	}
	vals, err := amf.Decode3(b)
	if err != nil {
		t.Fatal(err)
	}
	if len(vals) != 2 || vals[0] != "hi" || vals[1] != "hi" {
		t.Fatalf("%#v bytes=%x", vals, b)
	}
}

func TestAMF3_XMLDocument(t *testing.T) {
	in := amf.XMLDocument(`<?xml version="1.0"?>
<root>
  <item id="1">hello</item>
</root>`)
	b, err := amf.Encode3(in)
	if err != nil {
		t.Fatalf("Encode3: %v", err)
	}
	vals, err := amf.Decode3(b)
	if err != nil {
		t.Fatalf("Decode3: %v\nhex=%x", err, b)
	}
	if len(vals) != 1 {
		t.Fatalf("got %d values %#v", len(vals), vals)
	}
	got, ok := vals[0].(amf.XMLDocument)
	if !ok || got != in {
		t.Fatalf("got %#v", vals[0])
	}
}

func TestAMF3_XML(t *testing.T) {
	in := amf.XML(`<node id="1">hello</node>`)
	b, err := amf.Encode3(in)
	if err != nil {
		t.Fatalf("Encode3: %v", err)
	}
	vals, err := amf.Decode3(b)
	if err != nil {
		t.Fatalf("Decode3: %v\nhex=%x", err, b)
	}
	if len(vals) != 1 {
		t.Fatalf("got %d values %#v", len(vals), vals)
	}
	got, ok := vals[0].(amf.XML)
	if !ok || got != in {
		t.Fatalf("got %#v", vals[0])
	}
}

func TestAMF3_Date(t *testing.T) {
	in := amf.Date{Millis: 2000}
	got, ok := roundTrip3(t, in).(amf.Date)
	if !ok || got.Millis != 2000 {
		t.Fatalf("got %#v", got)
	}
}

func TestAMF3_Array(t *testing.T) {
	in := amf.Array{
		Dense: []amf.Value{true, false},
		Assoc: amf.Object{"a": amf.Integer(1)},
	}
	got, ok := roundTrip3(t, in).(amf.Array)
	if !ok {
		t.Fatalf("type %T", got)
	}
	if len(got.Dense) != 2 || got.Dense[0] != true || got.Dense[1] != false {
		t.Fatalf("dense %#v", got.Dense)
	}
	if got.Assoc["a"] != amf.Integer(1) {
		t.Fatalf("assoc %#v", got.Assoc)
	}
}

func TestAMF3_Object(t *testing.T) {
	in := amf.Object{"x": true, "n": amf.Integer(7)}
	got, ok := roundTrip3(t, in).(amf.Object)
	if !ok || got["x"] != true || got["n"] != amf.Integer(7) {
		t.Fatalf("got %#v", got)
	}
}

func TestAMF3_TypedObject(t *testing.T) {
	in := amf.TypedObject{
		ClassName: "Box",
		Keys:      []string{"w"},
		Values:    []amf.Value{amf.Integer(10)},
		Dynamic:   true,
		Fields:    amf.Object{"extra": false},
	}
	got, ok := roundTrip3(t, in).(amf.TypedObject)
	if !ok {
		t.Fatalf("type %T %#v", got, got)
	}
	if got.ClassName != "Box" || len(got.Keys) != 1 || got.Keys[0] != "w" || got.Values[0] != amf.Integer(10) {
		t.Fatalf("sealed %#v", got)
	}
	if !got.Dynamic || got.Fields["extra"] != false {
		t.Fatalf("dynamic %#v", got)
	}
}

func TestAMF3_ByteArray(t *testing.T) {
	in := amf.ByteArray{0x10, 0x20}
	got, ok := roundTrip3(t, in).(amf.ByteArray)
	if !ok || !reflect.DeepEqual([]byte(got), []byte(in)) {
		t.Fatalf("got %#v", got)
	}
}

func TestAMF3_VectorInt(t *testing.T) {
	in := amf.VectorInt{Fixed: true, Values: []int32{1, -2, 3}}
	got, ok := roundTrip3(t, in).(amf.VectorInt)
	if !ok || !reflect.DeepEqual(got, in) {
		t.Fatalf("got %#v", got)
	}
}

func TestAMF3_VectorUint(t *testing.T) {
	in := amf.VectorUint{Values: []uint32{1, 2, 0xffffffff}}
	got, ok := roundTrip3(t, in).(amf.VectorUint)
	if !ok || !reflect.DeepEqual(got.Values, in.Values) {
		t.Fatalf("got %#v", got)
	}
}

func TestAMF3_VectorDouble(t *testing.T) {
	in := amf.VectorDouble{Values: []float64{1.5, 2.5}}
	got, ok := roundTrip3(t, in).(amf.VectorDouble)
	if !ok || !reflect.DeepEqual(got.Values, in.Values) {
		t.Fatalf("got %#v", got)
	}
}

func TestAMF3_VectorObject(t *testing.T) {
	in := amf.VectorObject{
		TypeName: "*",
		Values:   []amf.Value{true, amf.Integer(2)},
	}
	got, ok := roundTrip3(t, in).(amf.VectorObject)
	if !ok || got.TypeName != "*" || len(got.Values) != 2 || got.Values[0] != true || got.Values[1] != amf.Integer(2) {
		t.Fatalf("got %#v", got)
	}
}

func TestAMF3_Dictionary(t *testing.T) {
	in := amf.Dictionary{
		Entries: []amf.DictEntry{
			{Key: "k", Value: amf.Integer(1)},
			{Key: true, Value: "v"},
		},
	}
	got, ok := roundTrip3(t, in).(amf.Dictionary)
	if !ok || len(got.Entries) != 2 {
		t.Fatalf("got %#v", got)
	}
	if got.Entries[0].Key != "k" || got.Entries[0].Value != amf.Integer(1) {
		t.Fatalf("entry0 %#v", got.Entries[0])
	}
	if got.Entries[1].Key != true || got.Entries[1].Value != "v" {
		t.Fatalf("entry1 %#v", got.Entries[1])
	}
}

func TestAMF3_ViaAMF0Marker(t *testing.T) {
	amf3Int, err := amf.Encode3(amf.Integer(1))
	if err != nil {
		t.Fatal(err)
	}
	b := append([]byte{amf.AMF0AVMPlus}, amf3Int...)
	cmd, err := amf.Encode([]amf.Value{"_result"})
	if err != nil {
		t.Fatal(err)
	}
	raw := append(cmd, b...)
	raw = append(raw, amf.AMF0Null)
	vals, err := amf.Decode0(raw)
	if err != nil {
		t.Fatal(err)
	}
	if vals[0] != "_result" || vals[1] != amf.Integer(1) || vals[2] != nil {
		t.Fatalf("%#v", vals)
	}
}

func TestEncodeU29(t *testing.T) {
	cases := []struct {
		v    uint32
		want []byte
	}{
		{0, []byte{0x00}},
		{0x7f, []byte{0x7f}},
		{0x80, []byte{0x81, 0x00}},
		{0x3fff, []byte{0xff, 0x7f}},
		{0x4000, []byte{0x81, 0x80, 0x00}},
	}
	for _, c := range cases {
		got := amf.EncodeU29(c.v)
		if !reflect.DeepEqual(got, c.want) {
			t.Fatalf("%d: got %x want %x", c.v, got, c.want)
		}
	}
}
