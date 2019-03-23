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
	closed    bool
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
		closed:    false,
		idleTimer: time.NewTimer(conf.IdleDuration),
		conf:      conf,
		codec:     codec,
	}
	return s
}

// 读取数据
func (s *Session) daemonReadPacket() {
	buf := bytes.NewBuffer(nil)
	for !s.closed {
		b, err := s.codec.ReadPacket(s.cr)
		if err != nil {
			fmt.Printf("session: read packet failed: %v\n", err)
			s.closed = true
			return
		}

		p, err := s.codec.UnmarshalPacket(b)
		if err != nil {
			fmt.Printf("session: unmarshal packet failed: %v\n", err)
			s.closed = true
			return
		}

		// 写入读缓冲
		s.ReadCh <- p
		s.idleTimer.Reset(s.conf.IdleDuration) // 重设空闲 timer
		buf.Reset()
	}
}

// 写入响应
func (s *Session) daemonWritePacket() {
	for !s.closed {
		if p, ok := <-s.WriteCh; ok {
			// write to buffer
			buf := s.codec.MarshalPacket(*p)
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

			// flush
			if s.closed || s.cw.Buffered() <= 0 {
				return
			}
			if err := s.cw.Flush(); err != nil {
				logx.Error("flush failed: %v", err)
				s.cw.Reset(s.conn)
				return
			}
		}
	}
}

// 对外保留的写数据方法
func (s *Session) Write(p *Packet) error {
	if !s.closed {
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
	if !s.closed {
		s.closed = true
		s.conn.Close() // 主动关闭连接
		close(s.ReadCh)
		close(s.WriteCh)
		fmt.Println("session closed")
	}
	return nil
}

func (s *Session) IsClosed() bool {
	return s.closed
}
