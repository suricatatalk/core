package main

import (
	// "log"
	"encoding/json"
	"fmt"
	"math/rand"
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
		"",
		time.Now().Unix(),
		time.Now().Unix(),
		"sohlich@gmail.com",
		[]Room{},
		[]Session{},
		[]string{},
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

	speakers := []Speaker{
		{
			ID:           bson.NewObjectId(),
			ImageURL:     "http://www.dotgo.eu/images/speakers/robpike.png",
			FirstName:    "Rob",
			LastName:     "Pike",
			Organization: "Google Inc.",
			URLs:         []string{"https://twitter.com/rob_pike"},
		},
		{
			ID:           bson.NewObjectId(),
			ImageURL:     "http://www.dotgo.eu/images/speakers/veronicalopez.png",
			FirstName:    "Verónica",
			LastName:     "López",
			Organization: "Ardan Labs",
			URLs:         []string{"https://twitter.com/maria_fibonacci"},
		},
		{
			ID:           bson.NewObjectId(),
			ImageURL:     "http://www.dotgo.eu/images/speakers/francesc-campoy-flores.png",
			FirstName:    "Francesc",
			LastName:     "Flores",
			Organization: "Google Inc.",
			URLs:         []string{"https://twitter.com/francesc"},
		},
	}

	startTime := time.Unix(1451635200, 0)

	rooms := []Room{
		{"U51/202", "#00ffff", "", "Workshop lab"},
		{"U51/109", "#56167d", "", "Conference room"},
		{"U51/207", "#89b524", "", "Presentation room"},
	}

	event := &Event{
		ID:         bson.NewObjectId(),
		EventToken: "",
		Name:       "Open Zlin Fake Conference",
		FromDate:   startTime.Unix(),
		ToDate:     startTime.Add(time.Duration(8) * time.Hour).Unix(),
		CreatedBy:  "sohlich@gmail.com",
	}

	dataLang := []string{"Mongo", "Ruby", "Golang", "JavaScript", "TypeScript", "Kotlin", "Groovy"}
	dataSessionType := []string{"Introduction", "Hardcore", "Talk", "Deep Dive"}

	sessions := make([]Session, 0)

	for index, room := range rooms {
		sessionBegin := startTime
		for sessionBegin.Unix() < startTime.Add(time.Duration(8)*time.Hour).Unix() {
			//Add some time
			sessionEnd := sessionBegin.Add(time.Duration(rand.Intn(80)) * time.Minute)

			session := Session{}
			if index%4 == 0 {
				session.HasDetail = false
			} else {
				session.HasDetail = true
			}
			session.Name = fmt.Sprintf("%s %s", dataLang[rand.Intn(len(dataLang)-1)], dataSessionType[rand.Intn(len(dataSessionType)-1)])
			session.From = sessionBegin.Unix()
			session.To = sessionEnd.Unix()
			session.Description = "Lorem Ipsum Dolore"
			session.Room = room.Name
			session.Speaker = []string{speakers[rand.Intn(len(speakers)-1)].ID.Hex()}
			sessionBegin = sessionEnd.Add(time.Duration(10) * time.Minute)
			sessions = append(sessions, session)
		}
	}

	speakerIds := make([]string, 0)
	for _, sp := range speakers {
		speakerIds = append(speakerIds, sp.ID.Hex())
	}

	event.Rooms = rooms
	event.Speakers = speakerIds
	event.Sessions = sessions
	event.Description = "First fake conference with super program"

	speakerOutput, _ := json.Marshal(speakers)

	fmt.Println(string(speakerOutput))

	// output, _ := json.Marshal(event)

	// fmt.Println(string(output))

	// sessions := []Session{
	// 	{
	// 		"U51/202",
	// 		"Mongo Workshop",
	// 		generateToken(4),
	// 		time.Now().Unix(),
	// 		time.Now().Add(3 * time.Hour).Unix(),
	// 		false,
	// 	}, {
	// 		"U51/202",
	// 		"Redis Workshop",
	// 		generateToken(4),
	// 		time.Now().Add(3 * time.Hour).Unix(),
	// 		time.Now().Add(6 * time.Hour).Unix(),
	// 		false,
	// 	}, {
	// 		"U51/207",
	// 		"Mongo Design Patterns",
	// 		generateToken(4),
	// 		time.Now().Unix(),
	// 		time.Now().Add(6 * time.Hour).Unix(),
	// 		false,
	// 	},
	// }

	// mongo.InsertEvent(event)

}
