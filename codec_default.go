package tron

import (
	"bufio"
	"encoding/binary"
)

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
