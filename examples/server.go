package main

import (
	"fmt"
	"time"
	"tron"
)

func main() {
	serverConf := tron.NewConfig(16*1024, 16*1024, 100, 100, 1000, 5*time.Second)
	s := tron.NewServer("localhost:8080", serverConf, serverPackHandler)
	s.ListenAndServe()

	time.Sleep(100000 * time.Second)
}

func serverPackHandler(worker *tron.Client, p *tron.Packet) {
	fmt.Printf("[client:%s] -> [server:%s]: %s\n", worker.LocalAddr(), worker.RemoteAddr(), p.Data)
	if string(p.Data) == "ping" {
		pongPack := tron.NewPacket([]byte("pong"))
		respCh, err := worker.DirectWrite(pongPack)
		if err != nil {
			fmt.Println(err)
			return
		}
		for resp := range respCh {
			fmt.Println("[resp]:", resp)
		}
	}
}
