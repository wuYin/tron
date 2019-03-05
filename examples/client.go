package main

import (
	"fmt"
	"net"
	"time"
	"tron"
)

func main() {
	addr, err := net.ResolveTCPAddr("tcp4", "localhost:8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	clientConf := tron.NewConfig(16*1024, 16*1024, 100, 100)
	r := tron.NewReconnectTaskManager(5*time.Second, 3)
	manager := tron.NewClientsManager(r)
	conn, err := net.DialTCP("tcp4", nil, addr)
	if err != nil {
		panic(err)
	}

	cli := tron.NewClient(conn, clientConf, clientPackHandler)
	cli.Run()
	manager.Add(cli)

	go func() {
		for i := 0; i < 5; i++ {
			pingPack := tron.NewPacket(1, []byte("ping"))
			cli.DirectWrite(pingPack)
			time.Sleep(1 * time.Second)
		}
	}()

	time.Sleep(100000 * time.Second)
}

func clientPackHandler(cli *tron.Client, p *tron.Packet) {
	fmt.Printf("[server:%s] -> [client:%s]: %s\n",
		cli.RemoteAddr()[len(cli.RemoteAddr())-4:],
		cli.LocalAddr()[len(cli.LocalAddr())-4:],
		string(p.Data))
}
