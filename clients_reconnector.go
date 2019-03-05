package tron

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type reconnectTask struct {
	client    *Client
	retried   int
}

func newReconnectTask(cli *Client) *reconnectTask {
	return &reconnectTask{
		client:    cli,
		retried:   0,
	}
}

// 尝试重连
func (t *reconnectTask) reconnect() bool {
	t.retried++

	succ, err := t.client.reconnect()
	if !succ || err != nil {
		return false
	}
	return true
}

type ReconnectTaskManager struct {
	taskTickers map[string]*time.Timer // 任务定时器
	timeout     time.Duration          // 重连超时
	maxRetry    int                    // 最大尝试次数
	lock        sync.Mutex
}

func NewReconnectTaskManager(timeout time.Duration, maxRetry int) *ReconnectTaskManager {
	return &ReconnectTaskManager{
		taskTickers: make(map[string]*time.Timer),
		timeout:     timeout,
		maxRetry:    maxRetry,
	}
}

// 准备重连任务
func (m ReconnectTaskManager) prepare(cli *Client) {
	m.lock.Lock()
	defer m.lock.Unlock()

	// 任务已存在
	if _, ok := m.taskTickers[cli.RemoteAddr()]; ok {
		return
	}
	newTask := newReconnectTask(cli)
	m.exec(newTask)
}

// 执行重连任务
func (m *ReconnectTaskManager) exec(task *reconnectTask) {
	serverAddr := task.client.RemoteAddr()
	sAddr := port(serverAddr)
	lAddr := port(task.client.LocalAddr())
	taskTicker := time.AfterFunc(m.timeout, func() {
		succ := task.reconnect()
		if succ {
			if _, ok := m.taskTickers[serverAddr]; ok {
				delete(m.taskTickers, serverAddr)
			}
			fmt.Printf("[client:%s] -> [server:%s] reconnected succ\n", lAddr, sAddr)
		} else {
			ticker := m.taskTickers[serverAddr]
			// 超出重试次数
			if task.retried >= m.maxRetry {
				fmt.Printf("[client:%s] -> [server:%s] try reconnected %d times > max %d times\n", lAddr, sAddr, task.retried, m.maxRetry)
				ticker.Stop()
				delete(m.taskTickers, serverAddr)
				return
			}

			// 二次规避重试策略
			next := math.Pow(2, float64(task.retried))
			ticker.Reset(time.Duration(next) * m.timeout)
			fmt.Printf("[client:%s] -> [server:%s] try reconnecting...  %d times\n", lAddr, sAddr, task.retried)
		}
	})

	m.taskTickers[serverAddr] = taskTicker
}
