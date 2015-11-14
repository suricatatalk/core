package main

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Vote struct {
	Client    string
	VoteTime  int64
	Operation int
}

type Question struct {
	ID           bson.ObjectId `bson:"_id"`
	SessionToken string
	EventToken   string
	Question     string
	Vote         int
	CreateTime   int64
}

type Session struct {
	Room         string
	Name         string
	SessionToken string
	From         int64
	To           int64
	Finished     bool
}

type Room struct {
	Name        string
	NameHash    string
	Description string
}

type Event struct {
	ID         bson.ObjectId `bson:"_id"`
	EventToken string
	Name       string
	FromDate   int64
	ToDate     int64
	CreatedBy  string
	Rooms      []Room
	Sessions   []Session
}

type EventStorage interface {
	InsertEvent(event *Event) error
	UpdateEvent(event *Event) error
	DeleteEvent(eventID string) error
	EventByToken(token string) (*Event, error)
}

type QuestionStorage interface {
	InsertQuestion(question *Question) error
	QuestionById(questionID string) (*Question, error)
	VoteQuestion(questionID string) error
	QuestionsByEventAndSession(eventtoken, sessionToken string) ([]Question, error)
}

type DataStorage interface {
	EventStorage
	QuestionStorage
	OpenSession() error
	CloseSession()
}

type MgoDataStorage struct {
	connectionString string
	database         string
	events           string
	questions        string
	mgoSession       *mgo.Session
	mgoDB            *mgo.Database
	mgoEvents        *mgo.Collection
	mgoQuestions     *mgo.Collection
}

func NewMgoStorage() *MgoDataStorage {
	return &MgoDataStorage{
		connectionString: "localhost:27017",
		database:         "surikata",
		events:           "events",
		questions:        "questions",
	}

}

func (a *MgoDataStorage) OpenSession() error {
	var err error
	a.mgoSession, err = mgo.Dial(a.connectionString)
	if err != nil {
		return err
	}
	a.mgoDB = a.mgoSession.DB(a.database)
	a.mgoEvents = a.mgoDB.C(a.events)
	a.mgoQuestions = a.mgoDB.C(a.questions)

	a.mgoEvents.EnsureIndex(mgo.Index{
		Key:        []string{"eventtoken"},
		Unique:     true,
		Background: true,
	})
	a.mgoQuestions.EnsureIndex(mgo.Index{
		Key:        []string{"eventtoken"},
		Background: true,
	})
	a.mgoQuestions.EnsureIndex(mgo.Index{
		Key:        []string{"sessionname"},
		Background: true,
	})
	return nil
}

func (a *MgoDataStorage) CloseSession() {
	a.mgoSession.Close()
}

func (m *MgoDataStorage) InsertEvent(event *Event) error {
	event.ID = bson.NewObjectId()
	event.EventToken = generateToken(8)
	fillTokens(event)
	return m.mgoEvents.Insert(event)
}

func (m *MgoDataStorage) UpdateEvent(event *Event) error {
	fillTokens(event)
	return m.mgoEvents.UpdateId(event.ID, event)
}

func (m *MgoDataStorage) DeleteEvent(eventId string) error {
	return m.mgoEvents.RemoveId(bson.ObjectIdHex(eventId))
}

func (m *MgoDataStorage) EventByToken(token string) (*Event, error) {
	result := &Event{}
	err := m.mgoEvents.Find(bson.M{"eventtoken": token}).One(result)
	return result, err
}

func (m *MgoDataStorage) InsertQuestion(question *Question) error {
	question.ID = bson.NewObjectId()
	return m.mgoQuestions.Insert(question)
}

func (m *MgoDataStorage) VoteQuestion(questionId string) error {
	return m.mgoQuestions.UpdateId(bson.ObjectIdHex(questionId), bson.M{"$inc": bson.M{"vote": 1}})
}

func (m *MgoDataStorage) QuestionById(questionID string) (*Question, error) {
	result := &Question{}
	err := m.mgoQuestions.FindId(bson.ObjectIdHex(questionID)).One(result)
	return result, err
}

func (m *MgoDataStorage) QuestionsByEventAndSession(eventToken, sessiontToken string) ([]Question, error) {
	result := make([]Question, 0)
	err := m.mgoQuestions.Find(bson.M{"eventtoken": eventToken, "sessiontoken": sessiontToken}).All(&result)
	return result, err
}

func generateToken(length int) string {
	token := uuid.NewV4()
	sha := sha256.Sum256(token.Bytes())
	return hex.EncodeToString(sha[:(length / 2)])
}

func fillTokens(event *Event) {
	if event.Rooms != nil {
		for i := 0; i < len(event.Rooms); i++ {
			if len(event.Rooms[i].NameHash) == 0 {
				event.Rooms[i].NameHash = generateToken(4)
			}
		}
	}
	if event.Sessions != nil {
		for i := 0; i < len(event.Sessions); i++ {
			if len(event.Sessions[i].SessionToken) == 0 {
				event.Sessions[i].SessionToken = generateToken(4)
			}
		}
	}
}
