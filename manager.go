package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

type EventManager interface {
	RegisterConnection(eventId string, conn *websocket.Conn)
	RemoveConnection(eventId string, conn *websocket.Conn)
	GetConnByEvent(eventId string) []*websocket.Conn
}

type Notifier interface {
	SendJsonByEventID(eventID string, object interface{}) []error
}

type eventConn struct {
	eventId string
	conn    *websocket.Conn
}

type MapEventManager struct {
	*sync.Mutex
	connMap map[string]map[*websocket.Conn]bool
}

func (m *MapEventManager) RegisterConnection(eventId string, conn *websocket.Conn) {
	m.Lock()
	if m.connMap[eventId] == nil {
		m.connMap[eventId] = make(map[*websocket.Conn]bool)
	}
	m.connMap[eventId][conn] = true
	m.Unlock()
}

func (m *MapEventManager) RemoveConnection(eventId string, conn *websocket.Conn) {
	m.Lock()
	if m.connMap[eventId] != nil {
		delete(m.connMap[eventId], conn)
	}
	m.Unlock()
}

func (m *MapEventManager) GetConnByEvent(eventId string) []*websocket.Conn {
	keys := make([]*websocket.Conn, 0)
	m.Lock()
	if m.connMap[eventId] != nil {
		for k := range m.connMap[eventId] {
			keys = append(keys, k)
		}
	}
	m.Unlock()
	return keys
}

func (m *MapEventManager) SendJsonByEventID(eventID string, object interface{}) []error {
	errors := make([]error, 0)
	connections := m.GetConnByEvent(eventID)
	m.Lock()
	for _, conn := range connections {
		err := conn.WriteJSON(object)
		if err != nil {
			errors = append(errors, err)
		}
	}
	m.Unlock()
	return errors
}

func NewEventManager() *MapEventManager {
	return newMapEventManager()
}

func newMapEventManager() *MapEventManager {
	manager := &MapEventManager{
		&sync.Mutex{},
		make(map[string]map[*websocket.Conn]bool),
	}
	return manager
}
