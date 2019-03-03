package tron

import (
	"logx"
	"net"
)

type Server struct {
	address string
	living  bool
	handler func(worker *Client, p *Packet)
	conf    *Config
}

func NewServer(addr string, conf *Config, f func(worker *Client, p *Packet)) *Server {
	s := &Server{
		address: addr,
		handler: f,
		living:  true,
		conf:    conf,
	}
	return s
}

func (s *Server) ListenAndServe() error {
	addr, err := net.ResolveTCPAddr("tcp4", s.address)
	if err != nil {
		logx.Error(err)
		return err
	}

	l, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		logx.Error(err)
		return err
	}

	go s.run(l)
	return nil
}

func (s *Server) run(l *net.TCPListener) error {
	for s.living {
		conn, err := l.AcceptTCP()
		if err != nil {
			logx.Error(err)
			continue
		}
		serverWorker := NewClient(conn, s.conf, s.handler)
		serverWorker.Run()
	}
	return nil
}
