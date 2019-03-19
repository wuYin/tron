package tron

import (
	"bufio"
	"encoding/binary"
)

// 将拼包和解包行为定义为接口方便第三方实现
type Codec interface {
	// 从连接缓冲区中读取数据
	ReadPacket(r *bufio.Reader) ([]byte, error)

	// 解包
	UnmarshalPacket(buf []byte) (*Packet, error)

	// 拼包
	MarshalPacket(p Packet) []byte
}

// 默认 codec
type DefaultCodec struct{}

func NewDefaultCodec() *DefaultCodec {
	return &DefaultCodec{}
}

func (c *DefaultCodec) ReadPacket(r *bufio.Reader) ([]byte, error) {
	var packLen int32
	err := read(r, binary.BigEndian, &packLen)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, packLen)
	curLen := 0
	for {
		n, err := r.Read(buf)
		if err != nil {
			return nil, err
		}
		curLen += n

		if int32(curLen) < packLen { // 未读完
			continue
		} else {
			buf = buf[:curLen]
			break
		}
	}
	return buf, nil
}

func (c *DefaultCodec) MarshalPacket(p Packet) []byte {
	hData := MarshalHeader(p.Header)
	return append(hData, p.Data...)
}

func (c *DefaultCodec) UnmarshalPacket(b []byte) (*Packet, error) {
	h, err := UnmarshalHeader(b)
	if err != nil {
		return nil, err
	}

	data := b[HEADER_LEN : HEADER_LEN+h.DataLen]
	h.DataLen = int32(len(data))
	return &Packet{Header: h, Data: data}, nil
}
