package main

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/suricatatalk/gate/auth"
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
		log.Error(err)
		c.AbortWithStatus(403)
		return
	}
	userJSON, _ := json.Marshal(user)
	c.Request.Header.Set(TokenHeader, string(userJSON))
}
