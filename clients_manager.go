package tron

import (
	"sync"
	"time"
)

type ClientsManager struct {
	reconnector *ReconnectTaskManager
	addr2client map[string]*Client // client.address -> Client
	lock        sync.Mutex
}

func NewClientsManager(r *ReconnectTaskManager) *ClientsManager {
	m := &ClientsManager{
		reconnector: r,
		addr2client: make(map[string]*Client),
		lock:        sync.Mutex{},
	}

	go m.manage()
	return m
}

func (m *ClientsManager) Add(newClient *Client) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.addr2client[newClient.RemoteAddr()] = newClient
}

func (m *ClientsManager) manage() {
	tick := time.NewTicker(5 * time.Second)
	for {
		for _, cli := range m.addr2client {
			if !cli.Living() {
				m.tryReconnect(cli)
			}
		}
		<-tick.C
	}
}

// 尝试重连
func (m *ClientsManager) tryReconnect(cli *Client) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.reconnector.prepare(cli)
}
