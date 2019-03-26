package main

import (
	"fmt"
	"time"
	"tron"
)

func main() {
	serverConf := tron.NewDefaultConf(1 * time.Minute)
	codec := tron.NewDefaultCodec()
	s := tron.NewServer("localhost:8080", serverConf, codec, serverPackHandler)
	s.ListenAndServe()

	time.Sleep(2 * time.Second)
	s.Shutdown()
}

func serverPackHandler(worker *tron.Client, p *tron.Packet) {
	fmt.Printf("[client:%s] -> [server:%s]: %s\n",
		tron.SplitPort(worker.RemoteAddr()),
		tron.SplitPort(worker.LocalAddr()),
		p.Data)
	if string(p.Data) == "ping" {
		pongPack := tron.NewRespPacket(p.Header.Seq, []byte("pong"))
		if _, err := worker.AsyncWrite(pongPack); err != nil {
			fmt.Printf("worker write failed: %v\n", err)
			return
		}
	}
}
