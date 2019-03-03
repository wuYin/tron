package tron

import (
	"reflect"
	"testing"
)

func TestRW(t *testing.T) {
	p1 := NewPacket('1', []byte("A"))
	buf := MarshalPacket(*p1)
	p2, err := UnmarshalPacket(buf)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p1, p2) {
		t.Fail()
	}
}
