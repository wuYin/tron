package tron

import (
	"fmt"
	"sync"
	"time"
)

type Group struct {
	ID  string
	Key string
}

func NewGroup(id, key string) *Group {
	return &Group{
		ID:  id,
		Key: key,
	}
}

type ClientsManager struct {
	client2group  map[string]*Group    // client.address -> Group
	group2clients map[string][]*Client // group_id -> Group
	addr2client   map[string]*Client   // client.address -> Client
	lock          sync.Mutex
}

func NewClientsManager() *ClientsManager {
	m := &ClientsManager{
		client2group:  make(map[string]*Group),
		group2clients: make(map[string][]*Client),
		addr2client:   make(map[string]*Client),
		lock:          sync.Mutex{},
	}

	go m.manage()
	return m
}

func (m *ClientsManager) Join(g *Group, newClient *Client) {
	m.lock.Lock()
	defer m.lock.Unlock()

	clients, ok := m.group2clients[g.ID]
	if !ok {
		clients = make([]*Client, 0, 10)
	}

	m.group2clients[g.ID] = append(clients, newClient)
	m.client2group[newClient.RemoteAddr()] = g
	m.addr2client[newClient.RemoteAddr()] = newClient
}

func (m *ClientsManager) manage() {
	tick := time.NewTicker(1 * time.Second)
	for {
		for _, cli := range m.addr2client {
			if !cli.Living() {
				fmt.Println("retry connecting...")
			}
			fmt.Println("connect living...")
		}
		<-tick.C
	}
}
