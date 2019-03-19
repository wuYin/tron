package tron

// 将拼包和解包行为定义为接口方便第三方实现
type Codec interface {
	MarshalPacket(p Packet) []byte
	UnmarshalPacket(buf []byte) (*Packet, error)
}

// 默认 codec
type DefaultCodec struct {
	Codec
}

func NewDefaultCodec() *DefaultCodec {
	return &DefaultCodec{}
}

func (c *DefaultCodec) MarshalPacket(p Packet) []byte {
	return MarshalPacket(p)
}

func (c *DefaultCodec) UnmarshalPacket(buf []byte) (*Packet, error) {
	return UnmarshalPacket(buf)
}
