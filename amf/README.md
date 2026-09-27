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
// AMF0
vals, err := amf.Decode(b)
v, rest, err := amf.Decode0(b)
b, err := amf.Encode(vals)
b, err := amf.Encode0(v)

// AMF3
vals, err := amf.Decode3All(b)
v, rest, err := amf.Decode3(b)
b, err := amf.Encode3All(vals)
b, err := amf.Encode3(v)
```

---

## 共用类型（AMF0 + AMF3 同名）

### Null

| | AMF0 | AMF3 |
|--|------|------|
| marker | `0x05` | `0x01` |
| Go | `nil` | `nil` |

```go
amf.Encode0(nil)   / amf.Decode0
amf.Encode3(nil)   / amf.Decode3
```

### Undefined

| | AMF0 | AMF3 |
|--|------|------|
| marker | `0x06` | `0x00` |
| Go | `amf.Undefined{}` | `amf.Undefined{}` |

```go
amf.Encode0(amf.Undefined{})
amf.Encode3(amf.Undefined{})
```

### Boolean

| | AMF0 | AMF3 |
|--|------|------|
| marker | `0x01` + 1 byte | `0x02` false / `0x03` true |
| Go | `bool` | `bool` |

```go
amf.Encode0(true)
amf.Encode3(false)
```

### Number / Double

IEEE-754 双精度。AMF0 叫 Number，AMF3 叫 Double，Go 侧统一为 `float64`。

| | AMF0 | AMF3 |
|--|------|------|
| marker | `0x00` | `0x05` |
| Go | `float64` | `float64` |

```go
amf.Encode0(1.5)
amf.Encode3(1.5)
```

### String

| | AMF0 | AMF3 |
|--|------|------|
| marker | `0x02`（≤65535）/ `0x0c` LongString | `0x06`（U29 + UTF-8，可引用） |
| Go | `string` | `string` |

```go
amf.Encode0("connect")
amf.Encode3("connect")
```

### Object

动态键值表。Enhanced RTMP：写优先 Object；读可与 ECMAArray 互通。

| | AMF0 | AMF3 |
|--|------|------|
| marker | `0x03` … `00 00 09` | `0x0a` 匿名 dynamic traits |
| Go | `amf.Object` (`map[string]Value`) | `amf.Object` |

```go
o := amf.Object{"app": "live"}
amf.Encode0(o)
amf.Encode3(o)
```

### Date

毫秒时间戳。AMF0 另有时区（分钟）；AMF3 无时区（`Timezone=0`）。

| | AMF0 | AMF3 |
|--|------|------|
| marker | `0x0b` | `0x08` |
| Go | `amf.Date{Millis, Timezone}` | `amf.Date{Millis}` |

```go
amf.Encode0(amf.Date{Millis: 1e6, Timezone: 480})
amf.Encode3(amf.Date{Millis: 1e6})
```

### XMLDocument

| | AMF0 | AMF3 |
|--|------|------|
| marker | `0x0f` | `0x07` (xml-doc) |
| Go | `amf.XMLDocument` | `amf.XMLDocument` |

```go
amf.Encode0(amf.XMLDocument("<a/>"))
amf.Encode3(amf.XMLDocument("<a/>"))
```

### TypedObject

带类名 / traits 的对象。

| | AMF0 | AMF3 |
|--|------|------|
| marker | `0x10` | `0x0a`（有 class 或 sealed keys） |
| Go | `amf.TypedObject`（`ClassName` + `Fields`） | `amf.TypedObject`（+ `Keys`/`Values`/`Dynamic`） |

```go
// AMF0
amf.Encode0(amf.TypedObject{ClassName: "NetConnection", Fields: amf.Object{"code": "ok"}})

// AMF3
amf.Encode3(amf.TypedObject{
    ClassName: "Box",
    Keys: []string{"w"}, Values: []amf.Value{amf.Integer(10)},
    Dynamic: true, Fields: amf.Object{"extra": false},
})
```

---

## 仅 AMF0

### ECMAArray — `0x08` — `amf.ECMAArray`

关联数组。读侧请用 `AsObject`。

```go
amf.Encode0(amf.ECMAArray{"width": 512.0})
```

### StrictArray — `0x0a` — `amf.StrictArray`

定长稠密数组。

```go
amf.Encode0(amf.StrictArray{"a", 1.0, true})
```

### LongString — `0x0c`

长度 > 65535 的字符串；编解码仍是 Go `string`（`Encode0` 自动升级 marker）。

### Unsupported — `0x0d` — `amf.Unsupported`

```go
amf.Encode0(amf.Unsupported{})
```

### Reference — `0x07`

指向前面已出现的复杂对象。当前**只解码**（编码始终内联写出）。

### AVMPlus — `0x11`

下一个值按 AMF3 解码（RTMP Command Extended / Data Extended 常用）。

### Movieclip `0x04` / Recordset `0x0e`

规范保留，编解码返回错误。

---

## 仅 AMF3

### Integer — `0x04` — `amf.Integer`（`int32`，29-bit 有符号）

```go
amf.Encode3(amf.Integer(-1))
```

超出 29-bit 范围的 `int`/`int64` 会改写为 Double。

### XML — `0x0b` — `amf.XML`

与 `XMLDocument`（xml-doc）不同 marker。

```go
amf.Encode3(amf.XML("<node/>"))
```

### Array — `0x09` — `amf.Array`

稠密 `Dense` + 可选关联 `Assoc`。

```go
amf.Encode3(amf.Array{
    Dense: []amf.Value{true, false},
    Assoc: amf.Object{"a": amf.Integer(1)},
})
```

### ByteArray — `0x0c` — `amf.ByteArray`

```go
amf.Encode3(amf.ByteArray{0x10, 0x20})
```

### VectorInt — `0x0d` — `amf.VectorInt`

```go
amf.Encode3(amf.VectorInt{Fixed: true, Values: []int32{1, -2}})
```

### VectorUint — `0x0e` — `amf.VectorUint`

```go
amf.Encode3(amf.VectorUint{Values: []uint32{1, 2}})
```

### VectorDouble — `0x0f` — `amf.VectorDouble`

```go
amf.Encode3(amf.VectorDouble{Values: []float64{1.5, 2.5}})
```

### VectorObject — `0x10` — `amf.VectorObject`

```go
amf.Encode3(amf.VectorObject{TypeName: "*", Values: []amf.Value{true}})
```

### Dictionary — `0x11` — `amf.Dictionary`

```go
amf.Encode3(amf.Dictionary{Entries: []amf.DictEntry{
    {Key: "k", Value: amf.Integer(1)},
}})
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
