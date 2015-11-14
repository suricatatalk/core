package main

import (
	"testing"

	"github.com/gorilla/websocket"
)

func TestMapManager(t *testing.T) {

	eventId := "abcd234"
	sessionId := "1234"
	mergedToken := mergeToken(eventId, sessionId)
	conn := &websocket.Conn{}
	manager := newMapEventManager()
	manager.RegisterConnection(eventId, sessionId, conn)

	if len(manager.connMap) == 0 {
		t.Error("websocket not added")
	}

	websockets := manager.GetConnByEventSession(eventId, sessionId)
	if len(websockets) == 0 {
		t.Error("websockets not obtain")
	}

	manager.RemoveConnection(eventId, sessionId, conn)

	if len(manager.connMap[mergedToken]) != 0 {
		t.Error("websocket not added")
	}

}
