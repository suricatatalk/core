package main

import (
	// "log"
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
		time.Now().Unix(),
		time.Now().Unix(),
		"sohlich@gmail.com",
		[]Room{},
		[]Session{},
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

func TestInserCompleteEvent(t *testing.T) {
	mongo = createMgoStorage()

	rooms := []Room{
		{"U51/202", "Workshop lab.", ""},
		{"U51/207", "Presentation room.", ""},
	}

	sessions := []Session{
		{
			"U51/202",
			"Mongo Workshop",
			generateToken(4),
			time.Now().Unix(),
			time.Now().Add(3 * time.Hour).Unix(),
			false,
		}, {
			"U51/202",
			"Redis Workshop",
			generateToken(4),
			time.Now().Add(3 * time.Hour).Unix(),
			time.Now().Add(6 * time.Hour).Unix(),
			false,
		}, {
			"U51/207",
			"Mongo Design Patterns",
			generateToken(4),
			time.Now().Unix(),
			time.Now().Add(6 * time.Hour).Unix(),
			false,
		},
	}

	event := &Event{
		bson.NewObjectId(),
		"1234",
		"Java Intro",
		time.Now().Unix(),
		time.Now().Unix(),
		"sohlich@gmail.com",
		rooms,
		sessions,
	}

	mongo.InsertEvent(event)

}
