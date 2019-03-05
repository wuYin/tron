package tron

import (
	"github.com/wuYin/logx"
	"math"
	"sync"
	"time"
)

type reconnectTask struct {
	client  *Client
	retried int
}

func NewReconnectTask(cli *Client) *reconnectTask {
	return &reconnectTask{
		client:  cli,
		retried: 0,
	}
}

// 尝试重连
func (t *reconnectTask) reconnect() bool {
	t.retried++

	succ, err := t.client.reconnect()
	if !succ || err != nil {
		logx.Error(succ, err)
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

	newTask := NewReconnectTask(cli)
	m.exec(newTask)
}

// 执行重连任务
func (m *ReconnectTaskManager) exec(task *reconnectTask) {
	addr := task.client.RemoteAddr()
	taskTicker := time.AfterFunc(m.timeout, func() {
		succ := task.reconnect()
		if succ {
			if _, ok := m.taskTickers[addr]; ok {
				delete(m.taskTickers, addr)
			}
			logx.Debug("reconnect %s succ", addr)
		} else {
			ticker := m.taskTickers[addr]
			// 超出重试次数
			if task.retried > m.maxRetry {
				logx.Debug("tried %s %d times, more than max: %d", task.client.RemoteAddr(), task.retried, m.maxRetry)
				ticker.Stop()
				delete(m.taskTickers, addr)
				return
			}

			// 二次规避重试策略
			next := math.Pow(2, float64(task.retried))
			ticker.Reset(time.Duration(next) * m.timeout)
			logx.Debug("reconnect %s %d times...", addr, task.retried)
		}
	})

	m.taskTickers[addr] = taskTicker
}
