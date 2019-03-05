package main

import (
	"fmt"
	"net"
	"tron"
)

func main() {
	addr, err := net.ResolveTCPAddr("tcp4", "localhost:8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	conn, err := net.DialTCP("tcp4", nil, addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	clientConf := tron.NewConfig(16*1024, 16*1024, 100, 100)
	cli := tron.NewClient(conn, clientConf, clientPackHandler)
	cli.Run()

	pingPack := tron.NewPacket(1, []byte("ping"))
	cli.DirectWrite(pingPack)

	wait(10)
}

func clientPackHandler(cli *tron.Client, p *tron.Packet) {
	fmt.Printf("[server:%s] -> [client:%s]: %s\n", trim(cli.RemoteAddr()), trim(cli.LocalAddr()), string(p.Data))
}
