package main

import (
	"bufio"
	"encoding/binary"
	"io"
	"tron"
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

func (c *DefaultCodec) MarshalPacket(p tron.Packet) []byte {
	hData := tron.MarshalHeader(p.Header)
	return append(hData, p.Data...)
}

func (c *DefaultCodec) UnmarshalPacket(b []byte) (*tron.Packet, error) {
	h, err := tron.UnmarshalHeader(b)
	if err != nil {
		return nil, err
	}

	data := b[tron.HEADER_LEN : tron.HEADER_LEN+h.DataLen]
	h.DataLen = int32(len(data))
	return &tron.Packet{Header: h, Data: data}, nil
}

// 读取 byte / []byte
func read(r io.Reader, order binary.ByteOrder, data interface{}) error {
	buf, ok := data.([]byte)
	if ok {
		_, err := io.ReadFull(r, buf)
		return err
	}
	return binary.Read(r, order, data)
}

// 写入 byte / []byte
func write(w io.Writer, order binary.ByteOrder, data interface{}) error {
	buf, ok := data.([]byte)
	if ok {
		_, err := w.Write(buf)
		return err
	}
	return binary.Write(w, order, data)
}
