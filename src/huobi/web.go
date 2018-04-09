package main

import (
	"github.com/gin-gonic/gin"
)

func ginWeb() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(200, lastSendText)
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}
