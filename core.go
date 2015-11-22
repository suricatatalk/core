package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sohlich/etcd_registry"
)

var (
	mongo          DataStorage
	commMan        EventManager
	notifier       Notifier
	registryConfig registry.EtcdRegistryConfig = registry.EtcdRegistryConfig{
		EtcdEndpoints: []string{"http://127.0.0.1:4001"},
		ServiceName:   "core",
		InstanceName:  "core1",
		BaseUrl:       "127.0.0.1:8080",
	}
	registryClient *registry.EtcdReigistryClient
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func main() {
	var registryErr error
	registryClient, registryErr = registry.New(registryConfig)
	if registryErr != nil {
		log.Panic(registryErr)
	}

	registryClient.Register()

	r := gin.Default()

	mgoStorage := NewMgoStorage()
	mgoStorage.connectionString = os.Getenv("MONGODB_ADDON_URI")
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

	//Public
	r.POST("/question/:questionID", voteQuestion)
	r.POST("/question", postQuestion)
	r.GET("/event/:eventtoken/:session", eventWebsockHandler)
	r.GET("/event/:eventtoken", getEvent)

	//Admin
	r.POST("/event", upsertEvent(insertEvent))
	r.PUT("/event", upsertEvent(updateEvent))

	bind := fmt.Sprintf("0.0.0.0:%s", os.Getenv("PORT"))
	r.Run(bind)
}

func eventWebsockHandler(c *gin.Context) {
	eventToken := c.Params.ByName("eventtoken")
	sessitonToken := c.Params.ByName("session")

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade: %+v", err)
		return
	}

	commMan.RegisterConnection(eventToken, sessitonToken, conn)
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
	notifyChange(q.EventToken, q.SessionToken)
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
	notifyChange(question.EventToken, question.SessionToken)
	c.JSON(200, "OK")
}

func notifyChange(eventToken, sessionToken string) error {
	questions, err := mongo.QuestionsByEventAndSession(eventToken, sessionToken)
	if err != nil {
		return err
	}
	errSlice := notifier.SendJsonByEventAndSessionToken(eventToken, sessionToken, questions)
	if len(errSlice) == 0 {
		return nil
	} else {
		return errors.New("Err while sending update")
	}
}

type eventHandlerFunc func(c *gin.Context, event *Event)

func upsertEvent(handler eventHandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		event := &Event{}
		err := c.BindJSON(event)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, "Malformed json object")
			return
		}
		handler(c, event)
	}
}

func insertEvent(c *gin.Context, event *Event) {
	err := mongo.InsertEvent(event)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, "Cannot import event")
		return
	}
	c.JSON(http.StatusOK, event)
}

func updateEvent(c *gin.Context, event *Event) {
	err := mongo.UpdateEvent(event)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, "Cannot update event")
		return
	}
	c.JSON(http.StatusOK, event)
}

func getEvent(c *gin.Context) {
	eventToken := c.Params.ByName("eventtoken")
	event, err := mongo.EventByToken(eventToken)
	if err != nil {
		c.JSON(405, "Event not exist")
		return
	}
	c.JSON(200, event)
}
