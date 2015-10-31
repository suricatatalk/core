package main

import (
	"testing"

	"github.com/gorilla/websocket"
)

func TestEmptySlice(t *testing.T) {

	eventId := "abcd234"
	conn := &websocket.Conn{}
	manager := newMapEventManager()
	manager.RegisterConnection(eventId, conn)

	if len(manager.connMap) == 0 {
		t.Error("websocket not added")
	}

	websockets := manager.GetConnByEvent(eventId)
	if len(websockets) == 0 {
		t.Error("websockets not obtain")
	}

	manager.RemoveConnection(eventId, conn)

	if len(manager.connMap[eventId]) != 0 {
		t.Error("websocket not added")
	}

}
