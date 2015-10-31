package main

import (
	"testing"
	"time"

	"gopkg.in/mgo.v2/bson"
)

func TestInsertEvent(t *testing.T) {
	storage := NewMgoStorage()
	storage.OpenSession()
	defer storage.CloseSession()

	event := &Event{
		bson.NewObjectId(),
		"1234",
		"Java Intro",
		time.Now(),
		time.Now(),
		"sohlich@gmail.com",
	}
	storage.InsertEvent(event)
}
