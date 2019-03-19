package tron

import (
	"logx"
	"net"
	"time"
)

type Server struct {
	address   string
	handler   func(worker *Client, p *Packet)
	conf      *Config
	living    bool
	shutdown  chan struct{}
	keepAlive time.Duration
	codec     Codec
}

func NewServer(addr string, conf *Config, f func(worker *Client, p *Packet)) *Server {
	s := &Server{
		address:   addr,
		handler:   f,
		living:    true,
		conf:      conf,
		shutdown:  make(chan struct{}, 1), // 这里是缓冲 channel
		keepAlive: 5 * time.Second,
		codec:     NewDefaultCodec(),
	}
	return s
}

func (s *Server) ListenAndServe() error {
	addr, err := net.ResolveTCPAddr("tcp4", s.address)
	if err != nil {
		logx.Error(err)
		return err
	}

	listener, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		logx.Error(err)
		return err
	}

	liver := NewLiveListener(listener, s.shutdown, 5*time.Second) // 保持 5s 连接
	go s.run(liver)
	return nil
}

// 将服务器的连接关闭，不再接受新连接
func (s *Server) Shutdown() {
	s.living = false
	s.shutdown <- struct{}{} // 立刻停止
	logx.Debug("shutdown...")
}

// run 服务器直到手动不接受新连接
func (s *Server) run(l *LiveListener) error {
	for s.living {
		conn, err := l.AcceptTCP()
		if err != nil {
			logx.Error(err)
			continue
		}

		// 将连接分发给 server worker 处理
		serverWorker := NewClient(conn, s.conf, s.codec, s.handler)
		serverWorker.Run()
	}
	return nil
}
