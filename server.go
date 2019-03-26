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
	closed    bool
	closeCh   chan struct{}
	keepAlive time.Duration
	codec     Codec
}

func NewServer(addr string, conf *Config, serverCodec Codec, f func(worker *Client, p *Packet)) *Server {
	s := &Server{
		address:   addr,
		handler:   f,
		closed:    false,
		conf:      conf,
		closeCh:   make(chan struct{}, 1),
		keepAlive: 5 * time.Second,
		codec:     serverCodec,
	}
	return s
}

// 启动
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

	liver := NewLiveListener(listener, s.closeCh, 5*time.Second) // 保持 5s 连接
	go func(l *LiveListener) {
		for !s.closed {
			conn, err := l.AcceptTCP()
			if err != nil {
				logx.Error(err)
				continue
			}

			// 将连接分发给 server worker 处理
			serverWorker := NewClient(conn, s.conf, s.codec, s.handler)
			serverWorker.ReadWriteAndHandle()
		}
	}(liver)
	return nil
}

// 将服务器的连接关闭，不再接受新连接
func (s *Server) Shutdown() {
	s.closed = true
	s.closeCh <- struct{}{} // 立刻停止
	logx.Debug("shutdown...")
}
