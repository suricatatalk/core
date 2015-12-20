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
	ID           bson.ObjectId `bson:"_id" json:"id"`
	SessionToken string        `json:"sessionToken"`
	EventToken   string        `json:"eventToken"`
	Question     string        `json:"question"`
	Vote         int           `json:"vote"`
	CreateTime   int64         `json:"createTime"`
}

type Session struct {
	Room         string   `json:"room"`
	Name         string   `json:"name"`
	Speaker      []string `json:"speaker"`
	Description  string   `json:"description"`
	SessionToken string   `json:"sessionToken"`
	From         int64    `json:"from"`
	To           int64    `json:"to"`
	Finished     bool     `json:"finished"`
	HasDetail    bool     `json:"hasDetail"`
}

type Room struct {
	Name        string `json:"name"`
	Tint        string `json:"tint"`
	NameHash    string `json:"nameHash"`
	Description string `json:"description"`
}

type Speaker struct {
	ID           bson.ObjectId `bson:"_id" json:"id"`
	ImageURL     string        `json:"imageUrl"`
	FirstName    string        `json:"firstName"`
	LastName     string        `json:"lastName"`
	Organization string        `json:"organization"`
	URLs         []string      `json:"urls"`
	Bio          string        `json:"bio"`
}

type Event struct {
	ID          bson.ObjectId `bson:"_id",json:"id"`
	EventToken  string        `json:"eventToken"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	FromDate    int64         `json:"fromDate"`
	ToDate      int64         `json:"toDate"`
	CreatedBy   string        `json:"createdBy"`
	Rooms       []Room        `json:"rooms"`
	Sessions    []Session     `json:"sessions"`
	Speakers    []string      `json:"speakers"`
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
	VoteQuestion(questionID string, incBy int) error
	QuestionsByEventAndSession(eventtoken, sessionToken string) ([]Question, error)
}

type SpeakerStorage interface {
	InsertSpeaker(s *Speaker) error
	UpdateSpeaker(s *Speaker) error
	SpeakerById(hexId string) (*Speaker, error)
	SpeakersById(hexId []string) ([]*Speaker, error)
}

type DataStorage interface {
	EventStorage
	QuestionStorage
	SpeakerStorage
	OpenSession() error
	CloseSession()
}

type MgoDataStorage struct {
	connectionString string
	database         string
	events           string
	questions        string
	speakers         string
	mgoSession       *mgo.Session
	mgoDB            *mgo.Database
	mgoEvents        *mgo.Collection
	mgoQuestions     *mgo.Collection
	mgoSpeakers      *mgo.Collection
}

func NewMgoStorage() *MgoDataStorage {
	return &MgoDataStorage{
		connectionString: "localhost:27017",
		database:         "surikata",
		events:           "events",
		questions:        "questions",
		speakers:         "speakers",
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
	a.mgoSpeakers = a.mgoDB.C(a.speakers)

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

func (m *MgoDataStorage) InsertSpeaker(s *Speaker) error {
	s.ID = bson.NewObjectId()
	return m.mgoSpeakers.Insert(s)
}

func (m *MgoDataStorage) UpdateSpeaker(s *Speaker) error {
	_, err := m.mgoSpeakers.UpsertId(s.ID, s)
	if err != nil {
		return err
	}
	return nil
}

func (m *MgoDataStorage) SpeakerById(hexId string) (*Speaker, error) {
	s := &Speaker{}
	err := m.mgoSpeakers.FindId(bson.ObjectIdHex(hexId)).One(s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (m *MgoDataStorage) SpeakersById(hexIds []string) ([]*Speaker, error) {
	speakers := make([]*Speaker, 0)
	for _, id := range hexIds {
		s := &Speaker{}
		err := m.mgoSpeakers.FindId(bson.ObjectIdHex(id)).One(s)
		if err != nil {
			return speakers, err
		}
		speakers = append(speakers, s)
	}
	return speakers, nil
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

func (m *MgoDataStorage) VoteQuestion(questionId string, incBy int) error {
	return m.mgoQuestions.UpdateId(bson.ObjectIdHex(questionId), bson.M{"$inc": bson.M{"vote": incBy}})
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
