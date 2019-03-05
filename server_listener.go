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
	return listener
}

func (l *LiveListener) Accept() (*net.TCPConn, error) {
	for {
		conn, err := l.listener.AcceptTCP()
		select {
		case <-l.closeCh:
			if err := l.listener.Close(); err != nil {
				logx.Error(err)
			}
			return nil, ERR_SERVER_CLOSED
		default:
			// 建立连接前检查下服务器是否已关闭
		}
		if err != nil {
			return nil, err
		}
		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(l.keepAlive)
		return conn, nil
	}
}
