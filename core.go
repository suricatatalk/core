package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/itsjamie/gin-cors"
	"github.com/sohlich/etcd_discovery"
)

const (
	// ServiceName defines the service type
	// that will be registered in etcd service registry
	ServiceName = "core"

	// KeyLogly is a enviromental
	// variable for logging with loggly
	KeyLogly = "LOGLY_TOKEN"

	// TokenHeader is header with auth informations
	TokenHeader = "X-AUTH"
)

var (
	log            = logrus.StandardLogger()
	mongo          DataStorage
	commMan        EventManager
	notifier       Notifier
	registryConfig = discovery.EtcdRegistryConfig{
		ServiceName: ServiceName,
	}
	registryClient *discovery.EtcdReigistryClient
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func main() {

	// Load all configuration
	appCfg := &AppConfig{}
	mgoCfg := &MgoConfig{}
	etcdCfg := &EtcdConfig{}
	loadConfiguration(appCfg, mgoCfg, etcdCfg)

	var registryErr error
	log.Infof("Initializing service discovery client for %s", appCfg.Name)
	registryConfig.InstanceName = appCfg.Name
	registryConfig.BaseURL = fmt.Sprintf("%s:%s", appCfg.Host, appCfg.Port)
	registryConfig.EtcdEndpoints = []string{etcdCfg.Endpoint}
	registryClient, registryErr = discovery.New(registryConfig)
	if registryErr != nil {
		log.Panic(registryErr)
	}
	registryClient.Register()

	log.Infoln("Initializing mongo storage with credentials %s , %s", mgoCfg.URI, mgoCfg.DB)
	localMgo := NewMgoStorage()
	localMgo.connectionString = mgoCfg.URI
	localMgo.database = mgoCfg.DB
	mongo = localMgo

	eventConnManager := NewEventManager()
	commMan = eventConnManager
	notifier = eventConnManager
	err := mongo.OpenSession()
	if err != nil {
		log.Panicln(err)
	}

	r := gin.Default()
	log.Infoln("Configuring CORS Middleware")
	r.Use(logrusLogger())
	r.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET, PUT, POST, DELETE",
		RequestHeaders:  "Origin, Authorization, Content-Type, X-AUTH",
		ExposedHeaders:  "",
		MaxAge:          50 * time.Second,
		Credentials:     true,
		ValidateHeaders: false,
	}))

	//Public
	r.POST("/question/:questionID", voteQuestion)
	r.DELETE("/question/:questionID", voteQuestion)
	r.POST("/question", postQuestion)
	r.GET("/event/:eventtoken/:session", eventWebsockHandler)
	r.GET("/event/:eventtoken", getEvent)
	r.GET("/speaker/:speakerID", getSpeaker)
	//Admin
	authReqi := r.Group("/")
	authReqi.Use(authToken)
	authReqi.POST("/event", upsertEvent(insertEvent))
	authReqi.PUT("/event", upsertEvent(updateEvent))
	authReqi.POST("/speaker", upsertSpeaker(insertSpeaker))
	authReqi.PUT("/speaker", upsertSpeaker(updateSpeaker))

	bind := fmt.Sprintf(":%s", appCfg.Port)
	r.Run(bind)
}

func logrusLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Infof("%s:%s from %s", c.Request.Method, c.Request.URL.String(), c.Request.Header.Get("X-Forwarded-For"))
	}
}

func eventWebsockHandler(c *gin.Context) {
	log.Printf("Receiving WS request %s", c.Request.Header)
	eventToken := c.Params.ByName("eventtoken")
	sessitonToken := c.Params.ByName("session")

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("Failed to set websocket upgrade: %v", err)
		return
	}

	updateErr := notifyChangeForConnection(conn, eventToken, sessitonToken)
	if updateErr != nil {
		log.Errorln(updateErr)
	}
	commMan.RegisterConnection(eventToken, sessitonToken, conn)
}

func voteQuestion(c *gin.Context) {
	questionID := c.Params.ByName("questionID")

	log.Infof("voteQuestion: voting question %s", questionID)

	incBy := 1
	if c.Request.Method == "DELETE" {
		incBy = -1
	}
	err := mongo.VoteQuestion(questionID, incBy)

	if err != nil {
		log.Errorln(err)
		c.JSON(405, "Event not exist")
		return
	}
	q, qerr := mongo.QuestionById(questionID)
	if qerr != nil {
		log.Errorln(qerr)
		c.JSON(405, "Event not exist")
		return
	}
	updateErr := notifyChange(q.EventToken, q.SessionToken)
	if updateErr != nil {
		log.Errorln(updateErr)
	}
	c.JSON(200, q)
}

