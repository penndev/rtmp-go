package amf_test

import (
	"reflect"
	"testing"

	"github.com/penndev/rtmp/amf"
)

func roundTrip0(t *testing.T, v amf.Value) amf.Value {
	t.Helper()
	b, err := amf.Encode0([]amf.Value{v})
	if err != nil {
		t.Fatalf("Encode0(%T): %v", v, err)
	}
	vals, err := amf.Decode0(b)
	if err != nil {
		t.Fatalf("Decode0(%T): %v", v, err)
	}
	if len(vals) != 1 {
		t.Fatalf("Decode0(%T): got %d values %#v", v, len(vals), vals)
	}
	return vals[0]
}

func TestAMF0_Null(t *testing.T) {
	if got := roundTrip0(t, nil); got != nil {
		t.Fatalf("got %#v", got)
	}
}

func TestAMF0_Undefined(t *testing.T) {
	if got := roundTrip0(t, amf.Undefined{}); got != (amf.Undefined{}) {
		t.Fatalf("got %#v", got)
	}
}

func TestAMF0_Boolean(t *testing.T) {
	for _, want := range []bool{true, false} {
		if got := roundTrip0(t, want); got != want {
			t.Fatalf("got %v", got)
		}
	}
}

func TestAMF0_Number(t *testing.T) {
	if got := roundTrip0(t, 319.5); got != 319.5 {
		t.Fatalf("got %v", got)
	}
	// Integer encodes as Number in AMF0
	got := roundTrip0(t, amf.Integer(7))
	if got != 7.0 {
		t.Fatalf("Integer via AMF0 got %v (%T)", got, got)
	}
}

func TestAMF0_String(t *testing.T) {
	if got := roundTrip0(t, "connect"); got != "connect" {
		t.Fatalf("got %v", got)
	}
}

func TestAMF0_LongString(t *testing.T) {
	s := string(make([]byte, 70000))
	got := roundTrip0(t, s)
	gs, ok := got.(string)
	if !ok || len(gs) != 70000 {
		t.Fatalf("got %T len=%d", got, len(gs))
	}
}

func TestAMF0_Object(t *testing.T) {
	in := amf.Object{"app": "live", "objectEncoding": 0.0}
	got, ok := roundTrip0(t, in).(amf.Object)
	if !ok || got["app"] != "live" || got["objectEncoding"] != 0.0 {
		t.Fatalf("got %#v", got)
	}
}

func TestAMF0_ECMAArray(t *testing.T) {
	in := amf.ECMAArray{"width": 512.0, "stereo": true}
	got, ok := roundTrip0(t, in).(amf.ECMAArray)
	if !ok || got["width"] != 512.0 || got["stereo"] != true {
		t.Fatalf("got %#v", got)
	}
	o, ok := amf.AsObject(got)
	if !ok || o["width"] != 512.0 {
		t.Fatalf("AsObject %#v", o)
	}
}

func TestAMF0_StrictArray(t *testing.T) {
	in := amf.StrictArray{"a", 1.0, true, nil}
	got, ok := roundTrip0(t, in).(amf.StrictArray)
	if !ok || !reflect.DeepEqual([]amf.Value(got), []amf.Value(in)) {
		t.Fatalf("got %#v", got)
	}
}

func TestAMF0_Date(t *testing.T) {
	in := amf.Date{Millis: 1_000_000, Timezone: 480}
	got, ok := roundTrip0(t, in).(amf.Date)
	if !ok || got.Millis != in.Millis || got.Timezone != in.Timezone {
		t.Fatalf("got %#v", got)
	}
}

func TestAMF0_XMLDocument(t *testing.T) {
	in := amf.XMLDocument(`<?xml version="1.0"?>
<root>
  <item id="1">hello</item>
</root>`)
	b, err := amf.Encode0([]amf.Value{in})
	if err != nil {
		t.Fatalf("Encode0: %v", err)
	}
	vals, err := amf.Decode0(b)
	if err != nil {
		t.Fatalf("Decode0: %v", err)
	}
	if len(vals) != 1 {
		t.Fatalf("got %d values %#v", len(vals), vals)
	}
	got, ok := vals[0].(amf.XMLDocument)
	if !ok || got != in {
		t.Fatalf("got %#v", vals[0])
	}
}

func TestAMF0_TypedObject(t *testing.T) {
	in := amf.TypedObject{
		ClassName: "NetConnection",
		Fields:    amf.Object{"code": "ok"},
	}
	got, ok := roundTrip0(t, in).(amf.TypedObject)
	if !ok || got.ClassName != "NetConnection" || got.Fields["code"] != "ok" {
		t.Fatalf("got %#v", got)
	}
}

func TestAMF0_Unsupported(t *testing.T) {
	if got := roundTrip0(t, amf.Unsupported{}); got != (amf.Unsupported{}) {
		t.Fatalf("got %#v", got)
	}
}

func TestAMF0_Reference(t *testing.T) {
	objBytes, err := amf.Encode0([]amf.Value{amf.Object{"k": "v"}})
	if err != nil {
		t.Fatal(err)
	}
	raw := append(append([]byte{}, objBytes...), amf.AMF0Reference, 0x00, 0x00)
	vals, err := amf.Decode0(raw)
	if err != nil {
		t.Fatal(err)
	}
	o1, _ := vals[0].(amf.Object)
	o2, _ := vals[1].(amf.Object)
	if o1["k"] != "v" || o2["k"] != "v" {
		t.Fatalf("%#v", vals)
	}
}

func TestAMF0_AVMPlus(t *testing.T) {
	b := []byte{amf.AMF0AVMPlus, amf.AMF3True}
	vals, err := amf.Decode0(b)
	if err != nil || len(vals) != 1 || vals[0] != true {
		t.Fatalf("got %#v err %v", vals, err)
	}
}

func TestAMF0_ConnectCommand(t *testing.T) {
	payload, err := amf.Encode0([]amf.Value{
		"connect", 1.0,
		amf.Object{"app": "live", "objectEncoding": 0.0},
	})
	if err != nil {
		t.Fatal(err)
	}
	vals, err := amf.Decode0(payload)
	if err != nil {
		t.Fatal(err)
	}
	obj, ok := amf.AsObject(vals[2])
	if vals[0] != "connect" || vals[1] != 1.0 || !ok || obj["app"] != "live" {
		t.Fatalf("%#v", vals)
	}
}
