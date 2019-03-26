package tron

import "fmt"

// 数据交互包
// seq + dataLen + data
type Packet struct {
	Header *Header // 头部
	Data   []byte  // 包数据
}

func (p Packet) Seq() int32 {
	return p.Header.Seq
}

func (p Packet) String() (s string) {
	s = fmt.Sprintf("header: %+v", p.Header)
	s += fmt.Sprintf("data: `%s`", p.Data)
	return
}

// 响应包直接使用 req packet 的 seq
func NewRespPacket(seq int32, data []byte) *Packet {
	h := &Header{
		Seq:     seq,
		DataLen: int32(len(data)),
	}
	return &Packet{Header: h, Data: data}
}

// 请求包 seq 需要重新处理
func NewReqPacket(data []byte) *Packet {
	h := &Header{
		Seq:     -1,
		DataLen: int32(len(data)),
	}
	return &Packet{Header: h, Data: data}
}
