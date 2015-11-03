package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	mongo    DataStorage
	commMan  EventManager
	notifier Notifier
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func main() {
	r := gin.Default()

	mgoStorage := NewMgoStorage()
	connString := fmt.Sprintf("mongodb://%s:%s@%s:%s",
		os.Getenv("MONGODB_ADDON_USER"),
		os.Getenv("MONGODB_ADDON_PASSWORD"),
		os.Getenv("MONGODB_ADDON_HOST"),
		os.Getenv("MONGODB_ADDON_PORT"))
	mgoStorage.connectionString = connString
	mgoStorage.database = os.Getenv("MONGODB_ADDON_DB")
	mongo = mgoStorage
	mongo.OpenSession()

	eventConnManager := NewEventManager()
	commMan = eventConnManager
	notifier = eventConnManager
	err := mongo.OpenSession()
	if err != nil {
		log.Panic(err)
	}

	r.POST("/question/:questionID", voteQuestion)
	r.POST("/question", postQuestion)
	r.GET("/event/:eventtoken", eventWebsockHandler)
	r.Run("localhost:8080")
}

func eventWebsockHandler(c *gin.Context) {
	eventID := c.Params.ByName("eventtoken")
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade: %+v", err)
		return
	}
	commMan.RegisterConnection(eventID, conn)
}

func voteQuestion(c *gin.Context) {
	questionID := c.Params.ByName("questionID")
	err := mongo.VoteQuestion(questionID)
	if err != nil {
		c.JSON(405, "Event not exist")
		return
	}
	q, qerr := mongo.QuestionById(questionID)
	if qerr != nil {
		c.JSON(405, "Event not exist")
		return
	}
	notifyChange(q.EventToken)
	c.JSON(200, "OK")
}

func postQuestion(c *gin.Context) {
	question := &Question{}
	err := c.BindJSON(question)
	if err != nil {
		log.Println(err)
		c.JSON(405, "Cannot store the question")
		return
	}

	_, err = mongo.EventByToken(question.EventToken)
	if err != nil {
		c.JSON(405, "Event not exist")
		return
	}
	mongo.InsertQuestion(question)
	notifyChange(question.EventToken)
	c.JSON(200, "OK")
}

func notifyChange(eventtoken string) error {
	questions, err := mongo.QuestionsByEvent(eventtoken)
	if err != nil {
		return err
	}
	errSlice := notifier.SendJsonByEventID(eventtoken, questions)
	if len(errSlice) == 0 {
		return nil
	} else {
		return errors.New("Err while sending update")
	}
}
