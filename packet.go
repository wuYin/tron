package tron

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// 数据交互包
// seq + cate + dataLen + data + \r\n
type Packet struct {
	Seq  int32  // 序号
	Cate uint8  // 包类型
	Data []byte // 包数据
}

func NewPacket(cate uint8, data []byte) *Packet {
	return &Packet{
		Seq:  -1,
		Cate: cate,
		Data: data,
	}
}

const (
	HEAD_LEN = 4 + 1 + 4 // seq + cate + dataLen
	CRLF_LEN = 2         // \r\n
)

var (
	big  = binary.BigEndian
	CRLF = []byte{'\r', '\n'}
)

func MarshalPacket(p Packet) []byte {
	return encode(p)
}

func UnmarshalPacket(buf []byte) (*Packet, error) {
	return decode(buf)
}

// 拼包
func encode(p Packet) []byte {
	dataLen := len(p.Data)
	packLen := HEAD_LEN + dataLen + CRLF_LEN       // seq + cate + dataLen + data + \r\n
	w := bytes.NewBuffer(make([]byte, 0, packLen)) //
	write(w, big, p.Seq)
	write(w, big, p.Cate)
	write(w, big, uint32(dataLen))
	write(w, big, p.Data)
	write(w, big, CRLF)
	return w.Bytes()
}

// 拆包
func decode(buf []byte) (*Packet, error) {
	buf = bytes.TrimSuffix(buf, CRLF)
	r := bytes.NewReader(buf)

	p := new(Packet)
	if err := read(r, big, &p.Seq); err != nil {
		return nil, err
	}
	if err := read(r, big, &p.Cate); err != nil {
		return nil, err
	}

	var dataLen uint32
	if err := read(r, big, &dataLen); err != nil {
		return nil, err
	}

	if dataLen <= 0 {
		return nil, errors.New("packet no data")
	}
	if int(dataLen) != r.Len() {
		return nil, fmt.Errorf("data len %d not match %d", dataLen, r.Len())
	}

	// 读取指定长度的数据到缓冲区中
	p.Data = make([]byte, dataLen)
	if err := read(r, big, p.Data); err != nil {
		return nil, err
	}

	return p, nil
}

// 读取 byte / []byte
func read(r *bytes.Reader, order binary.ByteOrder, data interface{}) error {
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
