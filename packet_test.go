package tron

import (
	"testing"
)

func TestRW(t *testing.T) {
	codec := NewDefaultCodec()
	oldPack := NewPacket([]byte("a"))
	b := codec.MarshalPacket(*oldPack)

	if len(b) < PACK_LEN {
		t.Fatalf("bytes len invalid: %d %v", len(b), b)
	}
	b = b[PACK_LEN:] // 掐掉 packetLen

	newPack, err := codec.UnmarshalPacket(b)
	if err != nil {
		t.Fatalf("unmarshal packet failed: %v", err)
	}

	if newPack.Header.DataLen != 1 || newPack.Header.Seq != -1 {
		t.Fatalf("invalid unmarshaled header: %+v", newPack.Header)
	}
	if string(newPack.Data) != "a" {
		t.Fatalf("invalid unmarshaled packet data: %q", newPack.Data)
	}
}
