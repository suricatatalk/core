package main

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/sohlich/surikata_auth/auth"
)

func authToken(c *gin.Context) {
	token := c.Request.Header.Get(TokenHeader)
	if len(token) == 0 {
		c.AbortWithStatus(401)
		return
	}
	log.Infof("Getting token", token)

	user, err := auth.DecodeJwtToken(token)
	if err != nil {
		log.Errorln(err)
		c.AbortWithStatus(403)
		return
	}
	userJson, _ := json.Marshal(user)
	c.Request.Header.Set(TokenHeader, string(userJson))
}
