package tron

import (
	"sync"
	"time"
)

type ClientsManager struct {
	reconnector   *ReconnectTaskManager
	addr2Client   map[string]*Client   // remoteAddr -> client
	group2Clients map[string][]*Client // gid -> clients
	allClients    map[string]*Client   // localAddr -> gid
	lock          sync.Mutex
}

func NewClientsManager(r *ReconnectTaskManager) *ClientsManager {
	m := &ClientsManager{
		reconnector:   r,
		addr2Client:   make(map[string]*Client),
		group2Clients: make(map[string][]*Client),
		allClients:    make(map[string]*Client),
		lock:          sync.Mutex{},
	}

	go m.manage()
	return m
}

func (m *ClientsManager) Add(g *ClientGroup, cli *Client) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.allClients[cli.LocalAddr()] = cli
	m.addr2Client[cli.RemoteAddr()] = cli
	m.group2Clients[g.Gid] = append(m.group2Clients[g.Gid], cli)
}

// 从特定组中选择依旧连接的 client
func (m *ClientsManager) FindClients(gid string, filter func(gid string, cli *Client) bool) []*Client {
	clis, ok := m.group2Clients[gid]
	if !ok {
		return nil
	}

	var clients []*Client
	for _, cli := range clis {
		if cli.IsClosed() || filter(gid, cli) {
			continue
		}
		clients = append(clients, cli)
	}
	return clients
}

func (m *ClientsManager) manage() {
	tick := time.NewTicker(5 * time.Second)
	for {
		for _, cli := range m.addr2Client {
			if cli.IsClosed() {
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
