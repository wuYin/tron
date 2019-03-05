package main

import (
	"fmt"
	"time"
	"tron"
)

func main() {
	serverConf := tron.NewConfig(16*1024, 16*1024, 100, 100)
	s := tron.NewServer("localhost:8080", serverConf, serverPackHandler)
	s.ListenAndServe()

	time.Sleep(100000 * time.Second)
}

func serverPackHandler(worker *tron.Client, p *tron.Packet) {
	fmt.Printf("[client:%s] -> [server:%s]: %s\n", worker.LocalAddr(), worker.RemoteAddr(), p.Data)
	if string(p.Data) == "ping" {
		pongPack := tron.NewPacket(1, []byte("pong"))
		if err := worker.DirectWrite(pongPack); err != nil {
			fmt.Println(err)
		}
	}
}
