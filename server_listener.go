package tron

import (
	"errors"
	"github.com/wuYin/logx"
	"net"
	"time"
)

var ERR_SERVER_CLOSED = errors.New("server closed")

// 手动维护的长连接连接器
type LiveListener struct {
	listener  *net.TCPListener
	closeCh   chan struct{} // 异步主动关闭连接
	keepAlive time.Duration // 保活时间
}

func NewLiveListener(l *net.TCPListener, ch chan struct{}, d time.Duration) *LiveListener {
	listener := &LiveListener{
		listener:  l,
		closeCh:   ch,
		keepAlive: d,
	}

	go func() {
		for {
			select {
			case <-listener.closeCh:
				if err := listener.listener.Close(); err != nil {
					logx.Debug("server close listener failed: %v", err)
					return
				}
				logx.Debug("server close listener succ")
			}
		}
	}()

	return listener
}

func (l *LiveListener) Accept() (*net.TCPConn, error) {
	conn, err := l.listener.AcceptTCP()
	if err != nil {
		return nil, err
	}
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(l.keepAlive)
	return conn, nil
}
