package tron

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"logx"
	"net"
	"time"
)

// 某个连接的会话信息
type Session struct {
	conn      *net.TCPConn
	cr        *bufio.Reader // 连接缓冲 reader
	cw        *bufio.Writer // 连接缓冲 writer
	ReadCh    chan *Packet  // 读请求的 channel
	WriteCh   chan *Packet  // 写响应的 channel
	living    bool
	idleTimer *time.Timer
	conf      *Config
	codec     Codec
}

func NewSession(conn *net.TCPConn, conf *Config, codec Codec) *Session {
	conn.SetReadBuffer(conf.ReadBufSize)
	conn.SetWriteBuffer(conf.WriteBufSize)
	s := &Session{
		conn:      conn,
		cr:        bufio.NewReaderSize(conn, conf.ReadBufSize),
		cw:        bufio.NewWriterSize(conn, conf.WriteBufSize),
		ReadCh:    make(chan *Packet, conf.ReadChanSize),
		WriteCh:   make(chan *Packet, conf.WriteChanSize),
		living:    true,
		idleTimer: time.NewTimer(conf.IdleDuration),
		conf:      conf,
		codec:     codec,
	}
	return s
}

// 读取数据
func (s *Session) ReadPacket() {
	buf := bytes.NewBuffer(nil)
	for s.living {
		b, err := s.codec.ReadPacket(s.cr)
		if err != nil {
			fmt.Println(err)
			s.living = false
			return
		}

		p, err := s.codec.UnmarshalPacket(b)
		if err != nil {
			fmt.Println(err)
			s.living = false
			return
		}

		// 写入读缓冲
		s.ReadCh <- p
		s.idleTimer.Reset(s.conf.IdleDuration) // 重设空闲 timer
		buf.Reset()
	}
}

// 写入响应
func (s *Session) WritePacket() {
	for s.living {
		if p, ok := <-s.WriteCh; ok {
			s.writeConn(*p)
			s.flush()
		}
	}
}

// 对外保留的写数据方法
func (s *Session) DirectWrite(p *Packet) error {
	if s.living {
		select {
		case s.WriteCh <- p:
			return nil
		default:
			return errors.New("write channel full")
		}
	}
	return errors.New("conn closed")
}

// 关闭当前连接
func (s *Session) Close() error {
	if s.living {
		s.living = false
		s.conn.Close() // 主动关闭连接
		close(s.ReadCh)
		close(s.WriteCh)
		fmt.Println("session closed")
	}
	return nil
}

func (s *Session) Living() bool {
	return s.living
}

func (s *Session) IsIdle() bool {
	select {
	case <-s.idleTimer.C:
		return true // 连接长时间空闲
	default:
		return false // 还没到超时时间
	}
}

// 真正写入数据流
func (s *Session) writeConn(p Packet) {
	buf := s.codec.MarshalPacket(p)
	if buf == nil || len(buf) == 0 {
		logx.Error("invalid packet: %+v", p)
		return
	}

	n, err := s.cw.Write(buf)
	if err != nil {
		logx.Error(err)
		if err == io.EOF { // 另一端主动关闭
			s.Close()
			return
		}
		if err == io.ErrShortWrite { // 未写完毕尝试重写
			s.cw.Write(buf[n:])
		}
	}
}

// 一次性写出缓存
func (s *Session) flush() {
	if !s.living || s.cw.Buffered() <= 0 {
		return
	}
	if err := s.cw.Flush(); err != nil {
		logx.Error("flush failed: %v", err)
		s.cw.Reset(s.conn)
		return
	}
}
