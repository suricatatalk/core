package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

type EventManager interface {
	RegisterConnection(eventToken, sessionToken string, conn *websocket.Conn)
	RemoveConnection(eventToken, sessionToken string, conn *websocket.Conn)
	GetConnByEventSession(eventToken, sessionToken string) []*websocket.Conn
}

type Notifier interface {
	SendJsonByEventAndSessionToken(eventToken, sessionToken string, object interface{}) []error
}

type eventConn struct {
	mergedToken string
	conn        *websocket.Conn
}

type MapEventManager struct {
	*sync.Mutex
	connMap map[string]map[*websocket.Conn]bool
}

func (m *MapEventManager) RegisterConnection(eventToken, sessionToken string, conn *websocket.Conn) {
	mergedToken := mergeToken(eventToken, sessionToken)
	m.Lock()
	if m.connMap[mergedToken] == nil {
		m.connMap[mergedToken] = make(map[*websocket.Conn]bool)
	}
	m.connMap[mergedToken][conn] = true
	m.Unlock()
}

func (m *MapEventManager) RemoveConnection(eventToken, sessionToken string, conn *websocket.Conn) {
	mergedToken := mergeToken(eventToken, sessionToken)
	m.Lock()
	if m.connMap[mergedToken] != nil {
		delete(m.connMap[mergedToken], conn)
	}
	m.Unlock()
}

func (m *MapEventManager) GetConnByEventSession(eventToken, sessionToken string) []*websocket.Conn {
	mergedToken := mergeToken(eventToken, sessionToken)
	keys := make([]*websocket.Conn, 0)
	m.Lock()
	if m.connMap[mergedToken] != nil {
		for k := range m.connMap[mergedToken] {
			keys = append(keys, k)
		}
	}
	m.Unlock()
	return keys
}

func (m *MapEventManager) SendJsonByEventAndSessionToken(eventToken, sessionToken string, object interface{}) []error {
	errors := make([]error, 0)
	connections := m.GetConnByEventSession(eventToken, sessionToken)
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

func mergeToken(eventToken, sessionToken string) string {
	return eventToken + sessionToken
}
