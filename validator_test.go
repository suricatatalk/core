package main

import (
	"testing"
	"time"

	"gopkg.in/mgo.v2/bson"
)

func TestValidateEventFromTo(t *testing.T) {

	now := time.Now()

	event := &Event{
		ID:         bson.NewObjectId(),
		EventToken: "",
		Name:       "Open Zlin Fake Conference",
		FromDate:   now.Unix(),
		ToDate:     now.Add(time.Duration(8) * time.Hour).Unix(),
		CreatedBy:  "sohlich@gmail.com",
	}

	err := ValidateEvent(event)

	if err != nil {
		t.Error(err)
		return
	}

	event.ToDate = now.Add(time.Duration(-8) * time.Hour).Unix()

	err = ValidateEvent(event)

	if err == nil {
		t.Error("Validator failed")
		return
	}

}

func TestValidateSessionRoomNotInEvent(t *testing.T) {

	now := time.Now()

	event := &Event{
		ID:         bson.NewObjectId(),
		EventToken: "",
		Name:       "Open Zlin Fake Conference",
		FromDate:   now.Unix(),
		ToDate:     now.Add(time.Duration(8) * time.Hour).Unix(),
		CreatedBy:  "sohlich@gmail.com",
	}

	rooms := []Room{
		{"U51/202", "#00ffff", "", "Workshop lab"},
		{"U51/109", "#56167d", "", "Conference room"},
		{"U51/207", "#89b524", "", "Presentation room"},
	}

	session := Session{
		Room:         "U51/203",
		Name:         "Test session",
		Speaker:      []string{"123456"},
		Description:  "This is test",
		SessionToken: "XYZ",
		From:         now.Unix(),
		To:           now.Add(1 * time.Hour).Unix(),
	}

	event.Rooms = rooms
	event.Sessions = []Session{session}

	err := ValidateEvent(event)

	if err == nil {
		t.Error("Validator failed")
	}

}

func TestValidateSessionInSequence(t *testing.T) {

	now := time.Now()

	event := &Event{
		ID:         bson.NewObjectId(),
		EventToken: "",
		Name:       "Open Zlin Fake Conference",
		FromDate:   now.Unix(),
		ToDate:     now.Add(time.Duration(8) * time.Hour).Unix(),
		CreatedBy:  "sohlich@gmail.com",
	}

	rooms := []Room{
		{"U51/202", "#00ffff", "", "Workshop lab"},
		{"U51/109", "#56167d", "", "Conference room"},
		{"U51/207", "#89b524", "", "Presentation room"},
	}

	session := Session{
		Room:         "U51/202",
		Name:         "Test session",
		Speaker:      []string{"123456"},
		Description:  "This is test",
		SessionToken: "XYZ",
		From:         now.Unix(),
		To:           now.Add(1 * time.Hour).Unix(),
	}

	session2 := Session{
		Room:         "U51/207",
		Name:         "Test session",
		Speaker:      []string{"123456"},
		Description:  "This is test",
		SessionToken: "XYZ",
		From:         now.Unix(),
		To:           now.Add(1 * time.Hour).Unix(),
	}

	event.Rooms = rooms
	event.Sessions = []Session{session, session2}

	err := ValidateEvent(event)

	if err != nil {
		t.Error("Validator failed for two different rooms same time")
	}

	session3 := Session{
		Room:         "U51/207",
		Name:         "Test session",
		Speaker:      []string{"123456"},
		Description:  "This is test",
		SessionToken: "XYZ",
		From:         now.Unix(),
		To:           now.Add(1 * time.Hour).Unix(),
	}

	event.Sessions = append(event.Sessions, session3)

	if err != nil {
		t.Error("Validator failed")
	}

}
