package rtmp

//  Version (8 bits): In C0, this field identifies the RTMP version
//  requested by the client. In S0, this field identifies the RTMP
//  version selected by the server. The version defined by this
//  specification is 3. Values 0-2 are deprecated values used by
//  earlier proprietary products; 4-31 are reserved for future
//  implementations; and 32-255 are not allowed (to allow
//  distinguishing RTMP from text-based protocols, which always start
//  with a printable character). A server that does not recognize the
//  client’s requested version SHOULD respond with 3. The client MAY
//  choose to degrade to version 3, or to abandon the handsha
const VERSION = 0x03
