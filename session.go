package tron

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"logx"
	"net"
)

// 某个连接的会话信息
type Session struct {
	conn    *net.TCPConn
	cr      *bufio.Reader // 连接缓冲 reader
	cw      *bufio.Writer // 连接缓冲 writer
	ReadCh  chan *Packet  // 读请求的 channel
	WriteCh chan *Packet  // 写响应的 channel
	living  bool
	conf    *Config
}

func NewSession(conn *net.TCPConn, conf *Config) *Session {
	conn.SetReadBuffer(conf.ReadBufSize)
	conn.SetWriteBuffer(conf.WriteBufSize)
	s := &Session{
		conn:    conn,
		cr:      bufio.NewReaderSize(conn, conf.ReadBufSize),
		cw:      bufio.NewWriterSize(conn, conf.WriteBufSize),
		ReadCh:  make(chan *Packet, conf.ReadChanSize),
		WriteCh: make(chan *Packet, conf.WriteChanSize),
		living:  true,
		conf:    conf,
	}
	return s
}

// 读取数据
func (s *Session) ReadPacket() {
	buf := bytes.NewBuffer(nil)
	for s.living {
		data, err := s.cr.ReadSlice(CRLF[0])
		if err != nil {
			if err == io.EOF { // 对方主动关闭
				s.Close()
				return
			}
			continue
		}

		n, err := buf.Write(data)
		if err != nil {
			logx.Error(err)
		}
		if n < len(data) {
			buf.Write(data[n:]) // 继续写
		}

		next, err := s.cr.ReadByte()
		if err != nil {
			logx.Error(err)
			if err == io.EOF {
				s.Close()
				return
			}
		}
		buf.Write([]byte{next})

		if next != CRLF[1] {
			logx.Error("invalid next byte")
			buf.Reset()
			continue
		}

		// 读取包数据
		p, err := UnmarshalPacket(buf.Bytes())
		if err != nil || p == nil {
			logx.Error("invalid packet data: %v", err)
			buf.Reset()
			continue
		}

		// 写入读缓冲
		s.ReadCh <- p
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

func (s *Session) writeConn(p Packet) {
	buf := MarshalPacket(p)
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
