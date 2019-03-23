package tron

import (
	"sync"
	"sync/atomic"
)

type SeqManager struct {
	maxSeq int32
	curSeq int32
	locks  []*sync.Mutex
	groups []map[int32]chan interface{}
}

const (
	MAX_CONCUR = 10 // 并发级别
)

// 并发级别决定分组
func NewSeqManager(maxSeq int32) *SeqManager {
	m := &SeqManager{
		maxSeq: maxSeq,
		curSeq: 0,
		locks:  make([]*sync.Mutex, MAX_CONCUR),
		groups: make([]map[int32]chan interface{}, MAX_CONCUR),
	}

	avgMaxSeq := maxSeq / MAX_CONCUR // 每个底层 manager 分配到的 seq 长度
	for i := 0; i < MAX_CONCUR; i++ {
		m.locks[i] = &sync.Mutex{}
		m.groups[i] = make(map[int32]chan interface{}, avgMaxSeq) // 分配内存直接使用
	}

	return m
}

// 记录一个新的 seq 及其响应 channel
func (m *SeqManager) AddSeq(nextSeq int32, respCh chan interface{}) {
	l, g := m.group(nextSeq)
	l.Lock()
	g[nextSeq] = respCh
	l.Unlock()
}

// 取出指定 seq 及其 channel 来发送响应
func (m *SeqManager) RemoveSeq(oldSeq int32, res interface{}) {
	l, g := m.group(oldSeq)
	l.Lock()
	defer l.Unlock()

	if ch, ok := g[oldSeq]; ok {
		ch <- res
		close(ch)
		delete(g, oldSeq)
	}
}

// 获取下一个可分配的 seq
func (m *SeqManager) NextSeq() int32 {
	next := atomic.AddInt32(&m.curSeq, 1)
	return next % m.maxSeq // 轮回使用
}

func (m *SeqManager) group(seq int32) (lock *sync.Mutex, group map[int32]chan interface{}) {
	g := seq % MAX_CONCUR
	lock = m.locks[g]
	group = m.groups[g]
	return
}
