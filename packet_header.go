package tron

import (
	"bytes"
	"encoding/binary"
	"io"
)

type Header struct {
	Seq     int32
	DataLen int32
}

const (
	PACK_LEN   = 4     // packet 总长度
	HEADER_LEN = 4 + 4 // seq  + dataLen
)

func MarshalHeader(h *Header) []byte {
	packLen := PACK_LEN + HEADER_LEN + h.DataLen
	buf := bytes.NewBuffer(make([]byte, 0, packLen))
	write(buf, binary.BigEndian, int32(HEADER_LEN+h.DataLen)) // packet length
	write(buf, binary.BigEndian, h.Seq)
	write(buf, binary.BigEndian, h.DataLen)
	return buf.Bytes()
}

func UnmarshalHeader(b []byte) (*Header, error) {
	r := bytes.NewReader(b)
	h := &Header{}
	if err := read(r, binary.BigEndian, &h.Seq); err != nil {
		return nil, err
	}
	if err := read(r, binary.BigEndian, &h.DataLen); err != nil {
		return nil, err
	}
	return h, nil
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
