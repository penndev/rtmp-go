
https://www.adobe.com/content/dam/acom/en/devnet/pdf/amf0-file-format-specification.pdf

# AMFO 说明

操作消息格式（AMF）是一种紧凑的二进制格式，用于序列化ActionScript(对象图)数据结构。

AMF于2001年在Flash Player 6中引入，并随Flash Player 7中的ActionScript 2.0引入而保持不变。此格式的版本标头设置为0，因此该格式的版本称为AMF 0。

AMF 0中有16个核心类型标记。类型标记的长度为1个字节，描述了可能跟随的编码数据的类型。 

# AMF0 数据类型

number-marker           =0x00           
boolean-marker          =0x01          
string-marker           =0x02           
object-marker           =0x03           
movieclip-marker 		=0x04 ;保留类型未使用; reserved, not supported 
null-marker             =0x05             
undefined-marker        =0x06        
reference-marker        =0x07        
ecma-array-marker       =0x08       
object-end-marker       =0x09       
strict-array-marker     =0x0A     
date-marker             =0x0B             
long-string-marker      =0x0C      
unsupported-marker      =0x0D      
recordset-marker 		=0x0E ;保留类型未使用; reserved, not supported xml-document-marker     =0x0F     
typed-object-marker     =0x10 
;拓展的数据类型。因为amf3出现
avmplus-object-marker   =0x11 ; change amf0 to amf3

类型标记后面可以跟实际编码的类型数据，或者如果标记代表单个可能的值（例如null），则无需进一步编码信息。

object-end-type只能显示为标记对象类型或typed-object-type的一组属性的结尾，或仅表示ECMA数组关联段的结尾。

# 数字类型 number-marker

header 0x00 (1 byte)

body number (8 byte)

AMF 0数字类型用于编码ActionScript数字。 数字类型标记后面的数据始终是网络字节顺序的8字节IEEE-754双精度浮点值（低位存储器中的符号位）。

number-type = number-marker DOUB

# 布尔类型 boolean-marker

header 0x01 (1 byte)

body 0=false 0<>true (1 byte)

布尔类型标记后跟无符号字节； 零字节值表示false，而非零字节值（通常为1）表示true。

# 字符类型 string-marker

header 0x02 (1 byte)

body-length (2 byte)

body (ASSIC UTF-8 max-length 65535)

AMF中的所有字符串都使用UTF-8编码； 但是，字节长度的头格式可能会有所不同。 AMF 0字符串类型使用标准字节长度的标头（即U16）。 对于需要多于65535字节才能以UTF-8编码的长字符串，应使用AMF 0长字符串类型。

# 对象类型 object-marker           

header 0x03 (1 byte)

@@@@@@@@@@@@

AMF 0对象类型用于编码匿名ActionScript对象。 没有注册类的任何类型化对象都应视为匿名ActionScript对象。 如果同一对象实例出现在对象图中，则应使用AMF 0通过引用发送它。使用引用类型可以减少冗余信息被序列化和循环引用的无限循环。

# 空类型 null-marker             

header 0x05 (1 byte)

# 未定义类型 undefined-marker

header 0x06 (1 byte)

没有任何其他数据

# 参考类型 reference-marker 

header 0x07 (1 byte)

body-length (2 byte)

# 数组类型 ecma-array-marker      

header 0x08 (1 byte)

body-length (4 byte)

此类型与匿名对象非常相似。

# 对象结束类型 object-end-marker

header 0x09 (1 byte)

body-length (4 byte)

# 严格数组类型  Strict Array Type

header 0x0A (1 byte)

body-length (4 byte)

# 日期类型 date-marker

header 0x0B (1 byte)

body-length (2 byte)

从1970年1月1日午夜（UTC）时区开始经过的毫秒数将ActionScript日期序列化。 尽管这种类型的设计为时区偏移量信息保留了空间，但不应填充或使用它，因为在网络上序列化日期时更改时区是非常规的。 建议根据需要独立查询时区。

# 长字符类型 long-string-marker

header 0x0c (1 byte)

body-length (4 byte)

# 不支持的类型 unsupported-marker

header 0x0D (1 byte)

# 记录类型 recordset-marker

header 0x0E

# XML类型 xml-document-marker

header 0x0F

# 强对象类型

header 0x10