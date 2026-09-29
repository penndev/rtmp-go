# amf

Action Message Format（AMF0 / AMF3）编解码库。

依据：

- [Adobe AMF0](https://rtmp.veriskope.com/pdf/amf0-file-format-specification.pdf)
- [Adobe AMF3](https://rtmp.veriskope.com/pdf/amf3-file-format-spec.pdf)
- [veovera/enhanced-rtmp](https://github.com/veovera/enhanced-rtmp)

约定：

- 同一概念在 AMF0 / AMF3 中使用**相同 Go 类型名**（见下表「共用」列）
- 写侧优先用 `Object`；读侧用 `AsObject` 同时接受 `Object` / `ECMAArray`
- AMF0 的 `0x11`（avmplus）表示下一个值按 AMF3 解码

## API

```go
// 解码 / 编码整段（连续多个值）
vals, err := amf.Decode0(b) // AMF0；遇 0x11 则下一个值按 AMF3
vals, err := amf.Decode3(b) // AMF3
b, err := amf.Encode0(vals) // AMF0
b, err := amf.Encode3(vals) // AMF3
```

编解码成对使用（以 AMF0 为例）：

```go
b, err := amf.Encode0([]amf.Value{"connect"})
if err != nil { ... }
vals, err := amf.Decode0(b)
if err != nil { ... }
// vals[0] == "connect"
```

---

## 共用类型（AMF0 + AMF3 同名）

### Null

| | AMF0 | AMF3 |
|--|------|------|
| marker | `0x05` | `0x01` |
| Go | `nil` | `nil` |

```go
b, _ := amf.Encode0([]amf.Value{nil})
vals, _ := amf.Decode0(b) // vals[0] == nil

b, _ = amf.Encode3([]amf.Value{nil})
vals, _ = amf.Decode3(b) // vals[0] == nil
```

### Undefined

| | AMF0 | AMF3 |
|--|------|------|
| marker | `0x06` | `0x00` |
| Go | `amf.Undefined{}` | `amf.Undefined{}` |

```go
b, _ := amf.Encode0([]amf.Value{amf.Undefined{}})
vals, _ := amf.Decode0(b) // vals[0] == amf.Undefined{}

b, _ = amf.Encode3([]amf.Value{amf.Undefined{}})
vals, _ = amf.Decode3(b)
```

### Boolean

| | AMF0 | AMF3 |
|--|------|------|
| marker | `0x01` + 1 byte | `0x02` false / `0x03` true |
| Go | `bool` | `bool` |

```go
b, _ := amf.Encode0([]amf.Value{true})
vals, _ := amf.Decode0(b) // vals[0] == true

b, _ = amf.Encode3([]amf.Value{false})
vals, _ = amf.Decode3(b) // vals[0] == false
```

### Number / Double

IEEE-754 双精度。AMF0 叫 Number，AMF3 叫 Double，Go 侧统一为 `float64`。

| | AMF0 | AMF3 |
|--|------|------|
| marker | `0x00` | `0x05` |
| Go | `float64` | `float64` |

```go
b, _ := amf.Encode0([]amf.Value{1.5})
vals, _ := amf.Decode0(b) // vals[0] == 1.5

b, _ = amf.Encode3([]amf.Value{1.5})
vals, _ = amf.Decode3(b)
```

### String

| | AMF0 | AMF3 |
|--|------|------|
| marker | `0x02`（≤65535）/ `0x0c` LongString | `0x06`（U29 + UTF-8，可引用） |
| Go | `string` | `string` |

```go
b, _ := amf.Encode0([]amf.Value{"connect"})
vals, _ := amf.Decode0(b) // vals[0] == "connect"

b, _ = amf.Encode3([]amf.Value{"connect"})
vals, _ = amf.Decode3(b)
```

### Object

动态键值表。Enhanced RTMP：写优先 Object；读可与 ECMAArray 互通。

| | AMF0 | AMF3 |
|--|------|------|
| marker | `0x03` … `00 00 09` | `0x0a` 匿名 dynamic traits |
| Go | `amf.Object` (`map[string]Value`) | `amf.Object` |

```go
o := amf.Object{"app": "live", "flashVer": "FMLE/3.0"}
b, _ := amf.Encode0([]amf.Value{o})
vals, _ := amf.Decode0(b)
obj := vals[0].(amf.Object) // obj["app"] == "live"

b, _ = amf.Encode3([]amf.Value{o})
vals, _ = amf.Decode3(b)
```

### Date

毫秒时间戳。AMF0 另有时区（分钟）；AMF3 无时区（`Timezone=0`）。

| | AMF0 | AMF3 |
|--|------|------|
| marker | `0x0b` | `0x08` |
| Go | `amf.Date{Millis, Timezone}` | `amf.Date{Millis}` |

```go
b, _ := amf.Encode0([]amf.Value{amf.Date{Millis: 1e6, Timezone: 480}})
vals, _ := amf.Decode0(b)
d := vals[0].(amf.Date) // d.Millis == 1e6, d.Timezone == 480

b, _ = amf.Encode3([]amf.Value{amf.Date{Millis: 1e6}})
vals, _ = amf.Decode3(b)
```

### XMLDocument

AMF 里的 XML Document：载荷是一整段 XML 文本。Go 类型是 `string` 的别名，用 `amf.XMLDocument(...)` 包一下，避免和普通 `string`、AMF3 的 `XML` 搞混。

| | AMF0 | AMF3 |
|--|------|------|
| marker | `0x0f` | `0x07` (xml-doc) |
| Go | `amf.XMLDocument` | `amf.XMLDocument` |

```go
doc := amf.XMLDocument(`<?xml version="1.0"?>
<root>
  <item id="1">hello</item>
</root>`)

b, err := amf.Encode0([]amf.Value{doc})
if err != nil { ... }
vals, err := amf.Decode0(b)
if err != nil { ... }
got := vals[0].(amf.XMLDocument) // 内容与 doc 相同

// AMF3 同理
b, err = amf.Encode3([]amf.Value{doc})
vals, err = amf.Decode3(b)
got = vals[0].(amf.XMLDocument)
```

### TypedObject

带类名 / traits 的对象。

| | AMF0 | AMF3 |
|--|------|------|
| marker | `0x10` | `0x0a`（有 class 或 sealed keys） |
| Go | `amf.TypedObject`（`ClassName` + `Fields`） | `amf.TypedObject`（+ `Keys`/`Values`/`Dynamic`） |

```go
// AMF0
in := amf.TypedObject{ClassName: "NetConnection", Fields: amf.Object{"code": "ok"}}
b, _ := amf.Encode0([]amf.Value{in})
vals, _ := amf.Decode0(b)
out := vals[0].(amf.TypedObject) // out.ClassName == "NetConnection"

// AMF3
in3 := amf.TypedObject{
    ClassName: "Box",
    Keys: []string{"w"}, Values: []amf.Value{amf.Integer(10)},
    Dynamic: true, Fields: amf.Object{"extra": false},
}
b, _ = amf.Encode3([]amf.Value{in3})
vals, _ = amf.Decode3(b)
```

---

## 仅 AMF0

### ECMAArray — `0x08` — `amf.ECMAArray`

关联数组。读侧请用 `AsObject`。

```go
b, _ := amf.Encode0([]amf.Value{amf.ECMAArray{"width": 512.0}})
vals, _ := amf.Decode0(b)
arr := vals[0].(amf.ECMAArray) // arr["width"] == 512.0
```

### StrictArray — `0x0a` — `amf.StrictArray`

定长稠密数组。

```go
b, _ := amf.Encode0([]amf.Value{amf.StrictArray{"a", 1.0, true}})
vals, _ := amf.Decode0(b)
arr := vals[0].(amf.StrictArray) // len == 3
```

### LongString — `0x0c`

长度 > 65535 的字符串；编解码仍是 Go `string`（`Encode0` 自动升级 marker）。

### Unsupported — `0x0d` — `amf.Unsupported`

```go
b, _ := amf.Encode0([]amf.Value{amf.Unsupported{}})
vals, _ := amf.Decode0(b) // vals[0] == amf.Unsupported{}
```

### Reference — `0x07`

指向前面已出现的复杂对象。当前**只解码**（编码始终内联写出）。

### AVMPlus — `0x11`

下一个值按 AMF3 解码（RTMP Command Extended / Data Extended 常用）。

```go
// 混合：前面 AMF0，碰到 0x11 后按 AMF3
raw := append([]byte{0x11}, mustEncode3([]amf.Value{amf.Integer(1)})...)
vals, err := amf.Decode0(raw) // vals[0] == amf.Integer(1)
```

### Movieclip `0x04` / Recordset `0x0e`

规范保留，编解码返回错误。

---

## 仅 AMF3

### Integer — `0x04` — `amf.Integer`（`int32`，29-bit 有符号）

```go
b, _ := amf.Encode3([]amf.Value{amf.Integer(-1)})
vals, _ := amf.Decode3(b) // vals[0] == amf.Integer(-1)
```

超出 29-bit 范围的 `int`/`int64` 会改写为 Double。

### XML — `0x0b` — `amf.XML`

与 `XMLDocument`（xml-doc）不同 marker。同样是 XML 文本，只是 AMF3 里另一种类型。

```go
b, _ := amf.Encode3([]amf.Value{amf.XML(`<node id="1"/>`)})
vals, _ := amf.Decode3(b)
x := vals[0].(amf.XML) // `<node id="1"/>`
```

### Array — `0x09` — `amf.Array`

稠密 `Dense` + 可选关联 `Assoc`。

```go
in := amf.Array{
    Dense: []amf.Value{true, false},
    Assoc: amf.Object{"a": amf.Integer(1)},
}
b, _ := amf.Encode3([]amf.Value{in})
vals, _ := amf.Decode3(b)
arr := vals[0].(amf.Array)
```

### ByteArray — `0x0c` — `amf.ByteArray`

```go
b, _ := amf.Encode3([]amf.Value{amf.ByteArray{0x10, 0x20}})
vals, _ := amf.Decode3(b)
ba := vals[0].(amf.ByteArray) // []byte{0x10, 0x20}
```

### VectorInt — `0x0d` — `amf.VectorInt`

```go
b, _ := amf.Encode3([]amf.Value{amf.VectorInt{Fixed: true, Values: []int32{1, -2}}})
vals, _ := amf.Decode3(b)
```

### VectorUint — `0x0e` — `amf.VectorUint`

```go
b, _ := amf.Encode3([]amf.Value{amf.VectorUint{Values: []uint32{1, 2}}})
vals, _ := amf.Decode3(b)
```

### VectorDouble — `0x0f` — `amf.VectorDouble`

```go
b, _ := amf.Encode3([]amf.Value{amf.VectorDouble{Values: []float64{1.5, 2.5}}})
vals, _ := amf.Decode3(b)
```

### VectorObject — `0x10` — `amf.VectorObject`

```go
b, _ := amf.Encode3([]amf.Value{amf.VectorObject{TypeName: "*", Values: []amf.Value{true}}})
vals, _ := amf.Decode3(b)
```

### Dictionary — `0x11` — `amf.Dictionary`

```go
b, _ := amf.Encode3([]amf.Value{amf.Dictionary{Entries: []amf.DictEntry{
    {Key: "k", Value: amf.Integer(1)},
}}})
vals, _ := amf.Decode3(b)
```

---

## 类型对照速查

| Go 类型 | AMF0 | AMF3 |
|---------|------|------|
| `nil` | Null | Null |
| `Undefined` | Undefined | Undefined |
| `bool` | Boolean | False/True |
| `float64` | Number | Double |
| `string` | String / LongString | String |
| `Object` | Object | Object (anon dynamic) |
| `Date` | Date | Date |
| `XMLDocument` | XML Document | XML Doc |
| `TypedObject` | Typed Object | Object (traits) |
| `ECMAArray` | ECMA Array | — |
| `StrictArray` | Strict Array | — (可用 `Array`) |
| `Unsupported` | Unsupported | — |
| `Integer` | → Number | Integer |
| `XML` | — | XML |
| `Array` | — | Array |
| `ByteArray` | — | ByteArray |
| `VectorInt/Uint/Double/Object` | — | Vector* |
| `Dictionary` | — | Dictionary |
