package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	mongo   DataStorage
	commMan EventManager
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func main() {
	r := gin.Default()

	mongo = NewMgoStorage()
	mongo.OpenSession()

	commMan = NewEventManager()
	err := mongo.OpenSession()
	if err != nil {
		log.Panic(err)
	}

	r.POST("/question", postQuestion)
	r.GET("/event/:eventtoken", eventEndpoint)
	r.Run("localhost:8888")
}

func eventEndpoint(c *gin.Context) {
	eventId := c.Params.ByName("eventtoken")
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade: %+v", err)
		return
	}
	commMan.RegisterConnection(eventId, conn)
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
	sendUpdate(question.EventToken)
	c.JSON(200, "OK")
}

func sendUpdate(eventId string) error {
	connections := commMan.GetConnByEvent(eventId)
	questions, err := mongo.QuestionsByEvent(eventId)
	if err != nil {
		return err
	}

	for _, conn := range connections {
		conn.WriteJSON(questions)
	}
	return nil
}
