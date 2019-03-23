package tron

import (
	"fmt"
	"net"
	"time"
)

type Client struct {
	conn       *net.TCPConn                 // 原生连接
	session    *Session                     // 连接会话
	heartbeat  int64                        // 最后心跳时间
	localAddr  string                       // 本端地址
	remoteAddr string                       // 对端地址
	handler    func(cli *Client, p *Packet) // 包处理函数
	conf       *Config                      // 共享配置
	codec      Codec
}

func NewClient(conn *net.TCPConn, conf *Config, codec Codec, f func(cli *Client, p *Packet)) *Client {
	session := NewSession(conn, conf, codec)
	cli := &Client{
		conn:      conn,
		heartbeat: time.Now().Unix(),
		session:   session,
		handler:   f,
		conf:      conf,
		codec:     codec,
	}
	return cli
}

// 从连接中读数据，处理包，写回数据
func (c *Client) Run() {
	lAddr, _ := c.conn.LocalAddr().(*net.TCPAddr)
	rAddr, _ := c.conn.RemoteAddr().(*net.TCPAddr)
	c.localAddr = fmt.Sprintf("%s:%d", lAddr.IP, lAddr.Port)
	c.remoteAddr = fmt.Sprintf("%s:%d", rAddr.IP, rAddr.Port)

	// 读写连接
	go c.session.ReadPacket()
	go c.session.WritePacket()

	// 处理接收到的包
	go c.handle()
}

// 写入新 pack
func (c *Client) DirectWrite(newPack *Packet) (chan interface{}, error) {
	nextSeq, respCh := c.fillSeq(newPack)
	c.conf.SeqManager.RegisterSeq(nextSeq, respCh)
	return respCh, c.session.DirectWrite(newPack)
}

// 定期检测连接是否存活
func (c *Client) Ping(heartbeat *Packet, timeout time.Duration) error {
	v, err := c.SyncWriteAndRead(heartbeat, timeout)
	if err != nil {
		return fmt.Errorf("ping: %v", err)
	}

	pong, ok := v.(int64)
	if !ok {
		return fmt.Errorf("ping: invalid pong data type")
	}
	if pong > c.heartbeat { // 避免 packet 延迟
		c.heartbeat = pong
	}
	return nil
}

// 同步请求，用于检测心跳等
func (c *Client) SyncWriteAndRead(newPack *Packet, timeout time.Duration) (interface{}, error) {
	respCh, err := c.DirectWrite(newPack)
	if err != nil {
		return nil, err
	}
	select {
	case <-time.After(timeout):
		return nil, fmt.Errorf("sync write: %d timeout", timeout)
	case resp := <-respCh:
		return resp, nil
	}
}

// 填充新的 seq
func (c *Client) fillSeq(newPack *Packet) (int32, chan interface{}) {
	if newPack.Header.Seq > 0 {
		return newPack.Header.Seq, nil
	}

	nextSeq := c.conf.SeqManager.NextSeq()
	respCh := make(chan interface{}, 1)
	newPack.Header.Seq = nextSeq

	return nextSeq, respCh
}

// 处理完毕
func (c *Client) Detach(seq int32, resp interface{}) {
	c.conf.SeqManager.RemoveSeq(seq, resp)
}

// 分发处理收取到的包
func (c *Client) handle() {
	for c.session != nil && c.session.living {
		if p, ok := <-c.session.ReadCh; ok {
			if c.handler != nil {
				go c.handler(c, p)
			}
		}
	}
}

func (c *Client) LocalAddr() string {
	return c.localAddr
}

func (c *Client) RemoteAddr() string {
	return c.remoteAddr
}

func (c *Client) Living() bool {
	return c.session.Living()
}

// 检测当前连接是否闲置
func (c *Client) IsIdle() bool {
	return c.session.IsIdle()
}

// 尝试重连
func (c *Client) reconnect() (bool, error) {
	newConn, err := net.DialTCP("tcp4", nil, c.conn.RemoteAddr().(*net.TCPAddr))
	if err != nil {
		return false, err
	}

	c.conn = newConn
	c.session = NewSession(newConn, c.conf, c.codec) // 建立连接
	c.Run()                                          // 重启
	return true, nil
}
