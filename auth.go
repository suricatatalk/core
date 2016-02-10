package main

import (
	"encoding/json"

	"github.com/sohlich/natsproxy"
	"github.com/suricatatalk/gate/auth"
)

func authToken(c *natsproxy.Context) {
	token := c.Request.Header.Get(TokenHeader)
	log.Info(token)
	if len(token) == 0 {
		log.Error("Token header not found")
		c.JSON(401, "Not authenticated")
		c.Abort()
		return
	}

	user, err := auth.DecodeJwtToken(token)
	if err != nil {
		log.Errorf("Jwt token cannot be decoded %s", err)
		c.JSON(403, "Access forbidden")
		c.Abort()
		return
	}
	userJSON, _ := json.Marshal(user)
	c.Request.Header.Set(TokenHeader, string(userJSON))
}
