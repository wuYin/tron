package tron

import (
	"sync"
	"time"
)

type ClientsManager struct {
	reconnector   *ReconnectTaskManager
	addr2Client   map[string]*Client // remoteAddr -> client
	reconnected   map[string]bool
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
		reconnected:   make(map[string]bool),
		lock:          sync.Mutex{},
	}

	go m.daemonReconnect()

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

func (m *ClientsManager) daemonReconnect() {
	tick := time.NewTicker(5 * time.Second)
	for {
		for _, cli := range m.addr2Client {
			if !cli.IsClosed() {
				continue
			}

			// 已尝试过重连
			if m.reconnected[cli.RemoteAddr()] {
				continue
			}
			m.reconnector.reconnect(cli)
			m.reconnected[cli.RemoteAddr()] = true
		}
		<-tick.C
	}
}
