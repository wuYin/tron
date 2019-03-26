package tron

import (
	"bufio"
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
