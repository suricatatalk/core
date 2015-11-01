package main

import (
	"testing"
	"time"

	"gopkg.in/mgo.v2/bson"
)

func cleanUp(store *MgoDataStorage) {
	store.mgoDB.DropDatabase()
	store.CloseSession()
}

func createMgoStorage() *MgoDataStorage {
	mongo := NewMgoStorage()
	mongo.database = "surikata_test"
	mongo.OpenSession()
	return mongo
}

func TestInsertEvent(t *testing.T) {
	storage := createMgoStorage()
	defer cleanUp(storage)

	event := &Event{
		bson.NewObjectId(),
		"1234",
		"Java Intro",
		time.Now(),
		time.Now(),
		"sohlich@gmail.com",
	}
	storage.InsertEvent(event)

	n, err := storage.mgoEvents.Count()
	if err != nil {
		t.Error(err)
		t.Error("Cannot query storage")
		return
	}

	if n == 0 {
		t.Error("Event do not insert")
	}
}
