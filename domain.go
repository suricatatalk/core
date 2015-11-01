package main

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Vote struct {
	Client    string
	VoteTime  time.Time
	Operation int
}

type Question struct {
	ID          bson.ObjectId `bson:"_id"`
	SessionName string
	EventToken  string
	Question    string
	Vote        int
	CreateTime  time.Time
}

// type Session struct {
// 	Name     string
// 	From     time.Time
// 	To       time.Time
// 	Finished bool
// }

type Event struct {
	ID         bson.ObjectId `bson:"_id"`
	EventToken string
	Name       string
	FromDate   time.Time
	ToDate     time.Time
	CreatedBy  string
}

type DataStorage interface {
	OpenSession() error
	CloseSession()
	InsertEvent(event *Event) error
	UpdateEvent(event *Event) error
	DeleteEvent(eventID string) error
	EventByToken(token string) (*Event, error)
	InsertQuestion(question *Question) error
	QuestionById(questionID string) (*Question, error)
	VoteQuestion(questionID string) error
	QuestionsByEvent(eventID string) ([]Question, error)
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
	return m.mgoEvents.Insert(event)
}

func (m *MgoDataStorage) UpdateEvent(event *Event) error {
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

func (m *MgoDataStorage) QuestionsByEvent(eventToken string) ([]Question, error) {
	result := make([]Question, 0)
	err := m.mgoQuestions.Find(bson.M{"eventtoken": eventToken}).All(&result)
	return result, err
}
