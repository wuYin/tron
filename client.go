package tron

import (
	"fmt"
	"net"
	"time"
)

type Client struct {
	conn      *net.TCPConn                 // 原生连接
	session   *Session                     // 连接会话
	heartbeat int64                        // 最后心跳时间
	handler   func(cli *Client, p *Packet) // 包处理函数
	conf      *Config                      // 共享配置
	codec     Codec
}

func NewClient(conn *net.TCPConn, conf *Config, workerCodec Codec, f func(cli *Client, p *Packet)) *Client {
	session := NewSession(conn, conf, workerCodec)
	cli := &Client{
		conn:      conn,
		heartbeat: time.Now().Unix(),
		session:   session,
		handler:   f,
		conf:      conf,
		codec:     workerCodec,
	}
	return cli
}

// 从连接中读数据，处理包，写回数据
func (c *Client) ReadWriteAndHandle() {
	// 读写连接
	go c.session.daemonReadPacket()
	go c.session.daemonWritePacket()

	// 处理接收到的包
	go c.handle()
}

// 异步写
func (c *Client) AsyncWrite(p *Packet) (chan interface{}, error) {
	if p.Header.Seq >= 0 {
		return nil, c.session.Write(p) // worker 的响应直接写回
	}

	// 请求的 packet 将 seq 写入
	p.Header.Seq = c.conf.SeqManager.NextSeq()
	respCh := make(chan interface{}, 1)
	c.conf.SeqManager.AddSeq(p.Header.Seq, respCh)
	return respCh, c.session.Write(p)
}

// 同步写
func (c *Client) SyncWrite(newPack *Packet, timeout time.Duration) (interface{}, error) {
	respCh, err := c.AsyncWrite(newPack)
	if err != nil {
		return nil, err
	}
	select {
	case <-time.After(timeout):
		return nil, fmt.Errorf("sync write: %.fs timeout", timeout.Seconds())
	case resp := <-respCh:
		return resp, nil
	}
}

// client 收到 worker 发出的响应数据
func (c *Client) NotifyReceived(seq int32, resp interface{}) {
	c.conf.SeqManager.RemoveSeq(seq, resp)
}

// 分发处理收取到的包
func (c *Client) handle() {
	for c.session != nil && !c.session.IsClosed() {
		if p, ok := <-c.session.ReadCh; ok {
			if c.handler != nil {
				go c.handler(c, p)
			}
		}
	}
}

func (c *Client) LocalAddr() string {
	return c.session.LocalAddr()
}

func (c *Client) RemoteAddr() string {
	return c.session.RemoteAddr()
}

func (c *Client) IsClosed() bool {
	return c.session.IsClosed()
}

// 尝试重连
func (c *Client) reconnect() (bool, error) {
	newConn, err := net.DialTCP("tcp4", nil, c.conn.RemoteAddr().(*net.TCPAddr))
	if err != nil {
		return false, err
	}

	c.conn = newConn
	c.session = NewSession(newConn, c.conf, c.codec) // 建立连接
	c.ReadWriteAndHandle()                           // 重启
	return true, nil
}
