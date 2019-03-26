package tron

import (
	"fmt"
	"math"
	"sync"
	"time"
	"timewheel"
)

type reconnectTask struct {
	client  *Client
	retried int
}

func newReconnectTask(cli *Client) *reconnectTask {
	return &reconnectTask{
		client:  cli,
		retried: 0,
	}
}

// 尝试重连
func (t *reconnectTask) connect() bool {
	t.retried++

	succ, err := t.client.reconnect()
	if !succ || err != nil {
		return false
	}
	return true
}

type ReconnectTaskManager struct {
	tw       *timewheel.TimeWheel // 任务时间轮
	timeout  time.Duration        // 初次尝试重连的超时时间
	maxRetry int                  // 最大尝试次数
	retrying map[string]bool      // 正在重连
	lock     sync.Mutex
}

var (
	retryPoints []int64
)

func NewReconnectTaskManager(timeout time.Duration, maxRetry int) *ReconnectTaskManager {
	m := &ReconnectTaskManager{
		tw:       timewheel.NewTimeWheel(10*time.Millisecond, 6000),
		timeout:  timeout,
		maxRetry: maxRetry,
		retrying: make(map[string]bool),
	}
	for i := 0; i < maxRetry; i++ {
		retryPoints = append(retryPoints, int64(math.Pow(2, float64(i)))) // 二次规避策略
	}

	return m
}

// 准备重连任务
func (m ReconnectTaskManager) reconnect(cli *Client) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.retrying[cli.RemoteAddr()]; ok { // 已经在尝试重连了...
		return
	}

	newTask := newReconnectTask(cli)
	go m.exec(newTask)
	m.retrying[cli.RemoteAddr()] = true
}

var (
	STATUS_SUCC      = 0
	STATUS_FAIL      = 1
	STATUS_OVERTRIED = 2
)

// 执行重连任务
func (m *ReconnectTaskManager) exec(task *reconnectTask) {
	serverAddr := task.client.RemoteAddr()
	sAddr := SplitPort(serverAddr)
	lAddr := SplitPort(task.client.LocalAddr())

	fmt.Printf("%v %v\n", m.timeout, retryPoints)
	tids, resChs := m.tw.AfterPoints(m.timeout, retryPoints, func() interface{} {
		if succ := task.connect(); !succ {
			if task.retried >= m.maxRetry {
				fmt.Printf("[client:%s] -> [server:%s] over tried %d times\n", lAddr, sAddr, task.retried)
				return STATUS_OVERTRIED
			}
			fmt.Printf("[client:%s] -> [server:%s] try reconnecting...  %d times\n", lAddr, sAddr, task.retried)
			return STATUS_FAIL
		}
		fmt.Printf("[client:%s] -> [server:%s] reconnected succ\n", lAddr, sAddr)
		return STATUS_SUCC
	})

	doneCh := make(chan int, 1)
	for _, resCh := range resChs {
		go func(ch chan interface{}) {
			for v := range ch {
				doneCh <- v.(int)
			}
		}(resCh)
	}

	for status := range doneCh {
		switch status {
		case STATUS_FAIL:
			continue
		case STATUS_SUCC: // 重连成功
			for _, tid := range tids {
				m.tw.Cancel(tid) // 取消剩余任务
			}
			delete(m.retrying, task.client.RemoteAddr())
		case STATUS_OVERTRIED:
			delete(m.retrying, task.client.RemoteAddr())
		}
	}
	close(doneCh)
}
