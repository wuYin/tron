package tron

import (
	"fmt"
	"net"
)

type Client struct {
	conn       *net.TCPConn
	session    *Session
	localAddr  string
	remoteAddr string
	handler    func(cli *Client, p *Packet) // 包处理函数
	conf       *Config
}

func NewClient(conn *net.TCPConn, conf *Config, f func(cli *Client, p *Packet)) *Client {
	session := NewSession(conn, conf)
	cli := &Client{
		conn:    conn,
		session: session,
		handler: f,
		conf:    conf,
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

// 暴露的直接写数据方法
func (c *Client) DirectWrite(p *Packet) error {
	return c.session.DirectWrite(p)
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
	c.session = NewSession(newConn, c.conf) // 建立连接
	c.Run()                                 // 重启
	return true, nil
}
