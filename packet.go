package tron

// 数据交互包
// seq + dataLen + data
type Packet struct {
	Header *Header // 头部
	Data   []byte  // 包数据
}

func NewPacket(data []byte) *Packet {
	h := &Header{Seq: -1, DataLen: int32(len(data))}
	return &Packet{Header: h, Data: data}
}