func postQuestion(c *gin.Context) {
	question := &Question{}
	err := c.BindJSON(question)
	if err != nil {
		log.Errorln(err)
		c.JSON(405, "Cannot store the question")
		return
	}

	log.Infof("postQuestion: posting question %s", question)

	_, err = mongo.EventByToken(question.EventToken)
	if err != nil {
		log.Errorln(err)
		c.JSON(405, "Event not exist")
		return
	}
	mongo.InsertQuestion(question)
	updateErr := notifyChange(question.EventToken, question.SessionToken)
	if updateErr != nil {
		log.Errorln(updateErr)
	}
	c.JSON(200, question)
}

func notifyChange(eventToken, sessionToken string) error {
	questions, err := mongo.QuestionsByEventAndSession(eventToken, sessionToken)
	if err != nil {
		return err
	}
	errSlice := notifier.SendJsonByEventAndSessionToken(eventToken, sessionToken, questions)
	if len(errSlice) > 0 {
		return errors.New("Err while sending update")
	}
	return nil
}

func notifyChangeForConnection(conn *websocket.Conn, eventToken, sessionToken string) error {
	questions, err := mongo.QuestionsByEventAndSession(eventToken, sessionToken)
	if err != nil {
		return err
	}
	sendErr := conn.WriteJSON(questions)
	if sendErr != nil {
		return sendErr
	}
	return nil
}

// Event handlers
type eventHandlerFunc func(c *gin.Context, event *Event)

func upsertEvent(handler eventHandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		event := &Event{}
		err := c.BindJSON(event)
		if err != nil {
			log.Errorln(err)
			c.JSON(http.StatusBadRequest, "Malformed json object")
			return
		}

		err = ValidateEvent(event)
		if err != nil {
			log.Errorln(err)
			c.JSON(http.StatusBadRequest, "Event not valid")
			return
		}

		handler(c, event)
	}
}

func insertEvent(c *gin.Context, event *Event) {
	log.Infof("insertEvent : inserting event %s", event)
	err := mongo.InsertEvent(event)
	if err != nil {
		log.Errorln(err)
		c.JSON(http.StatusInternalServerError, "Cannot import event")
		return
	}
	c.JSON(http.StatusOK, event)
}

func updateEvent(c *gin.Context, event *Event) {
	log.Infof("updateEvent : inserting event %s", event)
	err := mongo.UpdateEvent(event)
	if err != nil {
		log.Errorln(err)
		c.JSON(http.StatusInternalServerError, "Cannot update event")
		return
	}
	c.JSON(http.StatusOK, event)
}

func getEvent(c *gin.Context) {
	eventToken := c.Params.ByName("eventtoken")

	log.Infof("getEvent : getting event %s", eventToken)

	event, err := mongo.EventByToken(eventToken)
	if err != nil {
		log.Errorln(err)
		c.JSON(405, "Event not exist")
		return
	}

	log.Infoln("Getting speakers fro event %s", event.ID.Hex())
	speakers, spErr := mongo.SpeakersById(event.Speakers)
	if spErr != nil {
		humanError := fmt.Sprintf("Speaker not found reason: %s", spErr.Error())
		log.Errorln(humanError)
		c.JSON(500, fmt.Sprintf("Speaker not found reason: %s", spErr.Error()))
		return
	}

	// Fill response with all necessary data
	output := struct {
		Speakers []*Speaker `json:"speakers"`
		Event    *Event     `json:"event"`
	}{
		speakers,
		event,
	}

	log.Infoln("Event in response %v", output)
	c.JSON(200, output)
}

// Speakers handlers

type speakerHandlerFunc func(c *gin.Context, speaker *Speaker)

func upsertSpeaker(handler speakerHandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		speaker := &Speaker{}
		err := c.BindJSON(speaker)
		if err != nil {
			log.Errorln(err)
			c.AbortWithStatus(405)
			return
		}
		handler(c, speaker)
	}
}

func insertSpeaker(c *gin.Context, speaker *Speaker) {
	err := mongo.InsertSpeaker(speaker)
	if err != nil {
		log.Errorln(err)
		c.AbortWithError(500, err)
		return
	}
	c.JSON(http.StatusOK, speaker)
}

func updateSpeaker(c *gin.Context, speaker *Speaker) {
	err := mongo.UpdateSpeaker(speaker)
	if err != nil {
		log.Errorln(err)
		c.AbortWithError(500, err)
		return
	}
	c.JSON(http.StatusOK, speaker)
}

func getSpeaker(c *gin.Context) {
	speakerID := c.Params.ByName("speakerID")
	event, err := mongo.SpeakerById(speakerID)
	if err != nil {
		c.JSON(405, "Speaker not exist")
		return
	}
	c.JSON(200, event)
}
